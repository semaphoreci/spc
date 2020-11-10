package pipelines

import (
	"reflect"
	"strconv"
)

//
// Looks up a value in the pipeline by following a simple JSON path.
//
func (p *Pipeline) Lookup(path []string) (res interface{}) {
	defer func() {
		if r := recover(); r != nil {
			res = nil
		}
	}()

	return getNested(*p, path)
}

func getNested(obj interface{}, path []string) interface{} {
	res := obj

	for _, p := range path {
		res = get(res, p)
	}

	return res
}

func get(obj interface{}, name string) interface{} {
	v := reflect.ValueOf(obj)

	if index, err := strconv.Atoi(name); err == nil {
		return v.Index(index)
	}

	return reflect.Indirect(v).FieldByName(name)
}
