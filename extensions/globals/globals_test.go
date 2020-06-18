package globals_test

import (
	"testing"

	"github.com/onsi/ginkgo/extensions/globals"
	"github.com/onsi/ginkgo/internal/global"
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
