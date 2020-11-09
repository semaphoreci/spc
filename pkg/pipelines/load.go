package pipelines

import (
	"io/ioutil"

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

	pipeline := &Pipeline{}
	err = pipeline.UnmarshalJSON(jsonData)
	if err != nil {
		return nil, err
	}

	return pipeline, nil
}
