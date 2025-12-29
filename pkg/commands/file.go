package commands

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type File struct {
	FilePath    string
	ParentPath  []string
	YamlPath    string
	Commands    []string
}

func (f *File) Extract() error {
	// Get the full path to the commands file in the repository
	absoluteFilePath, err := f.getAbsoluteFilePath()
	if err != nil {
		return fmt.Errorf("failed to resolved the file path for file %s, error: %w", absoluteFilePath, err)
	}

	// Open the file
	file, err := os.Open(filepath.Clean((absoluteFilePath)))
	if err != nil {
		return fmt.Errorf("failed to open the commands_file at %s, error: %w", absoluteFilePath, err)
	}
	defer file.Close()

	// Read the file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		f.Commands = append(f.Commands, line)
	}
	
	// Check for scanning errors
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}
	
	// If no commands were read, return an error indicating that the file is empty
	if len(f.Commands) == 0 {
		return fmt.Errorf("the commands_file at location %s is empty", absoluteFilePath)
	}

	return nil
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