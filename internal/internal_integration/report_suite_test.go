package internal_integration_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/internal/interrupt_handler"
	. "github.com/onsi/ginkgo/v2/internal/test_helpers"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
)

var _ = Describe("Sending reports to ReportBeforeSuite and ReportAfterSuite nodes", func() {
	var failInReportBeforeSuiteA, failInReportAfterSuiteA, interruptSuiteB bool
	var fixture func()

	BeforeEach(func() {
		failInReportBeforeSuiteA = false
		failInReportAfterSuiteA = false
		interruptSuiteB = false
		conf.RandomSeed = 17
		fixture = func() {
			BeforeSuite(rt.T("before-suite", func() {
				outputInterceptor.AppendInterceptedOutput("out-before-suite")
			}))
			ReportBeforeSuite(func(report Report) {
				rt.RunWithData("report-before-suite-A", "report", report)
				writer.Print("gw-report-before-suite-A")
				outputInterceptor.AppendInterceptedOutput("out-report-before-suite-A")
				if failInReportBeforeSuiteA {
					F("fail in report-before-suite-A")
				}
			})
			ReportBeforeSuite(func(report Report) {
				rt.RunWithData("report-before-suite-B", "report", report)
				writer.Print("gw-report-before-suite-B")
				outputInterceptor.AppendInterceptedOutput("out-report-before-suite-B")
			})
			Context("container", func() {
				It("A", rt.T("A"))
				It("B", rt.T("B", func() {
					F("fail in B")
				}))
				It("C", rt.T("C"))
				PIt("D", rt.T("D"))
			})
			ReportAfterSuite("Report A", func(report Report) {
				rt.RunWithData("report-after-suite-A", "report", report)
				writer.Print("gw-report-after-suite-A")
				outputInterceptor.AppendInterceptedOutput("out-report-after-suite-A")
				if failInReportAfterSuiteA {
					F("fail in report-after-suite-A")
				}
			})
			ReportAfterSuite("Report B", func(report Report) {
				if interruptSuiteB {
					interruptHandler.Interrupt(interrupt_handler.InterruptCauseSignal)
					time.Sleep(time.Hour)
				}
				rt.RunWithData("report-after-suite-B", "report", report)
				writer.Print("gw-report-after-suite-B")
				outputInterceptor.AppendInterceptedOutput("out-report-after-suite-B")
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
					"report-before-suite-A", "report-before-suite-B",
					"before-suite",
					"A", "B", "C",
					"after-suite",
					"report-after-suite-A", "report-after-suite-B",
				))
			})

			It("reports on the report procs", func() {
				Ω(reporter.Did.FindByLeafNodeType(types.NodeTypeReportBeforeSuite)).Should(HavePassed(
					types.NodeTypeReportBeforeSuite,
					CapturedGinkgoWriterOutput("gw-report-before-suite-A"),
					CapturedStdOutput("out-report-before-suite-A"),
				))

				Ω(reporter.Did.Find("Report A")).Should(HavePassed(
					types.NodeTypeReportAfterSuite,
					CapturedGinkgoWriterOutput("gw-report-after-suite-A"),
					CapturedStdOutput("out-report-after-suite-A"),
				))

				Ω(reporter.Did.Find("Report B")).Should(HavePassed(
					types.NodeTypeReportAfterSuite,
					CapturedGinkgoWriterOutput("gw-report-after-suite-B"),
					CapturedStdOutput("out-report-after-suite-B"),
				))
			})

			It("passes the report in to each ReportBeforeSuite", func() {
				reportA := rt.DataFor("report-before-suite-A")["report"].(types.Report)
				reportB := rt.DataFor("report-before-suite-B")["report"].(types.Report)

				for _, report := range []types.Report{reportA, reportB} {
					Ω(report.SuiteDescription).Should(Equal("happy-path"))
					Ω(report.SuiteSucceeded).Should(BeTrue())
					Ω(report.SuiteConfig.RandomSeed).Should(Equal(int64(17)))
					Ω(report.PreRunStats.SpecsThatWillRun).Should(Equal(3))
					Ω(report.PreRunStats.TotalSpecs).Should(Equal(4))
				}

				Ω(len(reportB.SpecReports)-len(reportA.SpecReports)).Should(Equal(1), "Report B includes the invocation of ReportAfterSuite A")
				Ω(Reports(reportB.SpecReports).FindByLeafNodeType(types.NodeTypeReportBeforeSuite)).Should(Equal(reporter.Did.FindByLeafNodeType(types.NodeTypeReportBeforeSuite)))
			})

			It("passes the report in to each ReportAfterSuite", func() {
				reportA := rt.DataFor("report-after-suite-A")["report"].(types.Report)
				reportB := rt.DataFor("report-after-suite-B")["report"].(types.Report)

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

				Ω(len(reportB.SpecReports)-len(reportA.SpecReports)).Should(Equal(1), "Report B includes the invocation of ReportAfterSuite A")
				Ω(Reports(reportB.SpecReports).Find("Report A")).Should(Equal(reporter.Did.Find("Report A")))
			})
		})

		Context("when a ReportBeforeSuite node fails", func() {
			BeforeEach(func() {
				failInReportBeforeSuiteA = true
				success, _ := RunFixture("report-before-suite-A-fails", fixture)
				Ω(success).Should(BeFalse())
			})

			It("doesn't run any specs - just reporting functions", func() {
				Ω(rt).Should(HaveTracked(
					"report-before-suite-A", "report-before-suite-B",
					"report-after-suite-A", "report-after-suite-B",
				))
			})

			It("reports on the failure, to Ginkgo's reporter and any subsequent reporters", func() {
				Ω(reporter.Did.FindByLeafNodeType(types.NodeTypeReportBeforeSuite)).Should(HaveFailed(
					types.NodeTypeReportBeforeSuite,
					"fail in report-before-suite-A",
					CapturedGinkgoWriterOutput("gw-report-before-suite-A"),
					CapturedStdOutput("out-report-before-suite-A"),
				))

				reportB := rt.DataFor("report-before-suite-B")["report"].(types.Report)
				Ω(Reports(reportB.SpecReports).FindByLeafNodeType(types.NodeTypeReportBeforeSuite)).Should(Equal(reporter.Did.FindByLeafNodeType(types.NodeTypeReportBeforeSuite)))
			})
		})

		Context("when a ReportAfterSuite node fails", func() {
			BeforeEach(func() {
				failInReportAfterSuiteA = true
				success, _ := RunFixture("report-after-suite-A-fails", fixture)
				Ω(success).Should(BeFalse())
			})

			It("keeps running subseuqent reporting functions", func() {
				Ω(rt).Should(HaveTracked(
					"report-before-suite-A", "report-before-suite-B",
					"before-suite",
					"A", "B", "C",
					"after-suite",
					"report-after-suite-A", "report-after-suite-B",
				))
			})

			It("reports on the failure, to Ginkgo's reporter and any subsequent reporters", func() {
				Ω(reporter.Did.Find("Report A")).Should(HaveFailed(
					types.NodeTypeReportAfterSuite,
					"fail in report-after-suite-A",
					CapturedGinkgoWriterOutput("gw-report-after-suite-A"),
					CapturedStdOutput("out-report-after-suite-A"),
				))

				reportB := rt.DataFor("report-after-suite-B")["report"].(types.Report)
				Ω(Reports(reportB.SpecReports).Find("Report A")).Should(Equal(reporter.Did.Find("Report A")))
			})
		})

		Context("when an interrupt is attempted in a ReportAfterSuiteNode", func() {
			BeforeEach(func() {
				interruptSuiteB = true
				success, _ := RunFixture("report-after-suite-B-interrupted", fixture)
				Ω(success).Should(BeFalse())
			})

			It("interrupts and bails", func() {
				Ω(rt).Should(HaveTracked(
					"report-before-suite-A", "report-before-suite-B",
					"before-suite",
					"A", "B", "C",
					"after-suite",
					"report-after-suite-A",
				))
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
						"report-before-suite-A", "report-before-suite-B",
						"before-suite",
						"A", "B", "C",
						"after-suite",
						"report-after-suite-A", "report-after-suite-B",
					))
				})

				It("passes the report in to each reporter, including information from other procs", func() {
					reportA := rt.DataFor("report-after-suite-A")["report"].(types.Report)
					reportB := rt.DataFor("report-after-suite-B")["report"].(types.Report)

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

					Ω(len(reportB.SpecReports)-len(reportA.SpecReports)).Should(Equal(1), "Report B includes the invocation of ReportAfterSuite A")
					Ω(Reports(reportB.SpecReports).Find("Report A")).Should(Equal(reporter.Did.Find("Report A")))
				})

				It("tells the other procs that the ReportBeforeSuite has completed", func() {
					Ω(client.BlockUntilReportBeforeSuiteCompleted()).Should(Equal(types.SpecStatePassed))
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
						"report-before-suite-A", "report-before-suite-B",
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

			Context("when a ReportBeforeSuite fails", func() {
				BeforeEach(func() {
					// proc 2 has reported back and exited
					client.PostSuiteDidEnd(otherNodeReport)
					close(exitChannels[2])
					failInReportBeforeSuiteA = true
					success, _ := RunFixture("failure-in-report-before-suite-A", fixture)
					Ω(success).Should(BeFalse())
				})

				It("only runs the reporting nodes", func() {
					Ω(rt).Should(HaveTracked(
						"report-before-suite-A", "report-before-suite-B",
						"report-after-suite-A", "report-after-suite-B",
					))
				})

				It("tells the other procs that the ReportBeforeSuite failed", func() {
					Ω(client.BlockUntilReportBeforeSuiteCompleted()).Should(Equal(types.SpecStateFailed))
				})
			})
		})

		Context("on a non-primary proc", func() {
			var done chan interface{}
			BeforeEach(func() {
				done = make(chan interface{})
				go func() {
					conf.ParallelProcess = 2
					success, _ := RunFixture("non-primary proc", fixture)
					Ω(success).Should(BeFalse())
					close(done)
				}()
				Consistently(done).ShouldNot(BeClosed())

				Ω(rt).Should(HaveTrackedNothing(), "Nothing should run until we are cleared to go by proc1")
			})

			Context("the happy path", func() {
				BeforeEach(func() {
					// proc1 signals that its ReportBeforeSuites succeeded
					client.PostReportBeforeSuiteCompleted(types.SpecStatePassed)
					Eventually(done).Should(BeClosed())
				})

				It("does not run anything until the primary node finishes running the BeforeSuite node, then it runs the specs", func() {
					Ω(rt).Should(HaveTracked(
						"before-suite",
						"A", "B", "C",
						"after-suite",
					))
				})
			})

			Context("when the ReportBeforeSuite node fails", func() {
				BeforeEach(func() {
					// proc1 signals that its ReportBeforeSuites failed
					client.PostReportBeforeSuiteCompleted(types.SpecStateFailed)
					Eventually(done).Should(BeClosed())
				})

				It("does not run anything until it is notified of the failure, then it just exits without running anything", func() {
					Ω(rt).Should(HaveTrackedNothing())
				})
			})

			Context("when proc1 exits before reporting", func() {
				BeforeEach(func() {
					// proc1 signals that its ReportBeforeSuites failed
					close(exitChannels[1])
					Eventually(done).Should(BeClosed())
				})

				It("does not run anything until it is notified of the failure, then it just exits without running anything", func() {
					Ω(rt).Should(HaveTrackedNothing())
				})
			})
		})
	})
})
