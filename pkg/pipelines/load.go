package pipelines

import (
	"io/ioutil"

	gabs "github.com/Jeffail/gabs/v2"
	"github.com/ghodss/yaml"
)

func LoadFromYaml(path string) (*Pipeline, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	jsonData, err := yaml.YAMLToJSON(data)
	if err != nil {
		return nil, err
	}

	raw, err := gabs.ParseJSON(jsonData)
	if err != nil {
		return nil, err
	}

	return &Pipeline{raw: raw}, nil
}
