package example_test

import (
	"github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/extensions/table"
)

var _ = ginkgo.Describe("NodotFixture", func() {
	ginkgo.Describe("normal", func() {
		ginkgo.It("normal", func() {
			ginkgo.By("normal")
			ginkgo.By("normal")
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

	ginkgo.Measure("normal", func(b ginkgo.Benchmarker) {

	}, 2)

	table.DescribeTable("normal",
		func() {},
		table.Entry("normal"),
	)

	table.DescribeTable("normal",
		func() {},
		table.Entry("normal"),
	)
})
