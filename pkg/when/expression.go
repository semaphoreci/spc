package when

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	gabs "github.com/Jeffail/gabs/v2"
)

type WhenExpression struct {
	Expression string
	Path       []string
}

type Inputs struct {
	Requirments *gabs.Container

	Keywords  map[string]string `json:"keywords"`
	Functions []FunctionInput   `json:"functions"`
}

func (w *WhenExpression) ListNeededInputs() (*Inputs, error) {
	bytes, err := exec.Command("when", "list-inputs", w.Expression).Output()
	if err != nil {
		return nil, fmt.Errorf("Unprecessable when expression %s", string(bytes))
	}

	neededInputs, err := gabs.ParseJSON(bytes)
	if err != nil {
		return nil, fmt.Errorf("Unprocessable input list for when expressions %s", err.Error())
	}

	keywords := map[string]string{}
	functions := []FunctionInput{}

	return &Inputs{
		Requirments: neededInputs,
		Keywords:    keywords,
		Functions:   functions,
	}, nil
}

func (w *WhenExpression) Reduce(inputs *Inputs) error {
	inputBytes, err := json.Marshal(inputs)
	if err != nil {
		panic(err)
	}
	fmt.Printf("  Providing inputs: %s\n", string(inputBytes))

	err = ioutil.WriteFile("/tmp/inputs.json", inputBytes, os.ModePerm)
	if err != nil {
		panic(err)
	}

	bytes, err := exec.Command("when", "reduce", w.Expression, "--input", "/tmp/inputs.json").Output()
	if err != nil {
		return err
	}

	w.Expression = strings.TrimSpace(string(bytes))

	return nil
}
