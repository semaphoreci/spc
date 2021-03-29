package pipelines

// revive:disable:add-constant

import (
	"fmt"
	"testing"

	when "github.com/semaphoreci/spc/pkg/when"
	assert "github.com/stretchr/testify/assert"
)

func Test__WhenEvaluatorExtractAll(t *testing.T) {
	pipeline, err := LoadFromFile("../../test/fixtures/all_when_locations.yml")
	assert.Nil(t, err)

	e := newWhenEvaluator(pipeline)
	e.ExtractAll()

	for _, e1 := range e.list {
		fmt.Printf("%+v\n", e1)
	}

	assert.Equal(t, 11, len(e.list))
	assert.Equal(t, e.list, []when.WhenExpression{
		{
			Expression: "branch = 'master' and change_in('/lib')",
			Path:       []string{"auto_cancel", "queued", "when"},
			YamlPath:   "../../test/fixtures/all_when_locations.yml",
		},
		{
			Expression: "branch = 'master' and change_in('/app')",
			Path:       []string{"auto_cancel", "running", "when"},
			YamlPath:   "../../test/fixtures/all_when_locations.yml",
		},
		{
			Expression: "branch = 'master' and change_in('/lib')",
			Path:       []string{"fail_fast", "cancel", "when"},
			YamlPath:   "../../test/fixtures/all_when_locations.yml",
		},
		{
			Expression: "branch = 'master' and change_in('/app')",
			Path:       []string{"fail_fast", "stop", "when"},
			YamlPath:   "../../test/fixtures/all_when_locations.yml",
		},
		{
			Expression: "change_in('lib')",
			Path:       []string{"blocks", "0", "run", "when"},
			YamlPath:   "../../test/fixtures/all_when_locations.yml",
		},
		{
			Expression: "change_in('lib')",
			Path:       []string{"blocks", "1", "run", "when"},
			YamlPath:   "../../test/fixtures/all_when_locations.yml",
		},
		{
			Expression: "change_in('lib')",
			Path:       []string{"promotions", "0", "auto_promote", "when"},
			YamlPath:   "../../test/fixtures/all_when_locations.yml",
		},
		{
			Expression: "branch = 'master'",
			Path:       []string{"global_job_config", "priority", "0", "when"},
			YamlPath:   "../../test/fixtures/all_when_locations.yml",
		},
		{
			Expression: "branch = 'master'",
			Path:       []string{"blocks", "1", "task", "jobs", "0", "priority", "0", "when"},
			YamlPath:   "../../test/fixtures/all_when_locations.yml",
		},
		{
			Expression: "branch = 'master' and change_in('/lib')",
			Path:       []string{"queue", "0", "when"},
			YamlPath:   "../../test/fixtures/all_when_locations.yml",
		},
		{
			Expression: "branch = 'master' and change_in('/app')",
			Path:       []string{"queue", "1", "when"},
			YamlPath:   "../../test/fixtures/all_when_locations.yml",
		},
	})
}
