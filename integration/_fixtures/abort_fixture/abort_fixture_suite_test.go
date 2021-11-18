package abort_fixture_test

import (
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestAbortFixture(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "AbortFixture Suite")
}

var _ = Describe("top-level container", func() {
	It("runs and passes", func() {})
	It("aborts", func() {
		time.Sleep(time.Second)
		AbortSuite("this suite needs to end now!")
	})
	It("never runs", func() {
		time.Sleep(time.Hour)
		Fail("SHOULD NOT SEE THIS")
	})
	It("never runs either", func() {
		time.Sleep(time.Hour)
		Fail("SHOULD NOT SEE THIS")
	})
})
