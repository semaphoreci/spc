package pipelines

import (
	"fmt"
	"strconv"

	"github.com/Jeffail/gabs/v2"
	consolelogger "github.com/semaphoreci/spc/pkg/consolelogger"
	templates "github.com/semaphoreci/spc/pkg/templates"
)

// revive:disable:add-constant

type templateEvaluator struct {
	pipeline *Pipeline

	list []templates.Expression
}

func newTemplateEvaluator(p *Pipeline) *templateEvaluator {
	return &templateEvaluator{pipeline: p}
}

func (e *templateEvaluator) Run() error {
	var err error

	e.ExtractAll()

	e.displayFound()

	err = e.substituteValues()
	if err != nil {
		return err
	}

	err = e.updatePipeline()
	if err != nil {
		return err
	}

	return nil
}

func (e *templateEvaluator) ExtractAll() {
	e.ExtractTemplateExpression(e.pipeline.raw, []string{})
}

func (e *templateEvaluator) ExtractTemplateExpression(parent *gabs.Container, parentPath []string){
	path := []string{}

	switch parent.Data().(type) {

		case []interface{}:
			for childIndex, child := range parent.Children() {
				path = concatPaths(parentPath, []string{strconv.Itoa(childIndex)})
				e.ExtractTemplateExpression(child, path)
			}

		case map[string]interface{}:
			for key, child := range parent.ChildrenMap() {
				if key != "commands" {
					path = concatPaths(parentPath, []string{key})
					e.ExtractTemplateExpression(child, path)
				}
			}

		default:
			e.tryExtractingFromPath(parentPath)
	}
}

func (e *templateEvaluator) tryExtractingFromPath(path []string) {
	if !e.pipeline.PathExists(path) {
		return
	}

	value, ok := e.pipeline.raw.Search(path...).Data().(string)
	if !ok {
		return
	}

	if !templates.ContainsExpression(value) {
		return
	}

	expression := templates.Expression{
		Expression: value,
		Path:       path,
		YamlPath:   e.pipeline.yamlPath,
	}

	e.list = append(e.list, expression)
}

func (e *templateEvaluator) displayFound() {
	consolelogger.Infof("Found template expressions at %d locations.\n", len(e.list))
	consolelogger.EmptyLine()

	for index, item := range e.list {
		consolelogger.IncrementNesting()
		consolelogger.InfoNumberListLn(index+1, fmt.Sprintf("Location: %+v", item.Path))
		consolelogger.Infof("File: %s\n", item.YamlPath)
		consolelogger.Infof("Expression: %s\n", item.Expression)
		consolelogger.DecreaseNesting()
		consolelogger.EmptyLine()
	}
}

func (e *templateEvaluator) substituteValues() error {
	consolelogger.Infof("Substituting templates with their values.\n")
	consolelogger.EmptyLine()

	for index, item := range e.list {
		consolelogger.IncrementNesting()
		consolelogger.InfoNumberListLn(index+1, "Template Expression: "+item.Expression)

		err := e.list[index].Substitute()
		if err != nil {
			return err
		}

		consolelogger.Infof("Result: %s\n", e.list[index].Value)
		consolelogger.DecreaseNesting()
		consolelogger.EmptyLine()
	}

	return nil
}

func (e *templateEvaluator) updatePipeline() error {
	for index := range e.list {
		err := e.pipeline.UpdateField(e.list[index].Path, e.list[index].Value)

		if err != nil {
			return err
		}
	}

	return nil
}

func concatPaths(paths ...[]string) []string {
	if len(paths) == 0 {
		return []string{}
	}

	path := make([]string, 0, len(paths[0]))
	for _, p := range paths {
		path = append(path, p...)
	}

	return path
}
