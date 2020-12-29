package example_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
)

var _ = Describe("NormalFixture", func() {
	Describe("normal", func() {
		It("normal", func() {
			By("step 1")
			By("step 2")
		})
	})

	Context("normal", func() {
		It("normal", func() {

		})
	})

	When("normal", func() {
		It("normal", func() {

		})
	})

	It("normal", func() {

	})

	Specify("normal", func() {

	})

	Measure("normal", func(b Benchmarker) {

	}, 2)

	DescribeTable("normal",
		func() {},
		Entry("normal"),
	)

	DescribeTable("normal",
		func() {},
		Entry("normal"),
	)
})
