package spec_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/internal/spec"
	. "github.com/onsi/gomega"
)

var _ = Describe("ParallelizedIndexRange", func() {
	It("should return the correct index range", func() {
		var startIndex, count int

		startIndex, count = ParallelizedIndexRange(4, 2, 1)
		Ω(startIndex).Should(Equal(0))
		Ω(count).Should(Equal(2))

		startIndex, count = ParallelizedIndexRange(4, 2, 2)
		Ω(startIndex).Should(Equal(2))
		Ω(count).Should(Equal(2))

		startIndex, count = ParallelizedIndexRange(5, 2, 1)
		Ω(startIndex).Should(Equal(0))
		Ω(count).Should(Equal(3))

		startIndex, count = ParallelizedIndexRange(5, 2, 2)
		Ω(startIndex).Should(Equal(3))
		Ω(count).Should(Equal(2))

		startIndex, count = ParallelizedIndexRange(5, 3, 1)
		Ω(startIndex).Should(Equal(0))
		Ω(count).Should(Equal(2))

		startIndex, count = ParallelizedIndexRange(5, 3, 2)
		Ω(startIndex).Should(Equal(2))
		Ω(count).Should(Equal(2))

		startIndex, count = ParallelizedIndexRange(5, 3, 3)
		Ω(startIndex).Should(Equal(4))
		Ω(count).Should(Equal(1))

		startIndex, count = ParallelizedIndexRange(5, 4, 1)
		Ω(startIndex).Should(Equal(0))
		Ω(count).Should(Equal(1))

		startIndex, count = ParallelizedIndexRange(5, 4, 2)
		Ω(startIndex).Should(Equal(1))
		Ω(count).Should(Equal(1))

		startIndex, count = ParallelizedIndexRange(5, 4, 3)
		Ω(startIndex).Should(Equal(2))
		Ω(count).Should(Equal(1))

		startIndex, count = ParallelizedIndexRange(5, 4, 4)
		Ω(startIndex).Should(Equal(3))
		Ω(count).Should(Equal(2))

		startIndex, count = ParallelizedIndexRange(5, 5, 1)
		Ω(startIndex).Should(Equal(0))
		Ω(count).Should(Equal(1))

		startIndex, count = ParallelizedIndexRange(5, 5, 2)
		Ω(startIndex).Should(Equal(1))
		Ω(count).Should(Equal(1))

		startIndex, count = ParallelizedIndexRange(5, 5, 3)
		Ω(startIndex).Should(Equal(2))
		Ω(count).Should(Equal(1))

		startIndex, count = ParallelizedIndexRange(5, 5, 4)
		Ω(startIndex).Should(Equal(3))
		Ω(count).Should(Equal(1))

		startIndex, count = ParallelizedIndexRange(5, 5, 5)
		Ω(startIndex).Should(Equal(4))
		Ω(count).Should(Equal(1))

		startIndex, count = ParallelizedIndexRange(5, 6, 1)
		Ω(startIndex).Should(Equal(0))
		Ω(count).Should(Equal(1))

		startIndex, count = ParallelizedIndexRange(5, 6, 2)
		Ω(startIndex).Should(Equal(1))
		Ω(count).Should(Equal(1))

		startIndex, count = ParallelizedIndexRange(5, 6, 3)
		Ω(startIndex).Should(Equal(2))
		Ω(count).Should(Equal(1))

		startIndex, count = ParallelizedIndexRange(5, 6, 4)
		Ω(startIndex).Should(Equal(3))
		Ω(count).Should(Equal(1))

		startIndex, count = ParallelizedIndexRange(5, 6, 5)
		Ω(startIndex).Should(Equal(4))
		Ω(count).Should(Equal(1))

		startIndex, count = ParallelizedIndexRange(5, 6, 6)
		Ω(startIndex).Should(Equal(5))
		Ω(count).Should(Equal(0))

		startIndex, count = ParallelizedIndexRange(5, 7, 6)
		Ω(startIndex).Should(Equal(5))
		Ω(count).Should(Equal(0))

		startIndex, count = ParallelizedIndexRange(5, 7, 7)
		Ω(startIndex).Should(Equal(5))
		Ω(count).Should(Equal(0))
	})
})
