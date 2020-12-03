package changein

import (
	"fmt"
	"path"
	"strings"

	gabs "github.com/Jeffail/gabs/v2"
	environment "github.com/semaphoreci/spc/pkg/environment"
	logs "github.com/semaphoreci/spc/pkg/logs"
)

type ChangeInFunctionParser struct {
	raw      *gabs.Container
	whenPath []string
	yamlPath string
}

func Parse(whenPath []string, input *gabs.Container, yamlPath string) (*ChangeInFunction, error) {
	parser := ChangeInFunctionParser{
		raw:      input,
		whenPath: whenPath,
		yamlPath: yamlPath,
	}

	return parser.Execute()
}

func (p *ChangeInFunctionParser) Execute() (*ChangeInFunction, error) {
	track, err := p.TrackPipelineFile()
	if err != nil {
		return nil, err
	}

	onTags, err := p.OnTags()
	if err != nil {
		return nil, err
	}

	defaultRange, err := p.DefaultRange()
	if err != nil {
		return nil, err
	}

	commitRange, err := p.CommitRange()
	if err != nil {
		return nil, err
	}

	params := ChangeInFunctionParams{
		PathPatterns:         p.PathPatterns(),
		ExcludedPathPatterns: p.ExcludedPathPatterns(),
		DefaultBranch:        p.DefaultBranch(),
		TrackPipelineFile:    track,
		OnTags:               onTags,
		DefaultRange:         defaultRange,
		CommitRange:          commitRange,
	}

	return &ChangeInFunction{
		Workdir:  path.Dir(p.yamlPath),
		YamlPath: p.yamlPath,
		Location: logs.Location{
			File: p.yamlPath,
			Path: p.when.Path,
		},
		Params: params,
	}, nil
}

func (p *ChangeInFunctionParser) DefaultBranch() string {
	if p.functionParams().Exists("default_branch") {
		return p.functionParams().Search("default_branch").Data().(string)
	}

	return "master"
}

func (p *ChangeInFunctionParser) PathPatterns() []string {
	result := []string{}

	firstArg := p.raw.Search("params", "0")

	if _, ok := firstArg.Data().([]interface{}); ok {
		for _, p := range firstArg.Children() {
			result = append(result, p.Data().(string))
		}
	} else {
		result = append(result, firstArg.Data().(string))
	}

	return result
}

func (p *ChangeInFunctionParser) ExcludedPathPatterns() []string {
	result := []string{}

	if _, ok := p.functionParams().Search("exclude").Data().([]interface{}); ok {
		for _, p := range p.functionParams().Search("exclude").Children() {
			result = append(result, p.Data().(string))
		}
	}

	return result
}

func (p *ChangeInFunctionParser) TrackPipelineFile() (bool, error) {
	if p.functionParams().Exists("pipeline_file") {
		value, ok := p.functionParams().Search("pipeline_file").Data().(string)
		if !ok {
			return false, fmt.Errorf("unknown value type pipeline_file in change_in expression")
		}

		return p.ParseTrackPipelineFile(value)
	}

	return p.when.Path[0] != "promotions", nil
}

func (p *ChangeInFunctionParser) ParseTrackPipelineFile(val string) (bool, error) {
	switch val {
	case "track":
		return true, nil

	case "ignore":
		return false, nil
	}

	return false, fmt.Errorf("unknown value type pipeline_file in change_in expression")
}

func (p *ChangeInFunctionParser) functionParams() *gabs.Container {
	return p.raw.Search("params", "1")
}

func (p *ChangeInFunctionParser) OnTags() (bool, error) {
	if p.functionParams().Exists("on_tags") {
		value, ok := p.functionParams().Search("on_tags").Data().(bool)
		if !ok {
			return true, fmt.Errorf("unknown value type on_tags in change_in expression")
		}

		return value, nil
	}

	return true, nil
}

func (p *ChangeInFunctionParser) DefaultRange() (string, error) {
	if p.functionParams().Exists("default_range") {
		value, ok := p.functionParams().Search("default_range").Data().(string)
		if !ok {
			return "", fmt.Errorf("unknown value type default_range in change_in expression")
		}

		return value, nil
	}

	return p.fetchCommitRange(), nil
}

func (p *ChangeInFunctionParser) CommitRange() (string, error) {
	if p.raw.Exists("params", "1", "branch_range") {
		value, ok := p.functionParams().Search("branch_range").Data().(string)
		if !ok {
			return "", fmt.Errorf("unknown value type branch_range in change_in expression")
		}

		value = strings.ReplaceAll(value, "$SEMAPHORE_MERGE_BASE", environment.MergeBase())
		value = strings.ReplaceAll(value, "$SEMAPHORE_GIT_SHA", environment.CurrentGitSha())

		return value, nil
	}

	return p.fetchCommitRange(), nil
}

func (p *ChangeInFunctionParser) fetchCommitRange() string {
	commitRange := environment.GitCommitRange()
	if commitRange != "" {
		return commitRange
	}

	return fmt.Sprintf("%s...%s", p.DefaultBranch(), environment.CurrentGitSha())
}
