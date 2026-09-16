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

// A failing 'git merge-base' is not the same as git reporting that there is
// no merge base. Folding the two together resolves the condition to false
// while the repository is actually broken, and sends the operator off to look
// at branch topology.
//
// The clone here is complete, which is what makes the failure conclusive:
// there is no history left to fetch, so deepening cannot be the answer.
func TestDiffListFailsWhenMergeBaseFailsOnACompleteClone(t *testing.T) {
	setupCompleteClone(t)

	invocations := filepath.Join(t.TempDir(), "invocations")

	installGitShim(t, gitShim{
		failSubcommand: "merge-base",
		failExitCode:   128,
		failMessage:    "fatal: bad object HEAD",
		recordTo:       invocations,
	})

	list, err := DiffList(testCommitRange)
	if err == nil {
		t.Fatalf("expected a git failure to be reported, got diff list %v", list)
	}

	if errors.Is(err, ErrRangeUnresolvable) {
		t.Fatalf("a broken git must not be reported as an unresolvable range, got: %v", err)
	}

	if !strings.Contains(err.Error(), "merge-base") {
		t.Fatalf("expected the error to name the failing command, got: %v", err)
	}

	// git's own diagnostic has to survive into the error, otherwise every
	// cause arrives as a bare exit status.
	if !strings.Contains(err.Error(), "bad object HEAD") {
		t.Fatalf("expected git's diagnostic in the error, got: %v", err)
	}

	// Nothing to deepen, so nothing is attempted.
	if recorded := readFile(t, invocations); strings.Contains(recorded, "--deepen") {
		t.Fatalf("expected no deepening on a complete clone, git ran:\n%s", recorded)
	}
}

// The same failure in a shallow clone is not conclusive, because git reports
// an object that is merely outside the current depth the same way it reports
// a broken one. Deepening is tried first, and the range still must not end up
// silently false.
func TestDiffListKeepsDeepeningWhenMergeBaseFailsOnAShallowClone(t *testing.T) {
	setupShallowClone(t)
	withDeepenBudget(t, 2)

	invocations := filepath.Join(t.TempDir(), "invocations")

	installGitShim(t, gitShim{
		failSubcommand: "merge-base",
		failExitCode:   128,
		failMessage:    "fatal: Not a valid commit name deadbeef",
		recordTo:       invocations,
	})

	list, err := DiffList(testCommitRange)
	if err == nil {
		t.Fatalf("expected an error, got diff list %v", list)
	}

	if errors.Is(err, ErrRangeUnresolvable) {
		t.Fatalf("an inconclusive answer must not be reported as an unresolvable range, got: %v", err)
	}

	if recorded := readFile(t, invocations); !strings.Contains(recorded, "--deepen") {
		t.Fatalf("expected deepening to be attempted first, git ran:\n%s", recorded)
	}
}

// The final deepen has to count. Checking only at the top of the loop throws
// away whatever the last one fetched, which is the one most likely to have
// reached the merge base.
func TestDiffListResolvesWhenTheLastDeepenSucceeds(t *testing.T) {
	setupShallowClone(t)

	// The fixture needs exactly one deepen, so a budget of one means the
	// answer only becomes available after the very last one.
	withDeepenBudget(t, 1)

	list, err := DiffList(testCommitRange)
	if err != nil {
		t.Fatalf("expected the last deepen to be taken into account, got error: %v", err)
	}

	if !containsLine(list, "lib/feature.txt") {
		t.Fatalf("expected lib/feature.txt in the diff, got %v", list)
	}
}

// Running out of budget on a clone that is still shallow is not a finding
// that the branches share no history. It is a failure to find out, and must
// not be quietly turned into false.
func TestDiffListFailsWhenBudgetRunsOutWhileStillShallow(t *testing.T) {
	setupShallowClone(t)
	withDeepenBudget(t, 2)

	// Deepening reports success without fetching anything, so the clone stays
	// shallow and the range stays unresolved.
	installGitShim(t, gitShim{deepenIsNoop: true})

	list, err := DiffList(testCommitRange)
	if err == nil {
		t.Fatalf("expected an error when the budget runs out, got diff list %v", list)
	}

	if errors.Is(err, ErrRangeUnresolvable) {
		t.Fatalf("an exhausted budget must not be reported as an unresolvable range, got: %v", err)
	}

	if !strings.Contains(err.Error(), "still shallow") {
		t.Fatalf("expected the error to say the clone is still shallow, got: %v", err)
	}
}

// Deciding a range has no merge base costs a full round of deepen fetches,
// and the answer cannot change while the process runs. Pipelines gate many
// blocks on the same range, so repeating that work multiplies it by the
// number of conditions.
func TestDiffListMemoizesTheUnresolvableVerdict(t *testing.T) {
	setupOrphanClone(t)

	if _, err := DiffList(testCommitRange); !errors.Is(err, ErrRangeUnresolvable) {
		t.Fatalf("expected the range to be unresolvable, got: %v", err)
	}

	invocations := filepath.Join(t.TempDir(), "invocations")

	// Record from here on, without resetting what the package memorized.
	installRecordingShim(t, invocations)

	if _, err := DiffList(testCommitRange); !errors.Is(err, ErrRangeUnresolvable) {
		t.Fatalf("expected the memorized verdict to be returned, got: %v", err)
	}

	if recorded := readFile(t, invocations); strings.TrimSpace(recorded) != "" {
		t.Fatalf("expected no further git commands, git ran:\n%s", recorded)
	}
}

// setupCompleteClone builds the same fixture as setupShallowClone but clones
// the whole history, so that no failure can be blamed on missing objects.
func setupCompleteClone(t *testing.T) {
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
git clone -q "file://%[1]s" %[2]q
cd %[2]q && git checkout -q --detach
`, origin, clone))

	t.Chdir(clone)
	resetMemoizedGitState()

	if _, err := Fetch("master"); err != nil {
		t.Fatalf("failed to fetch master: %v", err)
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

	// A git subcommand to fail on every call, with this exit code and message.
	// Used to tell a broken git apart from a git that is working fine and
	// reporting an honest negative answer.
	failSubcommand string
	failExitCode   int
	failMessage    string

	// Report 'fetch --deepen' as successful without deepening anything, which
	// is what git does once there is nothing left to fetch.
	deepenIsNoop bool

	// Record every git invocation to this file.
	recordTo string

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

	record := ""
	if shim.recordTo != "" {
		record = fmt.Sprintf("echo \"$*\" >> %q\n", shim.recordTo)
	}

	brokenSubcommand := ""
	if shim.failSubcommand != "" {
		brokenSubcommand = fmt.Sprintf(`if [[ "$1" == %q ]]; then
  echo %q >&2
  exit %d
fi
`, shim.failSubcommand, shim.failMessage, shim.failExitCode)
	}

	deepenNoop := ""
	if shim.deepenIsNoop {
		deepenNoop = `if [[ "$*" == *"--deepen"* ]]; then
  exit 0
fi
`
	}

	body := fmt.Sprintf(`#!/bin/bash
%[5]s%[6]s%[7]sif [[ "$*" == *"--deepen"* ]]; then
  attempts=$(cat %[1]q 2>/dev/null || echo 0)
  attempts=$((attempts + 1))
  echo "$attempts" > %[1]q

  if %[2]s; then
%[3]s    echo "fatal: shallow file has changed since we read it" >&2
    exit 128
  fi
fi

exec %[4]q "$@"
`, counter, stillFailing, deepenBody, gitPath, record, brokenSubcommand, deepenNoop)

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

// installRecordingShim records git invocations without disturbing what the
// package has already memorized, which installGitShim deliberately resets.
func installRecordingShim(t *testing.T, recordTo string) {
	t.Helper()

	gitPath := lookupGit(t)
	shimDir := t.TempDir()

	body := fmt.Sprintf(`#!/bin/bash
echo "$*" >> %q
exec %q "$@"
`, recordTo, gitPath)

	// #nosec G306 -- the shim has to be executable
	if err := os.WriteFile(filepath.Join(shimDir, "git"), []byte(body), sentinelFileMode|0o111); err != nil {
		t.Fatalf("failed to write the recording shim: %v", err)
	}

	t.Setenv("PATH", shimDir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func readFile(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path) // #nosec G304 -- test-controlled path
	if err != nil {
		if os.IsNotExist(err) {
			return ""
		}

		t.Fatalf("failed to read %s: %v", path, err)
	}

	return string(content)
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
	unresolvableRanges = map[string]bool{}
}

// withDeepenBudget shrinks how many times the clone may be deepened, so that
// the end of the budget can be reached without building a repository with a
// hundred thousand commits in it.
func withDeepenBudget(t *testing.T, budget int) {
	t.Helper()

	original := MaxUnshallowIterations
	MaxUnshallowIterations = budget

	t.Cleanup(func() { MaxUnshallowIterations = original })
}

func containsLine(lines []string, wanted string) bool {
	for _, line := range lines {
		if strings.TrimSpace(line) == wanted {
			return true
		}
	}

	return false
}
