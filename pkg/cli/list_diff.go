package cli

import (
	"fmt"

	git "github.com/semaphoreci/spc/pkg/git"
	logs "github.com/semaphoreci/spc/pkg/logs"
	"github.com/spf13/cobra"
)

var listDiffCmd = &cobra.Command{
	Use: "list-diff",

	Run: func(cmd *cobra.Command, args []string) {
		defaultBranch := fetchOptionalStringFlag(cmd, "default-branch")
		defaultRange := fetchOptionalStringFlag(cmd, "default-range")
		branchRange := fetchOptionalStringFlag(cmd, "branch-range")
		onTags := fetchOptionalBoolFlag(cmd, "on-tags")

		logsPath := fetchRequiredStringFlag(cmd, "logs")
		logs.Open(logsPath)

		fmt.Printf("Listing diff for spc compiler...\n")

		gitSettings := git.NewGitSettings(defaultBranch, defaultRange, branchRange, onTags)
		diffList, err := git.DiffList(gitSettings.CommitRange())
		check(err)

		for _, file := range diffList {
			fmt.Println(file)
		}
	},
}

// revive:disable:deep-exit

func init() {

	listDiffCmd.Flags().String("default-branch", "", "default branch of repository")
	listDiffCmd.Flags().String("default-range", "", "default range for evaluation")
	listDiffCmd.Flags().String("branch-range", "", "branch range for evaluation")
	listDiffCmd.Flags().Bool("on-tags", true, "if commands is running on tags")

	listDiffCmd.Flags().String("logs", "", "path to the file where the compiler logs are streamed")

	rootCmd.AddCommand(listDiffCmd)
}
