package abort_fixture_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestAbortFixture(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "AbortFixture Suite")
}

var _ = Describe("top-level container", func() {
	It("runs and passes", func() {})
	It("aborts", func() {
		AbortSuite("this suite needs to end now!")
	})
	It("never runs", func() {
		Fail("SHOULD NOT SEE THIS")
	})
	It("never runs either", func() {
		Fail("SHOULD NOT SEE THIS")
	})
})
