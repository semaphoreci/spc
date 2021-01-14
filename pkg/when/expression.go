package when

import (
	"fmt"
	"time"

	gabs "github.com/Jeffail/gabs/v2"
	changein "github.com/semaphoreci/spc/pkg/when/changein"
	whencli "github.com/semaphoreci/spc/pkg/when/whencli"
)

type WhenExpression struct {
	Expression  string
	Path        []string
	YamlPath    string
	Requirments *gabs.Container
}

var TotalList int64
var TotalEval int64
var TotalReduce int64

func n() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func (w *WhenExpression) Eval() error {
	fmt.Println("")
	fmt.Println("*** Processing when expression ***")
	fmt.Printf("Expression: %v\n", w.Expression)
	fmt.Printf("From: %v\n", w.Path)

	fmt.Printf("Needs:\n")
	for _, need := range w.Requirments.Children() {
		fmt.Printf("  - %v\n", need)
	}

	start2 := n()
	reduceInputs := whencli.ReduceInputs{}

	for _, requirment := range w.ListChangeInFunctions(w.Requirments) {
		result, err := w.EvalFunction(requirment)
		if err != nil {
			return err
		}

		input := map[string]interface{}{}
		input["name"] = requirment.Search("name")
		input["params"] = requirment.Search("params")
		input["result"] = result

		reduceInputs.Functions = append(reduceInputs.Functions, input)
	}

	TotalEval += n() - start2

	start3 := n()
	result, err := whencli.Reduce(w.Expression, reduceInputs)
	if err != nil {
		return err
	}

	w.Expression = result
	TotalReduce += n() - start3

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
	fun, err := changein.Parse(w.Path, input, w.YamlPath)
	if err != nil {
		return false, err
	}

	return changein.Eval(fun)
}
