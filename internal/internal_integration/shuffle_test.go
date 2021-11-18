package internal_integration_test

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Shuffling Tests", func() {
	var fixture = func() {
		Describe("container-a", func() {
			It("a.1", rt.T("a.1"))
			It("a.2", rt.T("a.2"))
			It("a.3", rt.T("a.3"))
			It("a.4", rt.T("a.4"))
		})

		Describe("container-b", func() {
			It("b.1", rt.T("b.1"))
			It("b.2", rt.T("b.2"))
			It("b.3", rt.T("b.3"))
			It("b.4", rt.T("b.4"))
		})

		Describe("ordered-container", Ordered, func() {
			It("o.1", rt.T("o.1"))
			It("o.2", rt.T("o.2"))
			It("o.3", rt.T("o.3"))
			It("o.4", rt.T("o.4"))
		})

		It("top.1", rt.T("top.1"))
		It("top.2", rt.T("top.2"))
		It("top.3", rt.T("top.3"))
		It("top.4", rt.T("top.4"))
	}

	Describe("the default behavior", func() {
		It("shuffles top-level containers and its only, preserving the order of tests within containers", func() {
			orderings := []string{}
			uniqueOrderings := map[string]bool{}
			for i := 0; i < 10; i += 1 {
				conf.RandomSeed = int64(i)
				RunFixture("run", fixture)
				order := strings.Join(rt.TrackedRuns(), "")
				rt.Reset()

				Ω(order).Should(ContainSubstring("a.1a.2a.3a.4"), "order in containers should be preserved")
				Ω(order).Should(ContainSubstring("b.1b.2b.3b.4"), "order in containers should be preserved")
				Ω(order).Should(ContainSubstring("o.1o.2o.3o.4"), "order in containers should be preserved")
				orderings = append(orderings, order)
				uniqueOrderings[order] = true
			}
			Ω(orderings).Should(ContainElement(Not(ContainSubstring("top.1top.2top.3top.4"))), "top-level its should be randomized")
			Ω(uniqueOrderings).ShouldNot(HaveLen(1), "after 10 runs at least a few should be different!")
		})
	})

	Describe("when told to randomize all specs", func() {
		It("shuffles all its, but preserves ordered containers", func() {
			conf.RandomizeAllSpecs = true
			orderings := []string{}
			uniqueOrderings := map[string]bool{}
			for i := 0; i < 10; i += 1 {
				conf.RandomSeed = int64(i)
				RunFixture("run", fixture)
				order := strings.Join(rt.TrackedRuns(), "")
				rt.Reset()

				Ω(order).Should(ContainSubstring("o.1o.2o.3o.4"), "order in containers should be preserved")
				orderings = append(orderings, order)
				uniqueOrderings[order] = true
			}
			Ω(orderings).Should(ContainElement(Not(ContainSubstring("top.1top.2top.3top.4"))), "top-level its should be randomized")
			Ω(orderings).Should(ContainElement(Not(ContainSubstring("a.1a.2a.3a.4"))), "its in containers should be randomized")
			Ω(orderings).Should(ContainElement(Not(ContainSubstring("b.1b.2b.3b.4"))), "its in containers should be randomized")
			Ω(uniqueOrderings).ShouldNot(HaveLen(1), "after 10 runs at least a few should be different!")
		})
	})

	Describe("when given the same seed", func() {
		It("yields the same order", func() {
			for _, conf.RandomizeAllSpecs = range []bool{true, false} {
				uniqueOrderings := map[string]bool{}
				for i := 0; i < 10; i += 1 {
					conf.RandomSeed = 1138
					RunFixture("run", fixture)
					order := strings.Join(rt.TrackedRuns(), "")
					rt.Reset()
					uniqueOrderings[order] = true
				}

				Ω(uniqueOrderings).Should(HaveLen(1), "all orders are the same")
			}
		})
	})
})
