package when

import (
	gabs "github.com/Jeffail/gabs/v2"
	consolelogger "github.com/semaphoreci/spc/pkg/consolelogger"
	changein "github.com/semaphoreci/spc/pkg/when/changein"
	env "github.com/semaphoreci/spc/pkg/when/env"
	whencli "github.com/semaphoreci/spc/pkg/when/whencli"
)

type WhenExpression struct {
	Expression   string
	Path         []string
	YamlPath     string
	Requirments  *gabs.Container
	ReduceInputs whencli.ReduceInputs
}

func (w *WhenExpression) Eval() error {
	for _, requirment := range w.ListChangeInFunctions(w.Requirments) {
		result, err := w.EvaluateChangeInFunction(requirment)
		if err != nil {
			return err
		}

		input := map[string]interface{}{}
		input["name"] = w.functionName(requirment)
		input["params"] = w.functionParams(requirment)
		input["result"] = result

		w.ReduceInputs.Keywords = map[string]interface{}{}
		w.ReduceInputs.Functions = append(w.ReduceInputs.Functions, input)
	}

	for _, requirment := range w.ListEnvFunctions(w.Requirments) {
		result, err := w.EvaluateEnvFunction(requirment)
		if err != nil {
			return err
		}

		input := map[string]interface{}{}
		input["name"] = w.functionName(requirment)
		input["params"] = w.functionParams(requirment)
		input["result"] = result

		w.ReduceInputs.Keywords = map[string]interface{}{}
		w.ReduceInputs.Functions = append(w.ReduceInputs.Functions, input)
	}

	return nil
}

func (w *WhenExpression) ListChangeInFunctions(requirments *gabs.Container) []*gabs.Container {
	result := []*gabs.Container{}

	for _, input := range requirments.Children() {
		if w.IsChangeInFunction(input) {
			result = append(result, input)
		}
	}

	return result
}

func (w *WhenExpression) ListEnvFunctions(requirments *gabs.Container) []*gabs.Container {
	result := []*gabs.Container{}

	for _, input := range requirments.Children() {
		if w.IsEnvFunction(input) {
			result = append(result, input)
		}
	}

	return result
}

func (w *WhenExpression) IsChangeInFunction(input *gabs.Container) bool {
	elType := input.Search("type").Data().(string)
	if elType != "fun" {
		return false
	}

	if w.functionName(input) != "change_in" {
		return false
	}

	return true
}

func (w *WhenExpression) IsEnvFunction(input *gabs.Container) bool {
	elType := input.Search("type").Data().(string)
	if elType != "fun" {
		return false
	}

	if w.functionName(input) != "env" {
		return false
	}

	return true
}

func (w *WhenExpression) EvaluateChangeInFunction(input *gabs.Container) (bool, error) {
	consolelogger.EmptyLine()

	consolelogger.Infof("%s(%+v)\n", w.functionName(input), w.functionParams(input))

	fun, err := changein.Parse(w.Path, input, w.YamlPath)
	if err != nil {
		return false, err
	}

	return changein.Eval(fun)
}

func (w *WhenExpression) EvaluateEnvFunction(input *gabs.Container) (string, error) {
	consolelogger.EmptyLine()

	consolelogger.Infof("%s(%+v)\n", w.functionName(input), w.functionParams(input))

	fun, err := env.Parse(input)
	if err != nil {
		return "", err
	}

	return env.Eval(fun)
}

func (w *WhenExpression) functionName(input *gabs.Container) string {
	return input.Search("name").Data().(string)
}

func (w *WhenExpression) functionParams(input *gabs.Container) *gabs.Container {
	return input.Search("params")
}
