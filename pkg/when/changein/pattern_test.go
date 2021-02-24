package changein

import (
	"testing"

	assert "github.com/stretchr/testify/assert"
)

func Test__PatternMatch(t *testing.T) {
	var matched bool

	workdir := ".semaphore"

	matched = patternMatch("lib/a.txt", "/lib", workdir)
	assert.True(t, matched)

	matched = patternMatch("lib/package/a.txt", "/lib", workdir)
	assert.True(t, matched)

	matched = patternMatch("lib/b.txt", "/app", workdir)
	assert.False(t, matched)

	matched = patternMatch("lib/c.txt", "../lib", workdir)
	assert.True(t, matched)

	matched = patternMatch("lib/d.txt", "/lib/*.txt", workdir)
	assert.True(t, matched)

	matched = patternMatch("lib/e.txt", "/lib/**/*.txt", workdir)
	assert.True(t, matched)

	matched = patternMatch("lib/f.rb", "/lib/**/*.txt", workdir)
	assert.False(t, matched)

	matched = patternMatch("lib/g.txt", "../lib/**/*.txt", workdir)
	assert.True(t, matched)

	matched = patternMatch("lib/h.rb", "../lib/**/*.txt", workdir)
	assert.False(t, matched)
	
	matched = patternMatch("library/a.txt", "/lib/", workdir)
	assert.False(t, matched)
}
