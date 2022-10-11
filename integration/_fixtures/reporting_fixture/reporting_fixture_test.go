package reporting_fixture_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("reporting test", func() {
	It("passes", func() {
	})

	Describe("labelled tests", Label("dog"), func() {
		It("is labelled", Label("dog", "cat"), func() {
		})
	})

	It("fails", func() {
		GinkgoWriter.Print("some ginkgo-writer output")
		Fail("fail!")
	})

	It("panics", func() {
		panic("boom")
	})

	It("has a progress report", func() {
		GinkgoWriter.Print("some ginkgo-writer preamble")
		time.Sleep(300 * time.Millisecond)
		GinkgoWriter.Print("some ginkgo-writer postamble")
	}, PollProgressAfter(50*time.Millisecond))

	PIt("is pending", func() {

	})

	It("is skipped", func() {
		Skip("skip")
	})

	It("times out and fails during cleanup", func(ctx SpecContext) {
		<-ctx.Done()
		DeferCleanup(func() { Fail("double-whammy") })
		Fail("failure-after-timeout")
	}, NodeTimeout(time.Millisecond*100))
})
