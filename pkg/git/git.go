package git

import (
	"fmt"
	"os/exec"
	"strings"
)

//
// Fetching branches from Git remotes has a non-trivial performance impact.
// In this structure we store already fetched branches.
// If the branch was already fetched, the Fetch action will be a noop.
//
// Results of fetch are only memorized if there are no errors while fetching.
//
var fetchedBranches map[string][]byte

//
// Running and listing diffs has a non-trivial performance impact.
// In this structure we store already evaluated git diff outputs.
// If the diff is already evaluated for a commitRange range, the Diff action
// will be noop.
//
// Diff results are only memorized if there are no errors.
//
var evaluatedDiffs map[string][]string

func init() {
	fetchedBranches = map[string][]byte{}
	evaluatedDiffs = map[string][]string{}
}

func Fetch(name string) ([]byte, error) {
	if output, ok := fetchedBranches[name]; ok {
		return output, nil
	}

	flags := []string{"fetch", "origin", fmt.Sprintf("+refs/heads/%s:refs/heads/%s", name, name)}
	fmt.Printf("Running git %s\n", strings.Join(flags, " "))

	output, err := exec.Command("git", flags...).CombinedOutput()
	if err != nil {
		return output, err
	}

	fetchedBranches[name] = output
	return output, err
}

func Diff(commitRange string) ([]string, string, error) {
	if difflines, ok := evaluatedDiffs[commitRange]; ok {
		return difflines, "", nil
	}

	flags := []string{"diff", "--name-only", commitRange}
	fmt.Printf("Running git %s\n", strings.Join(flags, " "))

	bytes, err := exec.Command("git", flags...).CombinedOutput()
	if err != nil {
		return []string{}, string(bytes), err
	}

	difflines := strings.Split(strings.TrimSpace(string(bytes)), "\n")

	evaluatedDiffs[commitRange] = difflines
	return difflines, "", nil
}
