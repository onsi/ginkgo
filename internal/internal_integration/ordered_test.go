package internal_integration_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/internal/test_helpers"
	. "github.com/onsi/gomega"
)

var _ = Describe("Ordered", func() {
	Context("the happy path", func() {
		BeforeEach(func() {
			conf.RandomizeAllSpecs = true
			success, _ := RunFixture("ordered happy path", func() {
				Context("container", Ordered, func() {
					It("A", rt.T("A"))
					It("B", rt.T("B"))
					It("C", rt.T("C"))
					It("D", rt.T("D"))
					It("E", rt.T("E"))
					It("F", rt.T("F"))
					It("G", rt.T("G"))
					It("H", rt.T("H"))
				})
			})
			Ω(success).Should(BeTrue())
		})

		It("always preserves order", func() {
			Ω(rt).Should(HaveTracked("A", "B", "C", "D", "E", "F", "G", "H"))
			Ω(reporter.Did.Names()).Should(Equal([]string{"A", "B", "C", "D", "E", "F", "G", "H"}))
		})
	})

	Context("when a spec in an ordered container fails", func() {
		BeforeEach(func() {
			counter := 1
			success, _ := RunFixture("ordered happy path", func() {
				Context("outer container", func() {
					Context("container", Ordered, func() {
						It("A", rt.T("A"))
						It("B", rt.T("B"))
						It("C", rt.T("C", func() { F("fail") }))
						It("D", rt.T("D"))
						It("E", rt.T("E"))
					})
					Context("container", Ordered, func() {
						It("F", FlakeAttempts(3), rt.T("F", func() {
							if counter < 3 {
								counter++
								F("fail")
							}
						}))
						It("G", rt.T("G"))
					})
				})
			})
			Ω(success).Should(BeFalse())
		})

		It("skips all subsequent specs in the ordered container", func() {
			Ω(rt).Should(HaveTracked("A", "B", "C", "F", "F", "F", "G"))
		})

		It("reports on the tests appropriately", func() {
			Ω(reporter.Did.Find("A")).Should(HavePassed())
			Ω(reporter.Did.Find("B")).Should(HavePassed())
			Ω(reporter.Did.Find("C")).Should(HaveFailed("fail"))
			Ω(reporter.Did.Find("D")).Should(HaveBeenSkippedWithMessage("Spec skipped because an earlier spec in an ordered container failed"))
			Ω(reporter.Did.Find("E")).Should(HaveBeenSkippedWithMessage("Spec skipped because an earlier spec in an ordered container failed"))
			Ω(reporter.Did.Find("F")).Should(HavePassed())
			Ω(reporter.Did.Find("G")).Should(HavePassed())
		})

		It("supports FlakeAttempts inside ordered containers", func() {
			Ω(reporter.Did.Find("F")).Should(HavePassed(NumAttempts(3)))
			Ω(reporter.Did.Find("G")).Should(HavePassed(NumAttempts(1)))
		})
	})
})
