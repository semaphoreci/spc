package pipelines

import (
	"fmt"
	"strconv"

	consolelogger "github.com/semaphoreci/spc/pkg/consolelogger"
	parameters "github.com/semaphoreci/spc/pkg/parameters"
)

// revive:disable:add-constant

type parametersEvaluator struct {
	pipeline *Pipeline

	list []parameters.ParametersExpression
}

func newParametersEvaluator(p *Pipeline) *parametersEvaluator {
	return &parametersEvaluator{pipeline: p}
}

func (e *parametersEvaluator) Run() error {
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

func (e *parametersEvaluator) ExtractAll() {
	e.ExtractPipelineName()
	e.ExtractFromQueue()
	e.ExtractFromGlobalSecrets()
	e.ExtractFromSecrets()
}

func (e *parametersEvaluator) ExtractPipelineName() {
	e.tryExtractingFromPath([]string{"name"})
}

func (e *parametersEvaluator) ExtractFromSecrets() {
	for blockIndex, block := range e.pipeline.Blocks() {
		secrets := block.Search("task", "secrets").Children()

		for secretIndex := range secrets {
			e.tryExtractingFromPath([]string{
				"blocks",
				strconv.Itoa(blockIndex),
				"task",
				"secrets",
				strconv.Itoa(secretIndex),
				"name",
			})
		}
	}
}

func (e *parametersEvaluator) ExtractFromGlobalSecrets() {
	for index := range e.pipeline.GlobalSecrets() {
		e.tryExtractingFromPath([]string{"global_job_config", "secrets", strconv.Itoa(index), "name"})
	}
}

func (e *parametersEvaluator) ExtractFromQueue() {
	e.tryExtractingFromPath([]string{"queue", "name"})

	for index := range e.pipeline.QueueRules() {
		e.tryExtractingFromPath([]string{"queue", strconv.Itoa(index), "name"})
	}
}

func (e *parametersEvaluator) tryExtractingFromPath(path []string) {
	if !e.pipeline.PathExists(path) {
		return
	}

	value, ok := e.pipeline.raw.Search(path...).Data().(string)
	if !ok {
		return
	}

	if !parameters.ContainsParametersExpression(value) {
		return
	}

	expression := parameters.ParametersExpression{
		Expression: value,
		Path:       path,
		YamlPath:   e.pipeline.yamlPath,
	}

	e.list = append(e.list, expression)
}

func (e *parametersEvaluator) displayFound() {
	consolelogger.Infof("Found parameters expressions at %d locations.\n", len(e.list))
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

func (e *parametersEvaluator) substituteValues() error {
	consolelogger.Infof("Substituting parameters with their values .\n")
	consolelogger.EmptyLine()

	for index, item := range e.list {
		consolelogger.IncrementNesting()
		consolelogger.InfoNumberListLn(index+1, "Parameters Expression: "+item.Expression)

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

func (e *parametersEvaluator) updatePipeline() error {
	for index := range e.list {
		err := e.pipeline.UpdateField(e.list[index].Path, e.list[index].Value)

		if err != nil {
			return err
		}
	}

	return nil
}
