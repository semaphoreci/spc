package commands

import (
	"testing"

	assert "github.com/stretchr/testify/assert"
)

func Test__Extract(t *testing.T) {
	// If commands file does not exist, it retruns the error
	file := File{
		FilePath:   "non_existing_file.txt",
		ParentPath: []string{},
		YamlPath:   "../../test/fixtures/all_commands_file_locations.yml",
		Commands:   []string{},
	}
	err := file.Extract()

	assert.Error(t, err)

	expectedErrorMessage := "failed to open the commands_file at"
	assert.Contains(t, err.Error(), expectedErrorMessage)

	// If commands file is empty, it retruns the error
	file.FilePath = "empty_file.txt"
	err = file.Extract()

	assert.Error(t, err)

	expectedErrorMessage = "empty_file.txt is empty"
	assert.Contains(t, err.Error(), expectedErrorMessage)

	// Commands are read successfully from the valid file with relative path.
	file.FilePath = "valid_commands_file.txt"
	err = file.Extract()
	
	assert.Nil(t, err)

	expectedCommands := []string{"echo 1", "echo 12", "echo 123"}
	assert.Equal(t, file.Commands, expectedCommands)

	// Commands are read successfully from the valid file with absolute path.
	file.FilePath = "/../../test/fixtures/valid_commands_file.txt"
	file.Commands = []string{}
	err = file.Extract()
	
	assert.Nil(t, err)
	assert.Equal(t, file.Commands, expectedCommands)
}