package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/semaphoreci/spc/pkg/consolelogger"
	git "github.com/semaphoreci/spc/pkg/git"
	logs "github.com/semaphoreci/spc/pkg/logs"
	"github.com/spf13/cobra"
)

var listDiffCmd = &cobra.Command{
	Use: "list-diff",

	Run: func(cmd *cobra.Command, _ []string /*args*/) {
		consolelogger.Enabled = false

		defaultBranch := fetchOptionalStringFlag(cmd, "default-branch")
		defaultRange := fetchOptionalStringFlag(cmd, "default-range")
		branchRange := fetchOptionalStringFlag(cmd, "branch-range")

		gitDiffSet := git.NewDiffSet(defaultBranch, defaultRange, branchRange, true)

		if gitDiffSet.IsEvaluationNeeded() {
			println("Listing diffs for tags is not supported.")
			return
		}

		fetchNeeded, fetchTargets := gitDiffSet.IsGitFetchNeeded()

		if fetchNeeded {
			for _, fetchTarget := range fetchTargets {
				output, err := git.Fetch(fetchTarget)
				err = parseFetchError(fetchTarget, output, err)
				check(err)
			}
		}

		commitRange := gitDiffSet.CommitRange()

		diffList, err := git.DiffList(commitRange)
		check(parseDiffError(commitRange, err))

		for _, file := range diffList {
			fmt.Println(file)
		}
	},
}

// parseDiffError mirrors how change_in treats the two failure modes. A commit
// range with no merge base keeps listing nothing, as it always has, but says
// so on stderr rather than leaving the caller with silent empty output. A git
// command failure is reported and exits non-zero.
func parseDiffError(commitRange string, err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, git.ErrRangeUnresolvable) {
		fmt.Fprintf(
			os.Stderr,
			"warning: no merge base found for commit range '%s', listing no files\n",
			commitRange,
		)

		return nil
	}

	msg := fmt.Sprintf(
		"Failed to resolve the git diff for commit range '%s': %s",
		commitRange,
		err.Error(),
	)

	return &logs.ErrorInitializationFailed{Message: msg}
}

func parseFetchError(fetchTarget string, output string, err error) error {
	if strings.Contains(string(output), "couldn't find remote ref") {
		msg := fmt.Sprintf("Unknown git reference '%s'.", fetchTarget)
		err := logs.ErrorChangeInMissingBranch{Message: msg}
		return &err
	}

	return err
}

// revive:disable:deep-exit

func init() {
	listDiffCmd.Flags().String("default-branch", "", "default branch of repository")
	listDiffCmd.Flags().String("default-range", "", "default range for evaluation")
	listDiffCmd.Flags().String("branch-range", "", "branch range for evaluation")

	rootCmd.AddCommand(listDiffCmd)
}
