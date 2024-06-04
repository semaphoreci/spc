package cli

import (
	"fmt"
	"io/ioutil"

	"github.com/semaphoreci/spc/pkg/pipelines"
	"github.com/spf13/cobra"

	logs "github.com/semaphoreci/spc/pkg/logs"
)

var compileCmd = &cobra.Command{
	Use: "compile",

	Run: func(cmd *cobra.Command, _ []string /*args*/) {
		checkWhenInstalled()

		input := fetchRequiredStringFlag(cmd, "input")
		output := fetchRequiredStringFlag(cmd, "output")
		logsPath := fetchRequiredStringFlag(cmd, "logs")

		fmt.Printf("Evaluating template expressions in %s.\n\n", input)

		logs.Open(logsPath)
		logs.SetCurrentPipelineFilePath(input)

		ppl, err := pipelines.LoadFromFile(input)
		check(err)

		err = ppl.EvaluateTemplates()
		check(err)

		fmt.Printf("Evaluating change_in expressions in %s.\n\n", input)

		err = ppl.EvaluateChangeIns()
		check(err)

		yamlPpl, err := ppl.ToYAML()
		check(err)

		// #nosec
		err = ioutil.WriteFile(output, yamlPpl, 0644)
		check(err)
	},
}

func init() {
	compileCmd.Flags().String("input", "", "input pipeline YAML file")
	compileCmd.Flags().String("output", "", "output pipeline YAML file")
	compileCmd.Flags().String("logs", "", "path to the file where the compiler logs are streamed")

	rootCmd.AddCommand(compileCmd)
}
