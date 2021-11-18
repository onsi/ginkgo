package focused_fixture_test

import (
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("FocusedFixture", func() {
	FDescribe("focused", func() {
		It("focused", func() {

		})
	})

	FContext("focused", func() {
		It("focused", func() {

		})
	})

	FWhen("focused", func() {
		It("focused", func() {

		})
	})

	FIt("focused", func() {

	})

	It("focused", Focus, func() {

	})

	FSpecify("focused", func() {

	})

	FDescribeTable("focused",
		func() {},
		Entry("focused"),
	)

	DescribeTable("focused",
		func() {},
		FEntry("focused"),
	)

	Describe("not focused", func() {
		It("not focused", func() {

		})
	})

	Context("not focused", func() {
		It("not focused", func() {

		})
	})

	It("not focused", func() {

	})

	DescribeTable("not focused",
		func() {},
		Entry("not focused"),
	)
})
