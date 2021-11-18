package internal_integration_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/internal/interrupt_handler"
	. "github.com/onsi/ginkgo/v2/internal/test_helpers"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
)

var _ = Describe("Sending reports to ReportAfterSuite procs", func() {
	var failInReportAfterSuiteA, interruptSuiteB bool
	var fixture func()

	BeforeEach(func() {
		failInReportAfterSuiteA = false
		interruptSuiteB = false
		conf.RandomSeed = 17
		fixture = func() {
			BeforeSuite(rt.T("before-suite", func() {
				outputInterceptor.AppendInterceptedOutput("out-before-suite")
			}))
			Context("container", func() {
				It("A", rt.T("A"))
				It("B", rt.T("B", func() {
					F("fail in B")
				}))
				It("C", rt.T("C"))
				PIt("D", rt.T("D"))
			})
			ReportAfterSuite("Report A", func(report Report) {
				rt.RunWithData("report-A", "report", report)
				writer.Print("gw-report-A")
				outputInterceptor.AppendInterceptedOutput("out-report-A")
				if failInReportAfterSuiteA {
					F("fail in report-A")
				}
			})
			ReportAfterSuite("Report B", func(report Report) {
				if interruptSuiteB {
					interruptHandler.Interrupt(interrupt_handler.InterruptCauseTimeout)
					time.Sleep(100 * time.Millisecond)
				}
				rt.RunWithData("report-B", "report", report, "emitted-interrupt", interruptHandler.EmittedInterruptPlaceholderMessage())
				writer.Print("gw-report-B")
				outputInterceptor.AppendInterceptedOutput("out-report-B")
			})
			AfterSuite(rt.T("after-suite", func() {
				writer.Print("gw-after-suite")
				F("fail in after-suite")
			}))

		}
	})

	Context("when running in series", func() {
		BeforeEach(func() {
			conf.ParallelTotal = 1
			conf.ParallelProcess = 1
		})

		Context("the happy path", func() {
			BeforeEach(func() {
				success, _ := RunFixture("happy-path", fixture)
				Ω(success).Should(BeFalse())
			})

			It("runs all the functions", func() {
				Ω(rt).Should(HaveTracked(
					"before-suite",
					"A", "B", "C",
					"after-suite",
					"report-A", "report-B",
				))
			})

			It("reports on the report procs", func() {
				Ω(reporter.Did.Find("Report A")).Should(HavePassed(
					types.NodeTypeReportAfterSuite,
					CapturedGinkgoWriterOutput("gw-report-A"),
					CapturedStdOutput("out-report-A"),
				))

				Ω(reporter.Did.Find("Report B")).Should(HavePassed(
					types.NodeTypeReportAfterSuite,
					CapturedGinkgoWriterOutput("gw-report-B"),
					CapturedStdOutput("out-report-B"),
				))
			})

			It("passes the report in to each reporter", func() {
				reportA := rt.DataFor("report-A")["report"].(types.Report)
				reportB := rt.DataFor("report-B")["report"].(types.Report)

				for _, report := range []types.Report{reportA, reportB} {
					Ω(report.SuiteDescription).Should(Equal("happy-path"))
					Ω(report.SuiteSucceeded).Should(BeFalse())
					Ω(report.SuiteConfig.RandomSeed).Should(Equal(int64(17)))
					reports := Reports(report.SpecReports)
					Ω(reports.FindByLeafNodeType(types.NodeTypeBeforeSuite)).Should(HavePassed(CapturedStdOutput("out-before-suite")))
					Ω(reports.Find("A")).Should(HavePassed())
					Ω(reports.Find("B")).Should(HaveFailed("fail in B"))
					Ω(reports.Find("C")).Should(HavePassed())
					Ω(reports.Find("D")).Should(BePending())
					Ω(reports.FindByLeafNodeType(types.NodeTypeAfterSuite)).Should(HaveFailed("fail in after-suite", CapturedGinkgoWriterOutput("gw-after-suite")))
				}

				Ω(len(reportB.SpecReports)-len(reportA.SpecReports)).Should(Equal(1), "Report B includes the invocation of ReporteAfterSuite A")
				Ω(Reports(reportB.SpecReports).Find("Report A")).Should(Equal(reporter.Did.Find("Report A")))
			})
		})

		Context("when a ReportAfterSuite proc fails", func() {
			BeforeEach(func() {
				failInReportAfterSuiteA = true
				success, _ := RunFixture("report-A-fails", fixture)
				Ω(success).Should(BeFalse())
			})

			It("keeps running subseuqent reporting functions", func() {
				Ω(rt).Should(HaveTracked(
					"before-suite",
					"A", "B", "C",
					"after-suite",
					"report-A", "report-B",
				))
			})

			It("reports on the faitlure, to Ginkgo's reporter and any subsequent reporters", func() {
				Ω(reporter.Did.Find("Report A")).Should(HaveFailed(
					types.NodeTypeReportAfterSuite,
					"fail in report-A",
					CapturedGinkgoWriterOutput("gw-report-A"),
					CapturedStdOutput("out-report-A"),
				))

				reportB := rt.DataFor("report-B")["report"].(types.Report)
				Ω(Reports(reportB.SpecReports).Find("Report A")).Should(Equal(reporter.Did.Find("Report A")))
			})
		})

		Context("when an interrupt is attempted in a ReportAfterSuiteNode", func() {
			BeforeEach(func() {
				interruptSuiteB = true
				success, _ := RunFixture("report-B-interrupted", fixture)
				Ω(success).Should(BeFalse())
			})

			It("ignores the interrupt and soliders on", func() {
				Ω(rt).Should(HaveTracked(
					"before-suite",
					"A", "B", "C",
					"after-suite",
					"report-A", "report-B",
				))

				Ω(rt.DataFor("report-B")["report"]).ShouldNot(BeZero())
				Ω(rt.DataFor("report-B")["emitted-interrupt"]).Should(ContainSubstring("The running ReportAfterSuite node is at:\n%s", reporter.Did.Find("Report B").LeafNodeLocation.FileName))
			})
		})
	})

	Context("when running in parallel", func() {
		var otherNodeReport types.Report

		BeforeEach(func() {
			SetUpForParallel(2)

			otherNodeReport = types.Report{
				SpecReports: types.SpecReports{
					types.SpecReport{LeafNodeText: "E", LeafNodeLocation: cl, State: types.SpecStatePassed, LeafNodeType: types.NodeTypeIt},
					types.SpecReport{LeafNodeText: "F", LeafNodeLocation: cl, State: types.SpecStateSkipped, LeafNodeType: types.NodeTypeIt},
				},
			}
		})

		Context("on proc 1", func() {
			BeforeEach(func() {
				conf.ParallelProcess = 1
			})

			Context("the happy path", func() {
				BeforeEach(func() {
					// proc 2 has reported back and exited
					client.PostSuiteDidEnd(otherNodeReport)
					close(exitChannels[2])
					success, _ := RunFixture("happy-path", fixture)
					Ω(success).Should(BeFalse())
				})

				It("runs all the functions", func() {
					Ω(rt).Should(HaveTracked(
						"before-suite",
						"A", "B", "C",
						"after-suite",
						"report-A", "report-B",
					))
				})

				It("passes the report in to each reporter, including information from other procs", func() {
					reportA := rt.DataFor("report-A")["report"].(types.Report)
					reportB := rt.DataFor("report-B")["report"].(types.Report)

					for _, report := range []types.Report{reportA, reportB} {
						Ω(report.SuiteDescription).Should(Equal("happy-path"))
						Ω(report.SuiteSucceeded).Should(BeFalse())
						reports := Reports(report.SpecReports)
						Ω(reports.FindByLeafNodeType(types.NodeTypeBeforeSuite)).Should(HavePassed(CapturedStdOutput("out-before-suite")))
						Ω(reports.Find("A")).Should(HavePassed())
						Ω(reports.Find("B")).Should(HaveFailed("fail in B"))
						Ω(reports.Find("C")).Should(HavePassed())
						Ω(reports.Find("D")).Should(BePending())
						Ω(reports.Find("E")).Should(HavePassed())
						Ω(reports.Find("F")).Should(HaveBeenSkipped())
						Ω(reports.FindByLeafNodeType(types.NodeTypeAfterSuite)).Should(HaveFailed("fail in after-suite", CapturedGinkgoWriterOutput("gw-after-suite")))
					}

					Ω(len(reportB.SpecReports)-len(reportA.SpecReports)).Should(Equal(1), "Report B includes the invocation of ReporteAfterSuite A")
					Ω(Reports(reportB.SpecReports).Find("Report A")).Should(Equal(reporter.Did.Find("Report A")))
				})
			})

			Describe("waiting for reports from other procs", func() {
				It("blocks until the other procs have finished", func() {
					done := make(chan interface{})
					go func() {
						defer GinkgoRecover()
						success, _ := RunFixture("happy-path", fixture)
						Ω(success).Should(BeFalse())
						close(done)
					}()
					Consistently(done).ShouldNot(BeClosed())
					client.PostSuiteDidEnd(otherNodeReport)
					Consistently(done).ShouldNot(BeClosed())
					close(exitChannels[2])
					Eventually(done).Should(BeClosed())
				})
			})

			Context("when a non-primary proc disappears before it reports", func() {
				BeforeEach(func() {
					close(exitChannels[2]) //proc 2 disappears before reporting
					success, _ := RunFixture("disappearing-proc-2", fixture)
					Ω(success).Should(BeFalse())
				})

				It("does not run the ReportAfterSuite procs", func() {
					Ω(rt).Should(HaveTracked(
						"before-suite",
						"A", "B", "C",
						"after-suite",
					))
				})

				It("reports all the ReportAfterSuite procs as failed", func() {
					Ω(reporter.Did.Find("Report A")).Should(HaveFailed(types.GinkgoErrors.AggregatedReportUnavailableDueToNodeDisappearing().Error()))
					Ω(reporter.Did.Find("Report B")).Should(HaveFailed(types.GinkgoErrors.AggregatedReportUnavailableDueToNodeDisappearing().Error()))
				})
			})
		})

		Context("on a non-primary proc", func() {
			BeforeEach(func() {
				conf.ParallelProcess = 2
				success, _ := RunFixture("happy-path", fixture)
				Ω(success).Should(BeFalse())
			})

			It("does not run the ReportAfterSuite procs", func() {
				Ω(rt).Should(HaveTracked(
					"before-suite",
					"A", "B", "C",
					"after-suite",
				))
			})
		})
	})
})
