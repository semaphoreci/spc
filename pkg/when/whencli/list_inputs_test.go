package whencli

import (
	"testing"

	gabs "github.com/Jeffail/gabs/v2"
	assert "github.com/stretchr/testify/assert"
)

func Test__ListInputs(t *testing.T) {
	expressions := []string{
		"branch = 'master'",
		"change_in('lib')",
		"branch = ",
	}

	results, err := ListInputs(expressions)

	assert.Nil(t, err)
	assert.Equal(t, len(expressions), len(results))
	assert.Equal(t, []ListInputsResult{
		ListInputsResult{
			Expression: expressions[0],
			Error:      "",
			Inputs:     fromJSON(`[{"name": "branch", "type": "keyword"}]`),
		},
		ListInputsResult{
			Expression: expressions[1],
			Error:      "",
			Inputs:     fromJSON(`[{"name": "change_in", "params": ["lib"], "type": "fun"}]`),
		},
		ListInputsResult{
			Expression: expressions[2],
			Error:      "Invalid or incomplete expression at the end of the line.",
			Inputs:     fromJSON(`[]`),
		},
	}, results)
}

func fromJSON(s string) *gabs.Container {
	res, err := gabs.ParseJSON([]byte(s))
	if err != nil {
		panic(err)
	}

	return res
}
