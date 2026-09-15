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

func init() {
	fetchedBranches = map[string]string{}
	evaluatedDiffs = map[string][]string{}
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

const MaxUnshallowIterations = 10
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

// ErrRangeUnresolvable reports that no git command failed, but the commit
// range still has no merge base: the branches share no history, the base
// branch was recreated, or the merge base lies beyond the deepen budget.
// Deepening cannot fix any of those, so this is kept distinct from a genuine
// git failure, which callers treat far more severely.
var ErrRangeUnresolvable = errors.New("commit range is not resolvable")

func unshallow(commitRange string) error {
	consecutiveFailures := 0

	for i := 0; i < MaxUnshallowIterations; i++ {
		if canResolveCommitRange(commitRange) {
			return nil
		}

		err := deepen(InitialDeepenBy * int(math.Exp2(float64(i))))
		if err == nil {
			consecutiveFailures = 0
			continue
		}

		consecutiveFailures++
		if consecutiveFailures >= MaxConsecutiveDeepenFailures {
			return fmt.Errorf("failed to deepen the git clone while resolving commit range %s: %w", commitRange, err)
		}

		waitBeforeRetry(consecutiveFailures)
	}

	return fmt.Errorf("%w: %s", ErrRangeUnresolvable, commitRange)
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
	if err != nil {
		consolelogger.Infof("Git failed with %s\n", err.Error())
		consolelogger.Info(output)

		return err
	}

	return err
}

func canResolveCommitRange(commitRange string) bool {
	if needsMergeBase(commitRange) && !mergeBaseAvailable(commitRange) {
		return false
	}

	return diffIsResolvable(commitRange)
}

func needsMergeBase(commitRange string) bool {
	return strings.Contains(commitRange, threeDots)
}

func mergeBaseAvailable(commitRange string) bool {
	base, head := splitThreeDotRange(commitRange)

	output, err := run("merge-base", base, head)
	if err != nil {
		consolelogger.Info(output)
		return false
	}

	return strings.TrimSpace(output) != ""
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

func diffIsResolvable(commitRange string) bool {
	output, err := run("diff", "--shortstat", commitRange)
	if err != nil {
		consolelogger.Info(output)
	}

	return err == nil
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
