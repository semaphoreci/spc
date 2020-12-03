package changein

import (
	"fmt"
	"os/exec"
	"path"
	"strings"

	doublestar "github.com/bmatcuk/doublestar/v2"
	environment "github.com/semaphoreci/spc/pkg/environment"
	git "github.com/semaphoreci/spc/pkg/git"
	logs "github.com/semaphoreci/spc/pkg/logs"
)

type evaluator struct {
	function Function
	result   bool
	diffList []string
	err      error
}

func (e *evaluator) Run() (bool, erorr) {
	if e.isGitTag() {
		return e.evalForTags()
	}

	return e.evalForBranches()

}

func (e *evaluator) evalForTags() (bool, error) {
	fmt.Printf("Running on a tag, skipping evaluation")

	return e.function.ResultForGitTags(), nil
}

func (e *evaluator) isGitTag() bool {
	return environment.GitRefType() == environment.GitRefTypeTag
}

func (e *evaluator) evalForBranches() (bool, error) {
	err := e.FetchBranches()
	if err != nil {
		return false, err
	}

	diffSet, err = e.LoadDiffList()
	if err != nil {
		return false, err
	}

	return e.PatternMatchOnDiffList(), nil
}

func (f *ChangeInFunction) FetchBranches() error {
	if environment.CurrentBranch() != f.Params.DefaultBranch {
		base, _ := f.ParseCommitRange()

		output, err := git.Fetch(base)
		return f.ParseFetchError(base, string(output), err)
	}

	return nil
}

func (f *ChangeInFunction) ParseFetchError(name string, output string, err error) error {
	if err == nil {
		return nil
	}

	if strings.Contains(string(output), "couldn't find remote ref") {
		msg := fmt.Sprintf("Unknown git reference '%s'.", name)
		err := logs.ErrorChangeInMissingBranch{Message: msg, Location: f.Location}

		logs.Log(err)

		return &err
	}

	return err
}

func (e *evaluator) PatternMatchOnDiffList() bool {
	fmt.Printf("File Patterns: '%v'\n", f.Params.PathPatterns)
	fmt.Printf("Exclude Patterns: '%v'\n", f.Params.ExcludedPathPatterns)
	fmt.Printf("TrackPipelineFile: '%v'\n", f.Params.TrackPipelineFile)

	for _, diffLine := range f.diffList {
		fmt.Printf("  Checking diff line '%s'\n", diffLine)

		if f.Params.TrackPipelineFile && changeInPatternMatch(diffLine, "/"+f.YamlPath, f.Workdir) {
			fmt.Printf("    Matched tracked pipeline file %s\n", f.YamlPath)
			return true, nil
		}

		if f.MatchesPattern(diffLine) && !f.Excluded(diffLine) {
			return true, nil
		}
	}

	return false, nil
}

func (f *ChangeInFunction) MatchesPattern(diffLine string) bool {
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

func (f *ChangeInFunction) LoadDiffList() error {
	flags := []string{"diff", "--name-only", f.CommitRange()}
	fmt.Printf("  Running git %s\n", strings.Join(flags, " "))

	bytes, err := exec.Command("git", flags...).CombinedOutput()
	if err != nil {
		fmt.Println(string(bytes))

		return err
	}

	f.diffList = strings.Split(strings.TrimSpace(string(bytes)), "\n")

	return nil
}

func changeInPatternMatch(diffLine string, pattern string, workDir string) bool {
	pattern = preparePattern(pattern, workDir)
	diffLine = path.Clean("/" + diffLine)
	pattern = path.Clean(pattern)

	if strings.Contains(pattern, "*") {
		ok, err := doublestar.Match(pattern, diffLine)
		if err != nil {
			panic(err)
		}

		return ok
	}

	return strings.HasPrefix(diffLine, pattern)
}

func preparePattern(pattern, workDir string) string {
	if pattern[0] != '/' {
		return path.Join("/", workDir, pattern)
	}

	return pattern
}
