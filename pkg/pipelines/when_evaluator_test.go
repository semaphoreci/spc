package pipelines

import (
	"testing"

	when "github.com/semaphoreci/spc/pkg/when"
	assert "github.com/stretchr/testify/assert"
)

func Test__WhenEvaluatorExtractAll(t *testing.T) {
	pipeline, err := LoadFromFile("../../test/fixtures/all_when_locations.yml")
	assert.Nil(t, err)

	e := newWhenEvaluator(pipeline)

	assert.Equal(t, e, []when.WhenExpression{
		{},
	})
}
