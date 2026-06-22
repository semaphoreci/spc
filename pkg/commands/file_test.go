package commands

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	assert "github.com/stretchr/testify/assert"
)

const (
	testDirPerm  = 0o755
	testFilePerm = 0o644
)

func Test__Extract(t *testing.T) {
	// If commands file does not exist, it returns the error
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

	// If commands file is empty, it returns the error
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

// When the commands_file is not present in the working tree (e.g. a sparse
// checkout that only materializes the pipeline directory), its content is read
// from Git instead, so references outside the sparse paths keep working.
func Test__ExtractFromGitWhenMissingFromWorkingTree(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not available")
	}

	repoDir := t.TempDir()

	runGit := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = repoDir
		out, err := cmd.CombinedOutput()
		assert.NoError(t, err, string(out))
	}

	runGit("init", "-q")
	runGit("config", "user.email", "test@example.com")
	runGit("config", "user.name", "test")

	// The commands_file lives outside the pipeline directory.
	scriptsDir := filepath.Join(repoDir, "scripts")
	assert.NoError(t, os.MkdirAll(scriptsDir, testDirPerm))
	scriptPath := filepath.Join(scriptsDir, "build.sh")
	assert.NoError(t, os.WriteFile(scriptPath, []byte("echo a\necho b\necho c\n"), testFilePerm))

	runGit("add", ".")
	runGit("commit", "-q", "-m", "init")

	// Simulate a sparse checkout: the file exists in Git but not on disk.
	assert.NoError(t, os.Remove(scriptPath))

	originalWd, err := os.Getwd()
	assert.NoError(t, err)
	assert.NoError(t, os.Chdir(repoDir))
	defer func() { _ = os.Chdir(originalWd) }()

	expectedCommands := []string{"echo a", "echo b", "echo c"}

	// Absolute path (from repository root).
	absFile := File{
		FilePath:   "/scripts/build.sh",
		ParentPath: []string{},
		YamlPath:   ".semaphore/semaphore.yml",
		Commands:   []string{},
	}
	assert.NoError(t, absFile.Extract())
	assert.Equal(t, expectedCommands, absFile.Commands)

	// Relative path (relative to the pipeline YAML directory).
	relFile := File{
		FilePath:   "../scripts/build.sh",
		ParentPath: []string{},
		YamlPath:   ".semaphore/semaphore.yml",
		Commands:   []string{},
	}
	assert.NoError(t, relFile.Extract())
	assert.Equal(t, expectedCommands, relFile.Commands)
}
