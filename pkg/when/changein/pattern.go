package changein

import (
	"path"
	"strings"

	doublestar "github.com/bmatcuk/doublestar/v2"
)

func patternMatch(diffLine, pattern, workDir string) bool {
	if pattern[0] != '/' {
		pattern = path.Join("/", workDir, pattern)
	}

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
