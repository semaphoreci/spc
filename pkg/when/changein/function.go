package changein

import (
	"fmt"
	"path"

	gabs "github.com/Jeffail/gabs/v2"
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

func Parse(ast *gabs.Container) (*Function, error) {
	p := parser{ast: ast}

	return p.parse()
}

type parser struct {
	ast    *gabs.Container
	result Function
}

func (p *parser) parse() error {
	paths, err := p.parsePathParam()
	if err != nil {
		return err
	}

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

func (p *parser) parsePathParam() ([]string, error) {
	if !p.ast.Exists("params", "0") {
		return []string{}, fmt.Errorf("path parameter not found in change in expression")
	}

	result, ok := p.castToStringArray(p.ast.Search("params", "0").Data())
	if !ok {
		return []string{}, fmt.Errorf("uprocessable path parameter in change in expression")
	}

	return result, nil

}

func (p *parser) castToStringArray(obj interface{}) ([]string, error) {
	if value, ok := obj.(string); ok {
		return []string{value}, true
	}

	if values, ok := obj.([]string); ok {
		return values, true
	}

	return []string{}, false
}
