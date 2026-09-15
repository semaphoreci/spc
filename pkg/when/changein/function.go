package changein

import (
	"errors"
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

	if err := f.fetchRequiredBranches(); err != nil {
		return false, err
	}

	diffList, err := git.DiffList(f.GitDiffSet.CommitRange())
	if err != nil {
		return f.resolveDiffListError(err)
	}

	consolelogger.EmptyLine()
	consolelogger.Infof("Comparing change_in with git diff\n")

	result := f.HasMatchesInDiffList(diffList)

	consolelogger.EmptyLine()
	consolelogger.Infof("Result: %+v\n", result)

	return result, nil
}

func (f *Function) fetchRequiredBranches() error {
	fetchNeeded, fetchTargets := f.GitDiffSet.IsGitFetchNeeded()
	if !fetchNeeded {
		return nil
	}

	for _, fetchTarget := range fetchTargets {
		output, fetchErr := git.Fetch(fetchTarget)

		if err := f.parseFetchError(fetchTarget, output, fetchErr); err != nil {
			return err
		}
	}

	return nil
}

// resolveDiffListError separates the two ways resolving a commit range can go
// wrong. A range with no merge base is not a git failure and keeps resolving
// to false; anything else means git itself is broken and fails compilation.
func (f *Function) resolveDiffListError(err error) (bool, error) {
	if errors.Is(err, git.ErrRangeUnresolvable) {
		f.warnRangeUnresolvable()
		return false, nil
	}

	return false, f.gitFailureError(err)
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

// gitFailureError reports a git command failure as a structured error, which
// fails the whole compilation. Without it the condition would be silently
// resolved, which hides a broken git state behind a plain 'false' result.
//
// ErrorInitializationFailed is used on purpose: it is the type the workflow
// page already knows how to render. A type it does not recognise falls through
// to its "Unprocessable YAML file" template, which would blame the YAML for
// what is a git problem.
func (f *Function) gitFailureError(err error) error {
	msg := fmt.Sprintf(
		"Failed to resolve the git diff for commit range '%s': %s",
		f.GitDiffSet.CommitRange(),
		err.Error(),
	)

	gitErr := logs.ErrorInitializationFailed{Message: msg, Location: f.Location}

	logs.Log(gitErr)

	return &gitErr
}

// warnRangeUnresolvable reports a commit range whose two ends genuinely have
// no common ancestor, with the whole history in hand. No git command failed
// and deepening cannot help, so the condition resolves to false as it always
// has. The warning is what makes the situation visible, because the
// alternative - failing the compilation - would take down every pipeline in a
// repository in this state, including blocks that use no change_in at all.
//
// A clone that merely ran out of deepen budget never reaches here: that
// answer is inconclusive and is reported as a failure instead.
func (f *Function) warnRangeUnresolvable() {
	consolelogger.EmptyLine()
	consolelogger.Infof("WARNING: no merge base found for commit range '%s'.\n", f.GitDiffSet.CommitRange())
	consolelogger.Infof("WARNING: the full history is available and the two branches share no\n")
	consolelogger.Infof("WARNING: commit, so the base branch was most likely recreated.\n")
	consolelogger.Infof("WARNING: resolving this change_in to false.\n")
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
