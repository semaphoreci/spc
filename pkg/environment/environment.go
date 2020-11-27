package environment

import (
	"os"
	"os/exec"
	"strings"
)

const GitRefTypeTag = "tag"

func GitRefType() string {
	return os.Getenv("SEMAPHORE_GIT_REF_TYPE")
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
	value := os.Getenv("SEMAPHORE_GIT_COMMIT_RANGE")

	if value == "" {
		panic("SEMAPHORE_GIT_REF_TYPE not set")
	}

	return value
}
