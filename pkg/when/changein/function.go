package changein

import (
	"fmt"
	"strings"

	consolelogger "github.com/semaphoreci/spc/pkg/consolelogger"
	git "github.com/semaphoreci/spc/pkg/git"
	logs "github.com/semaphoreci/spc/pkg/logs"
)

type Function struct {
	Workdir  string
	YamlPath string
	Location logs.Location

	PathPatterns         []string
	ExcludedPathPatterns []string
	TrackPipelineFile    bool
	GitDiffSet           *git.DiffSet
}

func (f *Function) Eval() (bool, error) {
	if f.GitDiffSet.IsEvaluationNeeded() {
		consolelogger.Infof("Running on a tag, skipping evaluation\n")
		return f.GitDiffSet.OnTags, nil
	}

	fetchNeeded, fetchTargets := f.GitDiffSet.IsGitFetchNeeded()

	if fetchNeeded {
		for _, fetchTarget := range fetchTargets {
			output, err := git.Fetch(fetchTarget)
			err = f.parseFetchError(fetchTarget, output, err)

			if err != nil {
				return false, err
			}
		}
	}

	diffList, err := git.DiffList(f.GitDiffSet.CommitRange())
	if err != nil {
		return false, err
	}

	consolelogger.EmptyLine()
	consolelogger.Infof("Comparing change_in with git diff\n")

	result := f.HasMatchesInDiffList(diffList)

	consolelogger.EmptyLine()
	consolelogger.Infof("Result: %+v\n", result)

	return result, nil
}

func (f *Function) HasMatchesInDiffList(diffList []string) bool {
	for _, diffLine := range diffList {
		result := f.IsPatternMatchWith(diffLine)

		if result {
			consolelogger.Infof("(match) %s\n", diffLine)
		} else {
			consolelogger.Infof("(no match) %s\n", diffLine)
		}

		if result {
			return true
		}
	}

	return false
}

func (f *Function) IsPatternMatchWith(diffLine string) bool {
	if _, ok := f.IsDiffLineExcluded(diffLine); ok {
		return false
	}

	if _, ok := f.IsPipelineFileMatched(diffLine); ok {
		return true
	}

	if _, ok := f.IsPatternMatched(diffLine); ok {
		return true
	}

	return false
}

func (f *Function) IsDiffLineExcluded(diffLine string) (string, bool) {
	for _, pathPattern := range f.ExcludedPathPatterns {
		if patternMatch(diffLine, pathPattern, f.Workdir) {
			return pathPattern, true
		}
	}

	return "", false
}

func (f *Function) IsPatternMatched(diffLine string) (string, bool) {
	for _, pathPattern := range f.PathPatterns {
		if patternMatch(diffLine, pathPattern, f.Workdir) {
			return pathPattern, true
		}
	}

	return "", false
}

func (f *Function) IsPipelineFileMatched(diffLine string) (string, bool) {
	path := f.absoluteYAMLPath()

	return path, (f.TrackPipelineFile && patternMatch(diffLine, path, f.Workdir))
}

func (f *Function) absoluteYAMLPath() string {
	return "/" + f.YamlPath
}

func (f *Function) parseFetchError(fetchTarget string, output string, err error) error {
	if strings.Contains(string(output), "couldn't find remote ref") {
		msg := fmt.Sprintf("Unknown git reference '%s'.", fetchTarget)
		err := logs.ErrorChangeInMissingBranch{Message: msg, Location: f.Location}

		logs.Log(err)

		return &err
	}

	return err
}
