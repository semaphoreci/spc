package pipelines

import (
	"fmt"
	"strconv"

	consolelogger "github.com/semaphoreci/spc/pkg/consolelogger"
	logs "github.com/semaphoreci/spc/pkg/logs"
	when "github.com/semaphoreci/spc/pkg/when"
	whencli "github.com/semaphoreci/spc/pkg/when/whencli"
)

// revive:disable:add-constant

type whenEvaluator struct {
	pipeline *Pipeline

	list    []when.WhenExpression
	results []string
}

func newWhenEvaluator(p *Pipeline) *whenEvaluator {
	return &whenEvaluator{pipeline: p}
}

func (e *whenEvaluator) Run() error {
	var err error

	e.ExtractAll()

	e.displayFound()

	err = e.parse()
	if err != nil {
		return err
	}

	err = e.eval()
	if err != nil {
		return err
	}

	err = e.reduce()
	if err != nil {
		return err
	}

	err = e.updatePipeline()
	if err != nil {
		return err
	}

	return nil
}

func (e *whenEvaluator) updatePipeline() error {
	for index := range e.results {
		err := e.pipeline.UpdateWhenExpression(e.list[index].Path, e.results[index])

		if err != nil {
			return err
		}
	}

	return nil
}

func (e *whenEvaluator) eval() error {
	consolelogger.Infof("Evaluating when expressions.\n")
	consolelogger.EmptyLine()

	for index, condition := range e.list {
		consolelogger.IncrementNesting()
		consolelogger.InfoNumberListLn(index+1, "When Expression: "+condition.Expression)

		err := e.list[index].Eval()
		if err != nil {
			return err
		}

		consolelogger.DecreaseNesting()
		consolelogger.EmptyLine()
	}

	return nil
}

func (e *whenEvaluator) reduce() error {
	expressions := []string{}
	inputs := []whencli.ReduceInputs{}

	for index := range e.list {
		expressions = append(expressions, e.list[index].Expression)
		inputs = append(inputs, e.list[index].ReduceInputs)
	}

	results, err := whencli.Reduce(expressions, inputs)
	if err != nil {
		return err
	}

	e.results = results

	return nil
}

func (e *whenEvaluator) ExtractAll() {
	e.ExtractAutoCancel()
	e.ExtractFailFast()
	e.ExtractFromBlocks()
	e.ExtractFromPromotions()
	e.ExtractFromPriority()
	e.ExtractFromQueue()
}

func (e *whenEvaluator) parse() error {
	expressions := []string{}
	for _, e := range e.list {
		expressions = append(expressions, e.Expression)
	}

	requirments, err := whencli.ListInputs(expressions)
	if err != nil {
		return err
	}

	err = e.verifyParsed(requirments)
	if err != nil {
		return err
	}

	for index := range e.list {
		e.list[index].Requirments = requirments[index].Inputs
	}

	return nil
}

func (e *whenEvaluator) displayFound() {
	consolelogger.Infof("Found when expressions at %d locations.\n", len(e.list))
	consolelogger.EmptyLine()

	for index, condition := range e.list {
		consolelogger.IncrementNesting()
		consolelogger.InfoNumberListLn(index+1, fmt.Sprintf("Location: %+v", condition.Path))
		consolelogger.Infof("File: %s\n", condition.YamlPath)
		consolelogger.Infof("Expression: %s\n", condition.Expression)
		consolelogger.DecreaseNesting()
		consolelogger.EmptyLine()
	}
}

func (e *whenEvaluator) verifyParsed(requirments []whencli.ListInputsResult) error {
	var err error

	for index, r := range requirments {
		if r.Error != "" {
			loc := logs.Location{
				Path: e.list[index].Path,
				File: e.pipeline.yamlPath,
			}

			logError := logs.ErrorInvalidWhenExpression{
				Message:  r.Error,
				Location: loc,
			}

			logs.Log(logError)

			err = &logError
		}
	}

	return err
}

func (e *whenEvaluator) ExtractAutoCancel() {
	e.tryExtractingFromPath([]string{"auto_cancel", "queued", "when"})
	e.tryExtractingFromPath([]string{"auto_cancel", "running", "when"})
}

func (e *whenEvaluator) ExtractFailFast() {
	e.tryExtractingFromPath([]string{"fail_fast", "cancel", "when"})
	e.tryExtractingFromPath([]string{"fail_fast", "stop", "when"})
}

func (e *whenEvaluator) ExtractFromBlocks() {
	for index := range e.pipeline.Blocks() {
		e.tryExtractingFromPath([]string{"blocks", strconv.Itoa(index), "run", "when"})
		e.tryExtractingFromPath([]string{"blocks", strconv.Itoa(index), "skip", "when"})
	}
}

func (e *whenEvaluator) ExtractFromPromotions() {
	for index := range e.pipeline.Promotions() {
		e.tryExtractingFromPath([]string{"promotions", strconv.Itoa(index), "auto_promote", "when"})
	}
}

func (e *whenEvaluator) ExtractFromPriority() {
	for index := range e.pipeline.GlobalPriorityRules() {
		e.tryExtractingFromPath([]string{"global_job_config", "priority", strconv.Itoa(index), "when"})
	}

	for blockIndex, block := range e.pipeline.Blocks() {
		jobs := block.Search("task", "jobs").Children()

		for jobIndex, job := range jobs {
			priority := job.Search("priority").Children()

			for priorityIndex := range priority {
				e.tryExtractingFromPath([]string{
					"blocks",
					strconv.Itoa(blockIndex),
					"task",
					"jobs",
					strconv.Itoa(jobIndex),
					"priority",
					strconv.Itoa(priorityIndex),
					"when",
				})
			}
		}
	}
}

func (e *whenEvaluator) ExtractFromQueue() {
	for index := range e.pipeline.QueueRules() {
		e.tryExtractingFromPath([]string{"queue", strconv.Itoa(index), "when"})
	}
}

func (e *whenEvaluator) tryExtractingFromPath(path []string) {
	if !e.pipeline.PathExists(path) {
		return
	}

	value, ok := e.pipeline.raw.Search(path...).Data().(string)
	if !ok {
		return
	}

	expression := when.WhenExpression{
		Expression: value,
		Path:       path,
		YamlPath:   e.pipeline.yamlPath,
	}

	e.list = append(e.list, expression)
}
