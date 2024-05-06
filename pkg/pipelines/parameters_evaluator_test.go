package pipelines

// revive:disable:add-constant
// revive:disable:line-length-limit

import (
	"encoding/json"
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

	expectedExpressions := []templates.Expression{
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
			Path:       []string{"blocks", "1", "task", "secrets", "0", "name"},
			YamlPath:   "../../test/fixtures/all_parameters_locations.yml",
			Value:      nil,
		},
		{
			Expression: "${{parameters.DEPLOY_ENV}}_ecr",
			Path:       []string{"blocks", "1", "task", "secrets", "1", "name"},
			YamlPath:   "../../test/fixtures/all_parameters_locations.yml",
			Value:      nil,
		},
		{
			Expression: "${{parameters.DEPLOY_ENV}}_deploy_key",
			Path:       []string{"blocks", "2", "task", "secrets", "0", "name"},
			YamlPath:   "../../test/fixtures/all_parameters_locations.yml",
			Value:      nil,
		},
		{
			Expression: "${{parameters.DEPLOY_ENV}}_aws_creds",
			Path:       []string{"blocks", "2", "task", "secrets", "1", "name"},
			YamlPath:   "../../test/fixtures/all_parameters_locations.yml",
			Value:      nil,
		},
		{
			Expression: "${{parameters.DEPLOY_ENV}}_slack_token",
			Path:       []string{"after_pipeline", "task", "secrets", "0", "name"},
			YamlPath:   "../../test/fixtures/all_parameters_locations.yml",
			Value:      nil,
		},
		{
			Expression: "Deploy image to ${{parameters.DEPLOY_ENV}}",
			Path:       []string{"blocks", "2", "name"},
			YamlPath:   "../../test/fixtures/all_parameters_locations.yml",
			Value:      nil,
		},
		{
			Expression: "Deploy to ${{parameters.DEPLOY_ENV}} on ${{parameters.SERVER}}",
			Path:       []string{"blocks", "2", "task", "jobs", "0", "name"},
			YamlPath:   "../../test/fixtures/all_parameters_locations.yml",
			Value:      nil,
		},
		{
			Expression: "Notify on Slack: %{{parameters.SLACK_CHANNELS | splitList \",\"}}",
			Path:       []string{"after_pipeline", "task", "jobs", "0", "name"},
			YamlPath:   "../../test/fixtures/all_parameters_locations.yml",
			Value:      nil,
		},
		{
			Expression: "Ping ${{parameters.DEPLOY_ENV}} from %{{parameters.PARALLELISM}} jobs",
			Path:       []string{"after_pipeline", "task", "jobs", "1", "name"},
			YamlPath:   "../../test/fixtures/all_parameters_locations.yml",
			Value:      nil,
		},
		{
			Expression: "${{parameters.MACHINE_TYPE}}",
			Path:       []string{"agent", "machine", "type"},
			YamlPath:   "../../test/fixtures/all_parameters_locations.yml",
			Value:      nil,
		},
		{
			Expression: "${{parameters.OS_IMAGE}}",
			Path:       []string{"agent", "machine", "os_image"},
			YamlPath:   "../../test/fixtures/all_parameters_locations.yml",
			Value:      nil,
		},
		{
			Expression: "${{parameters.MACHINE_TYPE}}",
			Path:       []string{"blocks", "0", "task", "agent", "machine", "type"},
			YamlPath:   "../../test/fixtures/all_parameters_locations.yml",
			Value:      nil,
		},
		{
			Expression: "${{parameters.DEPLOY_ENV}}_test_container",
			Path:       []string{"blocks", "0", "task", "agent", "containers", "0", "name"},
			YamlPath:   "../../test/fixtures/all_parameters_locations.yml",
			Value:      nil,
		},
		{
			Expression: "${{parameters.DEPLOY_ENV}}_test_image",
			Path:       []string{"blocks", "0", "task", "agent", "containers", "0", "image"},
			YamlPath:   "../../test/fixtures/all_parameters_locations.yml",
			Value:      nil,
		},
		{
			Expression: "${{parameters.DEPLOY_ENV}}_api_key",
			Path:       []string{"blocks", "0", "task", "agent", "containers", "0", "secrets", "0", "name"},
			YamlPath:   "../../test/fixtures/all_parameters_locations.yml",
			Value:      nil,
		},
		{
			Expression: "%{{parameters.AWS_REGIONS | splitList \",\"}}",
			Path:       []string{"blocks", "2", "task", "jobs", "0", "matrix", "0", "values"},
			YamlPath:   "../../test/fixtures/all_parameters_locations.yml",
			Value:      nil,
		},
		{
			Expression: "%{{parameters.SLACK_CHANNELS | splitList \",\" }}",
			Path:       []string{"after_pipeline", "task", "jobs", "0", "matrix", "0", "values"},
			YamlPath:   "../../test/fixtures/all_parameters_locations.yml",
			Value:      nil,
		},
		{
			Expression: "%{{parameters.PARALLELISM | mul 2}}",
			Path:       []string{"blocks", "0", "task", "jobs", "0", "parallelism"},
			YamlPath:   "../../test/fixtures/all_parameters_locations.yml",
			Value:      nil,
		},
		{
			Expression: "%{{parameters.PARALLELISM | int64 }}",
			Path:       []string{"after_pipeline", "task", "jobs", "1", "parallelism"},
			YamlPath:   "../../test/fixtures/all_parameters_locations.yml",
			Value:      nil,
		},
	}

	assert.Equal(t, len(expectedExpressions), len(e.list))
	for i, e1 := range e.list {
		assert.Equal(t, expectedExpressions[i].Expression, e1.Expression)
		assert.Equal(t, expectedExpressions[i].Path, e1.Path)
		assert.Equal(t, expectedExpressions[i].YamlPath, e1.YamlPath)
	}
}

func Test__Run(t *testing.T) {
	pipeline, err := LoadFromFile("../../test/fixtures/all_parameters_locations.yml")
	assert.Nil(t, err)

	e := newParametersEvaluator(pipeline)

	os.Setenv("DEPLOY_ENV", "prod")
	os.Setenv("SERVER", "server_1")
	os.Setenv("AWS_REGIONS", "us-east-1,us-west-2")
	os.Setenv("MACHINE_TYPE", "e2-standard-2")
	os.Setenv("OS_IMAGE", "ubuntu2204")
	os.Setenv("PARALLELISM", "2")
	os.Setenv("SLACK_CHANNELS", "#engineering,#general")

	err = e.Run()
	assert.Nil(t, err)

	yamlResult, er := e.pipeline.ToYAML()
	assert.Nil(t, er)
	fmt.Printf("%s\n", yamlResult)

	assertValueOnPath(t, e, []string{"name"}, "Deploy to prod on server_1")
	assertValueOnPath(t, e, []string{"agent", "machine", "type"}, "e2-standard-2")
	assertValueOnPath(t, e, []string{"agent", "machine", "os_image"}, "ubuntu2204")
	assertValueOnPath(t, e, []string{"global_job_config", "secrets", "0", "name"}, "prod_deploy_key")
	assertValueOnPath(t, e, []string{"queue", "0", "name"}, "prod_deployment_queue")
	assertValueOnPath(t, e, []string{"queue", "1", "name"}, "MISSING_queue")
	assertValueOnPath(t, e, []string{"blocks", "0", "task", "agent", "machine", "type"}, "e2-standard-2")
	assertValueOnPath(t, e, []string{"blocks", "0", "task", "agent", "containers", "0", "name"}, "prod_test_container")
	assertValueOnPath(t, e, []string{"blocks", "0", "task", "agent", "containers", "0", "image"}, "prod_test_image")
	assertValueOnPath(t, e, []string{"blocks", "0", "task", "agent", "containers", "0", "secrets", "0", "name"}, "prod_api_key")
	assertValueOnPath(t, e, []string{"blocks", "0", "task", "jobs", "0", "parallelism"}, json.Number("4"))
	assertValueOnPath(t, e, []string{"blocks", "1", "task", "secrets", "0", "name"}, "prod_dockerhub")
	assertValueOnPath(t, e, []string{"blocks", "1", "task", "secrets", "1", "name"}, "prod_ecr")
	assertValueOnPath(t, e, []string{"blocks", "2", "name"}, "Deploy image to prod")
	assertValueOnPath(t, e, []string{"blocks", "2", "task", "secrets", "0", "name"}, "prod_deploy_key")
	assertValueOnPath(t, e, []string{"blocks", "2", "task", "secrets", "1", "name"}, "prod_aws_creds")
	assertValueOnPath(t, e, []string{"blocks", "2", "task", "jobs", "0", "name"}, "Deploy to prod on server_1")
	assertValueOnPath(t, e, []string{"blocks", "2", "task", "jobs", "0", "matrix", "0", "values"}, []interface{}{"us-east-1", "us-west-2"})
	assertValueOnPath(t, e, []string{"after_pipeline", "task", "jobs", "0", "name"}, "Notify on Slack: [\"#engineering\",\"#general\"]")
	assertValueOnPath(t, e, []string{"after_pipeline", "task", "jobs", "0", "matrix", "0", "values"}, []interface{}{"#engineering", "#general"})
	assertValueOnPath(t, e, []string{"after_pipeline", "task", "jobs", "1", "name"}, "Ping prod from 2 jobs")
	assertValueOnPath(t, e, []string{"after_pipeline", "task", "jobs", "1", "parallelism"}, json.Number("2"))

}

func assertValueOnPath(t *testing.T, e *parametersEvaluator, path []string, value interface{}) {
	field := e.pipeline.raw.Search(path...).Data()
	assert.Equal(t, value, field)
}
