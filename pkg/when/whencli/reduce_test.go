package whencli

import (
	"testing"

	assert "github.com/stretchr/testify/assert"
)

func Test__Reduce(t *testing.T) {
	expressions := []string{
		"branch = 'master'",
		"change_in('lib')",
	}

	inputs := []ReduceInputs{
		ReduceInputs{
			Keywords: map[string]interface{}{
				"branch": "master",
			},
			Functions: []interface{}{},
		},
		ReduceInputs{
			Keywords: map[string]interface{}{},
			Functions: []interface{}{
				fromJSON(`{"name": "change_in", "params": ["lib"], "result": false}`),
			},
		},
		ReduceInputs{
			Keywords:  map[string]interface{}{},
			Functions: []interface{}{},
		},
	}

	results, err := Reduce(expressions, inputs)

	assert.Nil(t, err)
	assert.Equal(t, len(expressions), len(results))

	assert.Equal(t, []string{
		"true",
		"false",
	}, results)
}
