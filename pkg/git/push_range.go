package git

import (
	"fmt"
	"math"
	"regexp"
	"strings"

	consolelogger "github.com/semaphoreci/spc/pkg/consolelogger"
)

// PushRangeDiff lists what changed between the commit a branch pointed to
// before a push and the commit it points to now.
type PushRangeDiff struct {
	Files []string

	// BaseAvailable is false when the previous commit could not be obtained.
	// Files then lists every file in the new commit, as if the push had
	// created the branch.
	BaseAvailable bool
}

// A push range is resolved once per process, like any other diff.
var evaluatedPushRanges map[string]PushRangeDiff

// Only a full object id can be fetched directly. Revisions such as 'sha^'
// are left to deepening.
var fullObjectID = regexp.MustCompile(`^[0-9a-f]{40}([0-9a-f]{24})?$`)

func init() {
	evaluatedPushRanges = map[string]PushRangeDiff{}
}

// PushRangeDiffList compares the two ends of a push range directly, not
// through their merge base. After a force push the previous commit is often
// not an ancestor of the new one, and a merge base diff misses what the push
// removed.
func PushRangeDiffList(commitRange string) (PushRangeDiff, error) {
	if diff, ok := evaluatedPushRanges[commitRange]; ok {
		return diff, nil
	}

	base, head := splitPushRange(commitRange)

	missing, err := obtainCommits(base, head)
	if err != nil {
		return PushRangeDiff{}, err
	}

	if contains(missing, head) {
		return PushRangeDiff{}, fmt.Errorf("commit %s is not available in the clone", head)
	}

	diff := PushRangeDiff{BaseAvailable: !contains(missing, base)}

	if diff.BaseAvailable {
		diff.Files, err = gitLines("diff", "--name-only", base, head)
	} else {
		diff.Files, err = gitLines("ls-tree", "-r", "--name-only", "--full-tree", head)
	}

	if err != nil {
		return PushRangeDiff{}, err
	}

	evaluatedPushRanges[commitRange] = diff

	return diff, nil
}

// obtainCommits makes the given revisions available locally and returns the
// ones that still are not. A commit is first fetched by its id alone, which
// also works for a commit no branch points to any more. Deepening comes
// second, for remotes that refuse such requests.
func obtainCommits(revs ...string) ([]string, error) {
	for _, rev := range missingCommits(revs) {
		if fullObjectID.MatchString(rev) {
			fetchCommit(rev)
		}
	}

	if err := deepenUntilPresent(revs); err != nil {
		return nil, err
	}

	return missingCommits(revs), nil
}

// deepenUntilPresent deepens a shallow clone until every revision is present
// or there is no history left to fetch.
func deepenUntilPresent(revs []string) error {
	for deepens := 0; deepens < MaxUnshallowIterations; deepens++ {
		missing := missingCommits(revs)
		if len(missing) == 0 || !isShallow() {
			return nil
		}

		if err := deepenWithRetries(InitialDeepenBy * int(math.Exp2(float64(deepens)))); err != nil {
			return fmt.Errorf(
				"failed to deepen the git clone while looking for %s: %w",
				strings.Join(missing, ", "),
				err,
			)
		}
	}

	return nil
}

// fetchCommit fetches one commit without its history: comparing two commits
// only needs their trees. A failure is not fatal, the caller still has
// deepening and the missing-base fallback.
func fetchCommit(id string) {
	output, err := run("fetch", "--no-tags", "--depth=1", "origin", id)
	if err != nil {
		consolelogger.Infof("Git failed with %s\n", err.Error())
		consolelogger.Info(output)
	}
}

func missingCommits(revs []string) []string {
	missing := []string{}

	for _, rev := range revs {
		if _, err := run("rev-parse", "--verify", "--quiet", rev+"^{commit}"); err != nil {
			missing = append(missing, rev)
		}
	}

	return missing
}

func gitLines(args ...string) ([]string, error) {
	output, err := run(args...)
	if err != nil {
		consolelogger.Info(output)

		return nil, fmt.Errorf("git %s failed: %s: %w", args[0], gitErrorDetail(output), err)
	}

	lines := []string{}

	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		if line != "" {
			lines = append(lines, line)
		}
	}

	return lines, nil
}

func splitPushRange(commitRange string) (string, string) {
	parts := splitCommitRange(commitRange)
	if len(parts) != 2 {
		sha := defaultBaseSha()

		return sha, sha
	}

	base := strings.TrimSpace(parts[0])
	head := strings.TrimSpace(parts[1])

	if base == "" {
		base = defaultBaseSha()
	}

	if head == "" {
		head = defaultBaseSha()
	}

	return base, head
}

func contains(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}

	return false
}
