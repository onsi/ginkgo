package example_test

import (
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("NormalFixture", func() {
	Describe("normal", Label("normal", "serial"), Pending, func() {
		It("normal", func() {
			By("step 1")
			By("step 2")
		})
	})

	Context("normal", func() {
		It("normal", Label("medium"), Label("slow"), func() {

		})
	})

	When("normal", func() {
		It("normal", func() {

		})
	})

	It("normal", Pending(), func() {

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
