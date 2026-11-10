package git

import (
	"testing"

	assert "github.com/stretchr/testify/assert"

	env "github.com/semaphoreci/spc/pkg/environment"
)

const (
	testDefaultBranch = "master"
	testPRBranch      = "dev"
)

func TestIsGitFetchNeeded(t *testing.T) {
	cases := []struct {
		name          string
		refType       string
		prSlug        string
		branchRange   string
		expectFetch   bool
		expectTargets []string
	}{
		{
			name:        "branch with commit range",
			refType:     env.GitRefTypeBranch,
			branchRange: "$SEMAPHORE_GIT_COMMIT_RANGE",
			expectFetch: false,
		},
		{
			name:        "branch with parent commit range",
			refType:     env.GitRefTypeBranch,
			branchRange: "$SEMAPHORE_GIT_SHA^...$SEMAPHORE_GIT_SHA",
			expectFetch: false,
		},
		{
			name:          "branch with default range",
			refType:       env.GitRefTypeBranch,
			expectFetch:   true,
			expectTargets: []string{testDefaultBranch},
		},
		{
			name:          "pull request with default range",
			refType:       env.GitRefTypePullRequest,
			expectFetch:   true,
			expectTargets: []string{testDefaultBranch, testPRBranch},
		},
		{
			name:          "pull request with commit range",
			refType:       env.GitRefTypePullRequest,
			branchRange:   "$SEMAPHORE_GIT_COMMIT_RANGE",
			expectFetch:   true,
			expectTargets: []string{testDefaultBranch, testPRBranch},
		},
		{
			name:          "pull request with parent commit range",
			refType:       env.GitRefTypePullRequest,
			branchRange:   "$SEMAPHORE_GIT_SHA^...$SEMAPHORE_GIT_SHA",
			expectFetch:   true,
			expectTargets: []string{testDefaultBranch, testPRBranch},
		},
		{
			name:        "forked pull request with commit range",
			refType:     env.GitRefTypePullRequest,
			prSlug:      "fork/repo",
			branchRange: "$SEMAPHORE_GIT_COMMIT_RANGE",
			expectFetch: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			branch := "feature"
			if tc.refType == env.GitRefTypePullRequest {
				branch = testDefaultBranch
			}

			t.Setenv("SEMAPHORE_GIT_REF_TYPE", tc.refType)
			t.Setenv("SEMAPHORE_GIT_BRANCH", branch)
			t.Setenv("SEMAPHORE_GIT_PR_BRANCH", testPRBranch)
			t.Setenv("SEMAPHORE_GIT_REPO_SLUG", "org/repo")
			t.Setenv("SEMAPHORE_GIT_PR_SLUG", "org/repo")
			if tc.prSlug != "" {
				t.Setenv("SEMAPHORE_GIT_PR_SLUG", tc.prSlug)
			}
			t.Setenv("SEMAPHORE_GIT_SHA", fallbackSha)
			t.Setenv("SEMAPHORE_GIT_COMMIT_RANGE", "abc...def")

			diffSet := NewDiffSet(testDefaultBranch, "", tc.branchRange, false)
			fetch, targets := diffSet.IsGitFetchNeeded()

			assert.Equal(t, tc.expectFetch, fetch)
			assert.Equal(t, tc.expectTargets, targets)
		})
	}
}
