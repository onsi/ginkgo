package internal_integration_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/onsi/ginkgo/v2/internal/interrupt_handler"
	. "github.com/onsi/ginkgo/v2/internal/test_helpers"
	"github.com/onsi/ginkgo/v2/types"
)

var _ = Describe("handling test aborts", func() {
	Describe("when BeforeSuite aborts", func() {
		BeforeEach(func() {
			success, _ := RunFixture("abort beforesuite", func() {
				BeforeSuite(rt.T("before-suite", func() {
					writer.Write([]byte("before-suite"))
					Abort("abort", cl)
				}))
				It("A", rt.T("A"))
				It("B", rt.T("B"))
				AfterSuite(rt.T("after-suite"))
			})
			Ω(success).Should(BeFalse())
		})

		It("reports a suite failure", func() {
			Ω(reporter.End).Should(BeASuiteSummary(false, NSpecs(2), NSkipped(0)))
		})

		It("reports a failure for the BeforeSuite", func() {
			Ω(reporter.Did.FindByLeafNodeType(types.NodeTypeBeforeSuite)).Should(HaveAborted("abort", cl, CapturedGinkgoWriterOutput("before-suite")))
			Ω(reporter.Did.FindByLeafNodeType(types.NodeTypeAfterSuite)).Should(HavePassed())
		})

		It("does not run any of the Its", func() {
			Ω(rt).ShouldNot(HaveRun("A"))
			Ω(rt).ShouldNot(HaveRun("B"))
		})

		It("does run the AfterSuite", func() {
			Ω(rt).Should(HaveTracked("before-suite", "after-suite"))
		})
	})

	Describe("when AfterSuite aborts", func() {
		BeforeEach(func() {
			success, _ := RunFixture("abort aftersuite", func() {
				BeforeSuite(rt.T("before-suite"))
				Describe("top-level", func() {
					It("A", rt.T("A"))
					It("B", rt.T("B"))
				})
				AfterSuite(rt.T("after-suite", func() {
					writer.Write([]byte("after-suite"))
					Abort("abort", cl)
				}))
			})
			Ω(success).Should(BeFalse())
		})

		It("reports a suite failure", func() {
			Ω(reporter.End).Should(BeASuiteSummary(false, NSpecs(2), NPassed(2)))
		})

		It("runs and reports on all the tests and reports a failure for the AfterSuite", func() {
			Ω(rt).Should(HaveTracked("before-suite", "A", "B", "after-suite"))
			Ω(reporter.Did.FindByLeafNodeType(types.NodeTypeBeforeSuite)).Should(HavePassed())
			Ω(reporter.Did.Find("A")).Should(HavePassed())
			Ω(reporter.Did.Find("B")).Should(HavePassed())
			Ω(reporter.Did.FindByLeafNodeType(types.NodeTypeAfterSuite)).Should(HaveAborted("abort", cl, CapturedGinkgoWriterOutput("after-suite")))
		})
	})

	Describe("individual test aborts", func() {
		Describe("when an It aborts", func() {
			BeforeEach(func() {
				success, _ := RunFixture("failed it", func() {
					BeforeSuite(rt.T("before-suite"))
					Describe("top-level", func() {
						It("A", rt.T("A", func() {
							writer.Write([]byte("running A"))
						}))
						It("B", rt.T("B", func() {
							writer.Write([]byte("running B"))
							Abort("abort", cl)
						}))
						It("C", rt.T("C"))
						It("D", rt.T("D"))
					})
					AfterEach(rt.T("after-each"))
					AfterSuite(rt.T("after-suite"))
				})
				Ω(success).Should(BeFalse())
			})

			It("reports a suite failure", func() {
				Ω(reporter.End).Should(BeASuiteSummary(false, NSpecs(4), NPassed(1), NFailed(1), NSkipped(2)))
			})

			It("does not run subsequent Its, the AfterEach, and the AfterSuite", func() {
				Ω(rt).Should(HaveTracked("before-suite", "A", "after-each", "B", "after-each", "after-suite"))
			})

			It("reports the It's abort and subsequent tests as skipped", func() {
				Ω(reporter.Did.Find("A")).Should(HavePassed(CapturedGinkgoWriterOutput("running A")))
				Ω(reporter.Did.Find("B")).Should(HaveAborted("abort", cl, CapturedGinkgoWriterOutput("running B")))
				Ω(reporter.Did.Find("C")).Should(HaveBeenSkipped())
				Ω(reporter.Did.Find("D")).Should(HaveBeenSkipped())
			})

			It("sets up the failure node location correctly", func() {
				report := reporter.Did.Find("B")
				Ω(report.Failure.FailureNodeContext).Should(Equal(types.FailureNodeIsLeafNode))
				Ω(report.Failure.FailureNodeType).Should(Equal(types.NodeTypeIt))
				Ω(report.Failure.FailureNodeLocation).Should(Equal(report.LeafNodeLocation))
			})
		})
	})

	Describe("when a test fails then an AfterEach aborts", func() {
		BeforeEach(func() {
			success, _ := RunFixture("failed it then after-each aborts", func() {
				BeforeSuite(rt.T("before-suite"))
				Describe("top-level", func() {
					It("A", rt.T("A"))
					It("B", rt.T("B", func() {
						writer.Write([]byte("running B"))
						F("fail")
					}))
					It("C", rt.T("C"))
					It("D", rt.T("D"))
				})
				ReportAfterEach(func(report SpecReport) {
					rt.Run("report-after-each")
					if report.State.Is(types.SpecStateFailed) {
						Abort("abort", cl)
					}
				})
				AfterSuite(rt.T("after-suite"))
			})
			Ω(success).Should(BeFalse())
		})

		It("reports a suite failure", func() {
			Ω(reporter.End).Should(BeASuiteSummary(false, NSpecs(4), NPassed(1), NFailed(1), NSkipped(2)))
		})

		It("does not run subsequent Its, the AfterEach, and the AfterSuite", func() {
			Ω(rt).Should(HaveTracked("before-suite", "A", "report-after-each", "B", "report-after-each", "report-after-each", "report-after-each", "after-suite"))
		})

		It("reports a failure and then aborts the rest of the suite", func() {
			Ω(reporter.Did.Find("A")).Should(HavePassed())
			Ω(reporter.Did.Find("B")).Should(HaveAborted("abort", cl, CapturedGinkgoWriterOutput("running B")))
			Ω(reporter.Did.Find("C")).Should(HaveBeenSkipped())
			Ω(reporter.Did.Find("D")).Should(HaveBeenSkipped())
		})
	})

	Describe("when running in parallel and a test aborts", func() {
		var c chan interface{}
		BeforeEach(func() {
			SetUpForParallel(2)
			c = make(chan interface{})
		})

		It("notifies the server of the abort", func() {
			Ω(client.ShouldAbort()).Should(BeFalse())
			success := RunFixtureInParallel("aborting in parallel", func(_ int) {
				It("A", func() {
					<-c
					Abort("abort")
				})

				It("B", func(ctx SpecContext) {
					close(c)
					select {
					case <-ctx.Done():
						rt.Run("dc-done")
					case <-time.After(interrupt_handler.ABORT_POLLING_INTERVAL * 2):
						rt.Run("dc-after")
					}
				})
			})
			Ω(success).Should(BeFalse())
			Ω(client.ShouldAbort()).Should(BeTrue())

			Ω(rt).Should(HaveTracked("dc-done")) //not dc-after
			Ω(reporter.Did.Find("A")).Should(HaveAborted("abort"))
			Ω(reporter.Did.Find("B")).Should(HaveBeenInterrupted(interrupt_handler.InterruptCauseAbortByOtherProcess))
		})

		It("does not interrupt cleanup nodes", func() {
			success := RunFixtureInParallel("aborting in parallel", func(_ int) {
				It("A", func() {
					<-c
					Abort("abort")
				})

				Context("B", func() {
					It("B", func() {
					})

					AfterEach(func(ctx SpecContext) {
						close(c)
						select {
						case <-ctx.Done():
							rt.Run("dc-done")
						case <-time.After(interrupt_handler.ABORT_POLLING_INTERVAL * 2):
							rt.Run("dc-after")
						}
					})
				})
			})
			Ω(success).Should(BeFalse())

			Ω(rt).Should(HaveTracked("dc-after")) //not dc-done
			Ω(reporter.Did.Find("A")).Should(HaveAborted("abort"))
			Ω(reporter.Did.Find("B")).Should(HavePassed())
		})

		It("does not start serial nodes if an abort occurs", func() {
			success := RunFixtureInParallel("aborting in parallel", func(proc int) {
				It("A", func() {
					time.Sleep(time.Millisecond * 50)
					if proc == 2 {
						rt.Run("aborting")
						Abort("abort")
					}
				})

				It("B", func() {
					time.Sleep(time.Millisecond * 50)
					if proc == 2 {
						rt.Run("aborting")
						Abort("abort")
					}
				})

				It("C", Serial, func() {
					rt.Run("C")
				})
			})
			Ω(success).Should(BeFalse())
			Ω(rt).Should(HaveTracked("aborting")) //just one aborting and we don't see C
		}, MustPassRepeatedly(10))

	})
})
