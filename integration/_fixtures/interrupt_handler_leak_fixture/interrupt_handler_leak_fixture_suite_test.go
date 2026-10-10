package interrupt_handler_leak_fixture_test

import (
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/extensions/globals"
	. "github.com/onsi/gomega"
)

// verifyNoInterruptHandlerGoroutine fails the test if an InterruptHandler
// goroutine is still alive shortly after the suite entry point returned.
// InterruptHandler.Stop is synchronous from the caller's perspective but the
// goroutines observe the closed channel asynchronously, hence the bounded poll.
func verifyNoInterruptHandlerGoroutine(t *testing.T) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		time.Sleep(50 * time.Millisecond)
		buf := make([]byte, 1<<20)
		n := runtime.Stack(buf, true)
		if !strings.Contains(string(buf[:n]), "registerForInterrupts") {
			return
		}
	}
	buf := make([]byte, 1<<20)
	n := runtime.Stack(buf, true)
	t.Fatalf("InterruptHandler goroutine(s) outlived the suite entry point:\n%s", string(buf[:n]))
}

func TestInterruptHandlerLeakFixture(t *testing.T) {
	RegisterFailHandler(Fail)
	switch {
	case os.Getenv("PREVIEW") == "true":
		PreviewSpecs("InterruptHandler Leak Fixture Suite")
		verifyNoInterruptHandlerGoroutine(t)
	case os.Getenv("RERUN") == "true":
		RunSpecs(t, "InterruptHandler Leak Fixture Suite")
		globals.Reset()
		RunSpecs(t, "InterruptHandler Leak Fixture Suite (second run)")
		verifyNoInterruptHandlerGoroutine(t)
	default:
		RunSpecs(t, "InterruptHandler Leak Fixture Suite")
		verifyNoInterruptHandlerGoroutine(t)
	}
}

var _ = Describe("a suite that runs a spec", func() {
	It("has a passing spec", func() {
		Expect(true).To(BeTrue())
	})
})
