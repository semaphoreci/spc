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

	// Whether DefaultRange is the range of the push itself, taken from
	// SEMAPHORE_GIT_COMMIT_RANGE, rather than one the pipeline set.
	defaultRangeFromPush bool
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

	defaultRangeFromPush := false

	if defaultRange == "" {
		defaultRange = fetchCommitRange(defaultBranch)
		defaultRangeFromPush = env.GitCommitRange() != ""
	}

	return &DiffSet{
		DefaultBranch:        defaultBranch,
		DefaultRange:         defaultRange,
		BranchRange:          branchRange,
		OnTags:               onTags,
		defaultRangeFromPush: defaultRangeFromPush,
	}
}

// UsesPushRange reports whether CommitRange is the push's own range: the
// commit the branch pointed to before the push and the one it points to now.
// Such a range is compared end to end rather than through a merge base.
func (d *DiffSet) UsesPushRange() bool {
	if !d.defaultRangeFromPush || d.runningOnPullRequest() {
		return false
	}

	return d.runningOnDefaultBranch() || d.BranchRange == "$SEMAPHORE_GIT_COMMIT_RANGE"
}

func (d *DiffSet) CommitRange() string {
	if d.runningOnPullRequest() {
		if d.runningOnForkedPullRequest() {
			return d.DefaultRange
		} else {
			return d.pullRequestRange()
		}
	} else {
		if d.runningOnDefaultBranch() {
			return d.DefaultRange
		} else {
			return d.branchRange()
		}
	}
}

func (d *DiffSet) IsEvaluationNeeded() bool {
	return d.runningOnGitTag()
}

func (d *DiffSet) IsGitFetchNeeded() (bool, []string) {
	// We don't need to fetch any branch, we are evaluating the
	// change in on the current branch. A pull request ignores branch_range
	// and diffs the target branch against the PR branch, so both still
	// need fetching even when branch_range names a base commit.
	if d.runningOnDefaultBranch() ||
		d.runningOnForkedPullRequest() ||
		(!d.runningOnPullRequest() && d.isBaseCommitSha()) {
		return false, nil
	}

	fetchTargets := make([]string, 1)

	commitRange := d.CommitRange()

	fetchTargets[0] = commitRangeBase(commitRange)

	if d.runningOnPullRequest() {
		fetchTargets = append(fetchTargets, commitRangeHead(commitRange))
	}

	return true, fetchTargets
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

func (d *DiffSet) runningOnGitTag() bool {
	return env.GitRefType() == env.GitRefTypeTag
}

func (d *DiffSet) runningOnPullRequest() bool {
	return env.GitRefType() == env.GitRefTypePullRequest
}

func (d *DiffSet) runningOnForkedPullRequest() bool {
	return d.runningOnPullRequest() &&
		env.PullRequestRepoSlug() != env.GitRepoSlug()
}

func (d *DiffSet) runningOnDefaultBranch() bool {
	return !d.runningOnPullRequest() &&
		env.CurrentBranch() == d.DefaultBranch
}

func (d *DiffSet) isBaseCommitSha() bool {
	return d.BranchRange == "$SEMAPHORE_GIT_COMMIT_RANGE" ||
		d.BranchRange == "$SEMAPHORE_GIT_SHA^...$SEMAPHORE_GIT_SHA"
}

// evaluating commit ranges

func fetchCommitRange(defaultBranch string) string {
	commitRange := env.GitCommitRange()
	if commitRange != "" {
		return commitRange
	}

	return fmt.Sprintf("%s...%s", defaultBranch, env.CurrentGitSha())
}

func (d *DiffSet) branchRange() string {
	if d.BranchRange == "$SEMAPHORE_GIT_COMMIT_RANGE" {
		return d.DefaultRange
	}
	if d.BranchRange == "$SEMAPHORE_GIT_SHA^...$SEMAPHORE_GIT_SHA" {
		return strings.ReplaceAll(d.BranchRange, "$SEMAPHORE_GIT_SHA", env.CurrentGitSha())
	}

	return standardBranchRange(d.BranchRange, d.DefaultBranch)
}

func standardBranchRange(branchRange string, defaultBranch string) string {
	branchRange = strings.ReplaceAll(branchRange, "$SEMAPHORE_MERGE_BASE", defaultBranch)
	branchRange = strings.ReplaceAll(branchRange, "$SEMAPHORE_GIT_SHA", env.CurrentGitSha())
	return branchRange
}

func (d *DiffSet) pullRequestRange() string {
	pullRequestRange := "$SEMAPHORE_MERGE_BASE...$SEMAPHORE_BRANCH_HEAD"
	pullRequestRange = strings.ReplaceAll(pullRequestRange, "$SEMAPHORE_MERGE_BASE", env.CurrentBranch())
	pullRequestRange = strings.ReplaceAll(pullRequestRange, "$SEMAPHORE_BRANCH_HEAD", env.PullRequestBranch())

	return pullRequestRange
}
