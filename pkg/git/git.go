package git

import (
	"fmt"
	"math"
	"os/exec"
	"strconv"
	"strings"

	consolelogger "github.com/semaphoreci/spc/pkg/consolelogger"
)

//
// Fetching branches from Git remotes has a non-trivial performance impact.
// In this structure we store already fetched branches.
// If the branch was already fetched, the Fetch action will be a noop.
//
// Results of fetch are only memorized if there are no errors while fetching.
//
var fetchedBranches map[string]string

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
		return []string{}, nil
	}

	list, _, err := Diff(commitRange)
	if err != nil {
		return list, err
	}

	return list, nil
}

const MaxUnshallowIterations = 10
const InitialDeepenBy = 100

func unshallow(commitRange string) error {
	for i := 0; i < MaxUnshallowIterations; i++ {
		if canResolveCommitRange(commitRange) {
			return nil
		}

		deepenBy := InitialDeepenBy * int(math.Exp2(float64(i)))

		err := deepen(deepenBy)
		if err != nil {
			return err
		}
	}

	return fmt.Errorf("commit range %s is not resolvable", commitRange)
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
