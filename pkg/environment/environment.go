package environment

import (
	"os"
	"os/exec"
	"strings"
)

const GitRefTypeTag = "tag"
const GitRefTypeBranch = "branch"

func GitRefType() string {
	value := os.Getenv("SEMAPHORE_GIT_REF_TYPE")
	if value != "" {
		return value
	}

	return GitRefTypeBranch
}

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

func MergeBase() string {
	value := os.Getenv("SEMAPHORE_MERGE_BASE")
	if value != "" {
		return value
	}

	return "master"
}
