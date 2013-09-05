package ginkgo

import (
	. "github.com/onsi/gomega"
)

func init() {
	Describe("Various Helpers", func() {
		Describe("parallelizedIndexRange", func() {
			It("should return the correct index range", func() {
				//Lazy TDD... :)
				var startIndex, count int

				startIndex, count = parallelizedIndexRange(4, 2, 1)
				Ω(startIndex).Should(Equal(0))
				Ω(count).Should(Equal(2))

				startIndex, count = parallelizedIndexRange(4, 2, 2)
				Ω(startIndex).Should(Equal(2))
				Ω(count).Should(Equal(2))

				startIndex, count = parallelizedIndexRange(5, 2, 1)
				Ω(startIndex).Should(Equal(0))
				Ω(count).Should(Equal(2))

				startIndex, count = parallelizedIndexRange(5, 2, 2)
				Ω(startIndex).Should(Equal(2))
				Ω(count).Should(Equal(3))

				startIndex, count = parallelizedIndexRange(5, 3, 1)
				Ω(startIndex).Should(Equal(0))
				Ω(count).Should(Equal(1))

				startIndex, count = parallelizedIndexRange(5, 3, 2)
				Ω(startIndex).Should(Equal(1))
				Ω(count).Should(Equal(1))

				startIndex, count = parallelizedIndexRange(5, 3, 3)
				Ω(startIndex).Should(Equal(2))
				Ω(count).Should(Equal(3))

				startIndex, count = parallelizedIndexRange(5, 4, 1)
				Ω(startIndex).Should(Equal(0))
				Ω(count).Should(Equal(1))

				startIndex, count = parallelizedIndexRange(5, 4, 2)
				Ω(startIndex).Should(Equal(1))
				Ω(count).Should(Equal(1))

				startIndex, count = parallelizedIndexRange(5, 4, 3)
				Ω(startIndex).Should(Equal(2))
				Ω(count).Should(Equal(1))

				startIndex, count = parallelizedIndexRange(5, 4, 4)
				Ω(startIndex).Should(Equal(3))
				Ω(count).Should(Equal(2))

				startIndex, count = parallelizedIndexRange(5, 5, 1)
				Ω(startIndex).Should(Equal(0))
				Ω(count).Should(Equal(1))

				startIndex, count = parallelizedIndexRange(5, 5, 2)
				Ω(startIndex).Should(Equal(1))
				Ω(count).Should(Equal(1))

				startIndex, count = parallelizedIndexRange(5, 5, 3)
				Ω(startIndex).Should(Equal(2))
				Ω(count).Should(Equal(1))

				startIndex, count = parallelizedIndexRange(5, 5, 4)
				Ω(startIndex).Should(Equal(3))
				Ω(count).Should(Equal(1))

				startIndex, count = parallelizedIndexRange(5, 5, 5)
				Ω(startIndex).Should(Equal(4))
				Ω(count).Should(Equal(1))

				startIndex, count = parallelizedIndexRange(5, 6, 1)
				Ω(startIndex).Should(Equal(0))
				Ω(count).Should(Equal(1))

				startIndex, count = parallelizedIndexRange(5, 6, 2)
				Ω(startIndex).Should(Equal(1))
				Ω(count).Should(Equal(1))

				startIndex, count = parallelizedIndexRange(5, 6, 3)
				Ω(startIndex).Should(Equal(2))
				Ω(count).Should(Equal(1))

				startIndex, count = parallelizedIndexRange(5, 6, 4)
				Ω(startIndex).Should(Equal(3))
				Ω(count).Should(Equal(1))

				startIndex, count = parallelizedIndexRange(5, 6, 5)
				Ω(startIndex).Should(Equal(4))
				Ω(count).Should(Equal(1))

				startIndex, count = parallelizedIndexRange(5, 6, 6)
				Ω(startIndex).Should(Equal(5))
				Ω(count).Should(Equal(0))

				startIndex, count = parallelizedIndexRange(5, 7, 6)
				Ω(startIndex).Should(Equal(5))
				Ω(count).Should(Equal(0))

				startIndex, count = parallelizedIndexRange(5, 7, 7)
				Ω(startIndex).Should(Equal(5))
				Ω(count).Should(Equal(0))
			})
		})
	})
}
