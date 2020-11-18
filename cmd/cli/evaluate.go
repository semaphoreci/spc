package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/ghodss/yaml"
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

		ppl.EvaluateChangeIns()

		jsonPpl, _ := ppl.MarshalJSON()
		yamlPpl, _ := yaml.JSONToYAML(jsonPpl)

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
