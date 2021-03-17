package when

import (
	gabs "github.com/Jeffail/gabs/v2"
	consolelogger "github.com/semaphoreci/spc/pkg/consolelogger"
	changein "github.com/semaphoreci/spc/pkg/when/changein"
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
		result, err := w.EvalFunction(requirment)
		if err != nil {
			return err
		}

		input := map[string]interface{}{}
		input["name"] = requirment.Search("name")
		input["params"] = requirment.Search("params")
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

func (w *WhenExpression) IsChangeInFunction(input *gabs.Container) bool {
	elType := input.Search("type").Data().(string)
	if elType != "fun" {
		return false
	}

	elName := input.Search("name").Data().(string)
	if elName != "change_in" {
		return false
	}

	return true
}

func (w *WhenExpression) EvalFunction(input *gabs.Container) (bool, error) {
	consolelogger.EmptyLine()

	consolelogger.Infof("%s(%+v)\n", input.Search("name").Data(), input.Search("params"))

	fun, err := changein.Parse(w.Path, input, w.YamlPath)
	if err != nil {
		return false, err
	}

	return changein.Eval(fun)
}
