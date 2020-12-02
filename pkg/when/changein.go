package when

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
	Params  ChangeInFunctionParams
	Workdir string

	YamlPath string
	Location logs.Location

	diffList []string
}

func (f *ChangeInFunction) Eval() (bool, error) {
	var err error

	err = f.Fetch()
	if err != nil {
		return false, err
	}

	err = f.LoadDiffList()
	if err != nil {
		return false, err
	}

	if environment.GitRefType() == environment.GitRefTypeTag {
		fmt.Printf("  Running on a tag, skipping evaluation\n")

		return f.Params.OnTags, nil
	}

	fmt.Printf("  File Patterns: '%v'\n", f.Params.PathPatterns)
	fmt.Printf("  Exclude Patterns: '%v'\n", f.Params.ExcludedPathPatterns)
	fmt.Printf("  TrackPipelineFile: '%v'\n", f.Params.TrackPipelineFile)

	for _, diffLine := range f.diffList {
		fmt.Printf("  Checking diff line '%s'\n", diffLine)

		if f.MatchesPattern(diffLine) && !f.Excluded(diffLine) {
			return true, nil
		}
	}

	return false, nil
}

func (f *ChangeInFunction) Fetch() error {
	if environment.CurrentBranch() != f.Params.DefaultBranch {
		base, _ := f.ParseCommitRange()

		output, err := git.Fetch(base)
		return f.ParseFetchError(base, string(output), err)
	}

	return nil
}

func (f *ChangeInFunction) ParseFetchError(name string, output string, err error) error {
	if err != nil {
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

func (f *ChangeInFunction) CommitRange() string {
	currentBranch := environment.CurrentBranch()

	if currentBranch == f.Params.DefaultBranch {
		return f.Params.DefaultRange
	} else {
		return f.Params.CommitRange
	}
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
