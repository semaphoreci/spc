package when

import (
	"fmt"
	"os/exec"
	"path"
	"strings"

	doublestar "github.com/bmatcuk/doublestar/v2"
	environment "github.com/semaphoreci/spc/pkg/environment"
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

type ChangeInFunction struct {
	Params   ChangeInFunctionParams
	Workdir  string
	YamlPath string

	diffList []string
}

func (f *ChangeInFunction) DefaultBranchExists() bool {
	err := exec.Command("git", "rev-parse", "--verify", f.Params.DefaultBranch).Run()

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

func (f *ChangeInFunction) LoadDiffList() {
	bytes, err := exec.Command("git", "diff", "--name-only", f.CommitRange()).CombinedOutput()
	if err != nil {
		fmt.Println(string(bytes))
		panic(err)
	}

	f.diffList = strings.Split(strings.TrimSpace(string(bytes)), "\n")
}

func (f *ChangeInFunction) CommitRange() string {
	currentBranch := environment.CurrentBranch()

	if currentBranch == f.Params.DefaultBranch {
		return f.Params.DefaultRange
	} else {
		return f.Params.CommitRange
	}
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
