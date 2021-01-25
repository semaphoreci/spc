package environment

import (
	"os"
	"os/exec"
	"strings"
)

const GitRefTypeTag = "tag"
const GitRefTypeBranch = "branch"
const GitRefTypePullRequest = "pull-request"

func GitRefType() string {
	value := os.Getenv("SEMAPHORE_GIT_REF_TYPE")
	if value != "" {
		return value
	}

	return GitRefTypeBranch
}

// In pipelines initiated by Pull Request, this environment variable
// points to branch that is the TARGET of pull request
func CurrentBranch() string {
	value := os.Getenv("SEMAPHORE_GIT_BRANCH")
	if value != "" {
		return value
	}

	gitBranch, err := exec.Command("git", "branch", "--show-current").CombinedOutput()
	if err != nil {
		panic(err)
	}

	return strings.TrimSpace(string(gitBranch))
}

// In pipelines initiated by Pull Requests, this environment variable
// points to branch that contains all the changes that should be merged
func PullRequestBranch() string {
	return os.Getenv("SEMAPHORE_GIT_PR_BRANCH")
}

func PullRequestRepoSlug() string {
	return os.Getenv("SEMAPHORE_GIT_PR_SLUG")
}

func GitRepoSlug() string {
	return os.Getenv("SEMAPHORE_GIT_REPO_SLUG")
}

func GitCommitRange() string {
	return os.Getenv("SEMAPHORE_GIT_COMMIT_RANGE")
}

func CurrentGitSha() string {
	value := os.Getenv("SEMAPHORE_GIT_SHA")
	if value != "" {
		return value
	}

	sha, err := exec.Command("git", "rev-parse", "HEAD").CombinedOutput()
	if err != nil {
		panic(err)
	}

	return strings.TrimSpace(string(sha))
}
