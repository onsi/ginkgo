package example_test

import (
	fooginkgo "github.com/onsi/ginkgo"
	footable "github.com/onsi/ginkgo/extensions/table"
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

	fooginkgo.Measure("normal", func(b fooginkgo.Benchmarker) {

	}, 2)

	footable.DescribeTable("normal",
		func() {},
		footable.Entry("normal"),
	)

	footable.DescribeTable("normal",
		func() {},
		footable.Entry("normal"),
	)
})
