package example_test

import (
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("unfocused", func() {
	FDescribe("focused", func() {
		It("focused", func() {
			By("focused")
			By("focused")
		})
	})

	FContext("focused", func() {
		It("focused", func() {

		})
	})

	FWhen("focused", func() {
		It("focused", func() {

		})
	})

	FIt("focused", func() {

	})

	FSpecify("focused", func() {

	})

	FDescribeTable("focused",
		func() {},
		Entry("focused"),
	)

	DescribeTable("focused",
		func() {},
		FEntry("focused"),
	)
})
