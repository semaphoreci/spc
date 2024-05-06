package templates

import (
	"os"
	"testing"

	assert "github.com/stretchr/testify/assert"
)

const Foo string = "Foo"

// revive:disable:add-constant
// revive:disable:line-length-limit

func Test__Substitute(t *testing.T) {
	os.Setenv("TEST_VAL_1", Foo)
	os.Setenv("TEST_VAL_2", "Bar")
	os.Setenv("TEST_VAL_3", "Baz")
	os.Setenv("TEST_VAL_4", "9,11")

	exp := Expression{
		Expression: "",
		Path:       []string{"semaphore.yml"},
		YamlPath:   "name",
	}

	// Only params expression with various number of whitespaces

	exp.Expression = "${{parameters.TEST_VAL_1}}"
	err := exp.Substitute()
	assert.Nil(t, err)
	assert.Equal(t, Foo, exp.Value)

	exp.Expression = "${{  parameters.TEST_VAL_1}}"
	err = exp.Substitute()
	assert.Nil(t, err)
	assert.Equal(t, Foo, exp.Value)

	exp.Expression = "${{  parameters.TEST_VAL_1  }}"
	err = exp.Substitute()
	assert.Nil(t, err)
	assert.Equal(t, Foo, exp.Value)

	// Text before and after  params expression

	exp.Expression = "Hello ${{parameters.TEST_VAL_3}}"
	err = exp.Substitute()
	assert.Nil(t, err)
	assert.Equal(t, "Hello Baz", exp.Value)

	exp.Expression = "${{parameters.TEST_VAL_3}} world"
	err = exp.Substitute()
	assert.Nil(t, err)
	assert.Equal(t, "Baz world", exp.Value)

	exp.Expression = "Hello ${{parameters.TEST_VAL_3}} world"
	err = exp.Substitute()
	assert.Nil(t, err)
	assert.Equal(t, "Hello Baz world", exp.Value)

	// Multiple params expressions

	exp.Expression = "Hello ${{parameters.TEST_VAL_1}} ${{parameters.TEST_VAL_2}}"
	err = exp.Substitute()
	assert.Nil(t, err)
	assert.Equal(t, "Hello Foo Bar", exp.Value)

	exp.Expression = "My name is ${{parameters.TEST_VAL_2}}, ${{parameters.TEST_VAL_1}} ${{parameters.TEST_VAL_2}}"
	err = exp.Substitute()
	assert.Nil(t, err)
	assert.Equal(t, "My name is Bar, Foo Bar", exp.Value)

	// If the env var is not present, the env var name is used

	exp.Expression = "${{ \"abc\" }}"
	err = exp.Substitute()
	assert.Nil(t, err)
	assert.Equal(t, "abc", exp.Value)

	exp.Expression = "Missing ${{parameters.THE_POINT}}"
	err = exp.Substitute()
	assert.Nil(t, err)
	assert.Equal(t, "Missing THE_POINT", exp.Value)

	exp.Expression = "%{{ parameters.THE_POINT | splitList \"_\" }}"
	err = exp.Substitute()
	assert.Nil(t, err)
	assert.Equal(t, []interface{}{"THE", "POINT"}, exp.Value)

	exp.Expression = "Missing %{{ parameters.THE_POINT | splitList \"_\" }}"
	err = exp.Substitute()
	assert.Nil(t, err)
	assert.Equal(t, "Missing [\"THE\",\"POINT\"]", exp.Value)

	exp.Expression = "${{ parameters.TEST_VAL_4 | splitList \",\" | join \".\" }}"
	err = exp.Substitute()
	assert.Nil(t, err)
	assert.Equal(t, "9.11", exp.Value)

	exp.Expression = "%{{ parameters.TEST_VAL_4 | splitList \",\" | join \".\"  }}"
	err = exp.Substitute()
	assert.Nil(t, err)
	assert.Equal(t, "9.11", exp.Value)

	exp.Expression = "%{{ parameters.TEST_VAL_4 | splitList \",\" | join \".\" | float64 }}"
	err = exp.Substitute()
	assert.Nil(t, err)
	assert.Equal(t, 9.11, exp.Value)

	// 9~11
	exp.Expression = "${{ parameters.TEST_VAL_4 | splitList \",\" | join \"~\" }}"
	err = exp.Substitute()
	assert.Nil(t, err)
	assert.Equal(t, "9~11", exp.Value)

	exp.Expression = "%{{ parameters.TEST_VAL_4 | splitList \",\" }} is a heck of a list!"
	err = exp.Substitute()
	assert.Nil(t, err)
	assert.Equal(t, "[\"9\",\"11\"] is a heck of a list!", exp.Value)

	exp.Expression = "${{ parameters.TEST_VAL_4 | splitList \",\" }} is a heck of a list!"
	err = exp.Substitute()
	assert.Nil(t, err)
	assert.Equal(t, "[9 11] is a heck of a list!", exp.Value)

	exp.Expression = "${{ \"${{,${{\" | splitList \",\" | join \" \" }} is a heck of a list!"
	err = exp.Substitute()
	assert.Nil(t, err)
	assert.Equal(t, "${{ ${{ is a heck of a list!", exp.Value)

	exp.Expression = "${{ \"${{,${{\" | splitList \",\" | join \"}}\" }} is a heck of a list!"
	err = exp.Substitute()
	assert.Error(t, err, "nested expressions are not supported")

	exp.Expression = "${{ \"${{parameters.TEST_VAL_1}}, ${{parameters.TEST_VAL_2}}\" | splitList \",\" }}"
	err = exp.Substitute()
	assert.Error(t, err, "nested expressions are not supported")

	exp.Expression = "%{{ \"${{parameters.TEST_VAL_1}}, ${{parameters.TEST_VAL_2}}\" | splitList \",\" }}"
	err = exp.Substitute()
	assert.Error(t, err, "nested expressions are not supported")
}
