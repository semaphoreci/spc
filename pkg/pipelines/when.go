package pipelines

import (
	"strconv"
)

func (p *Pipeline) ListWhenConditions() *WhenList {
	list := &WhenList{pipeline: p}

	list.AppendIfExists("auto_cancel", "queued", "when")
	list.AppendIfExists("auto_cancel", "running", "when")
	list.AppendIfExists("fail_fast", "cancel", "when")
	list.AppendIfExists("fail_fast", "stop", "when")

	for index, _ := range p.Blocks {
		i := strconv.Itoa(index)

		list.AppendIfExists("blocks", i, "skip", "when")
		list.AppendIfExists("blocks", i, "run", "when")
	}

	for index, _ := range p.Promotions {
		i := strconv.Itoa(index)

		list.AppendIfExists("promotions", i, "autoPromote", "when")
	}

	return list
}

type WhenListElement struct {
	expression string
	path       []string
}

type WhenList struct {
	pipeline *Pipeline
	list     []WhenListElement
}

func (w *WhenList) AppendIfExists(path ...string) {
	value := w.pipeline.Lookup(path)

	if value != nil {
		w.list = append(w.list, WhenListElement{
			expression: value.(string),
			path:       path,
		})
	}
}
