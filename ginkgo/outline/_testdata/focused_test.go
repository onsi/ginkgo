package example_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
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

	FMeasure("focused", func(b Benchmarker) {

	}, 2)

	FDescribeTable("focused",
		func() {},
		Entry("focused"),
	)

	DescribeTable("focused",
		func() {},
		FEntry("focused"),
	)
})
