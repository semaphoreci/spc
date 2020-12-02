package cli

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/semaphoreci/spc/pkg/pipelines"
	"github.com/spf13/cobra"

	logs "github.com/semaphoreci/spc/pkg/logs"
	when "github.com/semaphoreci/spc/pkg/when"
)

var evaluateCmd = &cobra.Command{
	Use: "evaluate",
}

var evaluateChangeInCmd = &cobra.Command{
	Use: "change-in",

	Run: func(cmd *cobra.Command, args []string) {
		checkWhenInstalled()

		input := fetchRequiredStringFlag(cmd, "input")
		output := fetchRequiredStringFlag(cmd, "output")
		logsPath := fetchRequiredStringFlag(cmd, "logs")

		logs.Open(logsPath)
		logs.SetCurrentPipelineFilePath(input)

		ppl, err := pipelines.LoadFromYaml(input)
		check(err)

		err = ppl.EvaluateChangeIns(input)
		check(err)

		yamlPpl, err := ppl.ToYAML()
		check(err)

		err = ioutil.WriteFile(output, yamlPpl, 0644)
		check(err)
	},
}

// revive:disable:deep-exit

func checkWhenInstalled() {
	if !when.IsInstalled() {
		fmt.Println("Error: Con't find the 'when' expression parser binary")
		fmt.Println()
		fmt.Println("Is it installed and available in $PATH?")

		os.Exit(1)
	}
}

func check(err error) {
	if err == nil {
		return
	}

	fmt.Println(err)

	if _, ok := err.(*logs.ErrorChangeInMissingBranch); ok {
		os.Exit(1)
	}

	panic(err)
}

// revive:disable:deep-exit

func init() {
	rootCmd.AddCommand(evaluateCmd)

	evaluateChangeInCmd.Flags().String("input", "", "input pipeline YAML file")
	evaluateChangeInCmd.Flags().String("output", "", "output pipeline YAML file")
	evaluateChangeInCmd.Flags().String("logs", "", "path to the file where the compiler logs are streamed")

	evaluateCmd.AddCommand(evaluateChangeInCmd)
}
