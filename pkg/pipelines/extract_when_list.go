package pipelines

import (
	"strconv"

	logs "github.com/semaphoreci/spc/pkg/logs"
	when "github.com/semaphoreci/spc/pkg/when"
	whencli "github.com/semaphoreci/spc/pkg/when/whencli"
)

// revive:disable:add-constant

type whenExtractor struct {
	pipeline *Pipeline
	list     []when.WhenExpression
}

func (e *whenExtractor) ExtractAll() {
	e.ExtractAutoCancel()
	e.ExtractFailFast()
	e.ExtractFromBlocks()
	e.ExtractFromPromotions()
	e.ExtractFromPriority()
	e.ExtractFromQueue()
}

func (e *whenExtractor) Parse() ([]when.WhenExpression, error) {
	expressions := []string{}
	for _, e := range e.list {
		expressions = append(expressions, e.Expression)
	}

	res := []when.WhenExpression{}

	requirments, err := whencli.ListInputs(expressions)
	if err != nil {
		return res, err
	}

	err = e.verifyParsed(requirments)
	if err != nil {
		return res, err
	}

	for index := range e.list {
		e.list[index].Requirments = requirments[index].Inputs
	}

	return e.list, nil
}

func (e *whenExtractor) verifyParsed(requirments []whencli.ListInputsResult) error {
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

func (e *whenExtractor) ExtractAutoCancel() {
	e.tryExtractingFromPath([]string{"auto_cancel", "queued", "when"})
	e.tryExtractingFromPath([]string{"auto_cancel", "running", "when"})
}

func (e *whenExtractor) ExtractFailFast() {
	e.tryExtractingFromPath([]string{"fail_fast", "cancel", "when"})
	e.tryExtractingFromPath([]string{"fail_fast", "stop", "when"})
}

func (e *whenExtractor) ExtractFromBlocks() {
	for index := range e.pipeline.Blocks() {
		e.tryExtractingFromPath([]string{"blocks", strconv.Itoa(index), "run", "when"})
		e.tryExtractingFromPath([]string{"blocks", strconv.Itoa(index), "skip", "when"})
	}
}

func (e *whenExtractor) ExtractFromPromotions() {
	for index := range e.pipeline.Promotions() {
		e.tryExtractingFromPath([]string{"promotions", strconv.Itoa(index), "auto_promote", "when"})
	}
}

func (e *whenExtractor) ExtractFromPriority() {
	for index := range e.pipeline.PriorityRules() {
		e.tryExtractingFromPath([]string{"priority", strconv.Itoa(index), "when"})
	}
}

func (e *whenExtractor) ExtractFromQueue() {
	for index := range e.pipeline.QueueRules() {
		e.tryExtractingFromPath([]string{"queue", strconv.Itoa(index), "when"})
	}
}

func (e *whenExtractor) tryExtractingFromPath(path []string) {
	if !e.pipeline.PathExists(path) {
		return
	}

	expression := when.WhenExpression{
		Expression: e.pipeline.GetStringFromPath(path),
		Path:       path,
		YamlPath:   e.pipeline.yamlPath,
	}

	e.list = append(e.list, expression)
}
