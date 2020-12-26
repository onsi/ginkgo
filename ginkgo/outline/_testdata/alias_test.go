package example_test

import (
	fooginkgo "github.com/onsi/ginkgo"
)

var _ = fooginkgo.Describe("NodotFixture", func() {
	fooginkgo.Describe("normal", func() {
		fooginkgo.It("normal", func() {

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

	fooginkgo.Measure("normal", func(b Benchmarker) {

	}, 2)

	fooginkgo.DescribeTable("normal",
		func() {},
		fooginkgo.Entry("normal"),
	)

	fooginkgo.DescribeTable("normal",
		func() {},
		fooginkgo.Entry("normal"),
	)
})
