package example_test

import (
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("PendingFixture", func() {
	PDescribe("pending", func() {
		It("pending", func() {
			By("pending")
			By("pending")
		})
	})

	PContext("pending", func() {
		It("pending", func() {

		})
	})

	PWhen("pending", func() {
		It("pending", func() {

		})
	})

	PIt("pending", func() {

	})

	PSpecify("pending", func() {

	})

	PDescribeTable("pending",
		func() {},
		Entry("pending"),
	)

	DescribeTable("pending",
		func() {},
		PEntry("pending"),
	)
})
