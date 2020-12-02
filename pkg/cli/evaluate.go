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
		input := fetchRequiredStringFlag(cmd, "input")
		output := fetchRequiredStringFlag(cmd, "output")
		logsPath := fetchRequiredStringFlag(cmd, "logs")

		if !when.IsInstalled() {
			fmt.Println("Error: Con't find the 'when' expression parser binary")
			fmt.Println()
			fmt.Println("Is it installed and available in $PATH?")

			os.Exit(1)
		}

		logs.Open(logsPath)
		logs.SetCurrentPipelineFilePath(input)

		ppl, err := pipelines.LoadFromYaml(input)
		if err != nil {
			panic(err)
		}

		err = ppl.EvaluateChangeIns(input)
		if err != nil {
			os.Exit(1)
		}

		yamlPpl, err := ppl.ToYAML()
		if err != nil {
			panic(err)
		}

		err = ioutil.WriteFile(output, yamlPpl, 0644)
		if err != nil {
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(evaluateCmd)

	evaluateChangeInCmd.Flags().String("input", "", "input pipeline YAML file")
	evaluateChangeInCmd.Flags().String("output", "", "output pipeline YAML file")
	evaluateChangeInCmd.Flags().String("logs", "", "path to the file where the compiler logs are streamed")

	evaluateCmd.AddCommand(evaluateChangeInCmd)
}
