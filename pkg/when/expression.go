package when

import (
	"encoding/json"
	"fmt"

	gabs "github.com/Jeffail/gabs/v2"
	changein "github.com/semaphoreci/spc/pkg/when/changein"
	whencli "github.com/semaphoreci/spc/pkg/when/whencli"
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

	for _, input := range w.ListChangeInFunctions(inputs) {
		result, err = w.EvalFunctions(input)
		if err != nil {
			return err
		}

		funInput := FunctionInput{
			Name:   "change_in",
			Params: input.Search("params"),
			Result: result,
		}

		inputs.Functions = append(inputs.Functions, funInput)
	}

	return w.Reduce(inputs)
}

func (w *WhenExpression) ListChangeInFunctions(inputs *Inputs) []*gabs.Container {
	result := []*gabs.Container{}

	for _, input := range inputs.Requirments.Children() {
		if IsChangeInFunction(input) {
			result = append(result, input)
		}

	}

	return result
}

func (w *WhenExpression) EvalFunction(input *Inputs) error {
	fun, err := changein.Parse(w.Path, input, w.YamlPath)
	if err != nil {
		return nil, err
	}

	return changein.Eval(fun)
}

func (w *WhenExpression) ListNeededInputs() (*Inputs, error) {
	neededInputs, err := whencli.ListInputs(w.Expression)
	if err != nil {
		return nil, err
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
		return err
	}

	result, err := whencli.Reduce(w.Expression, inputBytes)
	if err != nil {
		return err
	}

	w.Expression = result

	return nil
}
