package reporting_fixture_test

import (
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

	PIt("is pending", func() {

	})

	It("is skipped", func() {
		Skip("skip")
	})
})
