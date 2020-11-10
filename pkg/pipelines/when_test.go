package pipelines

import (
	"testing"

	assert "github.com/stretchr/testify/assert"
	require "github.com/stretchr/testify/assert"
)

func Test__ListWhenConditions(t *testing.T) {
	pipeline, err := LoadFromYaml("../../test/fixtures/when.yml")
	require.Nil(t, err)

	result := pipeline.ListWhenConditions().list

	assert.Equal(t, len(result), 2)

	assert.Equal(t, result[0].expression, "branch = 'master'")
	assert.Equal(t, result[1].expression, "change_in('lib')")
}
