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
	Expression string
	Inputs     ReduceInputs
}

func Reduce(expressions []string, inputs []ReduceInputs) ([]string, error) {
	inputPath := "/tmp/reduce-inputs"
	outputPath := "/tmp/reduce-output"

	err := ReducePrepareInput(expressions, inputs, inputPath)
	if err != nil {
		return []string{}, err
	}

	bytes, err := exec.Command("when", "reduce", "--input", inputPath, "output", outputPath).CombinedOutput()
	if err != nil {
		return []string{}, fmt.Errorf("Failed to reduce when expressions %s. Output: %s.", err, bytes)
	}

	exprs, err := ReduceLoadOutput(outputPath)
	if err != nil {
		return []string{}, fmt.Errorf("Failed to reduce when expressions %s. Output: %s.", err, bytes)
	}

	return exprs, nil
}

func ReducePrepareInput(expressions []string, inputs []ReduceInputs, path string) error {
	content := []ReduceElement{}

	for index := range expressions {
		content = append(content, ReduceElement{
			Expression: expressions[index],
			Inputs:     inputs[index],
		})
	}

	j, err := json.Marshal(content)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path, []byte(j), 0644)
	if err != nil {
		return err
	}

	return nil
}

func ReduceLoadOutput(path string) ([]string, error) {
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

	// return inputs.Children(), nil
	return []string{}, nil
}
