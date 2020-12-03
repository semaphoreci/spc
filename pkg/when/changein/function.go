package changein

import (
	"strings"

	environment "github.com/semaphoreci/spc/pkg/environment"
	logs "github.com/semaphoreci/spc/pkg/logs"
)

type ChangeInFunctionParams struct {
	PathPatterns         []string
	ExcludedPathPatterns []string
	DefaultBranch        string
	TrackPipelineFile    bool
	OnTags               bool
	DefaultRange         string
	CommitRange          string
}

type Function struct {
	Params  Params
	Workdir string

	YamlPath string
	Location logs.Location
}

func (f *ChangeInFunction) Eval() (bool, error) {
	e := evaluator{function: f}

	return e.Run()
}

func (f *ChangeInFunction) CommitRange() string {
	if environment.CurrentBranch() == f.Params.DefaultBranch {
		return f.Params.DefaultRange
	}

	return f.Params.CommitRange
}

func (f *ChangeInFunction) ParseCommitRange() (string, string) {
	var splitAt string

	if strings.Contains(f.CommitRange(), "...") {
		splitAt = "..."
	} else {
		splitAt = ".."
	}

	parts := strings.Split(f.CommitRange(), splitAt)

	return parts[0], parts[1]
}
