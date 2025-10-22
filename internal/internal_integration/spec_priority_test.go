package internal_integration_test

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("SpecPriority", func() {
	var fixture = func() {
		Describe("A", SpecPriority(10), func() {
			It("A1", rt.T("A1"), SpecPriority(20))
			It("A2", rt.T("A2"), SpecPriority(21))
		})

		Describe("B", SpecPriority(-10), func() {
			It("B1", rt.T("B1"))
			It("B2", rt.T("B2"))
		})

		Describe("C", func() {
			It("C1", rt.T("C1"), SpecPriority(15))
			It("C2", rt.T("C2"), SpecPriority(15))
			It("C3", rt.T("C3"), SpecPriority(16))

		})

		Describe("D", Ordered, SpecPriority(5), func() {
			It("D1", rt.T("D1"))
			It("D2", rt.T("D2"))
			It("D3", rt.T("D3"))
		})

		Describe("E", SpecPriority(-10), func() {
			It("E1", rt.T("E1"))
			It("E2", rt.T("E2"))
		})

	}

	Context("without randomize-all", func() {
		It("should respect spec priorities", func() {
			conf.RandomizeAllSpecs = false

			num812 := map[string]int{}
			for i := 0; i < 50; i++ {
				rt.Reset()
				conf.RandomSeed = int64(i + 1)
				success, _ := RunFixture("spec priority test", fixture)
				Ω(success).Should(BeTrue())
				runs := rt.TrackedRuns()
				Ω(runs[0:2]).Should(Equal([]string{"A1", "A2"}))
				Ω(runs[2:5]).Should(Equal([]string{"C1", "C2", "C3"}))
				Ω(runs[5:8]).Should(Equal([]string{"D1", "D2", "D3"}))
				Ω(runs[8:12]).Should(Or(
					Equal([]string{"B1", "B2", "E1", "E2"}),
					Equal([]string{"E1", "E2", "B1", "B2"}),
				))
				num812[strings.Join(runs[8:12], "")] += 1
			}
			Ω(len(num812)).Should(Equal(2))
		})
	})

	Context("with randomize-all", func() {
		It("should respect spec priorities", func() {
			conf.RandomizeAllSpecs = true
			num812 := map[string]int{}
			for i := 0; i < 1000; i++ {
				rt.Reset()
				conf.RandomSeed = int64(i + 1)
				success, _ := RunFixture("spec priority test", fixture)
				Ω(success).Should(BeTrue())
				runs := rt.TrackedRuns()
				Ω(runs[0:2]).Should(Equal([]string{"A2", "A1"}))
				Ω(runs[2:5]).Should(Or(
					Equal([]string{"C3", "C1", "C2"}),
					Equal([]string{"C3", "C2", "C1"}),
				))
				Ω(runs[5:8]).Should(Equal([]string{"D1", "D2", "D3"}))
				Ω(runs[8:12]).Should(ConsistOf("B1", "B2", "E1", "E2"))
				num812[strings.Join(runs[8:12], "")] += 1
			}
			Ω(len(num812)).Should(Equal(4 * 3 * 2 * 1))
		})
	})
})
