package whencli

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	gabs "github.com/Jeffail/gabs/v2"
)

type ReduceInputs struct {
	Keywords  map[string]interface{} `json:"keywords"`
	Functions []interface{}          `json:"functions"`
}

type ReduceElement struct {
	Expression string       `json:"expression"`
	Inputs     ReduceInputs `json:"inputs"`
}

func Reduce(expressions []string, inputs []ReduceInputs) ([]string, error) {
	inputPath := "/tmp/reduce-inputs"
	outputPath := "/tmp/reduce-output"

	err := ReducePrepareInput(expressions, inputs, inputPath)
	if err != nil {
		return []string{}, err
	}

	bytes, err := exec.Command("when", "reduce", "--input", inputPath, "--output", outputPath).CombinedOutput()
	if err != nil {
		return []string{}, fmt.Errorf("failed to reduce when expressions %s, output: %s", err, bytes)
	}

	exprs, err := ReduceLoadOutput(outputPath)
	if err != nil {
		return []string{}, fmt.Errorf("failed to reduce when expressions %s, Output: %s", err, bytes)
	}

	return exprs, nil
}

func ReducePrepareInput(expressions []string, inputs []ReduceInputs, path string) error {
	content := []ReduceElement{}

	for index := range expressions {
		if inputs[index].Functions == nil {
			inputs[index].Functions = []interface{}{}
		}

		if inputs[index].Keywords == nil {
			inputs[index].Keywords = map[string]interface{}{}
		}

		content = append(content, ReduceElement{
			Expression: expressions[index],
			Inputs:     inputs[index],
		})
	}

	j, err := json.Marshal(content)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path, []byte(j), os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

func ReduceLoadOutput(path string) ([]string, error) {
	// #nosec
	file, err := os.Open(path)
	if err != nil {
		return []string{}, err
	}
	defer file.Close()

	content, err := ioutil.ReadAll(file)
	if err != nil {
		return []string{}, err
	}

	inputs, err := gabs.ParseJSON(content)
	if err != nil {
		return []string{}, err
	}

	exprs := []string{}

	for index := range inputs.Children() {
		el := inputs.Children()[index]
		result := el.Search("result").Data().(string)
		errString := el.Search("error").Data().(string)

		if errString != "" {
			return []string{}, fmt.Errorf("unprocessable when expression %s", errString)
		}

		exprs = append(exprs, result)
	}

	return exprs, nil
}
