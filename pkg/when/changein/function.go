package changein

import (
	consolelogger "github.com/semaphoreci/spc/pkg/consolelogger"
	logs "github.com/semaphoreci/spc/pkg/logs"
)

type Function struct {
	Workdir  string
	YamlPath string
	Location logs.Location

	PathPatterns           []string
	ExcludedPathPatterns   []string
	DefaultBranch          string
	TrackPipelineFile      bool
	OnTags                 bool
	DefaultRange           string
	BranchRange            string
	PullRequestRange       string
	ForkedPullRequestRange string
	BaseIsCommitSha        bool
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

	if _, ok := f.IsPatternMacthed(diffLine); ok {
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

func (f *Function) IsPatternMacthed(diffLine string) (string, bool) {
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
