package focused_fixture_test

import (
	. "github.com/onsi/ginkgo"
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

	FIt("focused", func() {

	})

	FMeasure("focused", func(b Benchmarker) {

	}, 2)

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

	Measure("not focused", func(b Benchmarker) {

	}, 2)
})
