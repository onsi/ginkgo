package fail_fixture_test

import (
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("GinkgoTB", func() {
	It("synchronous failures with GinkgoTB().Fail", func() {
		GinkgoTB().Fail()
		println("NEVER SEE THIS")
	})

	DescribeTable("DescribeTable",
		func() {
			GinkgoTB().Fail()
		},
		Entry("a TableEntry constructed by Entry"),
	)

	It("tracks line numbers correctly when GinkgoTB().Helper() is called", func() {
		ginkgoTBHelper()
	})

	It("tracks the actual line number when no GinkgoTB helper is used", func() {
		ginkgoTBNoHelper()
	})
})

var ginkgoTBNoHelper = func() {
	GinkgoTB().Fail()
}
var ginkgoTBHelper = func() {
	t := GinkgoTB()
	t.Helper()
	t.Fail()
}
