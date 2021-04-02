package pipelines

import (
	"encoding/json"
	"strconv"
	"time"

	gabs "github.com/Jeffail/gabs/v2"
	"github.com/ghodss/yaml"
	consolelogger "github.com/semaphoreci/spc/pkg/consolelogger"
)

type Pipeline struct {
	raw      *gabs.Container
	yamlPath string
}

func n() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func (p *Pipeline) UpdateString(path []string, value string) error {
	_, err := p.raw.Set(value, path...)

	return err
}

func (p *Pipeline) EvaluateChangeIns() error {
	return newWhenEvaluator(p).Run()
}

func (p *Pipeline) SubstituteEnvVarsInDockerImages() error {
	consolelogger.Infof("Substituting aaaa\n")

	containers := p.raw.Search("agent", "containers").Children()

	for containerIndex := range containers {
		consolelogger.Infof("Substituting bbb\n")

		path := []string{"agent", "containers", strconv.Itoa(containerIndex), "image"}
		newValue := "hello"

		p.UpdateString(path, newValue)
	}

	for blockIndex := range p.Blocks() {
		consolelogger.Infof("Substituting ccc\n")

		path := []string{"blocks", strconv.Itoa(blockIndex), "agent", "containers"}

		containers := p.raw.Search(path...).Children()

		for containerIndex := range containers {
			consolelogger.Infof("Substituting %d", containerIndex)

			path := append(path, []string{strconv.Itoa(containerIndex), "image"}...)

			newValue := "hello"

			p.UpdateString(path, newValue)
		}
	}

	return nil
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

func (p *Pipeline) GlobalPriorityRules() []*gabs.Container {
	return p.raw.Search("global_job_config", "priority").Children()
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
