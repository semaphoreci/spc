package templates

import (
	"bytes"
	"encoding/json"
	"fmt"

	"os"
	"regexp"
	"strings"
	"text/template"

	"github.com/semaphoreci/spc/pkg/consolelogger"
	"github.com/semaphoreci/spc/pkg/parameters"
)

// revive:disable:add-constant

const expressionRegex = `\$(\{\{[^\}]+\}\})`

var templateFuncMap = template.FuncMap{
	"split": func(sep, orig string) []string {
		return strings.Split(orig, sep)
	},
	"join": func(sep string, arr []string) string {
		return strings.Join(arr, sep)
	},
	"toString": func(value interface{}) string {
		return fmt.Sprintf("\"%v\"", value)
	},
	"toJson": func(value interface{}) string {
		jsonValue, err := json.Marshal(value)
		if err != nil {
			return fmt.Sprintf("%v", value)
		}
		return string(jsonValue)
	},
}

type Expression struct {
	Expression string
	Path       []string
	YamlPath   string
	Parsed     string
	Value      interface{}
}

type EnvVars map[string]string

func ContainsExpression(value string) bool {
	regex := regexp.MustCompile(expressionRegex)
	return regex.MatchString(value)
}

func (exp *Expression) Substitute() error {
	exp.Parsed = strings.TrimSpace(exp.Expression)

	err := exp.substitutePlainParameters()
	if err != nil {
		return err
	}

	if !ContainsExpression(exp.Parsed) {
		// complex expression were not found
		// return the parsed value
		exp.Value = exp.Parsed
		return nil
	}

	consolelogger.EmptyLine()
	consolelogger.Infof("Complex expression found: %s\n", exp.Parsed)

	envValues, err := exp.traverseParameters()
	if err != nil {
		return err
	}

	templateString := strings.Replace(exp.Parsed, "${{", "{{", -1)
	consolelogger.Infof("Resolving template: %s\n", templateString)

	err = exp.substituteExpressions(envValues)
	if err != nil {
		return err
	}

	consolelogger.Infof("Value: %s\n", exp.Value)

	return nil
}

func (exp *Expression) substitutePlainParameters() error {
	if !parameters.ContainsParametersExpression(exp.Parsed) {
		return nil
	}

	parametersExpression := parameters.ParametersExpression{
		Expression: exp.Parsed,
		Path:       exp.Path,
		YamlPath:   exp.YamlPath,
		Value:      "",
	}

	err := parametersExpression.Substitute()
	if err != nil {
		return err
	}

	exp.Parsed = parametersExpression.Value
	return nil
}

func (exp *Expression) traverseParameters() (EnvVars, error) {
	parametersRegex := regexp.MustCompile(`parameters\.([a-zA-Z0-9_]+)`)
	allParameters := parametersRegex.FindAllStringSubmatch(exp.Parsed, -1)
	envValues := make(EnvVars, len(allParameters))

	for _, matchGroup := range allParameters {
		envName := matchGroup[1]
		envValue := os.Getenv(envName)

		consolelogger.Infof("Converting `parameters.%s` to `.%s`\n", envName, envName)

		if envValue == "" {
			consolelogger.Infof("\t** WARNING *** Environment variable %s not found.\n", envName)
			consolelogger.Infof("\tThe name of the environment variable will be used instead.\n")

			envValue = envName
		}

		updateRegex := regexp.MustCompile(`parameters\.` + envName)
		envValues[envName] = envValue

		exp.Parsed = updateRegex.ReplaceAllString(exp.Parsed, "."+envName)
	}

	return envValues, nil
}

func (exp *Expression) substituteExpressions(envValues EnvVars) error {
	consolelogger.EmptyLine()

	expressionRegex := regexp.MustCompile(expressionRegex)
	allExpressions := expressionRegex.FindAllStringSubmatch(exp.Parsed, -1)

	for _, matchGroup := range allExpressions {
		consolelogger.Infof("Resolving the value for: %s\n", matchGroup[1])

		expression := matchGroup[1]
		expressionValue, err := applyTemplate(expression, envValues)

		if err != nil {
			consolelogger.Infof("Unable to parse expression: %s\n", expression)
			consolelogger.Infof("Error: %s\n", err)
			return err
		}

		consolelogger.Infof("Value: %s\n", expressionValue)
		consolelogger.EmptyLine()

		// check if expression value is string
		if expressionValueAsString, ok := expressionValue.(string); ok {
			consolelogger.Infof("Expression value is a string: %s\n", expressionValue)
			exp.Value = strings.Replace(exp.Parsed, matchGroup[0], expressionValueAsString, 1)
		}

		// check if expression cover the whole expression
		if matchGroup[0] == strings.TrimSpace(exp.Parsed) {
			consolelogger.Infof("Expression value is not encapsulated in a string.\n")
			consolelogger.Infof("Its value will be injected literally into YAML.\n")
			consolelogger.Infof("Injected structure: %s\n", expressionValue)
			exp.Value = expressionValue
			return nil
		}

		// otherwise, dump back to JSON and encapsulate
		consolelogger.Infof("Expression value is not encapsulated in a string.\n")
		consolelogger.Infof("Its value will be dumped formatted according with Go fmt.Sprintf function.\n")
		consolelogger.Infof("Injected structure: %v\n", expressionValue)
		exp.Value = strings.Replace(exp.Parsed, matchGroup[0], fmt.Sprintf("%v", expressionValue), 1)
	}

	return nil
}

func applyTemplate(expression string, envVars EnvVars) (interface{}, error) {
	tmpl, err := template.New("expression").Funcs(templateFuncMap).Parse(expression)
	if err != nil {
		return nil, err
	}

	var buff bytes.Buffer
	var output interface{}

	err = tmpl.Execute(&buff, envVars)

	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(buff.Bytes(), &output)
	if err != nil {
		return nil, err
	}

	return output, nil
}
