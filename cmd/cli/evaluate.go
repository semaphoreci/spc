package main

import (
	"fmt"
	"os"

	"github.com/semaphoreci/spc/pkg/pipelines"
	"github.com/spf13/cobra"
)

var evaluateCmd = &cobra.Command{
	Use: "evaluate",
}

var evaluateChangeInCmd = &cobra.Command{
	Use: "change-in",

	Run: func(cmd *cobra.Command, args []string) {
		input := fetchRequiredStringFlag(cmd, "input")
		output := fetchRequiredStringFlag(cmd, "output")
		logs := fetchRequiredStringFlag(cmd, "logs")

		ppl, err := pipelines.LoadFromYaml(input)
		if err != nil {
			fmt.Printf("Writing failure to %s %s", logs, err.Error())

			os.Exit(1)
		}

		fmt.Printf("Pipeline %+v", ppl)

		fmt.Printf("Writing result to %s", output)
	},
}

func init() {
	rootCmd.AddCommand(evaluateCmd)

	evaluateChangeInCmd.Flags().String("input", "", "input pipeline YAML file")
	evaluateChangeInCmd.Flags().String("output", "", "output pipeline YAML file")
	evaluateChangeInCmd.Flags().String("logs", "", "path to the file where the compiler logs are streamed")

	evaluateCmd.AddCommand(evaluateChangeInCmd)
}
