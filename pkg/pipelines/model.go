package pipelines

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"

	gabs "github.com/Jeffail/gabs/v2"
	logs "github.com/semaphoreci/spc/pkg/logs"
	when "github.com/semaphoreci/spc/pkg/when"
)

type Pipeline struct {
	gabs.Container
}

func (p *Pipeline) EvaluateChangeIns(yamlPath string) error {
	fmt.Println("Evaluating start.")

	whenList := when.List(p)

	for _, w := range whenList {
		fmt.Printf("Processing when expression %s\n", w.Expression)
		fmt.Printf("  From: %v\n", w.Path)

		bytes, err := exec.Command("when", "list-inputs", w.Expression).Output()
		if err != nil {
			fmt.Println(string(bytes))
			panic(err)
		}

		fmt.Printf("  Inputs needed: %s\n", string(bytes))

		neededInputs, err := gabs.ParseJSON(bytes)
		if err != nil {
			panic(err)
		}

		m := map[string]string{}
		inputs := when.Inputs{Keywords: m}

		for _, input := range neededInputs.Children() {
			if !IsChangeInFunction(input) {
				continue
			}

			fun, err := NewChangeInFunctionFromWhenInputList(input, yamlPath)
			if err != nil {
				panic(err)
			}

			fmt.Println("  Checking if branch exists.")
			if !fun.DefaultBranchExists() {
				logs.Log(logs.ErrorChangeInMissingBranch{
					Message: "Unknown git reference 'random'.",
					Location: logs.Location{
						Path: w.Path,
					},
				})

				return fmt.Errorf("  Branch '%s' does not exists.", fun.Params.DefaultBranch)
			}

			hasChanges := fun.Eval()

			funInput := when.WhenFunctionInput{
				Name:   "change_in",
				Params: input.Search("params"),
				Result: hasChanges,
			}

			inputs.Functions = append(inputs.Functions, funInput)
		}

		inputBytes, err := json.Marshal(inputs)
		if err != nil {
			panic(err)
		}
		fmt.Printf("  Providing inputs: %s\n", string(inputBytes))

		err = ioutil.WriteFile("/tmp/inputs.json", inputBytes, 0644)
		if err != nil {
			panic(err)
		}

		bytes, err = exec.Command("when", "reduce", w.Expression, "--input", "/tmp/inputs.json").Output()
		if err != nil {
			panic(err)
		}

		fmt.Printf("  Reduced When Expression: %s\n", string(bytes))

		expr := strings.TrimSpace(string(bytes))

		p.Set(expr, w.Path...)
	}

	fmt.Println("Evaluating end.")
	return nil
}
