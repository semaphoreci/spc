package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var evaluateCmd = &cobra.Command{
	Use: "evaluate",
}

var evaluateChangeInCmd = &cobra.Command{
	Use: "change-in",

	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("evaluating change-ins")
	},
}

func init() {
	rootCmd.AddCommand(evaluateCmd)
	evaluateCmd.AddCommand(evaluateChangeInCmd)
}
