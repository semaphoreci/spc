package pipelines

import (
	"io/ioutil"

	gabs "github.com/Jeffail/gabs/v2"
	"github.com/ghodss/yaml"
)

func LoadFromYaml(path string) (*gabs.Container, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	jsonData, err := yaml.YAMLToJSON(data)
	if err != nil {
		return nil, err
	}

	pipeline, err := gabs.ParseJSON(jsonData)
	if err != nil {
		return nil, err
	}

	return pipeline, nil
}
