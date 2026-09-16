package git

import (
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	consolelogger "github.com/semaphoreci/spc/pkg/consolelogger"
)

// Fetching branches from Git remotes has a non-trivial performance impact.
// In this structure we store already fetched branches.
// If the branch was already fetched, the Fetch action will be a noop.
//
// Results of fetch are only memorized if there are no errors while fetching.
var fetchedBranches map[string]string

// Running and listing diffs has a non-trivial performance impact.
// In this structure we store already evaluated git diff outputs.
// If the diff is already evaluated for a commitRange range, the Diff action
// will be noop.
//
// Diff results are only memorized if there are no errors.
var evaluatedDiffs map[string][]string

// Deciding that a commit range has no merge base costs a full round of
// deepen fetches, and the answer cannot change while the process runs. Only
// this verdict is memorized: a git failure stays retryable.
var unresolvableRanges map[string]bool

func init() {
	fetchedBranches = map[string]string{}
	evaluatedDiffs = map[string][]string{}
	unresolvableRanges = map[string]bool{}
}

func Fetch(name string) (string, error) {
	if output, ok := fetchedBranches[name]; ok {
		return output, nil
	}

	output, err := run("fetch", "origin", fmt.Sprintf("+refs/heads/%s:refs/heads/%s", name, name))
	if err != nil {
		consolelogger.Infof("Git failed with %s\n", err.Error())
		consolelogger.Info(output)

		return output, err
	}

	fetchedBranches[name] = output
	return output, err
}

func Diff(commitRange string) ([]string, string, error) {
	if difflines, ok := evaluatedDiffs[commitRange]; ok {
		return difflines, "", nil
	}

	output, err := run("diff", "--name-only", commitRange)
	if err != nil {
		consolelogger.Infof("Git failed with %s\n", err.Error())
		consolelogger.Info(output)

		return []string{}, output, err
	}

	difflines := strings.Split(strings.TrimSpace(output), "\n")

	evaluatedDiffs[commitRange] = difflines
	return difflines, "", nil
}

func DiffList(commitRange string) ([]string, error) {
	err := unshallow(commitRange)
	if err != nil {
		return []string{}, err
	}

	list, _, err := Diff(commitRange)
	if err != nil {
		return list, err
	}

	return list, nil
}

// MaxUnshallowIterations is how many times the clone may be deepened while
// looking for a merge base. It is a variable so that tests can shrink the
// budget instead of building a repository deep enough to exhaust it.
var MaxUnshallowIterations = 10

const InitialDeepenBy = 100

// Deepening the clone can fail transiently, most notably when a concurrent git
// process rewrites .git/shallow while the fetch is in flight. Such a fetch
// usually applies before git aborts, so a failed deepen is retried instead of
// abandoning the commit range. Consecutive failures are capped so that a
// genuinely unreachable remote fails fast instead of retrying ten times.
const MaxConsecutiveDeepenFailures = 3

// Retries wait before trying again, so that a competing git process holding
// the shallow file has time to finish. Waiting grows with each failure.
const DeepenRetryBackoff = 500 * time.Millisecond

// ErrRangeUnresolvable reports that the whole history is available and the
// two commits still have no common ancestor: the branches share no history,
// or the base branch was recreated. Deepening cannot fix that, so it is kept
// distinct from a genuine git failure, which callers treat far more severely.
//
// A clone that merely ran out of deepen budget is NOT this error. That answer
// is inconclusive, because the history that would settle it was never
// fetched, and it is reported as a failure instead.
var ErrRangeUnresolvable = errors.New("commit range is not resolvable")

// The exit code git uses to say two commits have no common ancestor, as
// opposed to the command itself failing.
const noCommonAncestorExitCode = 1

// What git diff prints when a three dot range has no merge base. It exits
// 128 for that as well as for a genuine failure, so the message is the only
// way to tell the two apart.
const noMergeBaseMessage = "no merge base"

func unshallow(commitRange string) error {
	if unresolvableRanges[commitRange] {
		return unresolvableError(commitRange)
	}

	err := deepenUntilResolvable(commitRange)
	if errors.Is(err, ErrRangeUnresolvable) {
		unresolvableRanges[commitRange] = true
	}

	return err
}

func deepenUntilResolvable(commitRange string) error {
	for deepens := 0; deepens < MaxUnshallowIterations; deepens++ {
		resolvable, err := canResolveCommitRange(commitRange)
		if err != nil {
			return err
		}

		if resolvable {
			return nil
		}

		if err := deepenWithRetries(InitialDeepenBy * int(math.Exp2(float64(deepens)))); err != nil {
			return fmt.Errorf("failed to deepen the git clone while resolving commit range %s: %w", commitRange, err)
		}
	}

	return verdictAfterLastDeepen(commitRange)
}

// deepenWithRetries retries a failed deepen, waiting a little longer each
// time, so that a competing git process holding the shallow file has a chance
// to finish. Retrying does not consume a depth level: a deepen that never
// succeeded did not make the clone any deeper.
func deepenWithRetries(deepenBy int) error {
	var err error

	for attempt := 1; attempt <= MaxConsecutiveDeepenFailures; attempt++ {
		if err = deepen(deepenBy); err == nil {
			return nil
		}

		if attempt < MaxConsecutiveDeepenFailures {
			waitBeforeRetry(attempt)
		}
	}

	return err
}

// verdictAfterLastDeepen checks the range once more once the budget is spent.
// Without this the final deepen is never taken into account, and a range it
// just made reachable is still reported as unresolved.
func verdictAfterLastDeepen(commitRange string) error {
	resolvable, err := canResolveCommitRange(commitRange)
	if err != nil {
		return err
	}

	if resolvable {
		return nil
	}

	// Out of budget with a clone that is still shallow. The history that would
	// prove whether a common ancestor exists was never fetched, so this is a
	// failure to find out, not a finding that there is none.
	if isShallow() {
		return fmt.Errorf(
			"commit range %s is still unresolved after %d deepen iterations and the clone is still shallow",
			commitRange,
			MaxUnshallowIterations,
		)
	}

	return unresolvableError(commitRange)
}

func unresolvableError(commitRange string) error {
	return fmt.Errorf("%w: %s", ErrRangeUnresolvable, commitRange)
}

func isShallow() bool {
	output, err := run("rev-parse", "--is-shallow-repository")
	if err != nil {
		return false
	}

	return strings.TrimSpace(output) == "true"
}

func waitBeforeRetry(consecutiveFailures int) {
	backoff := DeepenRetryBackoff * time.Duration(consecutiveFailures)

	consolelogger.Infof(
		"Deepening the clone failed, retrying in %s (%d/%d)\n",
		backoff,
		consecutiveFailures,
		MaxConsecutiveDeepenFailures,
	)

	time.Sleep(backoff)
}

func deepen(numberOfCommits int) error {
	output, err := run("fetch", "origin", "--deepen", strconv.Itoa(numberOfCommits))
	if err == nil {
		return nil
	}

	consolelogger.Infof("Git failed with %s\n", err.Error())
	consolelogger.Info(output)

	// Carry git's own diagnostic into the error. Without it every cause -
	// the shallow file race, a refused connection, a missing repository -
	// reaches the user as a bare "exit status 128".
	return fmt.Errorf("%s: %w", gitErrorDetail(output), err)
}

// gitErrorDetail picks the line of git output worth putting in an error,
// preferring the line git itself marked as the failure.
func gitErrorDetail(output string) string {
	lines := strings.Split(strings.TrimSpace(output), "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "fatal:") || strings.HasPrefix(trimmed, "error:") {
			return trimmed
		}
	}

	return strings.TrimSpace(lines[len(lines)-1])
}

func exitCode(err error) int {
	var exitErr *exec.ExitError

	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}

	return -1
}

func canResolveCommitRange(commitRange string) (bool, error) {
	if needsMergeBase(commitRange) {
		available, err := mergeBaseAvailable(commitRange)
		if err != nil || !available {
			return false, err
		}
	}

	return diffIsResolvable(commitRange)
}

func needsMergeBase(commitRange string) bool {
	return strings.Contains(commitRange, threeDots)
}

// mergeBaseAvailable answers whether the two ends of the range have a common
// ancestor in the history fetched so far. It deliberately does not fold a
// failing git command into a negative answer: "there is no merge base" and
// "git could not tell me" lead to very different outcomes for the caller.
func mergeBaseAvailable(commitRange string) (bool, error) {
	base, head := splitThreeDotRange(commitRange)

	output, err := run("merge-base", base, head)
	if err == nil {
		return strings.TrimSpace(output) != "", nil
	}

	consolelogger.Info(output)

	// Exit code 1 is how git reports that the commits have no common ancestor
	// in the history available right now.
	if exitCode(err) == noCommonAncestorExitCode {
		return false, nil
	}

	// Anything else means merge-base could not do its job. In a shallow clone
	// that is most often an object that simply has not been fetched yet -
	// "Not a valid commit name" for a base commit outside the current depth -
	// which is exactly what deepening fixes, so it is not fatal while there
	// is still history left to fetch. If the budget runs out with the clone
	// still shallow, the caller reports that as a failure of its own.
	if isShallow() {
		return false, nil
	}

	return false, fmt.Errorf("git merge-base failed: %s: %w", gitErrorDetail(output), err)
}

func splitThreeDotRange(commitRange string) (string, string) {
	baseSha := defaultBaseSha()

	parts := strings.Split(commitRange, threeDots)
	if len(parts) != 2 {
		return baseSha, baseSha
	}

	base := strings.TrimSpace(parts[0])
	head := strings.TrimSpace(parts[1])

	if base == "" {
		base = baseSha
	}

	if head == "" {
		head = baseSha
	}

	return base, head
}

// diffIsResolvable answers whether git can diff the range as it stands.
//
// Unlike merge-base, git diff exits 128 both for a range with no merge base
// and for a genuine failure, so the exit code cannot separate them and the
// message has to. Treating every failure here as fatal would break every
// repository with an orphan branch.
func diffIsResolvable(commitRange string) (bool, error) {
	output, err := run("diff", "--shortstat", commitRange)
	if err == nil {
		return true, nil
	}

	consolelogger.Info(output)

	if strings.Contains(output, noMergeBaseMessage) {
		return false, nil
	}

	// As with merge-base, a shallow clone may simply not have the objects yet.
	if isShallow() {
		return false, nil
	}

	return false, fmt.Errorf("git diff failed: %s: %w", gitErrorDetail(output), err)
}

func run(args ...string) (string, error) {
	consolelogger.Infof("Running git %s\n", strings.Join(args, " "))

	output, err := exec.Command("git", args...).CombinedOutput()
	return string(output), err
}

func defaultBaseSha() string {
	sha := strings.TrimSpace(os.Getenv("SEMAPHORE_GIT_SHA"))
	if sha != "" {
		return sha
	}

	return "HEAD"
}
