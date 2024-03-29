package pipelines

import (
	"io/ioutil"

	gabs "github.com/Jeffail/gabs/v2"
	"github.com/ghodss/yaml"
)

func LoadFromFile(path string) (*Pipeline, error) {
	// #nosec
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	jsonData, err := yaml.YAMLToJSON(data)
	if err != nil {
		return nil, err
	}

	raw, _ := gabs.ParseJSON(jsonData)

	return &Pipeline{raw: raw, yamlPath: path}, nil
}
