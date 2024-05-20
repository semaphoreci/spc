package templates

import (
	"bytes"
	"encoding/json"
	"errors"

	"os"
	"regexp"
	"strings"
	"text/template"

	"github.com/42atomys/sprout"
	"github.com/semaphoreci/spc/pkg/consolelogger"
)

// revive:disable:add-constant

const expressionRegex = `([$%])({{([^(}})]+)}})`

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

func ContainsNextedExpressions(value string) bool {
	regex := regexp.MustCompile(expressionRegex)
	matches := regex.FindAllStringSubmatch(value, -1)

	for _, matchGroup := range matches {
		if regex.MatchString(matchGroup[3]) {
			return true
		}
	}
	return false
}

func (exp *Expression) Substitute() error {
	exp.Parsed = strings.TrimSpace(exp.Expression)

	if !ContainsExpression(exp.Parsed) {
		// complex expression were not found
		// return the parsed value
		exp.Value = exp.Parsed
		return nil
	}

	if ContainsNextedExpressions(exp.Parsed) {
		return errors.New("nested expressions are not supported")
	}

	envValues, err := exp.traverseParameters()
	if err != nil {
		return err
	}

	err = exp.substituteExpressions(envValues)
	if err != nil {
		return err
	}

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
		regexMatch, prefix, expression := matchGroup[0], matchGroup[1], matchGroup[2]
		expressionValue, err := applyTemplate(prefix, expression, envValues)

		if err != nil {
			consolelogger.Infof("Unable to parse expression: %s\n", expression)
			consolelogger.Infof("Error: %s\n", err)
			return err
		}

		consolelogger.Infof("Partial expression: %s\n", matchGroup[0])
		consolelogger.Infof("Partial expression value: %s\n", expressionValue)

		if exp.fullyMatchesRegex(regexMatch) {
			return exp.assignJSONObjectToValue(expressionValue)
		}

		err = exp.replaceValueInParsedString(expressionValue, regexMatch)
		if err != nil {
			return err
		}
	}

	exp.Value = exp.Parsed
	return nil
}

func (exp *Expression) fullyMatchesRegex(regexMatch string) bool {
	return strings.TrimSpace(exp.Parsed) == regexMatch
}

func (exp *Expression) assignJSONObjectToValue(expressionValue interface{}) error {
	consolelogger.Infof("Expression is used standalone (not encapsulated by a string).\n")
	consolelogger.Infof("Its value will be injected verbatim in the YAML file.\n")
	consolelogger.EmptyLine()

	exp.Value = expressionValue
	return nil
}

func (exp *Expression) replaceValueInParsedString(expressionValue interface{}, regexMatch string) error {
	if exprValueAsString, isString := expressionValue.(string); isString {
		consolelogger.Infof("Expression produces a string as a result of an expression.\n")
		consolelogger.Infof("Its value will be injected verbatim in the YAML file.\n")
		consolelogger.EmptyLine()

		exp.Parsed = strings.Replace(exp.Parsed, regexMatch, exprValueAsString, 1)
		return nil
	}

	consolelogger.Infof("Expression does not produce a string, but is not used standalone.\n")
	consolelogger.Infof("Its value will be serialized with JSON and injected in the string.\n")
	consolelogger.EmptyLine()

	exprValueAsJSON, err := json.Marshal(expressionValue)
	if err != nil {
		return err
	}

	exp.Parsed = strings.Replace(exp.Parsed, regexMatch, string(exprValueAsJSON), 1)
	return nil
}

func applyTemplate(prefix, expression string, envVars EnvVars) (interface{}, error) {
	if prefix == "%" {
		trailingEndRegex := regexp.MustCompile(`\s*}}$`)
		expression = trailingEndRegex.ReplaceAllString(expression, " | toJson }}")
	}

	tmpl, err := template.New("expression").Funcs(templateFuncMap()).Parse(expression)
	if err != nil {
		return nil, err
	}

	var buff bytes.Buffer
	var output interface{}

	err = tmpl.Execute(&buff, envVars)

	if err != nil {
		return nil, err
	}

	if prefix == "%" {
		decoder := json.NewDecoder(&buff)
		decoder.UseNumber()
		err = decoder.Decode(&output)
		if err != nil {
			return nil, err
		}

		return output, nil

	}

	return buff.String(), nil
}

func templateFuncMap() template.FuncMap {
	dateFuncs := []string{
		// default functions
		"default", "empty", "coalesce", "all", "any", "compact", "ternary",
		"fromJson", "toJson", "toPrettyJson", "toRawJson", "deepCopy",
		// encoding functions
		"b64enc", "b64dec", "b32enc", "b32dec",
		// data structure functions
		"list", "dict", "get", "set", "unset", "chunk",
		"hasKey", "pluck", "keys", "pick", "omit", "values", "concat", "dig",
		"merge", "mergeOverwrite", "append", "prepend", "reverse",
		"first", "rest", "last", "initial", "uniq", "without", "has", "slice",
		// regex functions
		"regexMatch", "regexFindAll", "regexFind", "regexReplaceAll",
		"regexReplaceAllLiteral", "regexSplit", "regexQuoteMeta",
		// string functions
		"ellipsis", "ellipsisBoth", "trunc", "trim", "upper", "lower",
		"title", "untitle", "substr", "repeat", "join", "sortAlpha",
		"trimAll", "trimSuffix", "trimPrefix", "nospace", "initials",
		"randAlphaNum", "randAlpha", "randAscii", "randNumeric",
		"swapcase", "shuffle", "snakecase", "camelcase", "kebabcase",
		"wrap", "wrapWith", "contains", "hasPrefix", "hasSuffix",
		"quote", "squote", "cat", "indent", "nindent", "replace",
		"plural", "sha1sum", "sha256sum", "adler32sum", "toString",
		"int64", "int", "float64", "seq", "toDecimal", "until", "untilStep",
		"split", "splitList", "splitn", "toStrings",
		// arithmetic functions
		"add1", "add", "sub", "div", "mod", "mul", "randInt",
		"add1f", "addf", "subf", "divf", "mulf",
		"max", "min", "maxf", "minf", "ceil", "floor", "round",
	}

	genericFuncMap := sprout.GenericFuncMap()
	funcMap := make(template.FuncMap, len(dateFuncs))

	for _, dateFunc := range dateFuncs {
		if genericFuncMap[dateFunc] != nil {
			funcMap[dateFunc] = genericFuncMap[dateFunc]
		}
	}

	return funcMap
}
