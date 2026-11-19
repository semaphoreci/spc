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

const (
	testPushRange         = "abc...def"
	testPushRangeVariable = "$SEMAPHORE_GIT_COMMIT_RANGE"
	testFeatureBranch     = "feature"
	testRepoSlug          = "org/repo"
	testPRSlugEnvVar      = "SEMAPHORE_GIT_PR_SLUG"
)

func TestUsesPushRange(t *testing.T) {
	cases := []struct {
		name         string
		refType      string
		branch       string
		prSlug       string
		commitRange  string
		defaultRange string
		branchRange  string
		expected     bool
	}{
		{
			name:        "default branch with the push range",
			refType:     env.GitRefTypeBranch,
			branch:      testDefaultBranch,
			commitRange: testPushRange,
			expected:    true,
		},
		{
			name:         "default branch with a range set in the pipeline",
			refType:      env.GitRefTypeBranch,
			branch:       testDefaultBranch,
			commitRange:  testPushRange,
			defaultRange: "HEAD~1..HEAD",
			expected:     false,
		},
		{
			name:     "default branch without a push range",
			refType:  env.GitRefTypeBranch,
			branch:   testDefaultBranch,
			expected: false,
		},
		{
			name:        "other branch measured by the push range",
			refType:     env.GitRefTypeBranch,
			branch:      testFeatureBranch,
			commitRange: testPushRange,
			branchRange: testPushRangeVariable,
			expected:    true,
		},
		{
			name:         "other branch measured by a range set in the pipeline",
			refType:      env.GitRefTypeBranch,
			branch:       testFeatureBranch,
			commitRange:  testPushRange,
			defaultRange: "HEAD~1..HEAD",
			branchRange:  testPushRangeVariable,
			expected:     false,
		},
		{
			name:        "other branch measured against the default branch",
			refType:     env.GitRefTypeBranch,
			branch:      testFeatureBranch,
			commitRange: testPushRange,
			expected:    false,
		},
		{
			name:        "pull request",
			refType:     env.GitRefTypePullRequest,
			branch:      testDefaultBranch,
			commitRange: testPushRange,
			expected:    false,
		},
		{
			name:        "forked pull request",
			refType:     env.GitRefTypePullRequest,
			branch:      testDefaultBranch,
			prSlug:      "fork/repo",
			commitRange: testPushRange,
			expected:    false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("SEMAPHORE_GIT_REF_TYPE", tc.refType)
			t.Setenv("SEMAPHORE_GIT_BRANCH", tc.branch)
			t.Setenv("SEMAPHORE_GIT_PR_BRANCH", testPRBranch)
			t.Setenv("SEMAPHORE_GIT_REPO_SLUG", testRepoSlug)
			t.Setenv(testPRSlugEnvVar, testRepoSlug)
			if tc.prSlug != "" {
				t.Setenv(testPRSlugEnvVar, tc.prSlug)
			}
			t.Setenv("SEMAPHORE_GIT_SHA", fallbackSha)
			t.Setenv("SEMAPHORE_GIT_COMMIT_RANGE", tc.commitRange)

			diffSet := NewDiffSet(testDefaultBranch, tc.defaultRange, tc.branchRange, false)

			assert.Equal(t, tc.expected, diffSet.UsesPushRange())
		})
	}
}
