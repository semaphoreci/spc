package pipelines

// revive:disable:add-constant
// revive:disable:line-length-limit

import (
	"fmt"
	"os"
	"testing"

	templates "github.com/semaphoreci/spc/pkg/templates"
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

	assert.Equal(t, 9, len(e.list))
	assert.Equal(t, e.list, []templates.Expression{
		{
			Expression: "Deploy to ${{parameters.DEPLOY_ENV}} on ${{parameters.SERVER}}",
			Path:       []string{"name"},
			YamlPath:   "../../test/fixtures/all_parameters_locations.yml",
			Value:      nil,
		},
		{
			Expression: "${{parameters.DEPLOY_ENV}}_deployment_queue",
			Path:       []string{"queue", "0", "name"},
			YamlPath:   "../../test/fixtures/all_parameters_locations.yml",
			Value:      nil,
		},
		{
			Expression: "${{parameters.MISSING}}_queue",
			Path:       []string{"queue", "1", "name"},
			YamlPath:   "../../test/fixtures/all_parameters_locations.yml",
			Value:      nil,
		},
		{
			Expression: "${{parameters.DEPLOY_ENV}}_deploy_key",
			Path:       []string{"global_job_config", "secrets", "0", "name"},
			YamlPath:   "../../test/fixtures/all_parameters_locations.yml",
			Value:      nil,
		},
		{
			Expression: "${{parameters.DEPLOY_ENV}}_dockerhub",
			Path:       []string{"blocks", "0", "task", "secrets", "0", "name"},
			YamlPath:   "../../test/fixtures/all_parameters_locations.yml",
			Value:      nil,
		},
		{
			Expression: "${{parameters.DEPLOY_ENV}}_ecr",
			Path:       []string{"blocks", "0", "task", "secrets", "1", "name"},
			YamlPath:   "../../test/fixtures/all_parameters_locations.yml",
			Value:      nil,
		},
		{
			Expression: "${{parameters.DEPLOY_ENV}}_deploy_key",
			Path:       []string{"blocks", "1", "task", "secrets", "0", "name"},
			YamlPath:   "../../test/fixtures/all_parameters_locations.yml",
			Value:      nil,
		},
		{
			Expression: "${{parameters.DEPLOY_ENV}}_aws_creds",
			Path:       []string{"blocks", "1", "task", "secrets", "1", "name"},
			YamlPath:   "../../test/fixtures/all_parameters_locations.yml",
			Value:      nil,
		},
		{
			Expression: "%{{parameters.AWS_REGIONS | split \",\" }}",
			Path:       []string{"blocks", "1", "task", "jobs", "0", "matrix", "0", "values"},
			YamlPath:   "../../test/fixtures/all_parameters_locations.yml",
			Value:      nil,
		},
	})
}

func Test__Run(t *testing.T) {
	pipeline, err := LoadFromFile("../../test/fixtures/all_parameters_locations.yml")
	assert.Nil(t, err)

	e := newParametersEvaluator(pipeline)

	os.Setenv("DEPLOY_ENV", "prod")
	os.Setenv("SERVER", "server_1")
	os.Setenv("AWS_REGIONS", "us-east-1,us-west-2")

	err = e.Run()
	assert.Nil(t, err)

	yamlResult, er := e.pipeline.ToYAML()
	assert.Nil(t, er)
	fmt.Printf("%s\n", yamlResult)

	assertValueOnPath(t, e, []string{"name"}, "Deploy to prod on server_1")
	assertValueOnPath(t, e, []string{"queue", "0", "name"}, "prod_deployment_queue")
	assertValueOnPath(t, e, []string{"queue", "1", "name"}, "MISSING_queue")
	assertValueOnPath(t, e, []string{"global_job_config", "secrets", "0", "name"}, "prod_deploy_key")
	assertValueOnPath(t, e, []string{"blocks", "0", "task", "secrets", "0", "name"}, "prod_dockerhub")
	assertValueOnPath(t, e, []string{"blocks", "0", "task", "secrets", "1", "name"}, "prod_ecr")
	assertValueOnPath(t, e, []string{"blocks", "1", "task", "secrets", "0", "name"}, "prod_deploy_key")
	assertValueOnPath(t, e, []string{"blocks", "1", "task", "secrets", "1", "name"}, "prod_aws_creds")
	assertValueOnPath(t, e, []string{"blocks", "1", "task", "jobs", "0", "matrix", "0", "values"}, []interface{}{"us-east-1", "us-west-2"})
}

func assertValueOnPath(t *testing.T, e *parametersEvaluator, path []string, value interface{}) {
	field := e.pipeline.raw.Search(path...).Data()
	assert.Equal(t, field, value)
}
