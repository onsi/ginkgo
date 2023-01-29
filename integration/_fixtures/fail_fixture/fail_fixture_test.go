package fail_fixture_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = It("handles top level failures", func() {
	Ω("a top level failure on line 12").Should(Equal("nope"))
	println("NEVER SEE THIS")
})

var _ = Describe("Exercising different failure modes", func() {
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

	It("times out", func(c context.Context) {
		<-c.Done()
	}, NodeTimeout(time.Millisecond*50))
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

var helper = func() {
	GinkgoHelper()
	Ω("a helper failed").Should(Equal("nope"))
}

var _ = It("tracks line numbers correctly when GinkgoHelper() is called", func() {
	helper()
})
