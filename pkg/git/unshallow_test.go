package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The tests below exercise the deepening loop against a real shallow clone,
// with a git shim on PATH that makes 'git fetch --deepen' fail the way a
// concurrent rewrite of .git/shallow does.

func TestDiffListRecoversFromTransientDeepenFailure(t *testing.T) {
	setupShallowClone(t)

	// A deepen that fails after it has already taken effect, which is how git
	// behaves when it aborts on a shallow file it no longer recognises.
	installGitShim(t, gitShim{failAttempts: 1, applyDeepen: true})

	list, err := DiffList("master...feature")
	if err != nil {
		t.Fatalf("expected the deepen to be retried, got error: %v", err)
	}

	if !containsLine(list, "lib/feature.txt") {
		t.Fatalf("expected lib/feature.txt in the diff, got %v", list)
	}
}

func TestDiffListReturnsErrorWhenDeepenKeepsFailing(t *testing.T) {
	setupShallowClone(t)

	// A deepen that never takes effect, e.g. an unreachable remote.
	installGitShim(t, gitShim{failAttempts: MaxUnshallowIterations, applyDeepen: false})

	list, err := DiffList("master...feature")
	if err == nil {
		t.Fatalf("expected an error when deepening keeps failing, got diff list %v", list)
	}

	if !strings.Contains(err.Error(), "failed to deepen the git clone") {
		t.Fatalf("expected a deepen failure error, got: %v", err)
	}

	if len(list) != 0 {
		t.Fatalf("expected an empty diff list alongside the error, got %v", list)
	}
}

// setupShallowClone builds an origin whose default branch has moved ahead of
// the feature branch, then makes a depth-1 clone of the feature branch the
// working directory. The merge base sits outside that depth, so resolving the
// commit range requires deepening the clone.
func setupShallowClone(t *testing.T) {
	t.Helper()

	root := t.TempDir()
	origin := filepath.Join(root, "origin")
	clone := filepath.Join(root, "clone")

	runScript(t, fmt.Sprintf(`
mkdir -p %[1]q && cd %[1]q
git init -q -b master .
mkdir lib && echo base > lib/base.txt
git add -A && git commit -qm base
git branch feature
for i in $(seq 1 5); do git commit -q --allow-empty -m "master $i"; done
git checkout -q feature
echo changed > lib/feature.txt
git add -A && git commit -qm "change in lib"
git checkout -q master
git clone -q --depth 1 --single-branch --branch feature "file://%[1]s" %[2]q
cd %[2]q && git checkout -q --detach
`, origin, clone))

	t.Chdir(clone)
	resetMemoizedGitState()

	// spc fetches the base branch before it resolves the commit range.
	if _, err := Fetch("master"); err != nil {
		t.Fatalf("failed to fetch master: %v", err)
	}
}

// The git shim has to be executable.
const shimFileMode = 0o755

type gitShim struct {
	// How many 'fetch --deepen' invocations to fail before letting them through.
	failAttempts int

	// Whether a failing invocation still performs the deepen before reporting
	// the failure.
	applyDeepen bool
}

// installGitShim puts a git wrapper on PATH that fails 'fetch --deepen'
// according to the given shim configuration.
func installGitShim(t *testing.T, shim gitShim) {
	t.Helper()

	gitPath := lookupGit(t)
	shimDir := t.TempDir()
	counter := filepath.Join(shimDir, "deepen-attempts")

	deepenBody := ""
	if shim.applyDeepen {
		deepenBody = fmt.Sprintf("    %q \"$@\" >/dev/null 2>&1\n", gitPath)
	}

	body := fmt.Sprintf(`#!/bin/bash
if [[ "$*" == *"--deepen"* ]]; then
  attempts=$(cat %[1]q 2>/dev/null || echo 0)
  attempts=$((attempts + 1))
  echo "$attempts" > %[1]q

  if [ "$attempts" -le %[2]d ]; then
%[3]s    echo "fatal: shallow file has changed since we read it" >&2
    exit 128
  fi
fi

exec %[4]q "$@"
`, counter, shim.failAttempts, deepenBody, gitPath)

	shimPath := filepath.Join(shimDir, "git")

	// #nosec G306 -- the shim has to be executable
	if err := os.WriteFile(shimPath, []byte(body), shimFileMode); err != nil {
		t.Fatalf("failed to write the git shim: %v", err)
	}

	t.Setenv("PATH", shimDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	resetMemoizedGitState()
}

func runScript(t *testing.T, script string) {
	t.Helper()

	lookupGit(t)

	cmd := exec.Command("bash", "-c", "set -e\n"+script)
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=test",
		"GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=test",
		"GIT_COMMITTER_EMAIL=test@example.com",
	)

	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build the test repository: %v\n%s", err, output)
	}
}

func lookupGit(t *testing.T) string {
	t.Helper()

	gitPath, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git is not available")
	}

	return gitPath
}

// The git package memoizes fetches and diffs process-wide, so every test has
// to start from a clean slate.
func resetMemoizedGitState() {
	fetchedBranches = map[string]string{}
	evaluatedDiffs = map[string][]string{}
}

func containsLine(lines []string, wanted string) bool {
	for _, line := range lines {
		if strings.TrimSpace(line) == wanted {
			return true
		}
	}

	return false
}
