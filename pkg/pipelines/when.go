package pipelines

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os/exec"
	"strconv"
	"strings"

	gjson "github.com/tidwall/gjson"
)

func (p *Pipeline) ListWhenConditions() *WhenList {
	list := &WhenList{pipeline: p}

	list.AppendIfExists("auto_cancel", "queued", "when")
	list.AppendIfExists("auto_cancel", "running", "when")
	list.AppendIfExists("fail_fast", "cancel", "when")
	list.AppendIfExists("fail_fast", "stop", "when")

	for index, _ := range p.Blocks {
		i := strconv.Itoa(index)

		list.AppendIfExists("blocks", i, "skip", "when")
		list.AppendIfExists("blocks", i, "run", "when")
	}

	for index, _ := range p.Promotions {
		i := strconv.Itoa(index)

		list.AppendIfExists("promotions", i, "auto_promote", "when")
	}

	return list
}

func (p *Pipeline) EvaluateChangeIns() {
	fmt.Println("Evaluating start.")

	whenList := p.ListWhenConditions()

	for _, w := range whenList.List {
		fmt.Println("Processing when expression:")
		fmt.Println(w.Expression)

		fmt.Println("From:")
		fmt.Println(w.Path)

		bytes, _ := exec.Command("when", "list-inputs", w.Expression).Output()
		output := string(bytes)

		fmt.Println("Inputs needed for this expression:")
		fmt.Println(output)

		neededInputs := gjson.Parse(output).Array()

		m := map[string]string{}
		inputs := WhenInputs{Keywords: m}

		for _, input := range neededInputs {
			elType := input.Get("type").String()
			if elType != "fun" {
				continue
			}

			elName := input.Get("name").String()
			if elName != "change_in" {
				continue
			}

			fmt.Println("Running git command")
			gitOpts := []string{"diff", "--name-only", "origin/master..HEAD"}

			fmt.Printf("git %s\n", strings.Join(gitOpts, " "))

			bytes, _ := exec.Command("git", gitOpts...).CombinedOutput()
			diffList := string(bytes)

			fmt.Println("Diff list:")
			fmt.Println(diffList)

			diffs := strings.Split(diffList, "\n")

			changes := false
			for _, filePath := range diffs {
				if filePath == input.Get("params").Array()[0].String() {
					changes = true
					break
				}
			}

			funInput := WhenFunctionInput{
				Name:   "change_in",
				Params: input.Get("params").Array(),
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
	}

	fmt.Println("Evaluating end.")
}

type WhenFunctionInput struct {
	Name   string      `json:"name"`
	Params interface{} `json:"params"`
	Result bool        `json:"result`
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

	pipeline *Pipeline
}

func (w *WhenList) AppendIfExists(path ...string) {
	value := w.pipeline.Lookup(path)

	if value != nil {
		w.List = append(w.List, WhenListElement{
			Expression: value.(string),
			Path:       path,
		})
	}
}
