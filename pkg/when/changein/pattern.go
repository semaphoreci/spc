package changein

import (
	"path"
	"strings"

	doublestar "github.com/bmatcuk/doublestar/v2"
)

func patternMatch(diffLine, pattern, workDir string) bool {
	pattern = cleanPattern(workDir, pattern)
	diffLine = path.Clean("/" + diffLine)

	if !strings.Contains(pattern, "*") {
		return strings.HasPrefix(diffLine, pattern)
	}

	ok, err := doublestar.Match(pattern, diffLine)
	if err != nil {
		panic(err)
	}

	return ok
}

func cleanPattern(workDir, pattern string) string {
	var cleanPattern string

	if pattern[0] != '/' {
		cleanPattern = path.Join("/", workDir, pattern)
	} else {
		cleanPattern = path.Clean(pattern)
	}

	if cleanPattern[len(pattern)-1] != '/' && pattern[len(pattern)-1] == '/' {
		cleanPattern += "/"
	}

	return cleanPattern
}
