package git

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// testCommitRange is the range every fixture below is built around: a feature
// branch measured against the default branch.
const testCommitRange = "master...feature"

// sentinelFileMode is the mode of the marker file that holds a failure window
// open.
const sentinelFileMode = 0o600

// The tests below exercise the deepening loop against a real shallow clone,
// with a git shim on PATH that makes 'git fetch --deepen' fail the way a
// concurrent rewrite of .git/shallow does.

func TestDiffListRecoversFromTransientDeepenFailure(t *testing.T) {
	setupShallowClone(t)

	// A deepen that fails after it has already taken effect, which is how git
	// behaves when it aborts on a shallow file it no longer recognises.
	installGitShim(t, gitShim{failAttempts: 1, applyDeepen: true})

	list, err := DiffList(testCommitRange)
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

	list, err := DiffList(testCommitRange)
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

// A commit range with no merge base is not a git failure: every deepen
// succeeds, there is simply no common ancestor to find. It has to stay
// distinguishable from a failing git command, because callers treat the two
// very differently.
func TestDiffListReportsUnresolvableRangeWhenNoGitCommandFails(t *testing.T) {
	setupOrphanClone(t)

	list, err := DiffList(testCommitRange)
	if err == nil {
		t.Fatalf("expected an unresolvable range error, got diff list %v", list)
	}

	if !errors.Is(err, ErrRangeUnresolvable) {
		t.Fatalf("expected ErrRangeUnresolvable, got: %v", err)
	}

	if strings.Contains(err.Error(), "failed to deepen") {
		t.Fatalf("an unresolvable range must not be reported as a deepen failure, got: %v", err)
	}
}

// The retry exists to outlast a competing git process holding .git/shallow,
// so it has to wait between attempts. A shim that fails for a fixed stretch of
// wall-clock time rather than a fixed number of attempts is what distinguishes
// a retry that backs off from one that just burns through its budget.
func TestDiffListRetryOutlastsATimedFailureWindow(t *testing.T) {
	setupShallowClone(t)

	// Long enough that the retries only get past it by waiting. With no
	// backoff the attempts all land inside the window and exhaust the budget,
	// which is the regression this pins.
	const failureWindow = 800 * time.Millisecond

	// applyDeepen is off on purpose: while the window is open no deepen takes
	// effect, so the range stays unresolvable and the loop has to keep retrying.
	installGitShim(t, gitShim{failWindow: failureWindow, applyDeepen: false})

	start := time.Now()

	list, err := DiffList(testCommitRange)
	if err != nil {
		t.Fatalf("expected the retry to outlast the failure window, got error: %v", err)
	}

	if elapsed := time.Since(start); elapsed < failureWindow {
		t.Fatalf("expected the retry to still be going after %s, finished in %s", failureWindow, elapsed)
	}

	if !containsLine(list, "lib/feature.txt") {
		t.Fatalf("expected lib/feature.txt in the diff, got %v", list)
	}
}

// setupOrphanClone builds an origin whose feature branch is an orphan, so it
// shares no history with the default branch and no amount of deepening will
// produce a merge base.
func setupOrphanClone(t *testing.T) {
	t.Helper()

	isolateGitConfig(t)

	root := t.TempDir()
	origin := filepath.Join(root, "origin")
	clone := filepath.Join(root, "clone")

	runScript(t, fmt.Sprintf(`
mkdir -p %[1]q && cd %[1]q
git init -q -b master .
mkdir lib && echo base > lib/base.txt
git add -A && git commit -qm base
for i in $(seq 1 3); do git commit -q --allow-empty -m "master $i"; done
git checkout -q --orphan feature
git rm -rqf . >/dev/null 2>&1 || true
mkdir lib && echo changed > lib/feature.txt
git add -A && git commit -qm "orphan root"
git checkout -q master
git clone -q --depth 1 --single-branch --branch feature "file://%[1]s" %[2]q
cd %[2]q && git checkout -q --detach
`, origin, clone))

	t.Chdir(clone)
	resetMemoizedGitState()

	if _, err := Fetch("master"); err != nil {
		t.Fatalf("failed to fetch master: %v", err)
	}
}

// setupShallowClone builds an origin whose default branch has moved ahead of
// the feature branch, then makes a depth-1 clone of the feature branch the
// working directory. The merge base sits outside that depth, so resolving the
// commit range requires deepening the clone.
func setupShallowClone(t *testing.T) {
	t.Helper()

	isolateGitConfig(t)

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

// isolateGitConfig keeps the developer's own git configuration out of the test
// repositories. Commit signing in particular fails these tests on machines
// where it is enabled globally but the agent cannot prompt.
func isolateGitConfig(t *testing.T) {
	t.Helper()

	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_SYSTEM", os.DevNull)
}

// The git shim has to be executable.
const shimFileMode = 0o755

type gitShim struct {
	// How many 'fetch --deepen' invocations to fail before letting them through.
	failAttempts int

	// How long to keep failing 'fetch --deepen' regardless of how many
	// attempts arrive. Takes precedence over failAttempts when set.
	//
	// The window is implemented with a sentinel file the test removes once it
	// elapses, rather than clock arithmetic inside the shim, because the shells
	// this runs on do not agree on how to read a sub-second timestamp.
	failWindow time.Duration

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

	// Fail while the window is open, or while the attempt count is still under
	// the limit, depending on how the shim is configured.
	stillFailing := fmt.Sprintf(`[ "$attempts" -le %d ]`, shim.failAttempts)

	if shim.failWindow > 0 {
		stillFailing = fmt.Sprintf(`[ -f %q ]`, openWindow(t, shimDir, shim.failWindow))
	}

	body := fmt.Sprintf(`#!/bin/bash
if [[ "$*" == *"--deepen"* ]]; then
  attempts=$(cat %[1]q 2>/dev/null || echo 0)
  attempts=$((attempts + 1))
  echo "$attempts" > %[1]q

  if %[2]s; then
%[3]s    echo "fatal: shallow file has changed since we read it" >&2
    exit 128
  fi
fi

exec %[4]q "$@"
`, counter, stillFailing, deepenBody, gitPath)

	shimPath := filepath.Join(shimDir, "git")

	// #nosec G306 -- the shim has to be executable
	if err := os.WriteFile(shimPath, []byte(body), shimFileMode); err != nil {
		t.Fatalf("failed to write the git shim: %v", err)
	}

	t.Setenv("PATH", shimDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	resetMemoizedGitState()
}

// openWindow creates a sentinel file and removes it once d has elapsed. The
// git shim fails for as long as the file is there, which gives the test a
// stretch of wall-clock time during which deepening cannot succeed no matter
// how many attempts arrive.
func openWindow(t *testing.T, dir string, d time.Duration) string {
	t.Helper()

	sentinel := filepath.Join(dir, "failure-window")

	if err := os.WriteFile(sentinel, []byte("open"), sentinelFileMode); err != nil {
		t.Fatalf("failed to open the failure window: %v", err)
	}

	timer := time.AfterFunc(d, func() { _ = os.Remove(sentinel) })
	t.Cleanup(func() { timer.Stop() })

	return sentinel
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
