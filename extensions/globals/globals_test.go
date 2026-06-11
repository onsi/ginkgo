package globals_test

import (
	"testing"

	"github.com/onsi/ginkgo/v2/extensions/globals"
	"github.com/onsi/ginkgo/v2/internal/global"
)

func TestGlobals(t *testing.T) {
	global.InitializeGlobals()
	oldSuite := global.Suite
	if oldSuite == nil {
		t.Error("global.Suite was nil")
	}

	globals.Reset()
	newSuite := global.Suite
	if newSuite == nil {
		t.Error("new global.Suite was nil")
	}

	if oldSuite == newSuite {
		t.Error("got the same suite but expected it to be different!")
	}
}

func TestResetAllowsRerunningSpecs(t *testing.T) {
	// RunSpecs refuses to run twice against the same suite, tracked by
	// global.SuiteDidRun. Reset must clear it alongside the suite itself, so a
	// program can run several suites sequentially in one process: register specs,
	// RunSpecs, Reset, register the next suite's specs, RunSpecs again.
	global.InitializeGlobals()

	global.SuiteDidRun = true
	globals.Reset()

	if global.SuiteDidRun {
		t.Error("expected Reset to clear SuiteDidRun so RunSpecs can run a fresh suite")
	}
}
