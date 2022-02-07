package skip_fixture_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = It("handles top level skips", func() {
	Skip("a top level skip on line 9")
	println("NEVER SEE THIS")
})

var _ = Describe("Exercising different skip modes", func() {
	It("synchronous skip", func() {
		Skip("a sync SKIP")
		println("NEVER SEE THIS")
	})
})

var _ = Describe("SKIP in a BeforeEach", func() {
	BeforeEach(func() {
		Skip("a BeforeEach SKIP")
		println("NEVER SEE THIS")
	})

	It("a SKIP BeforeEach", func() {
		println("NEVER SEE THIS")
	})
})

var _ = Describe("SKIP in an AfterEach", func() {
	AfterEach(func() {
		Skip("an AfterEach SKIP")
		println("NEVER SEE THIS")
	})

	It("a SKIP AfterEach", func() {
		Expect(true).To(BeTrue())
	})
})
