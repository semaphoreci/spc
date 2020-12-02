package pipelines

import (
	"encoding/json"
	"fmt"
	"strconv"

	gabs "github.com/Jeffail/gabs/v2"
	"github.com/ghodss/yaml"
	when "github.com/semaphoreci/spc/pkg/when"
)

type Pipeline struct {
	raw      *gabs.Container
	yamlPath string
}

func (p *Pipeline) EvaluateChangeIns() error {
	fmt.Println("Evaluating start.")

	for _, w := range p.ListWhenConditions() {
		err := w.Eval()
		if err != nil {
			return err
		}

		fmt.Printf("Reduced When Expression: %s\n", w.Expression)

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
	// revive:disable:add-constant

	list := []when.WhenExpression{}

	appendIfExists := func(path ...string) {
		value := p.raw.Search(path...)

		if value != nil {
			list = append(list, when.WhenExpression{
				Expression: value.Data().(string),
				Path:       path,
				YamlPath:   p.yamlPath,
			})
		}
	}

	appendIfExists("auto_cancel", "queued", "when")
	appendIfExists("auto_cancel", "running", "when")
	appendIfExists("fail_fast", "cancel", "when")
	appendIfExists("fail_fast", "stop", "when")

	for index := range p.Blocks() {
		appendIfExists("blocks", strconv.Itoa(index), "skip", "when")
		appendIfExists("blocks", strconv.Itoa(index), "run", "when")
	}

	for index := range p.Promotions() {
		appendIfExists("promotions", strconv.Itoa(index), "auto_promote", "when")
	}

	for index := range p.raw.Search("queue").Children() {
		appendIfExists("queue", strconv.Itoa(index), "when")
	}

	for index := range p.raw.Search("priority").Children() {
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
