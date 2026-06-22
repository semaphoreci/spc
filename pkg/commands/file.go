package commands

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type File struct {
	FilePath   string
	ParentPath []string
	YamlPath   string
	Commands   []string
}

func (f *File) Extract() error {
	// Get the full path to the commands file in the repository
	absoluteFilePath, err := f.getAbsoluteFilePath()
	if err != nil {
		return fmt.Errorf("failed to resolved the file path for file %s, error: %w", absoluteFilePath, err)
	}

	commands, err := f.readCommands(absoluteFilePath)
	if err != nil {
		return err
	}

	// If no commands were read, return an error indicating that the file is empty
	if len(commands) == 0 {
		return fmt.Errorf("the commands_file at location %s is empty", absoluteFilePath)
	}

	f.Commands = commands
	return nil
}

// readCommands reads the commands_file from the working tree. When the file is
// not present on disk - which happens when the repository was checked out with
// a sparse working tree (for example the pipeline initialization job, which
// only materializes the pipeline directory) - it falls back to reading the
// file content directly from Git. For partial (blobless) clones this fetches
// the blob on demand, so commands_file references outside the sparse paths
// keep working.
func (f *File) readCommands(absoluteFilePath string) ([]string, error) {
	file, err := os.Open(filepath.Clean(absoluteFilePath))
	if err != nil {
		if os.IsNotExist(err) {
			if commands, gitErr := f.readCommandsFromGit(); gitErr == nil {
				return commands, nil
			}
		}

		return nil, fmt.Errorf("failed to open the commands_file at %s, error: %w", absoluteFilePath, err)
	}
	defer file.Close()

	return readLines(file)
}

// readCommandsFromGit reads the commands_file content from the checked-out
// revision (HEAD) using `git show`, without requiring the file to be present
// in the working tree.
func (f *File) readCommandsFromGit() ([]string, error) {
	relPath, err := f.repoRelativePath()
	if err != nil {
		return nil, err
	}

	// #nosec G204 - relPath is derived from the pipeline definition that is
	// checked out in the repository, addressed via an explicit "HEAD:" revision.
	cmd := exec.Command("git", "show", "HEAD:"+relPath)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf(
			"failed to read commands_file %s from git: %s: %w",
			relPath, strings.TrimSpace(stderr.String()), err,
		)
	}

	return readLines(bytes.NewReader(stdout.Bytes()))
}

func readLines(reader io.Reader) ([]string, error) {
	var lines []string

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return lines, nil
}

func (f *File) getAbsoluteFilePath() (string, error) {
	// Get the path to git repository root on filesystem
	workingDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// If file path starts with '/' it is an absolute path from the root of git repository
	if strings.HasPrefix(f.FilePath, "/") {
		// Join the git repository root with the file path
		return filepath.Join(workingDir, f.FilePath), nil
	} else {
		// Else, join the git repository root with the directory path for YML file
		ymlDirPath := filepath.Join(workingDir, filepath.Dir(f.YamlPath))
		// and then File path is relative to that YML directory path
		return filepath.Join(ymlDirPath, f.FilePath), nil
	}
}

// repoRelativePath returns the commands_file path relative to the repository
// root, using forward slashes, so it can be addressed in Git (e.g. via
// `git show HEAD:<path>`). It mirrors the resolution rules of
// getAbsoluteFilePath: a leading '/' means a path from the repository root,
// otherwise the path is relative to the directory of the pipeline YAML.
func (f *File) repoRelativePath() (string, error) {
	var relPath string

	if strings.HasPrefix(f.FilePath, "/") {
		relPath = strings.TrimPrefix(f.FilePath, "/")
	} else {
		relPath = filepath.Join(filepath.Dir(f.YamlPath), f.FilePath)
	}

	relPath = filepath.ToSlash(filepath.Clean(relPath))

	if relPath == ".." || strings.HasPrefix(relPath, "../") {
		return "", fmt.Errorf("commands_file path %q escapes the repository root", relPath)
	}

	return relPath, nil
}
