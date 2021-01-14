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
	fmt.Println("Evaluating start.")

	start1 := n()

	list, err := p.ExtractWhenConditions()
	if err != nil {
		return err
	}

	TotalList = n() - start1

	start2 := n()

	for index := range list {
		err := list[index].Eval()
		if err != nil {
			return err
		}
	}

	TotalEval = n() - start2

	start3 := n()

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

	TotalReduce = n() - start3

	fmt.Println("Evaluating end.")

	fmt.Printf("Parse When Expressions:   %dms\n", TotalList)
	fmt.Printf("Evaluated change_in:      %dms\n", TotalEval)
	fmt.Printf("Reduce When Expressions:  %dms\n", TotalReduce)

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
