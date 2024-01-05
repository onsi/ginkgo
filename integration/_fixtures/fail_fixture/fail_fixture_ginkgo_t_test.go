package fail_fixture_test

import (
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("GinkgoT", func() {
	It("synchronous failures with GinkgoT().Fail", func() {
		GinkgoT().Fail()
		println("NEVER SEE THIS")
	})

	DescribeTable("DescribeTable",
		func() {
			GinkgoT().Fail()
		},
		Entry("a TableEntry constructed by Entry"),
	)

	It("tracks line numbers correctly when GinkgoT().Helper() is called", func() {
		ginkgoTHelper()
	})

	It("tracks the actual line number when no helper is used", func() {
		ginkgoTNoHelper()
	})
})

var ginkgoTNoHelper = func() {
	GinkgoT().Fail()
}
var ginkgoTHelper = func() {
	GinkgoT().Helper()
	GinkgoT().Fail()
}
