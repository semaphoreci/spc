package pipelines

import (
	"encoding/json"
	"fmt"
	"time"

	gabs "github.com/Jeffail/gabs/v2"
	"github.com/ghodss/yaml"
	when "github.com/semaphoreci/spc/pkg/when"
)

type Pipeline struct {
	raw      *gabs.Container
	yamlPath string
}

func n() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func (p *Pipeline) EvaluateChangeIns() error {
	fmt.Println("Evaluating start.")

	start2 := n()

	list, err := p.ExtractWhenConditions()
	if err != nil {
		return err
	}

	when.TotalList = n() - start2

	for _, w := range list {
		err := w.Eval()
		if err != nil {
			return err
		}

		fmt.Printf("Reduced When Expression: %s\n", w.Expression)

		p.raw.Set(w.Expression, w.Path...)
	}

	fmt.Println("Evaluating end.")

	fmt.Println("Total List")
	fmt.Println(when.TotalList)

	fmt.Println("Total Eval")
	fmt.Println(when.TotalEval)

	fmt.Println("Total Reduce")
	fmt.Println(when.TotalReduce)

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
