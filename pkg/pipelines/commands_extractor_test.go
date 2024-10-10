package pipelines

import (
	"testing"
	"fmt"
	"reflect"

	assert "github.com/stretchr/testify/assert"
	commands "github.com/semaphoreci/spc/pkg/commands"
)

func Test__findAll(t *testing.T) {
	yamlPath := "../../test/fixtures/all_commands_file_locations.yml"
	pipeline, err := LoadFromFile(yamlPath)
	assert.Nil(t, err)

	e := newCommandsExtractor(pipeline)
	e.findAll()

	for _, f := range e.files {
		fmt.Printf("%+v\n", f)
	}

	expectedFiles := []commands.File{
		{
			FilePath:   "valid_commands_file.txt",
			ParentPath: []string{"after_pipeline", "task", "epilogue", "always",},
			YamlPath:   yamlPath,
			Commands:   []string{},
		},
		{
			FilePath:   "valid_commands_file.txt",
			ParentPath: []string{"after_pipeline", "task", "epilogue", "on_fail",},
			YamlPath:   yamlPath,
			Commands:   []string{},
		},
		{
			FilePath:   "valid_commands_file.txt",
			ParentPath: []string{"after_pipeline", "task", "epilogue", "on_pass",},
			YamlPath:   yamlPath,
			Commands:   []string{},
		},
		{
			FilePath:   "valid_commands_file.txt",
			ParentPath: []string{"after_pipeline", "task", "jobs", "0",},
			YamlPath:   yamlPath,
			Commands:   []string{},
		},
		{
			FilePath:   "valid_commands_file.txt",
			ParentPath: []string{"after_pipeline", "task", "prologue",},
			YamlPath:   yamlPath,
			Commands:   []string{},
		},
		{
			FilePath:   "valid_commands_file.txt",
			ParentPath: []string{"blocks", "0", "task", "prologue",},
			YamlPath:   yamlPath,
			Commands:   []string{},
		},
		{
			FilePath:   "valid_commands_file.txt",
			ParentPath: []string{"blocks", "0", "task", "epilogue", "always",},
			YamlPath:   yamlPath,
			Commands:   []string{},
		},
		{
			FilePath:   "valid_commands_file.txt",
			ParentPath: []string{"blocks", "0", "task", "epilogue", "on_fail",},
			YamlPath:   yamlPath,
			Commands:   []string{},
		},
		{
			FilePath:   "valid_commands_file.txt",
			ParentPath: []string{"blocks", "0", "task", "epilogue", "on_pass",},
			YamlPath:   yamlPath,
			Commands:   []string{},
		},
		{
			FilePath:   "valid_commands_file.txt",
			ParentPath: []string{"blocks", "0", "task", "jobs", "0",},
			YamlPath:   yamlPath,
			Commands:   []string{},
		},
		{
			FilePath:   "valid_commands_file.txt",
			ParentPath: []string{"global_job_config", "prologue",},
			YamlPath:   yamlPath,
			Commands:   []string{},
		},
		{
			FilePath:   "valid_commands_file.txt",
			ParentPath: []string{"global_job_config", "epilogue", "always",},
			YamlPath:   yamlPath,
			Commands:   []string{},
		},
		{
			FilePath:   "valid_commands_file.txt",
			ParentPath: []string{"global_job_config", "epilogue", "on_fail",},
			YamlPath:   yamlPath,
			Commands:   []string{},
		},
		{
			FilePath:   "valid_commands_file.txt",
			ParentPath: []string{"global_job_config", "epilogue", "on_pass",},
			YamlPath:   yamlPath,
			Commands:   []string{},
		},
	}

	assert.Equal(t, len(expectedFiles), len(e.files))
	for _, f1 := range e.files {
		expectedFile := findFile(f1, expectedFiles)

		assert.Equal(t, expectedFile.FilePath, f1.FilePath)
		assert.Equal(t, expectedFile.ParentPath, f1.ParentPath)
		assert.Equal(t, expectedFile.YamlPath, f1.YamlPath)
	}
}

func findFile(file commands.File, expectedFiles []commands.File) commands.File {
	for _, e := range expectedFiles {
		if reflect.DeepEqual(e.ParentPath, file.ParentPath) {
			return e
		}
	}
	return commands.File{}
}

func Test__CommandsExtractorRun(t *testing.T) {
	yamlPath := "../../test/fixtures/all_commands_file_locations.yml"
	pipeline, err := LoadFromFile(yamlPath)
	assert.Nil(t, err)

	e := newCommandsExtractor(pipeline)
	err = e.Run()
	assert.Nil(t, err)

	yamlResult, er := e.pipeline.ToYAML()
	assert.Nil(t, er)
	fmt.Printf("%s\n", yamlResult)

	expectedCommands := []interface{}{"echo 1", "echo 12", "echo 123"}

	assertCommandsOnPath(t, e, []string{"after_pipeline", "task", "epilogue", "always", "commands"}, expectedCommands)
	assertCommandsOnPath(t, e, []string{"after_pipeline", "task", "epilogue", "on_fail", "commands"}, expectedCommands)
	assertCommandsOnPath(t, e, []string{"after_pipeline", "task", "epilogue", "on_pass", "commands"}, expectedCommands)
	assertCommandsOnPath(t, e, []string{"after_pipeline", "task", "jobs", "0", "commands"}, expectedCommands)
	assertCommandsOnPath(t, e, []string{"after_pipeline", "task", "prologue", "commands"}, expectedCommands)
	assertCommandsOnPath(t, e, []string{"blocks", "0", "task", "prologue", "commands"}, expectedCommands)
	assertCommandsOnPath(t, e, []string{"blocks", "0", "task", "epilogue", "always", "commands"}, expectedCommands)
	assertCommandsOnPath(t, e, []string{"blocks", "0", "task", "epilogue", "on_fail", "commands"}, expectedCommands)
	assertCommandsOnPath(t, e, []string{"blocks", "0", "task", "epilogue", "on_pass", "commands"}, expectedCommands)
	assertCommandsOnPath(t, e, []string{"blocks", "0", "task", "jobs", "0", "commands"}, expectedCommands)
	assertCommandsOnPath(t, e, []string{"global_job_config", "prologue", "commands"}, expectedCommands)
	assertCommandsOnPath(t, e, []string{"global_job_config", "epilogue", "always", "commands"}, expectedCommands)
	assertCommandsOnPath(t, e, []string{"global_job_config", "epilogue", "on_fail", "commands"}, expectedCommands)
	assertCommandsOnPath(t, e, []string{"global_job_config", "epilogue", "on_pass", "commands"}, expectedCommands)
}

func assertCommandsOnPath(t *testing.T, e *commandsExtractor, path []string, value interface{}) {
	field := e.pipeline.raw.Search(path...).Data()
	assert.Equal(t, value, field)
}