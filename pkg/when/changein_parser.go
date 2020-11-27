package when

import (
	"fmt"
	"path"

	gabs "github.com/Jeffail/gabs/v2"
	environment "github.com/semaphoreci/spc/pkg/environment"
)

type ChangeInFunctionParser struct {
	raw      *gabs.Container
	when     *WhenExpression
	yamlPath string
}

func (p *ChangeInFunctionParser) ParseFunction() (*ChangeInFunction, error) {
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

	params := ChangeInFunctionParams{
		PathPatterns:         p.PathPatterns(),
		ExcludedPathPatterns: p.ExcludedPathPatterns(),
		DefaultBranch:        p.DefaultBranch(),
		TrackPipelineFile:    track,
		OnTags:               onTags,
		DefaultRange:         defaultRange,
	}

	return &ChangeInFunction{
		Workdir:  path.Dir(p.yamlPath),
		YamlPath: p.yamlPath,
		Params:   params,
	}, nil
}

func (p *ChangeInFunctionParser) DefaultBranch() string {
	if p.raw.Exists("params", "1", "default_branch") {
		return p.raw.Search("params", "1", "default_branch").Data().(string)
	} else {
		return "master"
	}
}

func (p *ChangeInFunctionParser) PathPatterns() []string {
	result := []string{}

	if _, ok := p.raw.Search("params", "0").Data().([]interface{}); ok {
		for _, p := range p.raw.Search("params", "0").Children() {
			result = append(result, p.Data().(string))
		}
	} else {
		result = append(result, p.raw.Search("params", "0").Data().(string))
	}

	return result
}

func (p *ChangeInFunctionParser) ExcludedPathPatterns() []string {
	result := []string{}

	if _, ok := p.raw.Search("params", "1", "exclude").Data().([]interface{}); ok {
		for _, p := range p.raw.Search("params", "1", "exclude").Children() {
			result = append(result, p.Data().(string))
		}
	}

	return result
}

func (p *ChangeInFunctionParser) TrackPipelineFile() (bool, error) {
	if p.raw.Exists("params", "1", "pipeline_file") {
		value, ok := p.raw.Search("params", "1", "pipeline_file").Data().(string)
		if !ok {
			return false, fmt.Errorf("Unknown value type pipeline_file in change_in expression.")
		}

		switch value {
		case "track":
			return true, nil

		case "ignore":
			return false, nil

		default:
			return false, fmt.Errorf("Unknown value type pipeline_file in change_in expression.")
		}
	} else {
		if p.when.Path[0] == "promotions" {
			return false, nil
		} else {
			return true, nil
		}
	}
}

func (p *ChangeInFunctionParser) OnTags() (bool, error) {
	if p.raw.Exists("params", "1", "on_tags") {
		value, ok := p.raw.Search("params", "1", "on_tags").Data().(bool)
		if !ok {
			return true, fmt.Errorf("Unknown value type on_tags in change_in expression.")
		}

		return value, nil
	} else {
		return true, nil
	}
}

func (p *ChangeInFunctionParser) DefaultRange() (string, error) {
	if p.raw.Exists("params", "1", "default_range") {
		value, ok := p.raw.Search("params", "1", "default_range").Data().(string)
		if !ok {
			return "", fmt.Errorf("Unknown value type default_range in change_in expression.")
		}

		return value, nil
	} else {
		return environment.GitCommitRange(), nil
	}
}
