package failing_table_tests

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("FailingTableTests", func() {

	DescribeTable("a failing entry",
		func(val bool) { Ω(val).Should(BeTrue()) },
		Entry("passing", true),
		Entry("failing", false),
	)
})
