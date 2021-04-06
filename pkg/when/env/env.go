package env

import (
	"fmt"
	"os"

	gabs "github.com/Jeffail/gabs/v2"
	consolelogger "github.com/semaphoreci/spc/pkg/consolelogger"
)

type Function struct {
	VarName string
}

func Parse(ast *gabs.Container) (*Function, error) {
	firstArg := ast.Search("params", "0")

	if !firstArg.Exists() {
		return nil, fmt.Errorf("variable name not found in env function")
	}

	varName, ok := firstArg.Data().(string)
	if !ok {
		return nil, fmt.Errorf("invalid variable name in env function")
	}

	return &Function{VarName: varName}, nil
}

func Eval(fun *Function) (string, error) {
	value := os.Getenv(fun.VarName)

	consolelogger.Infof("Result: '%+v'\n", value)

	return value, nil
}
