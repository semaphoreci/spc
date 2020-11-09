package pipelines

import "fmt"

type Block struct {
	Skip string `yaml:"skip"`
	Run  string `yaml:"skip"`
}

type Pipeline struct {
	Blocks []Block `yaml:"blocks"`
}

func LoadFromYaml(path string) (*Pipeline, error) {
	return nil, fmt.Errorf("not found")
}
