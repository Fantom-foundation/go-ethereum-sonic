package tests

import (
	"github.com/ethereum/go-ethereum/params"
	"testing"
)

type TestMatcher struct {
	testMatcher
}

// Slow adds expected slow tests matching the pattern.
func (tm *TestMatcher) Slow(pattern string) {
	tm.testMatcher.slow(pattern)
}

// SkipLoad skips JSON loading of tests matching the pattern.
func (tm *TestMatcher) SkipLoad(pattern string) {
	tm.testMatcher.skipLoad(pattern)
}

// Fails adds an expected failure for tests matching the pattern.
//
//nolint:unused
func (tm *TestMatcher) Fails(pattern string, reason string) {
	tm.testMatcher.fails(pattern, reason)
}

func (tm *TestMatcher) Runonly(pattern string) {
	tm.testMatcher.runonly(pattern)
}

// Config defines chain config for tests matching the pattern.
func (tm *TestMatcher) Config(pattern string, cfg params.ChainConfig) {
	tm.testMatcher.config(pattern, cfg)
}

// FindSkip matches name against test skip patterns.
func (tm *TestMatcher) FindSkip(name string) (reason string, skipload bool) {
	return tm.testMatcher.findSkip(name)
}

// FindConfig returns the chain config matching defined patterns.
func (tm *TestMatcher) FindConfig(t *testing.T) *params.ChainConfig {
	return tm.testMatcher.findConfig(t)
}

// CheckFailure checks whether a failure is expected.
func (tm *TestMatcher) CheckFailure(t *testing.T, err error) error {
	return tm.testMatcher.checkFailure(t, err)
}

// Walk invokes its runTest argument for all subtests in the given directory.
//
// runTest should be a function of type func(t *testing.T, name string, x <TestType>),
// where TestType is the type of the test contained in test files.
func (tm *TestMatcher) Walk(t *testing.T, dir string, runTest interface{}) {
	tm.testMatcher.walk(t, dir, runTest)
}

func (tm *TestMatcher) RunTestFile(t *testing.T, path, name string, runTest interface{}) {
	tm.testMatcher.runTestFile(t, path, name, runTest)
}
