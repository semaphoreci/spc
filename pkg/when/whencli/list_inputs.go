package whencli

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	gabs "github.com/Jeffail/gabs/v2"
)

type ListInputsResult struct {
	Expression string
	Inputs     *gabs.Container
	Error      string
}

func ListInputs(expressions []string) ([]ListInputsResult, error) {
	var err error
	res := []ListInputsResult{}

	inputPath := "/tmp/when-expressions"
	outputPath := "/tmp/parsed-when-expressions"

	err = ListInputsPrepareInputFile(inputPath, expressions)
	if err != nil {
		return res, nil
	}

	output, err := exec.Command("when", "list-inputs", "--input", inputPath, "--output", outputPath).CombinedOutput()
	if err != nil {
		return res, fmt.Errorf("unprecessable when expressions %s", string(output))
	}

	results, err := ListInputsLoadResults(outputPath)
	if err != nil {
		return res, fmt.Errorf("unprocessable when expressions %s, when CLI output: %s", err.Error(), output)
	}

	return prepareResults(expressions, results)
}

func prepareResults(expressions []string, results *gabs.Container) ([]ListInputsResult, error) {
	result := []ListInputsResult{}

	for index, el := range results.Children() {
		result = append(result, ListInputsResult{
			Expression: expressions[index],
			Inputs:     el.Search("inputs"),
			Error:      el.Search("error").Data().(string),
		})
	}

	return result, nil
}

func ListInputsPrepareInputFile(path string, expressions []string) error {
	j, err := json.Marshal(expressions)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path, []byte(j), os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

func ListInputsLoadResults(path string) (*gabs.Container, error) {
	// #nosec
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	content, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	inputs, err := gabs.ParseJSON(content)
	if err != nil {
		return nil, err
	}

	return inputs, nil
}
