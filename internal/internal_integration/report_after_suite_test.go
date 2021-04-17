package internal_integration_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/internal/parallel_support"
	. "github.com/onsi/ginkgo/internal/test_helpers"
	"github.com/onsi/ginkgo/types"
	. "github.com/onsi/gomega"
)

var _ = Describe("Sending reports to ReportAfterSuite nodes", func() {
	var failInReportAfterSuiteA, interruptSuiteB bool
	var fixture func()

	BeforeEach(func() {
		failInReportAfterSuiteA = false
		interruptSuiteB = false
		conf.RandomSeed = 17
		fixture = func() {
			BeforeSuite(rt.T("before-suite", func() {
				outputInterceptor.InterceptedOutput = "out-before-suite"
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
				outputInterceptor.InterceptedOutput = "out-report-A"
				if failInReportAfterSuiteA {
					F("fail in report-A")
				}
			})
			ReportAfterSuite("Report B", func(report Report) {
				if interruptSuiteB {
					interruptHandler.Interrupt()
					time.Sleep(100 * time.Millisecond)
				}
				rt.RunWithData("report-B", "report", report, "emitted-interrupt", interruptHandler.EmittedInterruptMessage())
				writer.Print("gw-report-B")
				outputInterceptor.InterceptedOutput = "out-report-B"
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
			conf.ParallelNode = 1
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

			It("reports on the report nodes", func() {
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

		Context("when a ReportAfterSuite node fails", func() {
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
		var server *parallel_support.Server
		var client parallel_support.Client
		var exitChannels map[int]chan interface{}
		var otherNodeReport types.Report

		BeforeEach(func() {
			conf.ParallelTotal = 2
			server, client, exitChannels = SetUpServerAndClient(conf.ParallelTotal)
			conf.ParallelHost = server.Address()

			otherNodeReport = types.Report{
				SpecReports: types.SpecReports{
					types.SpecReport{NodeTexts: []string{"E"}, NodeLocations: []types.CodeLocation{cl}, State: types.SpecStatePassed, LeafNodeType: types.NodeTypeIt, LeafNodeLocation: cl},
					types.SpecReport{NodeTexts: []string{"F"}, NodeLocations: []types.CodeLocation{cl}, State: types.SpecStateSkipped, LeafNodeType: types.NodeTypeIt, LeafNodeLocation: cl},
				},
			}
		})

		AfterEach(func() {
			server.Close()
		})

		Context("on node 1", func() {
			BeforeEach(func() {
				conf.ParallelNode = 1
			})

			Context("the happy path", func() {
				BeforeEach(func() {
					// node 2 has reported back and exited
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

				It("passes the report in to each reporter, including information from other nodes", func() {
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

			Describe("waiting for reports from other nodes", func() {
				It("blocks until the other nodes have finished", func() {
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

			Context("when a non-primary node disappears before it reports", func() {
				BeforeEach(func() {
					close(exitChannels[2]) //node 2 disappears before reporting
					success, _ := RunFixture("disappearing-node-2", fixture)
					Ω(success).Should(BeFalse())
				})

				It("does not run the ReportAfterSuite nodes", func() {
					Ω(rt).Should(HaveTracked(
						"before-suite",
						"A", "B", "C",
						"after-suite",
					))
				})

				It("reports all the ReportAfterSuite nodes as failed", func() {
					Ω(reporter.Did.Find("Report A")).Should(HaveFailed(types.GinkgoErrors.AggregatedReportUnavailableDueToNodeDisappearing().Error()))
					Ω(reporter.Did.Find("Report B")).Should(HaveFailed(types.GinkgoErrors.AggregatedReportUnavailableDueToNodeDisappearing().Error()))
				})
			})
		})

		Context("on a non-primary node", func() {
			BeforeEach(func() {
				conf.ParallelNode = 2
				success, _ := RunFixture("happy-path", fixture)
				Ω(success).Should(BeFalse())
			})

			It("does not run the ReportAfterSuite nodes", func() {
				Ω(rt).Should(HaveTracked(
					"before-suite",
					"A", "B", "C",
					"after-suite",
				))
			})
		})
	})
})
