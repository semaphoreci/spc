package pipelines

import (
	"testing"

	assert "github.com/stretchr/testify/assert"
)

func Test__PipelineBlocks(t *testing.T) {
	pipeline, err := LoadFromFile("../../test/fixtures/all_when_locations.yml")

	assert.Nil(t, err)
	assert.Equal(t, 2, len(pipeline.Blocks()))
}

func Test__PipelinePromotions(t *testing.T) {
	pipeline, err := LoadFromFile("../../test/fixtures/all_when_locations.yml")

	assert.Nil(t, err)
	assert.Equal(t, 1, len(pipeline.Promotions()))
}

func Test__PipelineGlobalPriorityRules(t *testing.T) {
	pipeline, err := LoadFromFile("../../test/fixtures/all_when_locations.yml")

	assert.Nil(t, err)
	assert.Equal(t, 2, len(pipeline.GlobalPriorityRules()))

	pipeline2, err := LoadFromFile("../../test/fixtures/hello.yml")

	assert.Nil(t, err)
	assert.Equal(t, 0, len(pipeline2.GlobalPriorityRules()))
}
