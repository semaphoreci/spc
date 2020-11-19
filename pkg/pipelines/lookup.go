package pipelines

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/iancoleman/strcase"
)

//
// Looks up a value in the pipeline by following a simple JSON path.
//
func (p *Pipeline) Lookup(path []string) interface{} {
	snakeCasePath := []string{}

	for _, p := range path {
		snakeCasePath = append(snakeCasePath, strcase.ToCamel(p))
	}

	return getNested(p, snakeCasePath)
}

func getNested(obj interface{}, path []string) interface{} {
	res := obj

	for _, p := range path {
		res = get(res, p)
		if res == nil {
			return res
		}
	}

	return res
}

func get(obj interface{}, name string) interface{} {
	v := reflect.ValueOf(obj)

	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			return nil
		}

		return get(v.Elem().Interface(), name)

	case reflect.Struct:
		return reflect.Indirect(v).FieldByName(name).Interface()

	case reflect.Slice:
		index, err := strconv.Atoi(name)

		if err != nil {
			return nil
		}

		return reflect.Indirect(v).Index(index).Interface()
	}

	fmt.Println(v.Kind())
	panic("can't get path from object")
}

func (p *Pipeline) ChangeWhenExpression(path []string, value string) {
	snakeCasePath := []string{}

	for _, p := range path {
		snakeCasePath = append(snakeCasePath, strcase.ToCamel(p))
	}

	v := reflect.ValueOf(p)

	for _, p := range snakeCasePath {
		index, err := strconv.Atoi(p)

		if err == nil {
			v = v.Index(index)
		} else {
			v = v.FieldByName(p)
		}
	}

	v.SetString(value)
}
