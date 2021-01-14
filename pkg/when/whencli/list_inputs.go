package whencli

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	gabs "github.com/Jeffail/gabs/v2"
)

func ListInputs(expressions []string) ([]*gabs.Container, error) {
	var err error

	inputPath := "/tmp/when-expressions"
	outputPath := "/tmp/parsed-when-expressions"

	err = ListInputsPrepareInputFile(inputPath, expressions)
	if err != nil {
		return []*gabs.Container{}, nil
	}

	output, err := exec.Command("when", "list-inputs", "--input", inputPath, "--output", outputPath).CombinedOutput()
	if err != nil {
		return []*gabs.Container{}, fmt.Errorf("unprecessable when expressions %s", string(output))
	}

	result, err := ListInputsLoadResults(outputPath)
	if err != nil {
		return result, fmt.Errorf("unprocessable when expressions %s, when CLI output: %s", err.Error(), output)
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

func ListInputsLoadResults(path string) ([]*gabs.Container, error) {
	file, err := os.Open(path)
	if err != nil {
		return []*gabs.Container{}, err
	}
	defer file.Close()

	content, err := ioutil.ReadAll(file)
	if err != nil {
		return []*gabs.Container{}, err
	}

	inputs, err := gabs.ParseJSON(content)
	if err != nil {
		return []*gabs.Container{}, err
	}

	return inputs.Children(), nil
}
