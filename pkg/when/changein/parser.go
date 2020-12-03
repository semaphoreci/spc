package changein

import (
	"fmt"
	"path"
	"strings"

	gabs "github.com/Jeffail/gabs/v2"
	environment "github.com/semaphoreci/spc/pkg/environment"
	logs "github.com/semaphoreci/spc/pkg/logs"
)

func Parse(whenPath []string, ast *gabs.Container, yamlPath string) (*Function, error) {
	p := parser{
		ast:      ast,
		yamlPath: yamlPath,
		whenPath: whenPath,
	}

	return p.parse()
}

type parser struct {
	ast      *gabs.Container
	yamlPath string
	whenPath []string

	result Function
}

func (p *parser) parse() (*Function, error) {
	paths, err := p.PathPatterns()
	if err != nil {
		return nil, err
	}

	excludedPaths, err := p.ExcludedPathPatterns()
	if err != nil {
		return nil, err
	}

	track, err := p.TrackPipelineFile()
	if err != nil {
		return nil, err
	}

	onTags, err := p.OnTags()
	if err != nil {
		return nil, err
	}

	defaultBranch, err := p.DefaultBranch()
	if err != nil {
		return nil, err
	}

	defaultRange, err := p.DefaultRange(defaultBranch)
	if err != nil {
		return nil, err
	}

	commitRange, err := p.CommitRange(defaultBranch)
	if err != nil {
		return nil, err
	}

	location := logs.Location{
		File: p.yamlPath,
		Path: p.whenPath,
	}

	return &Function{
		Workdir:  path.Dir(p.yamlPath),
		YamlPath: p.yamlPath,
		Location: location,

		PathPatterns:         paths,
		ExcludedPathPatterns: excludedPaths,
		DefaultBranch:        defaultBranch,
		TrackPipelineFile:    track,
		OnTags:               onTags,
		DefaultRange:         defaultRange,
		CommitRange:          commitRange,
	}, nil
}

func (p *parser) PathPatterns() ([]string, error) {
	firstArg := p.ast.Search("params", "0")

	if !firstArg.Exists() {
		return []string{}, fmt.Errorf("path parameter not found in change in expression")
	}

	result, ok := p.castToStringArray(firstArg.Data())
	if !ok {
		return []string{}, fmt.Errorf("uprocessable path parameter in change in expression")
	}

	return result, nil
}

func (p *parser) DefaultBranch() (string, error) {
	defaultBranch, found, err := p.getStringParam("default_branch")
	if !found {
		return "master", nil
	}

	return defaultBranch, err
}

func (p *parser) ExcludedPathPatterns() ([]string, error) {
	excludedPaths, found := p.getParam("exclude")
	if !found {
		return []string{}, nil
	}

	result, ok := p.castToStringArray(excludedPaths)
	if !ok {
		return []string{}, fmt.Errorf("uprocessable exclude path parameter in change in expression")
	}

	return result, nil
}

func (p *parser) TrackPipelineFile() (bool, error) {
	pipelineFile, found, err := p.getStringParam("pipeline_file")
	if err != nil {
		return false, err
	}

	if !found {
		return p.whenPath[0] != "promotions", nil
	}

	switch pipelineFile {
	case "track":
		return true, nil

	case "ignore":
		return false, nil
	}

	return false, fmt.Errorf("unknown value type pipeline_file in change_in expression")
}

func (p *parser) OnTags() (bool, error) {
	onTags, found, err := p.getBoolParam("on_tags")
	if err != nil {
		return false, err
	}

	if !found {
		return true, nil
	}

	return onTags, err
}

func (p *parser) DefaultRange(defaultBranch string) (string, error) {
	defaultRange, found, err := p.getStringParam("default_range")
	if err != nil {
		return "", err
	}

	if !found {
		return p.fetchCommitRange(defaultBranch), nil
	}

	return defaultRange, err

}

func (p *parser) CommitRange(defaultBranch string) (string, error) {
	commitRange, found, err := p.getStringParam("branch_range")
	if err != nil {
		return "", err
	}

	if !found {
		return p.fetchCommitRange(defaultBranch), nil
	}

	commitRange = strings.ReplaceAll(commitRange, "$SEMAPHORE_MERGE_BASE", environment.MergeBase())
	commitRange = strings.ReplaceAll(commitRange, "$SEMAPHORE_GIT_SHA", environment.CurrentGitSha())

	return commitRange, nil
}

func (p *parser) fetchCommitRange(defaultBranch string) string {
	commitRange := environment.GitCommitRange()

	if commitRange != "" {
		return commitRange
	}

	return fmt.Sprintf("%s...%s", defaultBranch, environment.CurrentGitSha())
}

func (p *parser) getParam(path ...string) (interface{}, bool) {
	if p.ast.Exists(path...) {
		return p.ast.Search(path...).Data(), true
	} else {
		return nil, false
	}
}

func (p *parser) getStringParam(key string) (string, bool, error) {
	val, ok := p.getParam(key)
	if !ok {
		return "", false, nil
	}

	stringVal, ok := val.(string)
	if !ok {
		return "", true, fmt.Errorf("unknown value type %s in change_in expression", key)
	}

	return stringVal, true, nil
}

func (p *parser) getBoolParam(key string) (bool, bool, error) {
	val, ok := p.getParam(key)
	if !ok {
		return false, false, nil
	}

	boolVal, ok := val.(bool)
	if !ok {
		return false, true, fmt.Errorf("unknown value type %s in change_in expression", key)
	}

	return boolVal, true, nil
}

func (p *parser) castToStringArray(obj interface{}) ([]string, bool) {
	if value, ok := obj.(string); ok {
		return []string{value}, true
	}

	if values, ok := obj.([]string); ok {
		return values, true
	}

	return []string{}, false
}