package templates

import (
	"bytes"
	"encoding/json"

	"os"
	"regexp"
	"strings"
	"text/template"

	"github.com/42atomys/sprout"
	"github.com/semaphoreci/spc/pkg/consolelogger"
)

// revive:disable:add-constant

const expressionRegex = `([$%])({{[^(}})]+}})`

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
		prefix, expression := matchGroup[1], matchGroup[2]
		expressionValue, err := applyTemplate(prefix, expression, envValues)

		if err != nil {
			consolelogger.Infof("Unable to parse expression: %s\n", expression)
			consolelogger.Infof("Error: %s\n", err)
			return err
		}

		consolelogger.EmptyLine()
		consolelogger.Infof("Expression: %s\n", matchGroup[0])
		consolelogger.Infof("Expression value: %s\n", expressionValue)

		if matchGroup[0] == strings.TrimSpace(exp.Parsed) {
			consolelogger.Infof("Expression is used standalone (not encapsulated by a string).\n")
			consolelogger.Infof("Its value will be injected verbatim in the YAML file.\n")

			exp.Value = expressionValue
			return nil
		}

		if exprValueAsString, isString := expressionValue.(string); isString {
			consolelogger.Infof("Expression produces a string as an.\n")
			consolelogger.Infof("Its value will be injected verbatim in the YAML file.\n")

			exp.Parsed = strings.Replace(exp.Parsed, matchGroup[0], exprValueAsString, 1)
		} else {
			consolelogger.Infof("Expression does not produce a string, but is not used standalone.\n")
			consolelogger.Infof("Its value will be serialized with JSON and injected in the string.\n")

			exprValueAsJson, err := json.Marshal(expressionValue)
			if err != nil {
				return err
			}
			exp.Parsed = strings.Replace(exp.Parsed, matchGroup[0], string(exprValueAsJson), 1)
		}
	}

	exp.Value = exp.Parsed
	return nil
}

func applyTemplate(prefix, expression string, envVars EnvVars) (interface{}, error) {
	if prefix == "%" {
		trailingEndRegex := regexp.MustCompile(`\s*}}$`)
		expression = trailingEndRegex.ReplaceAllString(expression, "| toJson }}")
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
		err = json.Unmarshal(buff.Bytes(), &output)
		if err != nil {
			return nil, err
		}

		return output, nil

	}

	return buff.String(), nil
}

func templateFuncMap() template.FuncMap {
	dateFuncs := []string{
		// date functions
		// "ago", "date", "dateModify", "dateInZone", "duration", "durationRound",
		// "mustDateModify", "mustToDate", "now", "toDate", "unixEpoch",
		// default functions
		"default", "empty", "coalesce", "all", "any", "compact", "ternary",
		"fromJson", "toJson", "toPrettyJson", "toRawJson", "deepCopy",

		// encoding functions
		"b64enc", "b64dec", "b32enc", "b32dec",
		// data structure functions
		"list", "dict", "get", "set", "unset", "chunk", "mustChunk",
		"hasKey", "pluck", "keys", "pick", "omit", "values", "concat", "dig",
		"merge", "mergeOverwrite", "append", "prepend", "reverse",
		"first", "rest", "last", "initial", "uniq", "without", "has", "slice",
		// regex functions
		"regexMatch", "mustRegexMatch", "regexFindAll",
		"mustRegexFindAll", "regexFind", "mustRegexFind",
		"regexReplaceAll", "mustRegexReplaceAll",
		"regexReplaceAllLiteral", "mustRegexReplaceAllLiteral",
		"regexSplit", "mustRegexSplit", "regexQuoteMeta",
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
