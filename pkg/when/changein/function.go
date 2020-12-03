package changein

import (
	logs "github.com/semaphoreci/spc/pkg/logs"
)

type Function struct {
	Workdir  string
	YamlPath string
	Location logs.Location

	PathPatterns         []string
	ExcludedPathPatterns []string
	DefaultBranch        string
	TrackPipelineFile    bool
	OnTags               bool
	DefaultRange         string
	CommitRange          string
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
	for _, pathPattern := range f.ExcludedPathPatterns {
		if patternMatch(diffLine, pathPattern, f.Workdir) {
			return false
		}
	}

	if f.TrackPipelineFile && patternMatch(diffLine, f.absoluteYAMLPath(), f.Workdir) {
		return true
	}

	for _, pathPattern := range f.PathPatterns {
		if patternMatch(diffLine, pathPattern, f.Workdir) {
			return true
		}
	}

	return false
}

func (f *Function) absoluteYAMLPath() string {
	return "/" + f.YamlPath
}
