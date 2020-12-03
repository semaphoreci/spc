package whencli

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

type ReduceInputs struct {
	Keywords  map[string]interface{} `json:"keywords"`
	Functions []interface{}          `json:"functions"`
}

func Reduce(expression string, inputs ReduceInputs) (string, error) {
	path := "/tmp/input.json"

	inputs.Keywords = map[string]interface{}{}

	fmt.Printf("Inputs: \n")
	for _, f := range inputs.Functions {
		j, _ := json.Marshal(f)
		fmt.Printf("  - %s\n", j)
	}

	inputBytes, err := json.Marshal(inputs)
	if err != nil {
		return "", err
	}

	err = ioutil.WriteFile(path, inputBytes, os.ModePerm)
	if err != nil {
		return "", err
	}

	bytes, err := exec.Command("when", "reduce", expression, "--input", path).Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(bytes)), nil
}
