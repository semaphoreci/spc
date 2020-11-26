package pipelines

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os/exec"
	"strconv"
	"strings"

	gabs "github.com/Jeffail/gabs/v2"
	logs "github.com/semaphoreci/spc/pkg/logs"
)

func ListWhenConditions(p *gabs.Container) *WhenList {
	list := &WhenList{}

	list.AppendIfExists(p, "auto_cancel", "queued", "when")
	list.AppendIfExists(p, "auto_cancel", "running", "when")
	list.AppendIfExists(p, "fail_fast", "cancel", "when")
	list.AppendIfExists(p, "fail_fast", "stop", "when")

	for index, _ := range p.Search("blocks").Children() {
		i := strconv.Itoa(index)

		list.AppendIfExists(p, "blocks", i, "skip", "when")
		list.AppendIfExists(p, "blocks", i, "run", "when")
	}

	for index, _ := range p.Search("promotions").Children() {
		i := strconv.Itoa(index)

		list.AppendIfExists(p, "promotions", i, "auto_promote", "when")
	}

	for index, _ := range p.Search("queue").Children() {
		i := strconv.Itoa(index)

		list.AppendIfExists(p, "queue", i, "when")
	}

	for index, _ := range p.Search("priority").Children() {
		i := strconv.Itoa(index)

		list.AppendIfExists(p, "priority", i, "when")
	}

	return list
}

func EvaluateChangeIns(p *gabs.Container, yamlPath string) error {
	fmt.Println("Evaluating start.")

	whenList := ListWhenConditions(p)

	for _, w := range whenList.List {
		fmt.Printf("Processing when expression %s\n", w.Expression)
		fmt.Printf("  From: %v\n", w.Path)

		bytes, err := exec.Command("when", "list-inputs", w.Expression).Output()
		if err != nil {
			fmt.Println(string(bytes))
			panic(err)
		}

		fmt.Printf("  Inputs needed: %s\n", string(bytes))

		neededInputs, err := gabs.ParseJSON(bytes)
		if err != nil {
			panic(err)
		}

		m := map[string]string{}
		inputs := WhenInputs{Keywords: m}

		for _, input := range neededInputs.Children() {
			if !IsChangeInFunction(input) {
				continue
			}

			fun, err := NewChangeInFunctionFromWhenInputList(input, yamlPath)
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

			funInput := WhenFunctionInput{
				Name:   "change_in",
				Params: input.Search("params"),
				Result: hasChanges,
			}

			inputs.Functions = append(inputs.Functions, funInput)
		}

		inputBytes, err := json.Marshal(inputs)
		if err != nil {
			panic(err)
		}
		fmt.Printf("  Providing inputs: %s\n", string(inputBytes))

		err = ioutil.WriteFile("/tmp/inputs.json", inputBytes, 0644)
		if err != nil {
			panic(err)
		}

		bytes, err = exec.Command("when", "reduce", w.Expression, "--input", "/tmp/inputs.json").Output()
		if err != nil {
			panic(err)
		}

		fmt.Printf("  Reduced When Expression: %s\n", string(bytes))

		expr := strings.TrimSpace(string(bytes))

		p.Set(expr, w.Path...)
	}

	fmt.Println("Evaluating end.")
	return nil
}

type WhenFunctionInput struct {
	Name   string      `json:"name"`
	Params interface{} `json:"params"`
	Result bool        `json:"result"`
}

type WhenInputs struct {
	Keywords  map[string]string   `json:"keywords"`
	Functions []WhenFunctionInput `json:"functions"`
}

func IsChangeInFunction(input *gabs.Container) bool {
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

type WhenListElement struct {
	Expression string
	Path       []string
}

type WhenList struct {
	List []WhenListElement
}

func (w *WhenList) AppendIfExists(p *gabs.Container, path ...string) {
	value := p.Search(path...)

	if value != nil {
		w.List = append(w.List, WhenListElement{
			Expression: value.Data().(string),
			Path:       path,
		})
	}
}
