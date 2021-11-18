package example_test

import (
	. "github.com/onsi/ginkgo/v2"
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

	DescribeTable("normal",
		func() {},
		Entry("normal"),
	)

	DescribeTable("normal",
		func() {},
		Entry("normal"),
	)
})
