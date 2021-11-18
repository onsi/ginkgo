package example_test

import (
	"github.com/onsi/ginkgo/v2"
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

	ginkgo.DescribeTable("normal",
		func() {},
		ginkgo.Entry("normal"),
	)

	ginkgo.DescribeTable("normal",
		func() {},
		ginkgo.Entry("normal"),
	)
})
