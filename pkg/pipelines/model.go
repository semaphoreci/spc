package pipelines

import (
	"encoding/json"
	"fmt"
	"time"

	gabs "github.com/Jeffail/gabs/v2"
	"github.com/ghodss/yaml"
	when "github.com/semaphoreci/spc/pkg/when"
	whencli "github.com/semaphoreci/spc/pkg/when/whencli"
)

type Pipeline struct {
	raw      *gabs.Container
	yamlPath string
}

func n() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

var TotalList int64
var TotalEval int64
var TotalReduce int64

func (p *Pipeline) EvaluateChangeIns() error {
	list, err := p.ExtractWhenConditions()
	if err != nil {
		return err
	}

	p.displayFoundWhenExpressions(list)

	fmt.Println("Evaluating when expressions.\n")

	for index, condition := range list {
		fmt.Printf("%03d) %s\n", index+1, condition.Expression)

		err := list[index].Eval()
		if err != nil {
			return err
		}
	}

	expressions := []string{}
	inputs := []whencli.ReduceInputs{}

	for index := range list {
		expressions = append(expressions, list[index].Expression)
		inputs = append(inputs, list[index].ReduceInputs)
	}

	expressions, err = whencli.Reduce(expressions, inputs)
	if err != nil {
		return err
	}

	for index := range expressions {
		p.raw.Set(expressions[index], list[index].Path...)
	}

	return nil
}

func (p *Pipeline) Blocks() []*gabs.Container {
	return p.raw.Search("blocks").Children()
}

func (p *Pipeline) Promotions() []*gabs.Container {
	return p.raw.Search("blocks").Children()
}

func (p *Pipeline) PathExists(path []string) bool {
	return p.raw.Exists(path...)
}

func (p *Pipeline) GetStringFromPath(path []string) string {
	return p.raw.Search(path...).Data().(string)
}

func (p *Pipeline) PriorityRules() []*gabs.Container {
	return p.raw.Search("priority").Children()
}

func (p *Pipeline) QueueRules() []*gabs.Container {
	return p.raw.Search("queue").Children()
}

func (p *Pipeline) ExtractWhenConditions() ([]when.WhenExpression, error) {
	extractor := whenExtractor{pipeline: p}
	extractor.ExtractAll()

	return extractor.Parse()
}

func (p *Pipeline) displayFoundWhenExpressions(list []when.WhenExpression) {
	fmt.Printf("Found when expressions at %d locations:\n\n", len(list))

	for index, condition := range list {
		fmt.Printf("%03d) Location: %+v\n", index+1, condition.Path)
		fmt.Printf("     File: %s\n", condition.YamlPath)
		fmt.Printf("     Expression: %s\n", condition.Expression)
		fmt.Println()
	}
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
