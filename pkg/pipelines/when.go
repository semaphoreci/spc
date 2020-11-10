package pipelines

import "strconv"

func (p *Pipeline) ListWhenConditions() *WhenList {
	list := &WhenList{}

	list.AppendIfExists(p, []string{"AutoCancel", "Queued", "When"})
	list.AppendIfExists(p, []string{"AutoCancel", "Running", "When"})
	list.AppendIfExists(p, []string{"FailFast", "Cancel", "When"})
	list.AppendIfExists(p, []string{"FailFast", "Stop", "When"})

	for index, _ := range p.Blocks {
		i := strconv.Itoa(index)

		list.AppendIfExists(p, []string{"blocks", i, "Skip", "When"})
		list.AppendIfExists(p, []string{"blocks", i, "Run", "When"})
	}

	for index, _ := range p.Promotions {
		i := strconv.Itoa(index)

		list.AppendIfExists(p, []string{"promotions", i, "AutoPromote", "When"})
	}

	return list
}

type WhenListElement struct {
	expression string
	path       []string
}

type WhenList struct {
	list []WhenListElement
}

func (w *WhenList) AppendIfExists(pipeline *Pipeline, path []string) {
	value := pipeline.Lookup(path)

	if value != nil {
		w.list = append(w.list, WhenListElement{
			expression: value.(string),
			path:       path,
		})
	}
}
