package git

import (
	"fmt"
	"strings"

	env "github.com/semaphoreci/spc/pkg/environment"
)

const implicitBranchRange = "$SEMAPHORE_MERGE_BASE...$SEMAPHORE_GIT_SHA"
const implicitDefaultBranch = "master"

const threeDots = "..."
const twoDots = ".."

type DiffSet struct {
	DefaultBranch string
	DefaultRange  string
	BranchRange   string
	OnTags        bool
}

func NewDiffSet(
	defaultBranch string,
	defaultRange string,
	branchRange string,
	onTags bool,
) *DiffSet {

	if branchRange == "" {
		branchRange = implicitBranchRange
	}

	if defaultBranch == "" {
		defaultBranch = implicitDefaultBranch
	}

	if defaultRange == "" {
		defaultRange = fetchCommitRange(defaultBranch)
	}

	return &DiffSet{
		DefaultBranch: defaultBranch,
		DefaultRange:  defaultRange,
		BranchRange:   branchRange,
		OnTags:        onTags,
	}
}

func (r *DiffSet) CommitRange() string {
	if r.runningOnPullRequest() {
		if r.runningOnForkedPullRequest() {
			return r.DefaultRange
		} else {
			return r.pullRequestRange()
		}
	} else {
		if r.runningOnDefaultBranch() {
			return r.DefaultRange
		} else {
			return r.branchRange()
		}
	}
}

func (r *DiffSet) IsEvaluationNeeded() bool {
	return r.runningOnGitTag()
}

func (r *DiffSet) IsGitFetchNeeded() (bool, string) {
	// We don't need to fetch any branch, we are evaluating the
	// change in on the current branch.
	if r.runningOnDefaultBranch() ||
		r.runningOnForkedPullRequest() ||
		r.isBaseCommitSha() {
		return false, ""
	}

	commitRange := r.CommitRange()
	if r.runningOnPullRequest() {
		return true, commitRangeHead(commitRange)
	} else {
		return true, commitRangeBase(commitRange)
	}
}

// commit range helpers

func commitRangeBase(commitRange string) string {
	return splitCommitRange(commitRange)[0]
}

func commitRangeHead(commitRange string) string {
	return splitCommitRange(commitRange)[1]
}

func splitCommitRange(commitRange string) []string {
	var splitAt string

	if strings.Contains(commitRange, threeDots) {
		splitAt = threeDots
	} else {
		splitAt = twoDots
	}

	return strings.Split(commitRange, splitAt)
}

// running environment flags

func (e *DiffSet) runningOnGitTag() bool {
	return env.GitRefType() == env.GitRefTypeTag
}

func (r *DiffSet) runningOnPullRequest() bool {
	return env.GitRefType() == env.GitRefTypePullRequest
}

func (r *DiffSet) runningOnForkedPullRequest() bool {
	return r.runningOnPullRequest() &&
		env.PullRequestRepoSlug() != env.GitRepoSlug()
}

func (r *DiffSet) runningOnDefaultBranch() bool {
	return !r.runningOnPullRequest() &&
		env.CurrentBranch() == r.DefaultBranch
}

func (r *DiffSet) isBaseCommitSha() bool {
	return r.BranchRange == "$SEMAPHORE_GIT_COMMIT_RANGE" ||
		r.BranchRange == "$SEMAPHORE_GIT_SHA^...$SEMAPHORE_GIT_SHA"
}

// evaluating commit ranges

func fetchCommitRange(defaultBranch string) string {
	commitRange := env.GitCommitRange()
	if commitRange != "" {
		return commitRange
	}

	return fmt.Sprintf("%s...%s", defaultBranch, env.CurrentGitSha())
}

func (r *DiffSet) branchRange() string {
	if r.BranchRange == "$SEMAPHORE_GIT_COMMIT_RANGE" {
		return r.DefaultRange
	}
	if r.BranchRange == "$SEMAPHORE_GIT_SHA^...$SEMAPHORE_GIT_SHA" {
		return strings.ReplaceAll(r.BranchRange, "$SEMAPHORE_GIT_SHA", env.CurrentGitSha())
	}

	return standardBranchRange(r.BranchRange, r.DefaultBranch)
}

func standardBranchRange(branchRange string, defaultBranch string) string {
	branchRange = strings.ReplaceAll(branchRange, "$SEMAPHORE_MERGE_BASE", defaultBranch)
	branchRange = strings.ReplaceAll(branchRange, "$SEMAPHORE_GIT_SHA", env.CurrentGitSha())
	return branchRange
}

func (r *DiffSet) pullRequestRange() string {
	pullRequestRange := "$SEMAPHORE_MERGE_BASE...$SEMAPHORE_BRANCH_HEAD"
	pullRequestRange = strings.ReplaceAll(pullRequestRange, "$SEMAPHORE_MERGE_BASE", env.CurrentBranch())
	pullRequestRange = strings.ReplaceAll(pullRequestRange, "$SEMAPHORE_BRANCH_HEAD", env.PullRequestBranch())

	return pullRequestRange
}
