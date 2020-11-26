package when

import (
	"testing"

	assert "github.com/stretchr/testify/assert"
)

func Test__ChangeInPatternMatch(t *testing.T) {
	var matched bool

	matched = changeInPatternMatch("lib/a.txt", "/lib", ".semaphore")
	assert.True(t, matched)

	matched = changeInPatternMatch("lib/package/a.txt", "/lib", ".semaphore")
	assert.True(t, matched)

	matched = changeInPatternMatch("lib/a.txt", "/app", ".semaphore")
	assert.False(t, matched)

	matched = changeInPatternMatch("lib/a.txt", "../lib", ".semaphore")
	assert.True(t, matched)

	matched = changeInPatternMatch("lib/a.txt", "/lib/*.txt", ".semaphore")
	assert.True(t, matched)

	matched = changeInPatternMatch("lib/a.txt", "/lib/**/*.txt", ".semaphore")
	assert.True(t, matched)

	matched = changeInPatternMatch("lib/a.rb", "/lib/**/*.txt", ".semaphore")
	assert.False(t, matched)

	matched = changeInPatternMatch("lib/a.txt", "../lib/**/*.txt", ".semaphore")
	assert.True(t, matched)

	matched = changeInPatternMatch("lib/a.rb", "../lib/**/*.txt", ".semaphore")
	assert.False(t, matched)
}
