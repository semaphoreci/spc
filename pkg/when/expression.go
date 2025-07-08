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
	Requirements  *gabs.Container
	ReduceInputs whencli.ReduceInputs
}

func (w *WhenExpression) Eval() error {
	for _, requirement := range w.ListChangeInFunctions(w.Requirements) {
		result, err := w.EvalFunction(requirement)
		if err != nil {
			return err
		}

		input := map[string]interface{}{}
		input["name"] = w.functionName(requirement)
		input["params"] = w.functionParams(requirement)
		input["result"] = result

		w.ReduceInputs.Keywords = map[string]interface{}{}
		w.ReduceInputs.Functions = append(w.ReduceInputs.Functions, input)
	}

	return nil
}

func (w *WhenExpression) ListChangeInFunctions(requirements *gabs.Container) []*gabs.Container {
	result := []*gabs.Container{}

	for _, input := range requirements.Children() {
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

	if w.functionName(input) != "change_in" {
		return false
	}

	return true
}

func (w *WhenExpression) EvalFunction(input *gabs.Container) (bool, error) {
	consolelogger.EmptyLine()

	consolelogger.Infof("%s(%+v)\n", w.functionName(input), w.functionParams(input))

	fun, err := changein.Parse(w.Path, input, w.YamlPath)
	if err != nil {
		return false, err
	}

	return fun.Eval()
}

func (w *WhenExpression) functionName(input *gabs.Container) string {
	return input.Search("name").Data().(string)
}

func (w *WhenExpression) functionParams(input *gabs.Container) *gabs.Container {
	return input.Search("params")
}
