package pipelines

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os/exec"
	"strconv"
	"strings"

	gabs "github.com/Jeffail/gabs/v2"
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

	return list
}

func EvaluateChangeIns(p *gabs.Container) {
	fmt.Println("Evaluating start.")

	whenList := ListWhenConditions(p)

	for _, w := range whenList.List {
		fmt.Println("Processing when expression:")
		fmt.Println(w.Expression)

		fmt.Println("From:")
		fmt.Println(w.Path)

		bytes, err := exec.Command("when", "list-inputs", w.Expression).Output()
		if err != nil {
			fmt.Println(string(bytes))
			panic(err)
		}

		fmt.Println("Inputs needed for this expression:")
		fmt.Println(string(bytes))

		neededInputs, err := gabs.ParseJSON(bytes)
		if err != nil {
			panic(err)
		}

		m := map[string]string{}
		inputs := WhenInputs{Keywords: m}

		for _, input := range neededInputs.Children() {
			elType := input.Search("type").Data().(string)
			if elType != "fun" {
				continue
			}

			elName := input.Search("name").Data().(string)
			if elName != "change_in" {
				continue
			}

			defaultBranch := "master"
			if input.Exists("params", "1", "default_branch") {
				defaultBranch = input.Search("params", "1", "default_branch").Data().(string)
			}

			fmt.Println("Running git command")
			gitOpts := []string{"diff", "--name-only", fmt.Sprintf("origin/%s..HEAD", defaultBranch)}

			fmt.Printf("git %s\n", strings.Join(gitOpts, " "))

			bytes, _ := exec.Command("git", gitOpts...).CombinedOutput()
			diffList := string(bytes)

			fmt.Println("Diff list:")
			fmt.Println(diffList)

			diffs := strings.Split(diffList, "\n")

			changes := false
			for _, filePath := range diffs {
				if filePath == input.Search("params").Data().([]interface{})[0].(string) {
					changes = true
					break
				}
			}

			funInput := WhenFunctionInput{
				Name:   "change_in",
				Params: input.Search("params"),
				Result: changes,
			}

			inputs.Functions = append(inputs.Functions, funInput)
		}

		fmt.Println(inputs)
		inputBytes, err := json.Marshal(inputs)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(inputBytes))

		err = ioutil.WriteFile("/tmp/inputs.json", inputBytes, 0644)
		if err != nil {
			panic(err)
		}

		bytes, err = exec.Command("when", "reduce", w.Expression, "--input", "/tmp/inputs.json").Output()
		if err != nil {
			panic(err)
		}

		fmt.Println("Result:")
		fmt.Println(string(bytes))

		expr := strings.TrimSpace(string(bytes))

		p.Set(expr, w.Path...)
	}

	fmt.Println("Evaluating end.")
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
