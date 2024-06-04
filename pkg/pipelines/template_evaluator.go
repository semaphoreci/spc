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
	e.ExtractPipelineName()
	e.ExtractFromQueue()
	e.ExtractFromSecrets()
	e.ExtractFromBlockNames()
	e.ExtractFromJobNames()
	e.ExtractFromAgents()
	e.ExtractFromJobMatrices()
	e.ExtractFromParallelisms()
}

func (e *templateEvaluator) ExtractPipelineName() {
	e.tryExtractingFromPath([]string{"name"})
}

func (e *templateEvaluator) ExtractFromQueue() {
	e.tryExtractingFromPath([]string{"queue", "name"})

	for index := range e.pipeline.QueueRules() {
		e.tryExtractingFromPath([]string{"queue", strconv.Itoa(index), "name"})
	}
}

func (e *templateEvaluator) ExtractFromSecrets() {
	e.ExtractFromGlobalSecrets()
	e.ExtractFromBlockSecrets()
	e.ExtractFromAfterPipelineSecrets()
}

func (e *templateEvaluator) ExtractFromGlobalSecrets() {
	e.extractFromSecretsAt(e.pipeline.GlobalJobConfig(), []string{"global_job_config"})
}

func (e *templateEvaluator) ExtractFromBlockSecrets() {
	for blockIndex, block := range e.pipeline.Blocks() {
		e.extractFromSecretsAt(block.Search("task"), []string{"blocks", strconv.Itoa(blockIndex), "task"})
	}
}

func (e *templateEvaluator) ExtractFromAfterPipelineSecrets() {
	e.extractFromSecretsAt(e.pipeline.AfterPipelineTask(), []string{"after_pipeline", "task"})
}

func (e *templateEvaluator) ExtractFromJobMatrices() {
	e.ExtractFromBlockJobMatrices()
	e.ExtractFromAfterPipelineJobMatrices()
}

func (e *templateEvaluator) ExtractFromBlockJobMatrices() {
	for blockIndex, block := range e.pipeline.Blocks() {
		blockTaskPath := []string{"blocks", strconv.Itoa(blockIndex), "task"}

		for jobIndex, job := range block.Search("task", "jobs").Children() {
			jobPath := concatPaths(blockTaskPath, []string{"jobs", strconv.Itoa(jobIndex)})
			e.extractFromJobMatricesAt(job, jobPath)
		}
	}
}

func (e *templateEvaluator) ExtractFromAfterPipelineJobMatrices() {
	afterPipelineTask := e.pipeline.AfterPipelineTask()

	for jobIndex, job := range afterPipelineTask.Search("jobs").Children() {
		jobPath := []string{"after_pipeline", "task", "jobs", strconv.Itoa(jobIndex)}
		e.extractFromJobMatricesAt(job, jobPath)
	}
}

func (e *templateEvaluator) ExtractFromParallelisms() {
	e.ExtractFromBlockJobParallelisms()
	e.ExtractFromAfterPipelineParallelisms()
}

func (e *templateEvaluator) ExtractFromBlockJobParallelisms() {
	for blockIndex, block := range e.pipeline.Blocks() {
		blockTaskPath := []string{"blocks", strconv.Itoa(blockIndex), "task"}

		for jobIndex := range block.Search("task", "jobs").Children() {
			jobPath := concatPaths(blockTaskPath, []string{"jobs", strconv.Itoa(jobIndex)})
			e.tryExtractingFromPath(jobPath, []string{"parallelism"})
		}
	}
}

func (e *templateEvaluator) ExtractFromAfterPipelineParallelisms() {
	afterPipelineTask := e.pipeline.AfterPipelineTask()

	for jobIndex := range afterPipelineTask.Search("jobs").Children() {
		jobPath := []string{"after_pipeline", "task", "jobs", strconv.Itoa(jobIndex)}
		e.tryExtractingFromPath(jobPath, []string{"parallelism"})
	}
}

func (e *templateEvaluator) ExtractFromAgents() {
	e.ExtractFromTopLevelAgent()
	e.ExtractFromBlockAgents()
	e.ExtractFromAfterPipelineAgents()
}

func (e *templateEvaluator) ExtractFromTopLevelAgent() {
	e.extractFromAgentAt(e.pipeline.Agent(), []string{"agent"})
}

func (e *templateEvaluator) ExtractFromBlockAgents() {
	for blockIndex, block := range e.pipeline.Blocks() {
		agent := block.Search("task", "agent")
		agentPath := []string{"blocks", strconv.Itoa(blockIndex), "task", "agent"}
		e.extractFromAgentAt(agent, agentPath)
	}
}

func (e *templateEvaluator) ExtractFromAfterPipelineAgents() {
	afterPipelineTask := e.pipeline.AfterPipelineTask()
	if afterPipelineTask == nil {
		return
	}

	e.extractFromAgentAt(afterPipelineTask, []string{"after_pipeline", "task"})
}

func (e *templateEvaluator) ExtractFromBlockNames() {
	for blockIndex := range e.pipeline.Blocks() {
		e.tryExtractingFromPath([]string{"blocks", strconv.Itoa(blockIndex), "name"})
	}
}

func (e *templateEvaluator) ExtractFromJobNames() {
	e.ExtractFromBlockJobNames()
	e.ExtractFromAfterPipelineJobNames()
}

func (e *templateEvaluator) ExtractFromBlockJobNames() {
	for blockIndex, block := range e.pipeline.Blocks() {
		blockTaskPath := []string{"blocks", strconv.Itoa(blockIndex), "task"}
		for jobIndex := range block.Search("task", "jobs").Children() {
			e.tryExtractingFromPath(blockTaskPath, []string{"jobs", strconv.Itoa(jobIndex), "name"})
		}
	}
}

func (e *templateEvaluator) ExtractFromAfterPipelineJobNames() {
	afterPipelineTask := e.pipeline.AfterPipelineTask()

	for jobIndex := range afterPipelineTask.Search("jobs").Children() {
		e.tryExtractingFromPath([]string{"after_pipeline", "task", "jobs", strconv.Itoa(jobIndex), "name"})
	}
}

func (e *templateEvaluator) extractFromAgentAt(agent *gabs.Container, agentPath []string) {
	e.tryExtractingFromPath(agentPath, []string{"machine", "type"})
	e.tryExtractingFromPath(agentPath, []string{"machine", "os_image"})

	for containerIndex, container := range agent.Search("containers").Children() {
		containerPath := concatPaths(agentPath, []string{"containers", strconv.Itoa(containerIndex)})
		e.extractFromContainerAt(container, containerPath)
	}
}

func (e *templateEvaluator) extractFromContainerAt(container *gabs.Container, containerPath []string) {
	e.tryExtractingFromPath(containerPath, []string{"name"})
	e.tryExtractingFromPath(containerPath, []string{"image"})
	e.extractFromSecretsAt(container, containerPath)
}

func (e *templateEvaluator) extractFromSecretsAt(parent *gabs.Container, parentPath []string) {
	for secretIndex := range parent.Search("secrets").Children() {
		e.tryExtractingFromPath(parentPath, []string{"secrets", strconv.Itoa(secretIndex), "name"})
	}
}

func (e *templateEvaluator) extractFromJobMatricesAt(parent *gabs.Container, parentPath []string) {
	for matrixIndex := range parent.Search("matrix").Children() {
		e.tryExtractingFromPath(parentPath, []string{"matrix", strconv.Itoa(matrixIndex), "env_var"})
		e.tryExtractingFromPath(parentPath, []string{"matrix", strconv.Itoa(matrixIndex), "values"})
	}
}

func (e *templateEvaluator) tryExtractingFromPath(paths ...[]string) {
	path := concatPaths(paths...)
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
