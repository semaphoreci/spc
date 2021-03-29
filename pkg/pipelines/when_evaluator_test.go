package pipelines

// revive:disable:add-constant

import (
	"testing"

	when "github.com/semaphoreci/spc/pkg/when"
	assert "github.com/stretchr/testify/assert"
)

func Test__WhenEvaluatorExtractAll(t *testing.T) {
	pipeline, err := LoadFromFile("../../test/fixtures/all_when_locations.yml")
	assert.Nil(t, err)

	e := newWhenEvaluator(pipeline)
	e.ExtractAll()

	assert.Equal(t, 5, len(e.list))
	assert.Equal(t, e.list, []when.WhenExpression{
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
	})
}
