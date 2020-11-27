package pipelines

import (
	"encoding/json"
	"fmt"
	"strconv"

	gabs "github.com/Jeffail/gabs/v2"
	"github.com/ghodss/yaml"
	logs "github.com/semaphoreci/spc/pkg/logs"
	when "github.com/semaphoreci/spc/pkg/when"
)

type Pipeline struct {
	raw *gabs.Container
}

func (p *Pipeline) EvaluateChangeIns(yamlPath string) error {
	fmt.Println("Evaluating start.")

	for _, w := range p.ListWhenConditions() {
		fmt.Println("")
		fmt.Printf("Processing when expression %s\n", w.Expression)
		fmt.Printf("  From: %v\n", w.Path)

		inputs, err := w.ListNeededInputs()
		fmt.Printf("  Inputs needed: %v\n", inputs.Requirments.String())

		for _, input := range inputs.Requirments.Children() {
			if !when.IsChangeInFunction(input) {
				continue
			}

			fun, err := when.ParseChangeIn(&w, input, yamlPath)
			if err != nil {
				panic(err)
			}

			fmt.Println("  Checking if branch exists.")
			if !fun.DefaultBranchExists() {
				logs.Log(logs.ErrorChangeInMissingBranch{
					Message: "Unknown git reference 'random'.",
					Location: logs.Location{
						Path: w.Path,
					},
				})

				return fmt.Errorf("  Branch '%s' does not exists.", fun.Params.DefaultBranch)
			}

			hasChanges := fun.Eval()

			funInput := when.FunctionInput{
				Name:   "change_in",
				Params: input.Search("params"),
				Result: hasChanges,
			}

			inputs.Functions = append(inputs.Functions, funInput)
		}

		err = w.Reduce(inputs)
		if err != nil {
			panic(err)
		}

		fmt.Printf("  Reduced When Expression: %s\n", w.Expression)

		p.raw.Set(w.Expression, w.Path...)
	}

	fmt.Println("Evaluating end.")
	return nil
}

func (p *Pipeline) Blocks() []*gabs.Container {
	return p.raw.Search("blocks").Children()
}

func (p *Pipeline) Promotions() []*gabs.Container {
	return p.raw.Search("blocks").Children()
}

func (p *Pipeline) ListWhenConditions() []when.WhenExpression {
	list := []when.WhenExpression{}

	appendIfExists := func(path ...string) {
		value := p.raw.Search(path...)

		if value != nil {
			list = append(list, when.WhenExpression{Expression: value.Data().(string), Path: path})
		}
	}

	appendIfExists("auto_cancel", "queued", "when")
	appendIfExists("auto_cancel", "running", "when")
	appendIfExists("fail_fast", "cancel", "when")
	appendIfExists("fail_fast", "stop", "when")

	for index, _ := range p.Blocks() {
		appendIfExists("blocks", strconv.Itoa(index), "skip", "when")
		appendIfExists("blocks", strconv.Itoa(index), "run", "when")
	}

	for index, _ := range p.Promotions() {
		appendIfExists("promotions", strconv.Itoa(index), "auto_promote", "when")
	}

	for index, _ := range p.raw.Search("queue").Children() {
		appendIfExists("queue", strconv.Itoa(index), "when")
	}

	for index, _ := range p.raw.Search("priority").Children() {
		appendIfExists("priority", strconv.Itoa(index), "when")
	}

	return list
}

func (p *Pipeline) ToJSON() ([]byte, error) {
	return json.Marshal(p.raw)
}

func (p *Pipeline) ToYAML() ([]byte, error) {
	jsonPpl, err := p.ToJSON()
	if err != nil {
		return []byte{}, err
	}

	yamlPpl, err := yaml.JSONToYAML(jsonPpl)
	if err != nil {
		return []byte{}, err
	}

	return yamlPpl, nil
}
