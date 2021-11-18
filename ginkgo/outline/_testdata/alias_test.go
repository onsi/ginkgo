package example_test

import (
	fooginkgo "github.com/onsi/ginkgo/v2"
)

var _ = fooginkgo.Describe("NodotFixture", func() {
	fooginkgo.Describe("normal", func() {
		fooginkgo.It("normal", func() {
			fooginkgo.By("normal")
			fooginkgo.By("normal")

		})
	})

	fooginkgo.Context("normal", func() {
		fooginkgo.It("normal", func() {

		})
	})

	fooginkgo.When("normal", func() {
		fooginkgo.It("normal", func() {

		})
	})

	fooginkgo.It("normal", func() {

	})

	fooginkgo.Specify("normal", func() {

	})

	fooginkgo.DescribeTable("normal",
		func() {},
		fooginkgo.Entry("normal"),
	)

	fooginkgo.DescribeTable("normal",
		func() {},
		fooginkgo.Entry("normal"),
	)
})
