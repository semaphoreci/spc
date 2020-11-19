package pipelines

import (
	"testing"

	assert "github.com/stretchr/testify/assert"
	require "github.com/stretchr/testify/assert"
)

func Test__ListWhenConditions(t *testing.T) {
	pipeline, err := LoadFromYaml("../../test/fixtures/when.yml")
	require.Nil(t, err)

	result := ListWhenConditions(pipeline).List

	assert.Equal(t, len(result), 2)

	assert.Equal(t, result[0].Expression, "branch = 'master'")
	assert.Equal(t, result[0].Path, []string{"auto_cancel", "queued", "when"})

	assert.Equal(t, result[1].Expression, "change_in('lib')")
	assert.Equal(t, result[1].Path, []string{"blocks", "0", "skip", "when"})
}
