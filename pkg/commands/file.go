package commands

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
)

type File struct {
	FilePath    string
	ParentPath  []string
	YamlPath    string
	Commands    []string
}

func (f *File) Extract() error {
	// Resolve FilePath relative to YamlPath
	absoluteFilePath := filepath.Join(filepath.Dir(f.YamlPath), f.FilePath)

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