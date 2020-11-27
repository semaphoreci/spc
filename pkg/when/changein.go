package when

import (
	"fmt"
	"os/exec"
	"path"
	"strings"

	gabs "github.com/Jeffail/gabs/v2"
	doublestar "github.com/bmatcuk/doublestar/v2"
	environment "github.com/semaphoreci/spc/pkg/environment"
)

type ChangeInFunctionParams struct {
	PathPatterns         []string
	ExcludedPathPatterns []string
	DefaultBranch        string
	TrackPipelineFile    bool
	OnTags               bool
}

type ChangeInFunction struct {
	Params   ChangeInFunctionParams
	Workdir  string
	YamlPath string

	diffList []string
}

func NewChangeInFunctionFromWhenInputList(when *WhenExpression, input *gabs.Container, yamlPath string) (*ChangeInFunction, error) {
	params := ChangeInFunctionParams{
		PathPatterns:         []string{},
		ExcludedPathPatterns: []string{},
		DefaultBranch:        "master",
		TrackPipelineFile:    true,
		OnTags:               true,
	}

	if input.Exists("params", "1", "default_branch") {
		params.DefaultBranch = input.Search("params", "1", "default_branch").Data().(string)
	}

	if _, ok := input.Search("params", "0").Data().([]interface{}); ok {
		for _, p := range input.Search("params", "0").Children() {
			params.PathPatterns = append(params.PathPatterns, p.Data().(string))
		}
	} else {
		params.PathPatterns = append(params.PathPatterns, input.Search("params", "0").Data().(string))
	}

	if _, ok := input.Search("params", "1", "exclude").Data().([]interface{}); ok {
		for _, p := range input.Search("params", "1", "exclude").Children() {
			params.ExcludedPathPatterns = append(params.ExcludedPathPatterns, p.Data().(string))
		}
	}

	if input.Exists("params", "1", "pipeline_file") {
		value, ok := input.Search("params", "1", "pipeline_file").Data().(string)
		if !ok {
			return nil, fmt.Errorf("Unknown value type pipeline_file in change_in expression.")
		}

		switch value {
		case "track":
			params.TrackPipelineFile = true
		case "ignore":
			params.TrackPipelineFile = false
		default:
			return nil, fmt.Errorf("Unknown value type pipeline_file in change_in expression.")
		}
	} else {
		if when.Path[0] == "promotions" {
			params.TrackPipelineFile = false
		} else {
			params.TrackPipelineFile = true
		}
	}

	if input.Exists("params", "1", "on_tags") {
		value, ok := input.Search("params", "1", "on_tags").Data().(bool)
		if !ok {
			return nil, fmt.Errorf("Unknown value type on_tags in change_in expression.")
		}

		params.OnTags = value
	}

	fun := &ChangeInFunction{
		Workdir:  path.Dir(yamlPath),
		YamlPath: yamlPath,
		Params:   params,
	}

	return fun, nil
}

func (f *ChangeInFunction) DefaultBranchExists() bool {
	err := exec.Command(
		"git",
		"rev-parse",
		"--verify",
		f.Params.DefaultBranch,
	).Run()

	return err == nil
}

func (f *ChangeInFunction) Eval() bool {
	f.LoadDiffList()

	if environment.GitRefType() == environment.GitRefTypeTag {
		fmt.Printf("  Running on a tag, skipping evaluation\n")

		return f.Params.OnTags
	}

	fmt.Printf("  File Patterns: '%v'\n", f.Params.PathPatterns)
	fmt.Printf("  Exclude Patterns: '%v'\n", f.Params.ExcludedPathPatterns)
	fmt.Printf("  TrackPipelineFile: '%v'\n", f.Params.TrackPipelineFile)

	for _, diffLine := range f.diffList {
		fmt.Printf("  Checking diff line '%s'\n", diffLine)

		if f.MatchesPattern(diffLine) && !f.Excluded(diffLine) {
			return true
		}
	}

	return false
}

func (f *ChangeInFunction) MatchesPattern(diffLine string) bool {
	if f.Params.TrackPipelineFile && changeInPatternMatch(diffLine, "/"+f.YamlPath, f.Workdir) {
		fmt.Printf("    Matched tracked pipeline file %s\n", f.YamlPath)
		return true
	}

	for _, pathPattern := range f.Params.PathPatterns {
		if changeInPatternMatch(diffLine, pathPattern, f.Workdir) {
			fmt.Printf("    Matched pattern %s\n", pathPattern)
			return true
		}
	}

	return false
}

func (f *ChangeInFunction) Excluded(diffLine string) bool {
	for _, pathPattern := range f.Params.ExcludedPathPatterns {
		if changeInPatternMatch(diffLine, pathPattern, f.Workdir) {
			fmt.Printf("    Excluded with pattern %s\n", pathPattern)
			return true
		}
	}

	return false
}

func changeInPatternMatch(diffLine string, pattern string, workDir string) bool {
	if pattern[0] != '/' {
		pattern = path.Join("/", workDir, pattern)
	}

	diffLine = path.Clean("/" + diffLine)
	pattern = path.Clean(pattern)

	if strings.Contains(pattern, "*") {
		ok, err := doublestar.Match(pattern, diffLine)
		if err != nil {
			panic(err)
		}

		return ok
	} else {
		return strings.HasPrefix(diffLine, pattern)
	}

	return false
}

func (f *ChangeInFunction) LoadDiffList() {
	gitOpts := []string{
		"diff",
		"--name-only",
		fmt.Sprintf("%s..HEAD", f.Params.DefaultBranch),
	}

	bytes, err := exec.Command("git", gitOpts...).CombinedOutput()
	if err != nil {
		panic(err)
	}

	f.diffList = strings.Split(strings.TrimSpace(string(bytes)), "\n")
}
