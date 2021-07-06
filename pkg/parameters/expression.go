package parameters

import (
	"os"
	"regexp"

	consolelogger "github.com/semaphoreci/spc/pkg/consolelogger"
)

// revive:disable:add-constant

func findRegex() *regexp.Regexp {
	return regexp.MustCompile(`\$\{\{\s*parameters\.([a-zA-Z0-9_]+)\s*\}\}`)
}

func updateRegex(envName string) *regexp.Regexp {
	return regexp.MustCompile(`\$\{\{\s*parameters\.` + envName + `\s*\}\}`)
}

type ParametersExpression struct {
	Expression string
	Path       []string
	YamlPath   string
	Value      string
}

func ContainsParametersExpression(value string) bool {
	return findRegex().MatchString(value)
}

func (exp *ParametersExpression) Substitute() error {
	consolelogger.EmptyLine()

	exp.Value = exp.Expression

	allExpressions := findRegex().FindAllStringSubmatch(exp.Expression, -1)

	for _, matchGroup := range allExpressions {
		envName := matchGroup[1]
		consolelogger.Infof("Fetching the value for: %s\n", envName)

		envVal := os.Getenv(envName)
		if envVal == "" {
			consolelogger.Infof("\t** WARNING *** Environment variable %s not found.\n", envName)
			consolelogger.Infof("\tThe name of the environment variable will be used instead.\n")

			envVal = envName
		}

		consolelogger.Infof("Value: %s\n", envVal)
		consolelogger.EmptyLine()

		exp.Value = updateRegex(envName).ReplaceAllString(exp.Value, envVal)
	}

	return nil
}
