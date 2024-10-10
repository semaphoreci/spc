package pipelines

import (
	"fmt"
	"strconv"

	"github.com/Jeffail/gabs/v2"
	consolelogger "github.com/semaphoreci/spc/pkg/consolelogger"
	commands "github.com/semaphoreci/spc/pkg/commands"
)

// revive:disable:add-constant

type commandsExtractor struct {
	pipeline *Pipeline

	files []commands.File
}

func newCommandsExtractor(p *Pipeline) *commandsExtractor {
	return &commandsExtractor{pipeline: p}
}

func (e *commandsExtractor) Run() error {
	var err error

	e.findAll()

	e.displayFound()

	err = e.extractCommands()
	if err != nil {
		return err
	}

	err = e.updatePipeline()
	if err != nil {
		return err
	}

	return nil
}

func (e *commandsExtractor) findAll() {
	e.findCommandsFiles(e.pipeline.raw, []string{})
}

func (e *commandsExtractor) findCommandsFiles(parent *gabs.Container, parentPath []string){
	path := []string{}

	switch parent.Data().(type) {

		case []interface{}:
			for childIndex, child := range parent.Children() {
				path = concatPaths(parentPath, []string{strconv.Itoa(childIndex)})
				e.findCommandsFiles(child, path)
			}

		case map[string]interface{}:
			for key, child := range parent.ChildrenMap() {
				if key == "commands_file" {
					e.gatherCommandsFileData(child, parentPath)
				} else {
					path = concatPaths(parentPath, []string{key})
					e.findCommandsFiles(child, path)
				}
			}
	}
}

func (e *commandsExtractor) gatherCommandsFileData(element *gabs.Container, path []string) {
	file := commands.File{
		FilePath:   element.Data().(string),
		ParentPath: path,
		YamlPath:   e.pipeline.yamlPath,
		Commands:   []string{},
	}

	e.files = append(e.files, file)
}

func (e *commandsExtractor) displayFound() {
	consolelogger.Infof("Found commands_file fields at %d locations.\n", len(e.files))
	consolelogger.EmptyLine()

	for index, item := range e.files {
		itemPath := concatPaths(item.ParentPath, []string{"commands_file"})

		consolelogger.IncrementNesting()
		consolelogger.InfoNumberListLn(index+1, fmt.Sprintf("Location: %+v", itemPath))
		consolelogger.Infof("File: %s\n", item.YamlPath)
		consolelogger.Infof("The commands_file path: %s\n", item.FilePath)
		consolelogger.DecreaseNesting()
		consolelogger.EmptyLine()
	}
}

func (e *commandsExtractor) extractCommands() error {
	consolelogger.Infof("Extracting commands from commands_files.\n")
	consolelogger.EmptyLine()

	for index, item := range e.files {
		consolelogger.IncrementNesting()
		consolelogger.InfoNumberListLn(index+1, "The commands_file path: "+item.FilePath)

		err := e.files[index].Extract()
		if err != nil {
			return err
		}

		consolelogger.Infof("Extracted %d commands.\n", len(e.files[index].Commands))
		consolelogger.DecreaseNesting()
		consolelogger.EmptyLine()
	}

	return nil
}

func (e *commandsExtractor) updatePipeline() error {
	for _, item  := range e.files {
		
		cmdFilePath := concatPaths(item.ParentPath, []string{"commands_file"})

		err := e.pipeline.raw.Delete(cmdFilePath...)
		
		if err != nil {
			return err
		}

		cmdPath := concatPaths(item.ParentPath, []string{"commands"})

		_, err = e.pipeline.raw.Array(cmdPath...)

		if err != nil {
			return err
		}
		
		for _, command := range item.Commands{
			e.pipeline.raw.ArrayAppend(command, cmdPath...)

			if err != nil {
				return err
			}
		}
	}

	return nil
}