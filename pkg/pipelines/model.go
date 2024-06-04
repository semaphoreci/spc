package pipelines

import (
	"encoding/json"

	gabs "github.com/Jeffail/gabs/v2"
	"github.com/ghodss/yaml"
)

type Pipeline struct {
	raw      *gabs.Container
	yamlPath string
}

func (p *Pipeline) UpdateField(path []string, value interface{}) error {
	_, err := p.raw.Set(value, path...)

	return err
}

func (p *Pipeline) EvaluateChangeIns() error {
	return newWhenEvaluator(p).Run()
}

func (p *Pipeline) EvaluateTemplates() error {
	return newTemplateEvaluator(p).Run()
}

func (p *Pipeline) Blocks() []*gabs.Container {
	return p.raw.Search("blocks").Children()
}

func (p *Pipeline) Promotions() []*gabs.Container {
	return p.raw.Search("promotions").Children()
}

func (p *Pipeline) PathExists(path []string) bool {
	return p.raw.Exists(path...)
}

func (p *Pipeline) Agent() *gabs.Container {
	return p.raw.Search("agent")
}

func (p *Pipeline) GlobalJobConfig() *gabs.Container {
	return p.raw.Search("global_job_config")
}

func (p *Pipeline) GlobalPriorityRules() []*gabs.Container {
	return p.raw.Search("global_job_config", "priority").Children()
}

func (p *Pipeline) GlobalSecrets() []*gabs.Container {
	return p.raw.Search("global_job_config", "secrets").Children()
}

func (p *Pipeline) AfterPipelineTask() *gabs.Container {
	return p.raw.Search("after_pipeline", "task")
}

func (p *Pipeline) QueueRules() []*gabs.Container {
	return p.raw.Search("queue").Children()
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
