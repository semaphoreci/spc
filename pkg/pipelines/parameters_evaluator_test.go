package pipelines

// revive:disable:add-constant

import (
	"fmt"
	"os"
	"testing"

	parameters "github.com/semaphoreci/spc/pkg/parameters"
	assert "github.com/stretchr/testify/assert"
)

func Test__ParametersEvaluatorExtractAll(t *testing.T) {
	pipeline, err := LoadFromFile("../../test/fixtures/all_parameters_locations.yml")
	assert.Nil(t, err)

	e := newParametersEvaluator(pipeline)
	e.ExtractAll()

	for _, e1 := range e.list {
		fmt.Printf("%+v\n", e1)
	}

	assert.Equal(t, 8, len(e.list))
	assert.Equal(t, e.list, []parameters.ParametersExpression{
		{
			Expression: "Deploy to ${{parameters.DEPLOY_ENV}} on ${{parameters.SERVER}}",
			Path:       []string{"name"},
			YamlPath:   "../../test/fixtures/all_parameters_locations.yml",
			Value:      "",
		},
		{
			Expression: "${{parameters.DEPLOY_ENV}}_deployment_queue",
			Path:       []string{"queue", "0", "name"},
			YamlPath:   "../../test/fixtures/all_parameters_locations.yml",
			Value:      "",
		},
		{
			Expression: "${{parameters.MISSING}}_queue",
			Path:       []string{"queue", "1", "name"},
			YamlPath:   "../../test/fixtures/all_parameters_locations.yml",
			Value:      "",
		},
		{
			Expression: "${{parameters.DEPLOY_ENV}}_deploy_key",
			Path:       []string{"global_job_config", "secrets", "0", "name"},
			YamlPath:   "../../test/fixtures/all_parameters_locations.yml",
			Value:      "",
		},
		{
			Expression: "${{parameters.DEPLOY_ENV}}_dockerhub",
			Path:       []string{"blocks", "0", "task", "secrets", "0", "name"},
			YamlPath:   "../../test/fixtures/all_parameters_locations.yml",
			Value:      "",
		},
		{
			Expression: "${{parameters.DEPLOY_ENV}}_ecr",
			Path:       []string{"blocks", "0", "task", "secrets", "1", "name"},
			YamlPath:   "../../test/fixtures/all_parameters_locations.yml",
			Value:      "",
		},
		{
			Expression: "${{parameters.DEPLOY_ENV}}_deploy_key",
			Path:       []string{"blocks", "1", "task", "secrets", "0", "name"},
			YamlPath:   "../../test/fixtures/all_parameters_locations.yml",
			Value:      "",
		},
		{
			Expression: "${{parameters.DEPLOY_ENV}}_aws_creds",
			Path:       []string{"blocks", "1", "task", "secrets", "1", "name"},
			YamlPath:   "../../test/fixtures/all_parameters_locations.yml",
			Value:      "",
		},
	})
}

func Test__Run(t *testing.T) {
	pipeline, err := LoadFromFile("../../test/fixtures/all_parameters_locations.yml")
	assert.Nil(t, err)

	e := newParametersEvaluator(pipeline)

	os.Setenv("DEPLOY_ENV", "prod")
	os.Setenv("SERVER", "server_1")

	err = e.Run()
	assert.Nil(t, err)

	yaml_result, er := e.pipeline.ToYAML()
	assert.Nil(t, er)
	fmt.Printf("%s\n", yaml_result)

	assert_value_on_path(t, e, []string{"name"}, "Deploy to prod on server_1")
	assert_value_on_path(t, e, []string{"queue", "0", "name"}, "prod_deployment_queue")
	assert_value_on_path(t, e, []string{"queue", "1", "name"}, "MISSING_queue")
	assert_value_on_path(t, e, []string{"global_job_config", "secrets", "0", "name"}, "prod_deploy_key")
	assert_value_on_path(t, e, []string{"blocks", "0", "task", "secrets", "0", "name"}, "prod_dockerhub")
	assert_value_on_path(t, e, []string{"blocks", "0", "task", "secrets", "1", "name"}, "prod_ecr")
	assert_value_on_path(t, e, []string{"blocks", "1", "task", "secrets", "0", "name"}, "prod_deploy_key")
	assert_value_on_path(t, e, []string{"blocks", "1", "task", "secrets", "1", "name"}, "prod_aws_creds")
}

func assert_value_on_path(t *testing.T, e *parametersEvaluator, path []string, value string) {
	field, ok := e.pipeline.raw.Search(path...).Data().(string)
	if !ok {
		assert.Equal(t, "Invalid value after parsing at", path)
	}

	assert.Equal(t, field, value)
}
