package fail_fixture_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = It("handles top level failures", func() {
	Ω("a top level failure on line 9").Should(Equal("nope"))
	println("NEVER SEE THIS")
})

var _ = Describe("Excercising different failure modes", func() {
	It("synchronous failures", func() {
		Ω("a sync failure").Should(Equal("nope"))
		println("NEVER SEE THIS")
	})

	It("synchronous panics", func() {
		panic("a sync panic")
		println("NEVER SEE THIS")
	})

	It("synchronous failures with FAIL", func() {
		Fail("a sync FAIL failure")
		println("NEVER SEE THIS")
	})
})

var _ = Specify("a top level specify", func() {
	Fail("fail the test")
})

var _ = DescribeTable("a top level DescribeTable",
	func(x, y int) {
		Expect(x).To(Equal(y))
	},
	Entry("a TableEntry constructed by Entry", 2, 3),
)
