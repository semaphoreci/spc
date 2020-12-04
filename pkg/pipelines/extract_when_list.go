package pipelines

import (
	"strconv"

	when "github.com/semaphoreci/spc/pkg/when"
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
