package cli

import (
	"fmt"
	"strings"

	"github.com/semaphoreci/spc/pkg/consolelogger"
	git "github.com/semaphoreci/spc/pkg/git"
	logs "github.com/semaphoreci/spc/pkg/logs"
	"github.com/spf13/cobra"
)

var listDiffCmd = &cobra.Command{
	Use: "list-diff",

	Run: func(cmd *cobra.Command, args []string) {
		consolelogger.Enabled = false

		defaultBranch := fetchOptionalStringFlag(cmd, "default-branch")
		defaultRange := fetchOptionalStringFlag(cmd, "default-range")
		branchRange := fetchOptionalStringFlag(cmd, "branch-range")

		gitDiffSet := git.NewDiffSet(defaultBranch, defaultRange, branchRange, true)

		if gitDiffSet.IsEvaluationNeeded() {
			println("Listing diffs for tags is not supported.")
			return
		}

		fetchNeeded, fetchTarget := gitDiffSet.IsGitFetchNeeded()
		if fetchNeeded {
			output, err := git.Fetch(fetchTarget)
			err = parseFetchError(fetchTarget, output, err)
			check(err)
		}

		diffList, err := git.DiffList(gitDiffSet.CommitRange())
		check(err)

		for _, file := range diffList {
			fmt.Println(file)
		}
	},
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
