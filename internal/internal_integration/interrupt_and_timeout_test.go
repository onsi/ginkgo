package internal_integration_test

import (
	"context"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/internal/interrupt_handler"
	. "github.com/onsi/ginkgo/v2/internal/test_helpers"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
)

type TimeMap struct {
	m    map[string]time.Duration
	lock *sync.Mutex
}

func NewTimeMap() *TimeMap {
	return &TimeMap{
		m:    map[string]time.Duration{},
		lock: &sync.Mutex{},
	}
}

func (tm *TimeMap) Set(key string, d time.Duration) {
	tm.lock.Lock()
	defer tm.lock.Unlock()
	tm.m[key] = d
}

func (tm *TimeMap) Get(key string) time.Duration {
	tm.lock.Lock()
	defer tm.lock.Unlock()
	return tm.m[key]
}

var _ = Describe("Interrupts and Timeouts", func() {
	Describe("when it is interrupted in a BeforeSuite", func() {
		BeforeEach(func() {
			success, _ := RunFixture("interrupted test", func() {
				BeforeSuite(rt.T("before-suite", func() {
					interruptHandler.Interrupt(interrupt_handler.InterruptCauseSignal)
					time.Sleep(time.Hour)
				}))
				AfterSuite(rt.T("after-suite"))
				It("A", rt.T("A"))
				It("B", rt.T("B"))
			})
			Ω(success).Should(Equal(false))
		})

		It("runs the AfterSuite and skips all the tests", func() {
			Ω(rt).Should(HaveTracked("before-suite", "after-suite"))
			Ω(reporter.Did.FindByLeafNodeType(types.NodeTypeIt)).Should(BeZero())
		})

		It("reports the correct failure", func() {
			summary := reporter.Did.FindByLeafNodeType(types.NodeTypeBeforeSuite)
			Ω(summary.State).Should(Equal(types.SpecStateInterrupted))
			Ω(summary.Failure.Message).Should(ContainSubstring("Interrupted by User"))
			Ω(summary.Failure.ProgressReport.Message).Should(Equal("{{bold}}This is the Progress Report generated when the interrupt was received:{{/}}"))
			Ω(summary.Failure.ProgressReport.CurrentNodeType).Should(Equal(types.NodeTypeBeforeSuite))
		})

		It("emits a progress report", func() {
			Ω(reporter.ProgressReports).Should(HaveLen(1))
			pr := reporter.ProgressReports[0]
			Ω(pr.Message).Should(ContainSubstring("Interrupted by User"))
			Ω(pr.Message).Should(ContainSubstring("First interrupt received"))
			Ω(pr.CurrentNodeType).Should(Equal(types.NodeTypeBeforeSuite))
		})

		It("reports the correct statistics", func() {
			Ω(reporter.End).Should(BeASuiteSummary(false, NSpecs(2), NWillRun(2), NPassed(0), NSkipped(0), NFailed(0)))
		})

		It("reports the correct special failure reason", func() {
			Ω(reporter.End.SpecialSuiteFailureReasons).Should(ContainElement("Interrupted by User"))
		})
	})

	Describe("when aborted", func() {
		BeforeEach(func() {
			success, _ := RunFixture(CurrentSpecReport().LeafNodeText, func() {
				BeforeEach(rt.T("bef-outer"))
				Context("nested", func() {
					BeforeEach(rt.T("bef-inner"))
					It("A", rt.T("A", func() {
						interruptHandler.Interrupt(interrupt_handler.InterruptCauseAbortByOtherProcess)
						time.Sleep(time.Hour)
					}))

					It("B", rt.T("B"))
					AfterEach(rt.T("aft-inner"))
				})
				AfterEach(rt.T("aft-outer"))
			})
			Ω(success).Should(Equal(false))
		})

		It("interrupts the current spec and doesn't run any others", func() {
			Ω(rt).Should(HaveTracked("bef-outer", "bef-inner", "A", "aft-inner", "aft-outer"))
			Ω(reporter.Did.Find("A")).Should(HaveBeenInterrupted(interrupt_handler.InterruptCauseAbortByOtherProcess))
			Ω(reporter.Did.Find("B")).Should(HaveBeenSkipped())
		})

		It("does not include a progress report", func() {
			Ω(reporter.Did.Find("A").Failure.ProgressReport).Should(BeZero())
		})
	})

	Describe("when interrupted in a spec", func() {
		Context("with no SpecContext", func() {
			Context("when subsequent nodes exit right away", func() {
				BeforeEach(func() {
					success, _ := RunFixture(CurrentSpecReport().LeafNodeText, func() {
						BeforeEach(rt.T("bef-outer"))
						Context("nested", func() {
							BeforeEach(rt.T("bef-inner"))
							It("A", rt.T("A", func() {
								interruptHandler.Interrupt(interrupt_handler.InterruptCauseSignal)
								time.Sleep(time.Hour)
							}))

							It("B", rt.T("B"))
							AfterEach(rt.T("aft-inner"))
						})
						AfterEach(rt.T("aft-outer"))
					})
					Ω(success).Should(Equal(false))
				})

				It("starts cleaning up immediately and runs no other specs", func() {
					Ω(rt).Should(HaveTracked("bef-outer", "bef-inner", "A", "aft-inner", "aft-outer"))
					Ω(reporter.Did.Find("A")).Should(HaveBeenInterrupted(interrupt_handler.InterruptCauseSignal))
					Ω(reporter.Did.Find("B")).Should(HaveBeenSkipped())
				})

				It("includes a ProgressReport in the spec with a full stack trace", func() {
					Ω(reporter.Did.Find("A").Failure.ProgressReport.OtherGoroutines()).ShouldNot(BeEmpty())
				})

				It("emits a condensed ProgressReport with a shorter stack trace - note that it does not say anything about a leaked goroutine becuase the grace period is not enforced", func() {
					Ω(reporter.ProgressReports).Should(HaveLen(1))
					pr := reporter.ProgressReports[0]
					Ω(pr.Message).Should(ContainSubstring("Interrupted by User"))
					Ω(pr.Message).Should(ContainSubstring("First interrupt received"))
					Ω(pr.CurrentNodeType).Should(Equal(types.NodeTypeIt))
					Ω(pr.OtherGoroutines()).Should(BeEmpty())
				})
			})

			Context("when subsequent nodes get stuck", func() {
				BeforeEach(func(_ SpecContext) {
					conf.GracePeriod = time.Millisecond * 100
					success, _ := RunFixture(CurrentSpecReport().LeafNodeText, func() {
						BeforeEach(rt.T("bef-outer"))
						Context("nested", func() {
							BeforeEach(rt.T("bef-inner"))
							It("A", rt.T("A", func() {
								interruptHandler.Interrupt(interrupt_handler.InterruptCauseSignal)
								time.Sleep(time.Hour)
							}))

							It("B", rt.T("B"))
							AfterEach(rt.T("aft-inner", func() {
								time.Sleep(time.Hour)
							}))
						})
						AfterEach(rt.T("aft-outer"))
					})
					Ω(success).Should(Equal(false))
				}, NodeTimeout(time.Second))

				It("waits for a grace-period interval, only", func() {
					Ω(rt).Should(HaveTracked("bef-outer", "bef-inner", "A", "aft-inner", "aft-outer"))
					Ω(reporter.Did.Find("A")).Should(HaveBeenInterrupted(interrupt_handler.InterruptCauseSignal))
					Ω(reporter.Did.Find("B")).Should(HaveBeenSkipped())
					Ω(reporter.Did.Find("A").Failure.ProgressReport.CurrentNodeType).Should(Equal(types.NodeTypeIt))
					Ω(reporter.ProgressReports).Should(HaveLen(1))
				})
			})
		})

		Context("with a SpecContext", func() {
			Context("when the node exits before the grace period elapses", func() {
				BeforeEach(func(_ SpecContext) {
					success, _ := RunFixture(CurrentSpecReport().LeafNodeText, func() {
						BeforeEach(rt.T("bef-outer"))
						Context("nested", func() {
							BeforeEach(rt.T("bef-inner"))
							It("A", rt.TSC("A", func(c SpecContext) {
								interruptHandler.Interrupt(interrupt_handler.InterruptCauseSignal)
								select {
								case <-c.Done():
									F("bam")
								case <-time.After(time.Hour):
								}
							}))

							It("B", rt.T("B"))
							AfterEach(rt.T("aft-inner"))
						})
						AfterEach(rt.T("aft-outer"))
					})
					Ω(success).Should(Equal(false))
				}, NodeTimeout(time.Second))

				It("cancels the context, captures any subsequent failures, proceeds to clean up and runs no other specs; and it doesn't emit a grace-period related Progress Report", func() {
					Ω(rt).Should(HaveTracked("bef-outer", "bef-inner", "A", "aft-inner", "aft-outer"))
					Ω(reporter.Did.Find("A")).Should(HaveBeenInterrupted(interrupt_handler.InterruptCauseSignal))
					Ω(reporter.Did.Find("A").Failure.AdditionalFailure).Should(HaveFailed("bam"))

					Ω(reporter.Did.Find("B")).Should(HaveBeenSkipped())

					Ω(reporter.Did.Find("A").Failure.ProgressReport.OtherGoroutines()).ShouldNot(BeEmpty())

					Ω(reporter.ProgressReports).Should(HaveLen(1))
					pr := reporter.ProgressReports[0]
					Ω(pr.Message).Should(ContainSubstring("Interrupted by User"))
					Ω(pr.Message).Should(ContainSubstring("First interrupt received"))
					Ω(pr.CurrentNodeType).Should(Equal(types.NodeTypeIt))
					Ω(pr.OtherGoroutines()).Should(BeEmpty())
				})
			})

			Context("when the node does not exit before the grace period elapses", func() {
				BeforeEach(func(_ SpecContext) {
					success, _ := RunFixture(CurrentSpecReport().LeafNodeText, func() {
						BeforeEach(rt.T("bef-outer"))
						Context("nested", func() {
							BeforeEach(rt.T("bef-inner"))
							It("A", rt.TSC("A", func(c SpecContext) {
								interruptHandler.Interrupt(interrupt_handler.InterruptCauseSignal)
								select {
								case <-c.Done():
									time.Sleep(time.Hour)
								case <-time.After(time.Hour):
								}
							}), GracePeriod(time.Millisecond*100))

							It("B", rt.T("B"))
							AfterEach(rt.T("aft-inner"))
						})
						AfterEach(rt.T("aft-outer"))
					})
					Ω(success).Should(Equal(false))
				}, NodeTimeout(time.Second))

				It("leaks the goroutine, continues with cleanup", func() {
					Ω(rt).Should(HaveTracked("bef-outer", "bef-inner", "A", "aft-inner", "aft-outer"))
					Ω(reporter.Did.Find("A")).Should(HaveBeenInterrupted(interrupt_handler.InterruptCauseSignal))
					Ω(reporter.Did.Find("B")).Should(HaveBeenSkipped())

					Ω(reporter.Did.Find("A").Failure.ProgressReport.OtherGoroutines()).ShouldNot(BeEmpty())

					Ω(reporter.ProgressReports).Should(HaveLen(1))
					pr := reporter.ProgressReports[0]
					Ω(pr.Message).Should(ContainSubstring("Interrupted by User"))
					Ω(pr.Message).Should(ContainSubstring("First interrupt received"))
					Ω(pr.CurrentNodeType).Should(Equal(types.NodeTypeIt))
					Ω(pr.OtherGoroutines()).Should(BeEmpty())
				})
			})

			Context("when the node does not exit and the spec is interrupted again before the GracePeriod elapses", func() {
				BeforeEach(func(_ SpecContext) {
					success, _ := RunFixture(CurrentSpecReport().LeafNodeText, func() {
						BeforeEach(rt.T("bef-outer"))
						Context("nested", func() {
							BeforeEach(rt.T("bef-inner"))
							It("A", rt.TSC("A", func(c SpecContext) {
								interruptHandler.Interrupt(interrupt_handler.InterruptCauseSignal)
								select {
								case <-c.Done():
									time.Sleep(time.Millisecond * 100)
									interruptHandler.Interrupt(interrupt_handler.InterruptCauseSignal)
									time.Sleep(time.Hour)
								case <-time.After(time.Hour):
								}
							}), GracePeriod(time.Minute))

							It("B", rt.T("B"))
							AfterEach(rt.T("aft-inner"))
						})
						AfterEach(rt.T("aft-outer"))
						ReportAfterEach(func(_ SpecReport) { rt.Run("report-after-each") })
					})
					Ω(success).Should(Equal(false))
				}, NodeTimeout(time.Second))

				It("moves on past the goroutine but skips cleanup (since this is the second interrupt)", func() {
					Ω(rt).Should(HaveTracked("bef-outer", "bef-inner", "A", "report-after-each", "report-after-each"))
					Ω(reporter.Did.Find("A")).Should(HaveBeenInterrupted(interrupt_handler.InterruptCauseSignal))
					Ω(reporter.Did.Find("B")).Should(HaveBeenSkipped())

					Ω(reporter.Did.Find("A").Failure.ProgressReport.OtherGoroutines()).ShouldNot(BeEmpty())

					Ω(reporter.ProgressReports).Should(HaveLen(2))

					pr := reporter.ProgressReports[0]
					Ω(pr.Message).Should(ContainSubstring("Interrupted by User"))
					Ω(pr.Message).Should(ContainSubstring("First interrupt received"))
					Ω(pr.CurrentNodeType).Should(Equal(types.NodeTypeIt))
					Ω(pr.OtherGoroutines()).Should(BeEmpty())

					pr = reporter.ProgressReports[1]
					Ω(pr.Message).Should(ContainSubstring("Interrupted by User"))
					Ω(pr.Message).Should(ContainSubstring("Second interrupt received"))
					Ω(pr.CurrentNodeType).Should(Equal(types.NodeTypeIt))
					Ω(pr.OtherGoroutines()).Should(BeEmpty())
				})
			})

			Context("if subsequent nodes take a SpecContext but don't have a timeout", func() {
				var times *TimeMap
				BeforeEach(func(_ SpecContext) {
					times = NewTimeMap()
					success, _ := RunFixture(CurrentSpecReport().LeafNodeText, func() {
						BeforeEach(rt.T("bef-outer"))
						Context("nested", func() {
							BeforeEach(rt.T("bef-inner"))
							It("A", rt.TSC("A", func(c SpecContext) {
								interruptHandler.Interrupt(interrupt_handler.InterruptCauseSignal)
								<-c.Done()
							}))
							It("B", rt.T("B"))
							AfterEach(rt.TSC("aft-inner", func(c SpecContext) {
								t := time.Now()
								select {
								case <-c.Done():
									times.Set("aft-inner", time.Since(t))
									time.Sleep(time.Second)
								case <-time.After(time.Second):
								}
							}), GracePeriod(time.Millisecond*50))
						})
						AfterEach(rt.T("aft-outer"))
					})
					Ω(success).Should(Equal(false))
				}, NodeTimeout(time.Second))

				It("waits for a GracePeriod then interrupts, then waits for a grace period again, then leaks and lets the user know a leak occured", func() {
					Ω(rt).Should(HaveTracked("bef-outer", "bef-inner", "A", "aft-inner", "aft-outer"))
					Ω(reporter.Did.Find("A")).Should(HaveBeenInterrupted(interrupt_handler.InterruptCauseSignal))
					Ω(reporter.Did.Find("B")).Should(HaveBeenSkipped())

					Ω(reporter.Did.Find("A").Failure.ProgressReport.OtherGoroutines()).ShouldNot(BeEmpty())

					Ω(times.Get("aft-inner")).Should(BeNumerically("~", time.Millisecond*50, time.Millisecond*25))

					Ω(reporter.ProgressReports).Should(HaveLen(2))

					pr := reporter.ProgressReports[0]
					Ω(pr.Message).Should(ContainSubstring("Interrupted by User"))
					Ω(pr.Message).Should(ContainSubstring("First interrupt received"))
					Ω(pr.CurrentNodeType).Should(Equal(types.NodeTypeIt))
					Ω(pr.OtherGoroutines()).Should(BeEmpty())

					pr = reporter.ProgressReports[1]
					Ω(pr.Message).Should(ContainSubstring("A running node failed to exit in time"))
					Ω(pr.CurrentNodeType).Should(Equal(types.NodeTypeAfterEach))
					Ω(pr.OtherGoroutines()).Should(BeEmpty())
				})
			})

			Context("if subsequent nodes take a SpecContext and have a timeout", func() {
				var times *TimeMap
				BeforeEach(func(_ SpecContext) {
					times = NewTimeMap()
					success, _ := RunFixture(CurrentSpecReport().LeafNodeText, func() {
						var t time.Time
						BeforeEach(rt.T("bef-outer", func() {
							t = time.Now()
						}))
						Context("nested", func() {
							BeforeEach(rt.T("bef-inner"))
							It("A", rt.TSC("A", func(c SpecContext) {
								interruptHandler.Interrupt(interrupt_handler.InterruptCauseSignal)
								<-c.Done()
							}))
							It("B", rt.T("B"))
							AfterEach(rt.TSC("aft-inner", func(c SpecContext) {
								select {
								case <-c.Done():
									times.Set("aft-inner", time.Since(t))
									time.Sleep(time.Second)
								case <-time.After(time.Second):
								}
							}), NodeTimeout(time.Millisecond*100), GracePeriod(time.Millisecond*50))
						})
						AfterEach(rt.T("aft-outer", func() {
							times.Set("aft-outer", time.Since(t))
						}))
					})
					Ω(success).Should(Equal(false))
				}, NodeTimeout(time.Second))

				It("waits for the timeout and then the grace period", func() {
					Ω(rt).Should(HaveTracked("bef-outer", "bef-inner", "A", "aft-inner", "aft-outer"))
					Ω(reporter.Did.Find("A")).Should(HaveBeenInterrupted(interrupt_handler.InterruptCauseSignal))
					Ω(reporter.Did.Find("B")).Should(HaveBeenSkipped())
					Ω(reporter.Did.Find("A").Failure.ProgressReport.OtherGoroutines()).ShouldNot(BeEmpty())

					Ω(times.Get("aft-inner")).Should(BeNumerically("~", time.Millisecond*100, time.Millisecond*50))
					Ω(times.Get("aft-outer")).Should(BeNumerically("~", times.Get("aft-inner")+time.Millisecond*50, time.Millisecond*25))

					Ω(reporter.ProgressReports).Should(HaveLen(2))

					pr := reporter.ProgressReports[0]
					Ω(pr.Message).Should(ContainSubstring("Interrupted by User"))
					Ω(pr.Message).Should(ContainSubstring("First interrupt received"))
					Ω(pr.CurrentNodeType).Should(Equal(types.NodeTypeIt))
					Ω(pr.OtherGoroutines()).Should(BeEmpty())

					pr = reporter.ProgressReports[1]
					Ω(pr.Message).Should(ContainSubstring("A running node failed to exit in time"))
					Ω(pr.CurrentNodeType).Should(Equal(types.NodeTypeAfterEach))
					Ω(pr.OtherGoroutines()).Should(BeEmpty())
				})
			})
		})

		Context("when interrupted twice", func() {
			BeforeEach(func(_ SpecContext) {
				success, _ := RunFixture(CurrentSpecReport().LeafNodeText, func() {
					BeforeEach(rt.T("bef-outer"))
					Context("nested", func() {
						BeforeEach(rt.T("bef-inner"))
						It("A", rt.TSC("A", func(c SpecContext) {
							interruptHandler.Interrupt(interrupt_handler.InterruptCauseSignal)
							<-c.Done()
						}))
						It("B", rt.T("B"))
						AfterEach(rt.TSC("aft-inner", func(c SpecContext) {
							interruptHandler.Interrupt(interrupt_handler.InterruptCauseSignal)
							<-c.Done()
						}))
					})
					AfterEach(rt.T("aft-outer"))
					ReportAfterEach(func(_ SpecReport) { rt.Run("report-after-each") })
					ReportAfterSuite("Report After Suite", func(_ Report) { rt.Run("report-after-suite") })
					AfterSuite(rt.T("after-suite"))
				})
				Ω(success).Should(Equal(false))
			}, NodeTimeout(time.Second))

			It("bails out on future cleanup nodes, but runs reporting nodes", func() {
				Ω(rt).Should(HaveTracked("bef-outer", "bef-inner", "A", "aft-inner", "report-after-each", "report-after-each", "report-after-suite"))
				Ω(reporter.Did.Find("A")).Should(HaveBeenInterrupted(interrupt_handler.InterruptCauseSignal))
				Ω(reporter.Did.Find("B")).Should(HaveBeenSkipped())
				Ω(reporter.Did.Find("A").Failure.ProgressReport.OtherGoroutines()).ShouldNot(BeEmpty())
				Ω(reporter.Did.Find("A").Failure.ProgressReport.CurrentNodeType).Should(Equal(types.NodeTypeIt))

				Ω(reporter.ProgressReports).Should(HaveLen(2))

				pr := reporter.ProgressReports[0]
				Ω(pr.Message).Should(ContainSubstring("Interrupted by User"))
				Ω(pr.Message).Should(ContainSubstring("First interrupt received"))
				Ω(pr.CurrentNodeType).Should(Equal(types.NodeTypeIt))
				Ω(pr.OtherGoroutines()).Should(BeEmpty())

				pr = reporter.ProgressReports[1]
				Ω(pr.Message).Should(ContainSubstring("Interrupted by User"))
				Ω(pr.Message).Should(ContainSubstring("Second interrupt received"))
				Ω(pr.CurrentNodeType).Should(Equal(types.NodeTypeAfterEach))
				Ω(pr.OtherGoroutines()).Should(BeEmpty())
			})
		})

		Context("when interrupted three times", func() {
			BeforeEach(func(_ SpecContext) {
				success, _ := RunFixture(CurrentSpecReport().LeafNodeText, func() {
					BeforeEach(rt.T("bef-outer"))
					Context("nested", func() {
						BeforeEach(rt.T("bef-inner"))
						It("A", rt.TSC("A", func(c SpecContext) {
							interruptHandler.Interrupt(interrupt_handler.InterruptCauseSignal)
							<-c.Done()
						}))
						It("B", rt.T("B"))
						AfterEach(rt.TSC("aft-inner", func(c SpecContext) {
							interruptHandler.Interrupt(interrupt_handler.InterruptCauseSignal)
							<-c.Done()
						}))
					})
					AfterEach(rt.T("aft-outer"))
					ReportAfterEach(func(_ SpecReport) {
						rt.Run("report-after-each")
						interruptHandler.Interrupt(interrupt_handler.InterruptCauseSignal)
						time.Sleep(time.Hour)
					})
					ReportAfterEach(func(_ SpecReport) { rt.Run("report-after-each-2") })
					ReportAfterSuite("Report After Suite", func(_ Report) { rt.Run("report-after-suite") })
					AfterSuite(rt.T("after-suite"))
				})
				Ω(success).Should(Equal(false))
			}, NodeTimeout(time.Second))

			It("bails out on everything and just exits ASAP", func() {
				Ω(rt).Should(HaveTracked("bef-outer", "bef-inner", "A", "aft-inner", "report-after-each"))
				Ω(reporter.Did.Find("A")).Should(HaveBeenInterrupted(interrupt_handler.InterruptCauseSignal))
				Ω(reporter.Did.Find("B")).Should(HaveBeenSkipped())
				Ω(reporter.Did.Find("A").Failure.ProgressReport.OtherGoroutines()).ShouldNot(BeEmpty())
				Ω(reporter.Did.Find("A").Failure.ProgressReport.CurrentNodeType).Should(Equal(types.NodeTypeIt))

				Ω(reporter.ProgressReports).Should(HaveLen(3))

				pr := reporter.ProgressReports[0]
				Ω(pr.Message).Should(ContainSubstring("Interrupted by User"))
				Ω(pr.Message).Should(ContainSubstring("First interrupt received"))
				Ω(pr.CurrentNodeType).Should(Equal(types.NodeTypeIt))
				Ω(pr.OtherGoroutines()).Should(BeEmpty())

				pr = reporter.ProgressReports[1]
				Ω(pr.Message).Should(ContainSubstring("Interrupted by User"))
				Ω(pr.Message).Should(ContainSubstring("Second interrupt received"))
				Ω(pr.CurrentNodeType).Should(Equal(types.NodeTypeAfterEach))
				Ω(pr.OtherGoroutines()).Should(BeEmpty())

				pr = reporter.ProgressReports[2]
				Ω(pr.Message).Should(ContainSubstring("Interrupted by User"))
				Ω(pr.Message).Should(ContainSubstring("Final interrupt received"))
				Ω(pr.CurrentNodeType).Should(Equal(types.NodeTypeReportAfterEach))
				Ω(pr.OtherGoroutines()).Should(BeEmpty())
			})
		})

		Context("when interrupted in a BeforeEach", func() {
			BeforeEach(func(_ SpecContext) {
				success, _ := RunFixture(CurrentSpecReport().LeafNodeText, func() {
					BeforeEach(rt.T("bef-outer"))
					Context("nested", func() {
						BeforeEach(rt.TSC("bef-inner", func(c SpecContext) {
							interruptHandler.Interrupt(interrupt_handler.InterruptCauseSignal)
							<-c.Done()
						}))

						Context("even more nested", func() {
							BeforeEach(rt.T("bef-even-innerer"))
							It("A", rt.T("A"))
							AfterEach(rt.T("aft-even-innerer"))
							ReportAfterEach(func(_ SpecReport) { rt.Run("report-after-each-even-innerer") })
						})
						AfterEach(rt.T("aft-inner"))
						ReportAfterEach(func(_ SpecReport) { rt.Run("report-after-each-inner") })
					})
					AfterEach(rt.T("aft-outer"))
					ReportAfterEach(func(_ SpecReport) { rt.Run("report-after-each-outer") })

					ReportAfterSuite("Report After Suite", func(_ Report) { rt.Run("report-after-suite") })
					AfterSuite(rt.T("after-suite"))
				})
				Ω(success).Should(Equal(false))
			}, NodeTimeout(time.Second))

			It("only cleans up at the matching nesting level", func() {
				Ω(rt).Should(HaveTracked("bef-outer", "bef-inner", "aft-inner", "aft-outer", "report-after-each-even-innerer", "report-after-each-inner", "report-after-each-outer", "after-suite", "report-after-suite"))
				Ω(reporter.Did.Find("A")).Should(HaveBeenInterrupted(interrupt_handler.InterruptCauseSignal))
				Ω(reporter.Did.Find("A").Failure.ProgressReport.OtherGoroutines()).ShouldNot(BeEmpty())
				Ω(reporter.Did.Find("A").Failure.ProgressReport.CurrentNodeType).Should(Equal(types.NodeTypeBeforeEach))

				Ω(reporter.ProgressReports).Should(HaveLen(1))

				pr := reporter.ProgressReports[0]
				Ω(pr.Message).Should(ContainSubstring("Interrupted by User"))
				Ω(pr.Message).Should(ContainSubstring("First interrupt received"))
				Ω(pr.CurrentNodeType).Should(Equal(types.NodeTypeBeforeEach))
				Ω(pr.OtherGoroutines()).Should(BeEmpty())
			})
		})

		Context("when a timeout has already occured", func() {
			BeforeEach(func() {
				success, _ := RunFixture(CurrentSpecReport().LeafNodeText, func() {
					It("A", func(c SpecContext) {
						rt.Run("A")
						time.Sleep(time.Millisecond * 100)
						interruptHandler.Interrupt(interrupt_handler.InterruptCauseSignal)
						time.Sleep(time.Second)
					}, NodeTimeout(time.Millisecond*10))
					AfterEach(rt.T("aft-1"))
				})
				Ω(success).Should(Equal(false))
			})

			It("does not overwrite the timeout", func() {
				Ω(rt).Should(HaveTracked("A", "aft-1"))
				Ω(reporter.Did.Find("A")).Should(HaveTimedOut())
			})
		})

		Context("when a failure has already occured", func() {
			BeforeEach(func() {
				success, _ := RunFixture(CurrentSpecReport().LeafNodeText, func() {
					It("A", rt.T("A", func() {
						F("boom")
					}))
					AfterEach(rt.T("aft-1", func() {
						interruptHandler.Interrupt(interrupt_handler.InterruptCauseSignal)
					}))
					AfterEach(rt.T("aft-2"))
				})
				Ω(success).Should(Equal(false))
			})

			It("does not overwrite the failure", func() {
				Ω(rt).Should(HaveTracked("A", "aft-1", "aft-2"))
				Ω(reporter.Did.Find("A")).Should(HaveFailed("boom"))
				Ω(reporter.Did.Find("A").Failure.ProgressReport).Should(BeZero())
			})
		})
	})

	Describe("when a node times out", func() {
		var times *TimeMap

		BeforeEach(func() {
			times = NewTimeMap()

			conf.GracePeriod = time.Millisecond * 200
			success, _ := RunFixture(CurrentSpecReport().LeafNodeText, func() {
				Context("container", func() {
					var t time.Time
					BeforeEach(rt.T("bef", func() { t = time.Now() }))

					Describe("when it exits in time", func() {
						It("A", rt.TSC("A", func(c SpecContext) {
							<-c.Done()
							rt.Run("A-cancelled")
							Fail("subsequent failure message")
						}), NodeTimeout(time.Millisecond*100))
					})

					Describe("with no configured grace period", func() {
						It("B", rt.TSC("B", func(c SpecContext) {
							<-c.Done()
							time.Sleep(time.Hour)
						}), NodeTimeout(time.Millisecond*100))
					})

					Describe("with a configured grace period", func() {
						It("C", rt.TSC("C", func(c SpecContext) {
							<-c.Done()
							time.Sleep(time.Hour)
						}), NodeTimeout(time.Millisecond*100), GracePeriod(time.Millisecond*50))
					})

					It("D", rt.T("D"))
					It("E", rt.T("E", func() { F("boom") }))

					AfterEach(rt.T("aft", func() { times.Set(CurrentSpecReport().LeafNodeText, time.Since(t)) }))
				})
			})
			Ω(success).Should(Equal(false))
		})

		It("runs all the specs - a node timeout is just a failure, not an interrupt", func() {
			Ω(rt).Should(HaveTracked(
				"bef", "A", "A-cancelled", "aft",
				"bef", "B", "aft",
				"bef", "C", "aft",
				"bef", "D", "aft",
				"bef", "E", "aft",
			))

			Ω(reporter.Did.Find("A")).Should(HaveTimedOut("A node timeout occurred"))
			Ω(reporter.Did.Find("A").Failure.AdditionalFailure).Should(HaveFailed("A node timeout occurred and then the following failure was recorded in the timedout node before it exited:\nsubsequent failure message"))
			Ω(reporter.Did.Find("B")).Should(HaveTimedOut())
			Ω(reporter.Did.Find("C")).Should(HaveTimedOut())
			Ω(reporter.Did.Find("D")).Should(HavePassed())
			Ω(reporter.Did.Find("E")).Should(HaveFailed())
		})

		It("times out after the configured NodeTimeout", func() {
			Ω(times.Get("A")).Should(BeNumerically("~", 100*time.Millisecond, 50*time.Millisecond))
		})

		It("waits for the grace-period if the node doesn't exit in time", func() {
			Ω(times.Get("B")).Should(BeNumerically("~", 300*time.Millisecond, 50*time.Millisecond))
			Ω(times.Get("C")).Should(BeNumerically("~", 150*time.Millisecond, 50*time.Millisecond))
		})

		It("attaches progress reports to the timout failures", func() {
			Ω(reporter.Did.Find("A").Failure.ProgressReport.LeafNodeText).Should(Equal("A"))
			Ω(reporter.Did.Find("A").Failure.ProgressReport.Message).Should(Equal("{{bold}}This is the Progress Report generated when the node timeout occurred:{{/}}"))
			Ω(reporter.Did.Find("B").Failure.ProgressReport.LeafNodeText).Should(Equal("B"))
			Ω(reporter.Did.Find("C").Failure.ProgressReport.LeafNodeText).Should(Equal("C"))
			Ω(reporter.Did.Find("D").Failure.ProgressReport).Should(BeZero())
			Ω(reporter.Did.Find("E").Failure.ProgressReport).Should(BeZero())
		})

		It("emits progress reports when it has to leak a goroutine", func() {
			Ω(reporter.ProgressReports).Should(HaveLen(2))
			Ω(reporter.ProgressReports[0].Message).Should(ContainSubstring("A running node failed to exit in time"))
			Ω(reporter.ProgressReports[0].LeafNodeText).Should(Equal("B"))
			Ω(reporter.ProgressReports[1].Message).Should(ContainSubstring("A running node failed to exit in time"))
			Ω(reporter.ProgressReports[1].LeafNodeText).Should(Equal("C"))
		})
	})

	Describe("when a BeforeSuite/AfterSuite node times out", func() {
		var times *TimeMap
		BeforeEach(func(_ SpecContext) {
			times = NewTimeMap()
			var t time.Time
			success, _ := RunFixture(CurrentSpecReport().LeafNodeText, func() {
				SynchronizedBeforeSuite(func() []byte {
					t = time.Now()
					rt.Run("befs-proc-1")
					return []byte("befs-all-proc")
				}, func(c SpecContext, b []byte) {
					rt.Run(string(b))
					<-c.Done()
					times.Set(string(b), time.Since(t))
				}, NodeTimeout(time.Millisecond*100))

				It("A", rt.T("A"))

				SynchronizedAfterSuite(func() {
					rt.Run("afts-all-proc")
				}, func(c SpecContext) {
					rt.Run("afts-proc-1")
					<-c.Done()
					times.Set("afts-proc-1", time.Since(t))
				}, NodeTimeout(time.Millisecond*200))
			})
			Ω(success).Should(Equal(false))
		}, NodeTimeout(time.Second*5))

		It("marks the nodes as timed out", func() {
			Ω(rt).Should(HaveTracked("befs-proc-1", "befs-all-proc", "afts-all-proc", "afts-proc-1"))
			Ω(reporter.Did.FindByLeafNodeType(types.NodeTypeSynchronizedBeforeSuite)).Should(HaveTimedOut())
			Ω(reporter.Did.FindByLeafNodeType(types.NodeTypeSynchronizedBeforeSuite).Failure.ProgressReport.CurrentNodeType).Should(Equal(types.NodeTypeSynchronizedBeforeSuite))
			Ω(reporter.Did.Find("A")).Should(BeZero())
			Ω(reporter.Did.FindByLeafNodeType(types.NodeTypeSynchronizedAfterSuite)).Should(HaveTimedOut())
			Ω(reporter.Did.FindByLeafNodeType(types.NodeTypeSynchronizedAfterSuite).Failure.ProgressReport.CurrentNodeType).Should(Equal(types.NodeTypeSynchronizedAfterSuite))

			Ω(times.Get("befs-all-proc")).Should(BeNumerically("~", time.Millisecond*100, time.Millisecond*50))

			Ω(times.Get("afts-proc-1")).Should(BeNumerically("~", times.Get("befs-all-proc")+time.Millisecond*200, time.Millisecond*50))
		})
	})

	Describe("when a (spec or suite) timeout elapses and the node has no SpecContext", func() {
		var times *TimeMap
		BeforeEach(func(_ SpecContext) {
			times = NewTimeMap()

			conf.Timeout = time.Millisecond * 200
			conf.GracePeriod = time.Millisecond * 100
			success, _ := RunFixture(CurrentSpecReport().LeafNodeText, func() {
				Context("container", func() {
					var t time.Time
					BeforeEach(rt.T("bef", func() { t = time.Now() }))

					It("A", rt.T("A", func() {
						time.Sleep(time.Hour)
					}))

					AfterEach(rt.TSC("aft-1", func(c SpecContext) {
						times.Set("A", time.Since(t))
						<-c.Done()
						times.Set("aft-1-cancel", time.Since(t))
						writer.Println("aft-1")
						time.Sleep(time.Hour)
					}))

					AfterEach(rt.TSC("aft-2", func(c SpecContext) {
						times.Set("aft-1-out", time.Since(t))
						<-c.Done()
						times.Set("aft-2-cancel", time.Since(t))
						writer.Println("aft-2")
						time.Sleep(time.Hour)
					}), GracePeriod(time.Millisecond*50))

					AfterEach(rt.TSC("aft-3", func(c SpecContext) {
						times.Set("aft-2-out", time.Since(t))
						<-c.Done()
						times.Set("aft-3-cancel", time.Since(t))
						writer.Println("aft-3")
						time.Sleep(time.Hour)
					}), GracePeriod(time.Millisecond*50), NodeTimeout(time.Millisecond*150))

					AfterEach(rt.T("aft-4", func() {
						times.Set("aft-3-out", time.Since(t))
						time.Sleep(time.Hour)
					}))

					AfterEach(rt.T("aft-5", func() {
						times.Set("aft-4-out", time.Since(t))
					}))
				})
			})
			Ω(success).Should(Equal(false))
		}, NodeTimeout(time.Second*5))

		It("timesout the node in question and proceeds with other nodes without waiting; subsequent nodes are subject to the timeout if present, otherwise the grace period", func() {
			Ω(rt).Should(HaveTracked("bef", "A", "aft-1", "aft-2", "aft-3", "aft-4", "aft-5"))

			dt := 50 * time.Millisecond
			gracePeriod := 100 * time.Millisecond
			Ω(times.Get("A")).Should(BeNumerically("~", 200*time.Millisecond, dt))
			Ω(times.Get("aft-1-cancel")).Should(BeNumerically("~", times.Get("A")+gracePeriod, dt))
			Ω(times.Get("aft-1-out")).Should(BeNumerically("~", times.Get("aft-1-cancel")+gracePeriod, dt))
			Ω(times.Get("aft-2-cancel")).Should(BeNumerically("~", times.Get("aft-1-out")+50*time.Millisecond, dt))
			Ω(times.Get("aft-2-out")).Should(BeNumerically("~", times.Get("aft-2-cancel")+50*time.Millisecond, dt))
			Ω(times.Get("aft-3-cancel")).Should(BeNumerically("~", times.Get("aft-2-out")+150*time.Millisecond, dt))
			Ω(times.Get("aft-3-out")).Should(BeNumerically("~", times.Get("aft-3-cancel")+50*time.Millisecond, dt))
			Ω(times.Get("aft-4-out")).Should(BeNumerically("~", times.Get("aft-3-out")+gracePeriod, dt))

			Ω(reporter.Did.Find("A")).Should(HaveTimedOut("A suite timeout occurred"))
			Ω(reporter.Did.Find("A").Failure.ProgressReport.LeafNodeText).Should(Equal("A"))

			Ω(reporter.ProgressReports).Should(HaveLen(3))
			Ω(reporter.ProgressReports[0].Message).Should(ContainSubstring("A running node failed to exit in time"))
			Ω(reporter.ProgressReports[0].CapturedGinkgoWriterOutput).Should(Equal("aft-1\n"))

			Ω(reporter.ProgressReports[1].Message).Should(ContainSubstring("A running node failed to exit in time"))
			Ω(reporter.ProgressReports[1].CapturedGinkgoWriterOutput).Should(Equal("aft-1\naft-2\n"))

			Ω(reporter.ProgressReports[2].Message).Should(ContainSubstring("A running node failed to exit in time"))
			Ω(reporter.ProgressReports[2].CapturedGinkgoWriterOutput).Should(Equal("aft-1\naft-2\naft-3\n"))

		})
	})

	Describe("the interaction of suite, spec, and node timeouts", func() {
		var times *TimeMap
		BeforeEach(func(_ SpecContext) {
			times = NewTimeMap()

			conf.Timeout = time.Millisecond * 450
			success, _ := RunFixture(CurrentSpecReport().LeafNodeText, func() {
				Context("container", func() {
					Context("when the node timeout is shortest", func() {
						BeforeEach(rt.TSC("bef-A", func(c SpecContext) {
							t := time.Now()
							<-c.Done()
							times.Set("bef-A", time.Since(t))
						}), NodeTimeout(time.Millisecond*100))

						It("A", rt.TSC("A", func(c SpecContext) {}), SpecTimeout(time.Millisecond*200))
					})

					Context("when the spec timeout is shortest", func() {
						BeforeEach(rt.TSC("bef-B", func(c SpecContext) {
							t := time.Now()
							<-c.Done()
							times.Set("bef-B", time.Since(t))
						}), NodeTimeout(time.Millisecond*250))

						It("B", rt.TSC("B", func(c SpecContext) {}), SpecTimeout(time.Millisecond*150))
					})

					Context("when the suite timeout is the shortest", func() {
						BeforeEach(rt.TSC("bef-C", func(c SpecContext) {
							t := time.Now()
							<-c.Done()
							times.Set("bef-C", time.Since(t))
						}), NodeTimeout(time.Millisecond*300))

						It("C", rt.TSC("C", func(c SpecContext) {}), SpecTimeout(time.Second))
					})
				})
			})
			Ω(success).Should(Equal(false))
		}, NodeTimeout(time.Second*5))

		It("should always favor the shorter timeout", func() {
			Ω(rt).Should(HaveTracked("bef-A", "bef-B", "bef-C"))
			Ω(reporter.Did.Find("A")).Should(HaveTimedOut("A node timeout occurred"))
			Ω(reporter.Did.Find("B")).Should(HaveTimedOut("A spec timeout occurred"))
			Ω(reporter.Did.Find("C")).Should(HaveTimedOut("A suite timeout occurred"))

			Ω(times.Get("bef-A")).Should(BeNumerically("~", time.Millisecond*100, 50*time.Millisecond))
			Ω(times.Get("bef-B")).Should(BeNumerically("~", time.Millisecond*150, 50*time.Millisecond))
			Ω(times.Get("bef-C")).Should(BeNumerically("~", time.Millisecond*200, 50*time.Millisecond))

			Ω(reporter.End.SpecialSuiteFailureReasons).Should(Equal([]string{"Suite Timeout Elapsed"}))
		})
	})

	Describe("using timeouts with Gomega's Eventually", func() {
		BeforeEach(func(ctx SpecContext) {
			success, _ := RunFixture(CurrentSpecReport().LeafNodeText, func() {
				Context("container", func() {
					It("A", rt.TSC("A", func(c SpecContext) {
						cl = types.NewCodeLocation(0)
						Eventually(func() string { return "foo" }).WithTimeout(time.Hour).WithContext(c).Should(Equal("bar"))
						rt.Run("A-dont-see") //never see this because Eventually fails which panics and ends execution
					}), SpecTimeout(time.Millisecond*200))
				})
			})
			Ω(success).Should(Equal(false))
		}, NodeTimeout(time.Second))

		It("doesn't get stuck because Eventually will exit and it includes the additional report provided by eventually", func() {
			Ω(rt).Should(HaveTracked("A"))
			Ω(reporter.Did.Find("A")).Should(HaveTimedOut(clLine(-1)))
			Ω(reporter.Did.Find("A")).Should(HaveTimedOut(`A spec timeout occurred`))
			Ω(reporter.Did.Find("A").Failure.AdditionalFailure).Should(HaveFailed(MatchRegexp("A spec timeout occurred and then the following failure was recorded in the timedout node before it exited:\nContext was cancelled after .*\nExpected\n    <string>: foo\nto equal\n    <string>: bar"), clLine(1)))
			Ω(reporter.Did.Find("A").Failure.ProgressReport.Message).Should(Equal("{{bold}}This is the Progress Report generated when the spec timeout occurred:{{/}}"))
			Ω(reporter.Did.Find("A").Failure.ProgressReport.AdditionalReports).Should(ConsistOf("Expected\n    <string>: foo\nto equal\n    <string>: bar"))
		})
	})

	Describe("when a suite timeout occurs", func() {
		BeforeEach(func(_ SpecContext) {
			conf.Timeout = time.Millisecond * 100
			success, _ := RunFixture(CurrentSpecReport().LeafNodeText, func() {
				Context("container", func() {
					It("A", rt.TSC("A", func(c SpecContext) {
						<-c.Done()
					}))

					It("B", rt.T("B"))
					It("C", rt.TSC("C"))

					AfterEach(rt.T("aft-each"))
				})

				AfterSuite(rt.T("aft-suite"))
			})
			Ω(success).Should(Equal(false))
		}, NodeTimeout(time.Second*5))

		It("skips all subsequent specs", func() {
			Ω(rt).Should(HaveTracked("A", "aft-each", "aft-suite"))
			Ω(reporter.Did.Find("A")).Should(HaveTimedOut())
			Ω(reporter.Did.Find("B")).Should(HaveBeenSkipped())
			Ω(reporter.Did.Find("C")).Should(HaveBeenSkipped())
			Ω(reporter.End.SpecialSuiteFailureReasons).Should(Equal([]string{"Suite Timeout Elapsed"}))
		})
	})

	Describe("passing contexts to DeferCleanups", func() {
		var times *TimeMap
		BeforeEach(func() {
			times = NewTimeMap()

			success, _ := RunFixture(CurrentSpecReport().LeafNodeText, func() {
				It("A", rt.T("A", func() {
					DeferCleanup(func(c context.Context, key string) {
						rt.Run(key)
						t := time.Now()
						<-c.Done()
						times.Set(key, time.Since(t))
					}, NodeTimeout(time.Millisecond*100), "dc-1")

					DeferCleanup(func(c SpecContext, key string) {
						rt.Run(key)
						t := time.Now()
						<-c.Done()
						times.Set(key, time.Since(t))
					}, NodeTimeout(time.Millisecond*100), "dc-2")

					DeferCleanup(func(c SpecContext, d context.Context, key string) {
						key = d.Value("key").(string) + key
						rt.Run(key)
						t := time.Now()
						<-c.Done()
						times.Set(key, time.Since(t))
					}, NodeTimeout(time.Millisecond*100), context.WithValue(context.Background(), "key", "dc"), "-3")

					DeferCleanup(func(d context.Context, key string) {
						key = d.Value("key").(string) + key
						rt.Run(key)
					}, context.WithValue(context.Background(), "key", "dc"), "-4")
				}))
			})
			Ω(success).Should(Equal(false))
		})

		It("should work", func() {
			Ω(rt).Should(HaveTracked("A", "dc-4", "dc-3", "dc-2", "dc-1"))
			Ω(reporter.Did.Find("A")).Should(HaveTimedOut())
			Ω(reporter.Did.Find("A").Failure.ProgressReport.CurrentNodeType).Should(Equal(types.NodeTypeCleanupAfterEach))

			Ω(times.Get("dc-1")).Should(BeNumerically("~", 100*time.Millisecond, 50*time.Millisecond))
			Ω(times.Get("dc-2")).Should(BeNumerically("~", 100*time.Millisecond, 50*time.Millisecond))
			Ω(times.Get("dc-3")).Should(BeNumerically("~", 100*time.Millisecond, 50*time.Millisecond))
		})
	})

	Describe("passing contexts to TableEntries", func() {
		var times *TimeMap
		BeforeEach(func() {
			times = NewTimeMap()

			success, _ := RunFixture(CurrentSpecReport().LeafNodeText, func() {
				Context("container", func() {
					DescribeTable("timeout table",
						func(c SpecContext, d context.Context, key string) {
							key = d.Value("key").(string) + key
							rt.Run(CurrentSpecReport().LeafNodeText)
							t := time.Now()
							<-c.Done()
							times.Set(key, time.Since(t))
						},
						func(d context.Context, key string) string {
							key = d.Value("key").(string) + key
							return key
						},
						Entry(nil, context.WithValue(context.Background(), "key", "entry-"), "1", NodeTimeout(time.Millisecond)*100),
						Entry(nil, context.WithValue(context.Background(), "key", "entry-"), "2", SpecTimeout(time.Millisecond)*150),
					)

					DescribeTable("timeout table",
						func(c context.Context, key string) {
							rt.Run(CurrentSpecReport().LeafNodeText)
							t := time.Now()
							<-c.Done()
							times.Set(key, time.Since(t))
						},
						func(key string) string {
							return key
						},
						Entry(nil, "entry-3", NodeTimeout(time.Millisecond)*100),
						Entry(nil, "entry-4", SpecTimeout(time.Millisecond)*150),
					)

					DescribeTable("timeout table",
						func(c context.Context, key string) {
							key = c.Value("key").(string) + key
							rt.Run(CurrentSpecReport().LeafNodeText + "-" + key)
						},
						func(d context.Context, key string) string {
							key = d.Value("key").(string) + key
							return key
						},
						Entry(nil, context.WithValue(context.Background(), "key", "entry-"), "5"),
						Entry(nil, context.WithValue(context.Background(), "key", "entry-"), "6"),
					)
				})
			})
			Ω(success).Should(Equal(false))
		})

		It("should work", func() {
			Ω(rt).Should(HaveTracked("entry-1", "entry-2", "entry-3", "entry-4", "entry-5-entry-5", "entry-6-entry-6"))
			Ω(reporter.Did.Find("entry-1")).Should(HaveTimedOut())
			Ω(reporter.Did.Find("entry-2")).Should(HaveTimedOut())
			Ω(reporter.Did.Find("entry-3")).Should(HaveTimedOut())
			Ω(reporter.Did.Find("entry-4")).Should(HaveTimedOut())
			Ω(reporter.Did.Find("entry-1").Failure.ProgressReport.CurrentNodeType).Should(Equal(types.NodeTypeIt))

			Ω(times.Get("entry-1")).Should(BeNumerically("~", 100*time.Millisecond, 50*time.Millisecond))
			Ω(times.Get("entry-2")).Should(BeNumerically("~", 150*time.Millisecond, 50*time.Millisecond))
			Ω(times.Get("entry-3")).Should(BeNumerically("~", 100*time.Millisecond, 50*time.Millisecond))
			Ω(times.Get("entry-4")).Should(BeNumerically("~", 150*time.Millisecond, 50*time.Millisecond))
		})
	})
})
