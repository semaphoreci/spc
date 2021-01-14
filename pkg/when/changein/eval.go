package changein

import (
	"fmt"
	"strings"

	environment "github.com/semaphoreci/spc/pkg/environment"
	git "github.com/semaphoreci/spc/pkg/git"
	logs "github.com/semaphoreci/spc/pkg/logs"
)

func Eval(fun *Function) (bool, error) {
	e := evaluator{function: fun}

	return e.Run()
}

type evaluator struct {
	function *Function
	result   bool
	diffList []string
	err      error
}

const ThreeDots = "..."
const TwoDots = ".."

func (e *evaluator) Run() (bool, error) {
	fmt.Println("Processing change_in function")
	fmt.Println("Params:")
	fmt.Printf("  - File Patterns: '%v'\n", e.function.PathPatterns)
	fmt.Printf("  - Exclude Patterns: '%v'\n", e.function.ExcludedPathPatterns)
	fmt.Printf("  - TrackPipelineFile: '%v'\n", e.function.TrackPipelineFile)

	if e.runningOnGitTag() {
		fmt.Printf("Running on a tag, skipping evaluation")
		return e.function.OnTags, nil
	}

	err := e.FetchBranches()
	if err != nil {
		return false, err
	}

	diffList, err := e.LoadDiffList()
	if err != nil {
		return false, err
	}

	result := e.function.HasMatchesInDiffList(diffList)

	return result, nil
}

func (e *evaluator) runningOnGitTag() bool {
	return environment.GitRefType() == environment.GitRefTypeTag
}

func (e *evaluator) runningOnPullRequest() bool {
	return environment.GitRefType() == environment.GitRefTypePullRequest
}

func (e *evaluator) runningOnDefaultBranch() bool {
	if e.runningOnPullRequest() {
		return false
	}

	return environment.CurrentBranch() == e.function.DefaultBranch
}

func (e *evaluator) CommitRangeBase() string {
	var splitAt string

	if strings.Contains(e.CommitRange(), ThreeDots) {
		splitAt = ThreeDots
	} else {
		splitAt = TwoDots
	}

	parts := strings.Split(e.CommitRange(), splitAt)

	return parts[0]
}

func (e *evaluator) CommitRangeHead() string {
	var splitAt string

	if strings.Contains(e.CommitRange(), ThreeDots) {
		splitAt = ThreeDots
	} else {
		splitAt = TwoDots
	}

	parts := strings.Split(e.CommitRange(), splitAt)

	return parts[1]
}

func (e *evaluator) FetchBranches() error {
	if e.runningOnDefaultBranch() {
		// We don't need to fetch any branch, we are evaluating the
		// change in on the current branch.
		return nil
	}

	if e.runningOnPullRequest() {
		pullRequestBranch := e.CommitRangeHead()

		result, err := git.Fetch(pullRequestBranch)
		if err != nil {
			return e.ParseFetchError(pullRequestBranch, string(result), err)
		}
	}

	base := e.CommitRangeBase()

	output, err := git.Fetch(base)
	if err != nil {
		return e.ParseFetchError(base, string(output), err)
	}

	return e.ParseFetchError(base, string(output), err)
}

func (e *evaluator) ParseFetchError(name string, output string, err error) error {
	if strings.Contains(string(output), "couldn't find remote ref") {
		msg := fmt.Sprintf("Unknown git reference '%s'.", name)
		err := logs.ErrorChangeInMissingBranch{Message: msg, Location: e.function.Location}

		logs.Log(err)

		return &err
	}

	return err
}

func (e *evaluator) LoadDiffList() ([]string, error) {
	list, _, err := git.Diff(e.CommitRange())
	if err != nil {
		return list, err
	}

	return list, nil
}

func (e *evaluator) CommitRange() string {
	if e.runningOnPullRequest() {
		return e.function.PullRequestRange
	} else {
		if e.runningOnDefaultBranch() {
			return e.function.DefaultRange
		} else {
			return e.function.BranchRange
		}
	}
}
