package changein

import (
	"testing"

	assert "github.com/stretchr/testify/assert"
)

func Test__ChangeInPatternMatch(t *testing.T) {
	var matched bool

	workdir := ".semaphore"

	matched = changeInPatternMatch("lib/a.txt", "/lib", workdir)
	assert.True(t, matched)

	matched = changeInPatternMatch("lib/package/a.txt", "/lib", workdir)
	assert.True(t, matched)

	matched = changeInPatternMatch("lib/b.txt", "/app", workdir)
	assert.False(t, matched)

	matched = changeInPatternMatch("lib/c.txt", "../lib", workdir)
	assert.True(t, matched)

	matched = changeInPatternMatch("lib/d.txt", "/lib/*.txt", workdir)
	assert.True(t, matched)

	matched = changeInPatternMatch("lib/e.txt", "/lib/**/*.txt", workdir)
	assert.True(t, matched)

	matched = changeInPatternMatch("lib/f.rb", "/lib/**/*.txt", workdir)
	assert.False(t, matched)

	matched = changeInPatternMatch("lib/g.txt", "../lib/**/*.txt", workdir)
	assert.True(t, matched)

	matched = changeInPatternMatch("lib/h.rb", "../lib/**/*.txt", workdir)
	assert.False(t, matched)
}
