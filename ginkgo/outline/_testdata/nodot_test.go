package example_test

import (
	"github.com/onsi/ginkgo"
)

var _ = ginkgo.Describe("NodotFixture", func() {
	ginkgo.Describe("normal", func() {
		ginkgo.It("normal", func() {

		})
	})

	ginkgo.Context("normal", func() {
		ginkgo.It("normal", func() {

		})
	})

	ginkgo.When("normal", func() {
		ginkgo.It("normal", func() {

		})
	})

	ginkgo.It("normal", func() {

	})

	ginkgo.Specify("normal", func() {

	})

	ginkgo.Measure("normal", func(b Benchmarker) {

	}, 2)

	ginkgo.DescribeTable("normal",
		func() {},
		ginkgo.Entry("normal"),
	)

	ginkgo.DescribeTable("normal",
		func() {},
		ginkgo.Entry("normal"),
	)
})
