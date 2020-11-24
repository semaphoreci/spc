package pipelines

import (
	"fmt"
	"os/exec"
	"path"
	"strings"

	doublestar "github.com/bmatcuk/doublestar/v2"
)

type ChangeInFunctionParams struct {
	PathPatterns  []string
	DefaultBranch string
}

type ChangeInFunction struct {
	Params  ChangeInFunctionParams
	Workdir string

	diffList []string
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

	for _, pathPattern := range f.Params.PathPatterns {
		if len(pathPattern) == 0 {
			continue
		}

		for _, diffLine := range f.diffList {
			if changeInPatternMatch(diffLine, pathPattern, f.Workdir) {
				fmt.Printf("Change In: Matched '%s' with '%s' in the git diff list\n", pathPattern, diffLine)

				return true
			}
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
