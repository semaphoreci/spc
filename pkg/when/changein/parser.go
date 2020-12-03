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
	defaultBranch := p.functionParams().Search("default_branch")

	if !defaultBranch.Exists() {
		return "master", nil
	}

	value, ok := defaultBranch.Data().(string)
	if !ok {
		return "", fmt.Errorf("uprocessable default branch flag in change in expression")
	}

	return value, nil
}

func (p *parser) ExcludedPathPatterns() ([]string, error) {
	excludedPaths := p.functionParams().Search("exclude")

	if !excludedPaths.Exists() {
		return []string{}, nil
	}

	result, ok := p.castToStringArray(excludedPaths.Data())
	if !ok {
		return []string{}, fmt.Errorf("uprocessable exclude path parameter in change in expression")
	}

	return result, nil
}

func (p *parser) TrackPipelineFile() (bool, error) {
	pipelineFile := p.functionParams().Search("pipeline_file")

	if !pipelineFile.Exists() {
		return p.whenPath[0] != "promotions", nil
	}

	value, ok := pipelineFile.Data().(string)
	if !ok {
		return false, fmt.Errorf("unknown value type pipeline_file in change_in expression")
	}

	switch value {
	case "track":
		return true, nil

	case "ignore":
		return false, nil
	}

	return false, fmt.Errorf("unknown value type pipeline_file in change_in expression")
}

func (p *parser) OnTags() (bool, error) {
	onTags := p.functionParams().Search("on_tags")

	if !onTags.Exists() {
		return true, nil
	}

	value, ok := onTags.Data().(bool)
	if !ok {
		return true, fmt.Errorf("unknown value type on_tags in change_in expression")
	}

	return value, nil
}

func (p *parser) DefaultRange(defaultBranch string) (string, error) {
	defaultRange := p.functionParams().Search("default_range")

	if !defaultRange.Exists() {
		return p.fetchCommitRange(defaultBranch), nil
	}

	value, ok := defaultRange.Data().(string)
	if !ok {
		return "", fmt.Errorf("unknown value type default_range in change_in expression")
	}

	return value, nil

}

func (p *parser) CommitRange(defaultBranch string) (string, error) {
	branchRange := p.functionParams().Search("branch_range")

	if !branchRange.Exists() {
		return p.fetchCommitRange(defaultBranch), nil
	}

	value, ok := branchRange.Data().(string)
	if !ok {
		return "", fmt.Errorf("unknown value type branch_range in change_in expression")
	}

	value = strings.ReplaceAll(value, "$SEMAPHORE_MERGE_BASE", environment.MergeBase())
	value = strings.ReplaceAll(value, "$SEMAPHORE_GIT_SHA", environment.CurrentGitSha())

	return value, nil
}

func (p *parser) fetchCommitRange(defaultBranch string) string {
	commitRange := environment.GitCommitRange()
	if commitRange != "" {
		return commitRange
	}

	return fmt.Sprintf("%s...%s", defaultBranch, environment.CurrentGitSha())
}

func (p *parser) functionParams() *gabs.Container {
	return p.ast.Search("params", "1")
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
