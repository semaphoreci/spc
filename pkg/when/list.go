package when

import (
	"strconv"

	gabs "github.com/Jeffail/gabs/v2"
	pipelines "github.com/semaphoreci/spc/pkg/pipelines"
)

type WhenExpression struct {
	Expression string
	Path       []string
}

func ListWhenConditions(p *pipelines.Pipeline) []WhenExpression {
	list := []WhenExpression{}

	appendIfExists := func(path ...string) {
		value := p.Search(path...)

		if value == nil {
			return
		}

		list = append(list, WhenExpression{
			Expression: value.Data().(string),
			Path:       path,
		})
	}

	appendIfExists(p, "auto_cancel", "queued", "when")
	appendIfExists(p, "auto_cancel", "running", "when")
	appendIfExists(p, "fail_fast", "cancel", "when")
	appendIfExists(p, "fail_fast", "stop", "when")

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

type FunctionInput struct {
	Name   string      `json:"name"`
	Params interface{} `json:"params"`
	Result bool        `json:"result"`
}

type Inputs struct {
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
