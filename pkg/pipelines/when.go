package pipelines

import (
	"fmt"
	"os/exec"
	"strconv"

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
	whenList := p.ListWhenConditions()

	for _, w := range whenList.List {
		fmt.Println("Evaluating start.")
		fmt.Println(w.Expression)
		fmt.Println(w.Path)

		cmd, _ := exec.Command("when", "list-inputs", w.Expression).Output()
		output := string(cmd)
		inputs := gjson.Parse(output)

		inputs.ForEach(func(key, value gjson.Result) bool {
			fmt.Println(key)
			fmt.Println(value)

			return true
		})

		fmt.Println("Evaluating end.")
	}
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
