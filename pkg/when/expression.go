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
	YamlPath   string
}

type Inputs struct {
	Requirments *gabs.Container

	Keywords  map[string]string `json:"keywords"`
	Functions []FunctionInput   `json:"functions"`
}

func (w *WhenExpression) Eval() error {
	fmt.Println("")
	fmt.Printf("Processing when expression %s\n", w.Expression)
	fmt.Printf("  From: %v\n", w.Path)

	inputs, err := w.ListNeededInputs()
	if err != nil {
		return err
	}

	fmt.Printf("  Inputs needed: %v\n", inputs.Requirments.String())

	err = w.EvalFunctions(inputs)
	if err != nil {
		return err
	}

	return w.Reduce(inputs)
}

func (w *WhenExpression) EvalFunctions(inputs *Inputs) error {
	for _, input := range inputs.Requirments.Children() {
		if !IsChangeInFunction(input) {
			continue
		}

		fun, err := ParseChangeIn(w, input, w.YamlPath)
		if err != nil {
			return err
		}

		hasChanges, err := fun.Eval()
		if err != nil {
			return err
		}

		funInput := FunctionInput{
			Name:   "change_in",
			Params: input.Search("params"),
			Result: hasChanges,
		}

		inputs.Functions = append(inputs.Functions, funInput)
	}

	return nil
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
