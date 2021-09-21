package internal_integration_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/internal/interrupt_handler"
	. "github.com/onsi/ginkgo/internal/test_helpers"
	"github.com/onsi/ginkgo/types"
	. "github.com/onsi/gomega"
)

var _ = Describe("Ordered", func() {
	Describe("simple ordered specs", func() {
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

	Describe("BeforeAll and AfterAll", func() {
		BeforeEach(func() {
			conf.RandomizeAllSpecs = true
		})

		Context("the happy path", func() {
			BeforeEach(func() {
				success, _ := RunFixture("ordered happy path", func() {
					BeforeEach(rt.T("BE1"))
					JustBeforeEach(rt.T("JBE1"))
					AfterEach(rt.T("AE1"))
					Context("container", Ordered, func() {
						BeforeEach(rt.T("BE2"))
						JustBeforeEach(rt.T("JBE2"))
						BeforeAll(rt.T("BA1"))
						BeforeEach(rt.T("BE3"))
						JustBeforeEach(rt.T("JBE3"))
						BeforeAll(rt.T("BA2"))
						BeforeEach(rt.T("BE4"))
						It("A", rt.T("A"))
						It("B", rt.T("B"))
						It("C", rt.T("C"))
						JustAfterEach(rt.T("JAE1"))
						AfterEach(rt.T("AE2"))
						AfterAll(rt.T("AA1"))
						AfterEach(rt.T("AE3"))
						JustAfterEach(rt.T("JAE2"))
						AfterAll(rt.T("AA2"))
						AfterEach(rt.T("AE4"))
						JustAfterEach(rt.T("JAE3"))
					})
					JustAfterEach(rt.T("JAE4"))
					AfterEach(rt.T("AE5"))
					BeforeEach(rt.T("BE5"))
					JustBeforeEach(rt.T("JBE4"))
				})
				Ω(success).Should(BeTrue())
			})

			It("runs the setup nodes just once and in the right order", func() {
				Ω(rt).Should(HaveTracked(
					"BE1", "BE5",
					"BA1", "BA2", "BE2", "BE3", "BE4",
					"JBE1", "JBE4", "JBE2", "JBE3",
					"A",
					"JAE1", "JAE2", "JAE3", "JAE4",
					"AE2", "AE3", "AE4",
					"AE1", "AE5",
					"BE1", "BE5",
					"BE2", "BE3", "BE4",
					"JBE1", "JBE4", "JBE2", "JBE3",
					"B",
					"JAE1", "JAE2", "JAE3", "JAE4",
					"AE2", "AE3", "AE4",
					"AE1", "AE5",
					"BE1", "BE5",
					"BE2", "BE3", "BE4",
					"JBE1", "JBE4", "JBE2", "JBE3",
					"C",
					"JAE1", "JAE2", "JAE3", "JAE4",
					"AE2", "AE3", "AE4", "AA1", "AA2",
					"AE1", "AE5",
				))
			})
		})

		Context("when there is only one test", func() {
			BeforeEach(func() {
				success, _ := RunFixture("ordered one test", func() {
					Context("container", Ordered, func() {
						BeforeEach(rt.T("BE"))
						BeforeAll(rt.T("BA"))
						It("A", rt.T("A"))
						AfterAll(rt.T("AA"))
						AfterEach(rt.T("AE"))
					})
				})
				Ω(success).Should(BeTrue())
			})

			It("runs the setup nodes just once and in the right order", func() {
				Ω(rt).Should(HaveTracked("BA", "BE", "A", "AE", "AA"))
			})
		})

		Context("when there are focused tests", func() {
			BeforeEach(func() {
				success, _ := RunFixture("ordered focused tests", func() {
					Context("container", Ordered, func() {
						BeforeEach(rt.T("BE"))
						BeforeAll(rt.T("BA"))
						It("A", rt.T("A"))
						FIt("B", rt.T("B"))
						FIt("C", rt.T("C"))
						AfterAll(rt.T("AA"))
						AfterEach(rt.T("AE"))
					})
				})
				Ω(success).Should(BeTrue())
			})

			It("runs the setup nodes just once and in the right order", func() {
				Ω(rt).Should(HaveTracked("BA", "BE", "B", "AE", "BE", "C", "AE", "AA"))
			})
		})

		Context("when there is nothing that will run", func() {
			BeforeEach(func() {
				success, _ := RunFixture("ordered nothing will run", func() {
					Context("container", Ordered, func() {
						BeforeEach(rt.T("BE"))
						BeforeAll(rt.T("BA"))
						PIt("A", rt.T("A"))
						AfterAll(rt.T("AA"))
						AfterEach(rt.T("AE"))
					})
				})
				Ω(success).Should(BeTrue())
			})

			It("does not run BeforeAll/AfterAll", func() {
				Ω(rt.TrackedRuns()).Should(BeEmpty())
			})
		})

		Context("when a failure occurs prior to the BeforeAll running", func() {
			BeforeEach(func() {
				success, _ := RunFixture("ordered failure before BeforeAll", func() {
					BeforeEach(rt.T("BE-outer", func() { F("fail") }))
					Context("container", Ordered, func() {
						BeforeEach(rt.T("BE"))
						BeforeAll(rt.T("BA"))
						It("A", rt.T("A"))
						AfterAll(rt.T("AA"))
						AfterEach(rt.T("AE"))
					})
					AfterEach(rt.T("AE-outer"))
				})
				Ω(success).Should(BeFalse())
			})

			It("does not run the BeforeAll or the AfterAll", func() {
				Ω(rt).Should(HaveTracked("BE-outer", "AE-outer"))
				Ω(reporter.Did.Find("A")).Should(HaveFailed(types.FailureNodeAtTopLevel, FailureNodeType(types.NodeTypeBeforeEach), "fail"))
			})
		})

		Context("when a failure occurs in a test", func() {
			BeforeEach(func() {
				success, _ := RunFixture("ordered failure in test", func() {
					BeforeEach(rt.T("BE-outer"))
					Context("container", Ordered, func() {
						BeforeEach(rt.T("BE"))
						BeforeAll(rt.T("BA"))
						It("A", rt.T("A", func() { F("fail") }))
						It("B", rt.T("B"))
						AfterAll(rt.T("AA"))
						AfterEach(rt.T("AE"))
					})
					AfterEach(rt.T("AE-outer"))
				})
				Ω(success).Should(BeFalse())
			})

			It("runs the AfterAll when that test ends and skips subsequent tests", func() {
				Ω(rt).Should(HaveTracked("BE-outer", "BA", "BE", "A", "AE", "AA", "AE-outer"))
				Ω(reporter.Did.Find("A")).Should(HaveFailed(types.FailureNodeIsLeafNode, FailureNodeType(types.NodeTypeIt), "fail"))
			})
		})

		Context("when a failure occurs in a BeforeAll", func() {
			BeforeEach(func() {
				success, _ := RunFixture("ordered failure in BeforeAll", func() {
					BeforeEach(rt.T("BE-outer"))
					Context("container", Ordered, func() {
						BeforeEach(rt.T("BE"))
						BeforeAll(rt.T("BA", func() { F("fail") }))
						It("A", rt.T("A"))
						It("B", rt.T("B"))
						AfterAll(rt.T("AA"))
						AfterEach(rt.T("AE"))
					})
					AfterEach(rt.T("AE-outer"))
				})
				Ω(success).Should(BeFalse())
			})

			It("still manages to run the AfterAll for the test, even if that means it runs it out of order with the AfterEach", func() {
				Ω(rt).Should(HaveTracked("BE-outer", "BA", "AE", "AA", "AE-outer"))
				Ω(reporter.Did.Find("A")).Should(HaveFailed(types.FailureNodeInContainer, FailureNodeType(types.NodeTypeBeforeAll), "fail"))
			})
		})

		Context("when a failure occurs in an AfterAll", func() {
			BeforeEach(func() {
				success, _ := RunFixture("ordered failure in BeforeAll", func() {
					BeforeEach(rt.T("BE-outer"))
					Context("container", Ordered, func() {
						BeforeEach(rt.T("BE"))
						BeforeAll(rt.T("BA"))
						It("A", rt.T("A"))
						It("B", rt.T("B"))
						AfterAll(rt.T("AA", func() { F("fail") }))
						AfterEach(rt.T("AE"))
					})
					AfterEach(rt.T("AE-outer"))
				})
				Ω(success).Should(BeFalse())
			})

			It("still manages to run the AfterAll for the test, even if that means it runs it out of order with the AfterEach", func() {
				Ω(rt).Should(HaveTracked(
					"BE-outer", "BA", "BE", "A", "AE", "AE-outer",
					"BE-outer", "BE", "B", "AE", "AA", "AE-outer",
				))

				Ω(reporter.Did.Find("B")).Should(HaveFailed(types.FailureNodeInContainer, FailureNodeType(types.NodeTypeAfterAll), "fail"))
			})
		})

		Context("when a failure occurs in an AfterEach", func() {
			BeforeEach(func() {
				success, _ := RunFixture("ordered failure in BeforeAll", func() {
					BeforeEach(rt.T("BE-outer"))
					Context("container", Ordered, func() {
						BeforeEach(rt.T("BE"))
						BeforeAll(rt.T("BA"))
						It("A", rt.T("A"))
						It("B", rt.T("B"))
						AfterAll(rt.T("AA"))
						AfterEach(rt.T("AE", func() { F("fail") }))
					})
					AfterEach(rt.T("AE-outer"))
				})
				Ω(success).Should(BeFalse())
			})

			It("still manages to run the AfterAll for the test, even if that means it runs it out of order with the AfterEach", func() {
				Ω(rt).Should(HaveTracked("BE-outer", "BA", "BE", "A", "AE", "AE-outer", "AA"))

				Ω(reporter.Did.Find("A")).Should(HaveFailed(types.FailureNodeInContainer, FailureNodeType(types.NodeTypeAfterEach), "fail"))
			})
		})

		Context("when an interruption occurs", func() {
			BeforeEach(func() {
				success, _ := RunFixture("ordered failure in BeforeAll", func() {
					BeforeEach(rt.T("BE-outer"))
					Context("container", Ordered, func() {
						BeforeEach(rt.T("BE"))
						BeforeAll(rt.T("BA"))
						It("A", rt.T("A", func() {
							interruptHandler.Interrupt(interrupt_handler.InterruptCauseSignal)
							time.Sleep(time.Minute)
						}))
						It("B", rt.T("B"))
						AfterAll(rt.T("AA"))
						AfterEach(rt.T("AE"))
					})
					AfterEach(rt.T("AE-outer"))
				})
				Ω(success).Should(BeFalse())
			})

			It("runs the AfterAll and skips subsequent tests", func() {
				Ω(rt).Should(HaveTracked("BE-outer", "BA", "BE", "A", "AE", "AA", "AE-outer"))

				Ω(reporter.Did.Find("A")).Should(HaveBeenInterrupted(interrupt_handler.InterruptCauseSignal))
			})
		})
	})

})
