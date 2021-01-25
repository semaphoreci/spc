package changein

import (
	"fmt"

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
}

func (f *Function) HasMatchesInDiffList(diffList []string) bool {
	for _, diffLine := range diffList {
		if f.IsPatternMatchWith(diffLine) {
			return true
		}
	}

	return false
}

func (f *Function) IsPatternMatchWith(diffLine string) bool {
	if pattern, ok := f.IsDiffLineExcluded(diffLine); ok {
		fmt.Printf("* Rejected by pattern: %s\n", pattern)
		return false
	}

	if pattern, ok := f.IsPipelineFileMatched(diffLine); ok {
		fmt.Printf("* Matched by pipeline file: %s\n", pattern)
		return true
	}

	if pattern, ok := f.IsPatternMacthed(diffLine); ok {
		fmt.Printf("* Matched by pattern: %s\n", pattern)
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
