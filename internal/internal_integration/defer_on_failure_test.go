package internal_integration_test

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/internal/interrupt_handler"
	. "github.com/onsi/ginkgo/v2/internal/test_helpers"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
)

var _ = Describe("DeferOnFailure", func() {
	// OF registers an on-failure callback that tracks its label (exercising DeferCleanup-style argument passing)
	OF := func(label string) func() {
		return func() {
			DeferOnFailure(rt.Run, label)
		}
	}
	// C registers a cleanup that tracks its label
	C := func(label string) func() {
		return func() {
			DeferCleanup(rt.Run, label)
		}
	}
	// both runs several registration functions in order
	both := func(fs ...func()) func() {
		return func() {
			for _, f := range fs {
				f()
			}
		}
	}

	Describe("when it runs", func() {
		Context("when the It fails", func() {
			BeforeEach(func() {
				success, _ := RunFixture("it fails", func() {
					Describe("top", func() {
						BeforeEach(rt.T("BE-outer", both(OF("OF-outer"), C("C-outer"))))
						Context("inner", func() {
							BeforeEach(rt.T("BE-inner", OF("OF-inner")))
							It("A", rt.T("A", both(C("C-A"), func() { F("boom") })))
							JustAfterEach(rt.T("JAE"))
							AfterEach(rt.T("AE"))
						})
						It("B", rt.T("B"))
					})
				})
				Ω(success).Should(BeFalse())
			})

			It("runs the callbacks newest-first, before any teardown, and only for the failing spec", func() {
				Ω(rt).Should(HaveTracked(
					"BE-outer", "BE-inner", "A", "OF-inner", "OF-outer", "JAE", "AE", "C-A", "C-outer",
					"BE-outer", "B", "C-outer",
				))
				Ω(reporter.Did.Find("A")).Should(HaveFailed("boom"))
				Ω(reporter.Did.Find("B")).Should(HavePassed())
			})

			It("records the callbacks in the spec's timeline", func() {
				var onFailureEvents []types.SpecEvent
				for _, event := range reporter.Did.Find("A").SpecEvents {
					if event.SpecEventType == types.SpecEventNodeStart && event.NodeType == types.NodeTypeOnFailureEach {
						onFailureEvents = append(onFailureEvents, event)
					}
				}
				Ω(onFailureEvents).Should(HaveLen(2))
				for _, event := range onFailureEvents {
					// the registration's code location points at the user's call to DeferOnFailure, not into Ginkgo
					Ω(event.CodeLocation.FileName).Should(HaveSuffix("defer_on_failure_test.go"))
				}
			})
		})

		Context("when a setup node fails", func() {
			BeforeEach(func() {
				success, _ := RunFixture("setup fails", func() {
					Describe("top", func() {
						BeforeEach(rt.T("BE-1", OF("OF-1")))
						BeforeEach(rt.T("BE-2", both(OF("OF-2"), func() { F("boom") })))
						It("A", rt.T("A"))
						AfterEach(rt.T("AE"))
					})
				})
				Ω(success).Should(BeFalse())
			})

			It("runs every callback registered so far, including the failing node's own, before teardown", func() {
				Ω(rt).Should(HaveTracked("BE-1", "BE-2", "OF-2", "OF-1", "AE"))
			})
		})

		Context("when the first failure happens in an AfterEach", func() {
			BeforeEach(func() {
				success, _ := RunFixture("aftereach fails", func() {
					Describe("top", func() {
						BeforeEach(rt.T("BE", both(OF("OF"), C("C-BE"))))
						AfterEach(rt.T("AE-outer"))
						Context("inner", func() {
							It("A", rt.T("A"))
							JustAfterEach(rt.T("JAE"))
							AfterEach(rt.T("AE-inner", func() { F("boom") }))
						})
					})
				})
				Ω(success).Should(BeFalse())
			})

			It("runs the callbacks right after the failing node, before the rest of teardown", func() {
				Ω(rt).Should(HaveTracked("BE", "A", "JAE", "AE-inner", "OF", "AE-outer", "C-BE"))
			})
		})

		Context("when the first failure happens in a DeferCleanup", func() {
			BeforeEach(func() {
				success, _ := RunFixture("cleanup fails", func() {
					BeforeEach(rt.T("BE", both(OF("OF"), C("C-1"), func() {
						DeferCleanup(func() {
							rt.Run("C-2")
							F("boom")
						})
					})))
					It("A", rt.T("A"))
				})
				Ω(success).Should(BeFalse())
			})

			It("runs the callbacks before the remaining cleanups", func() {
				Ω(rt).Should(HaveTracked("BE", "A", "C-2", "OF", "C-1"))
			})
		})

		Context("when the first failure happens in an AfterAll", func() {
			BeforeEach(func() {
				success, _ := RunFixture("afterall fails", func() {
					Describe("ordered", Ordered, func() {
						BeforeAll(rt.T("BA", both(OF("OF-BA"), C("C-BA"))))
						It("A", rt.T("A"))
						It("B", rt.T("B"))
						AfterAll(rt.T("AA", func() { F("boom") }))
					})
				})
				Ω(success).Should(BeFalse())
			})

			It("runs the callbacks before the remaining cleanups", func() {
				Ω(rt).Should(HaveTracked("BA", "A", "B", "AA", "OF-BA", "C-BA"))
			})
		})

		Context("when the spec fails more than once", func() {
			BeforeEach(func() {
				success, _ := RunFixture("multiple failures", func() {
					BeforeEach(rt.T("BE", OF("OF")))
					It("A", rt.T("A", func() { F("boom") }))
					AfterEach(rt.T("AE-1", func() { F("boom-1") }))
					AfterEach(rt.T("AE-2", func() { F("boom-2") }))
				})
				Ω(success).Should(BeFalse())
			})

			It("runs the callbacks only once, for the first failure", func() {
				Ω(rt).Should(HaveTracked("BE", "A", "OF", "AE-1", "AE-2"))
				report := reporter.Did.Find("A")
				Ω(report).Should(HaveFailed("boom"))
				Ω(report.AdditionalFailures).Should(HaveLen(2))
			})
		})

		Context("when the spec passes or is skipped", func() {
			BeforeEach(func() {
				success, _ := RunFixture("no failure", func() {
					Describe("top", func() {
						BeforeEach(rt.T("BE", OF("OF")))
						It("A", rt.T("A"))
						It("B", rt.T("B", func() { Skip("skip") }))
						It("C", rt.T("C"))
						AfterEach(rt.T("AE"))
					})
				})
				Ω(success).Should(BeTrue())
			})

			It("never runs the callbacks", func() {
				Ω(rt).Should(HaveTracked("BE", "A", "AE", "BE", "B", "AE", "BE", "C", "AE"))
			})
		})
	})

	Describe("scope", func() {
		Context("when registered in an It, BeforeEach or JustBeforeEach", func() {
			BeforeEach(func() {
				success, _ := RunFixture("each scope", func() {
					Describe("top", func() {
						JustBeforeEach(rt.T("JBE", OF("OF-JBE")))
						It("A", rt.T("A", OF("OF-A")))
						It("B", rt.T("B", func() { F("boom") }))
					})
				})
				Ω(success).Should(BeFalse())
			})

			It("covers only the current spec", func() {
				Ω(rt).Should(HaveTracked("JBE", "A", "JBE", "B", "OF-JBE"))
			})
		})

		Context("when registered in a BeforeAll", func() {
			BeforeEach(func() {
				success, _ := RunFixture("all scope", func() {
					Describe("ordered", Ordered, func() {
						BeforeAll(rt.T("BA", OF("OF-BA")))
						It("A", rt.T("A"))
						It("B", rt.T("B"))
						It("C", rt.T("C", func() { F("boom") }))
					})
					Describe("after", func() {
						It("D", rt.T("D", func() { F("boom") }))
					})
				})
				Ω(success).Should(BeFalse())
			})

			It("covers every spec in the ordered container, and nothing after it", func() {
				Ω(rt.TrackedRuns()).Should(ContainElements("BA", "A", "B", "C", "OF-BA", "D"))
				Ω(rt.TrackedRuns()).Should(ContainElement("OF-BA"))
				runs := rt.TrackedRuns()
				count := 0
				for i, run := range runs {
					if run == "OF-BA" {
						count++
						Ω(runs[i-1]).Should(Equal("C"))
					}
				}
				Ω(count).Should(Equal(1))
			})
		})

		Context("when registered in a BeforeEach marked OncePerOrdered", func() {
			BeforeEach(func() {
				success, _ := RunFixture("once per ordered scope", func() {
					Describe("top", func() {
						BeforeEach(rt.T("BE", OF("OF-BE")), OncePerOrdered)
						Context("ordered", Ordered, func() {
							It("A", rt.T("A"))
							It("B", rt.T("B"))
							It("C", rt.T("C", func() { F("boom") }))
						})
					})
				})
				Ω(success).Should(BeFalse())
			})

			It("covers every spec in the ordered container", func() {
				Ω(rt).Should(HaveTracked("BE", "A", "B", "C", "OF-BE"))
			})
		})

		Context("when an ordered container is configured to ContinueOnFailure", func() {
			BeforeEach(func() {
				success, _ := RunFixture("continue on failure", func() {
					Describe("ordered", Ordered, ContinueOnFailure, func() {
						BeforeAll(rt.T("BA", OF("OF-BA")))
						It("A", rt.T("A", func() { F("boom-A") }))
						It("B", rt.T("B"))
						It("C", rt.T("C", func() { F("boom-C") }))
					})
				})
				Ω(success).Should(BeFalse())
			})

			It("fires once for each failing spec", func() {
				Ω(rt).Should(HaveTracked("BA", "A", "OF-BA", "B", "C", "OF-BA"))
			})
		})

		Context("when registered in a BeforeSuite", func() {
			BeforeEach(func() {
				success, _ := RunFixture("suite scope", func() {
					BeforeSuite(rt.T("BS", OF("OF-BS")))
					Describe("top", func() {
						BeforeEach(rt.T("BE", OF("OF-BE")))
						It("A", rt.T("A"))
						It("B", rt.T("B", func() { F("boom") }))
						It("C", rt.T("C", func() { F("boom") }))
					})
					AfterSuite(rt.T("AS", func() { F("boom") }))
				})
				Ω(success).Should(BeFalse())
			})

			It("covers every spec, runs after narrower-scoped callbacks, and covers suite-level nodes", func() {
				Ω(rt).Should(HaveTracked(
					"BS",
					"BE", "A",
					"BE", "B", "OF-BE", "OF-BS",
					"BE", "C", "OF-BE", "OF-BS",
					"AS", "OF-BS",
				))
			})
		})

		Context("when registered in the all-processes function of a SynchronizedBeforeSuite", func() {
			BeforeEach(func() {
				success, _ := RunFixture("synchronized suite scope", func() {
					SynchronizedBeforeSuite(func() []byte {
						rt.Run("SBS-1")
						return nil
					}, func(_ []byte) {
						rt.Run("SBS-all")
						OF("OF-SBS")()
					})
					It("A", rt.T("A", func() { F("boom") }))
				})
				Ω(success).Should(BeFalse())
			})

			It("covers every spec", func() {
				Ω(rt).Should(HaveTracked("SBS-1", "SBS-all", "A", "OF-SBS"))
			})
		})

		Context("when the all-processes function of a SynchronizedAfterSuite fails", func() {
			BeforeEach(func() {
				success, _ := RunFixture("synchronized after suite fails", func() {
					BeforeSuite(OF("OF-BS"))
					It("A", rt.T("A"))
					SynchronizedAfterSuite(func() {
						rt.Run("SAS-all")
						F("boom")
					}, func() {
						rt.Run("SAS-1")
					})
				})
				Ω(success).Should(BeFalse())
			})

			It("runs the callbacks before process #1's teardown", func() {
				Ω(rt).Should(HaveTracked("A", "SAS-all", "OF-BS", "SAS-1"))
			})
		})

		Context("when a BeforeSuite fails after registering a callback", func() {
			BeforeEach(func() {
				success, _ := RunFixture("before suite fails", func() {
					BeforeSuite(rt.T("BS", both(OF("OF-BS"), func() { F("boom") })))
					It("A", rt.T("A"))
				})
				Ω(success).Should(BeFalse())
			})

			It("runs the callback", func() {
				Ω(rt).Should(HaveTracked("BS", "OF-BS"))
			})
		})
	})

	Describe("what counts as a failure", func() {
		Context("a panic", func() {
			BeforeEach(func() {
				RunFixture("panic", func() {
					BeforeEach(OF("OF"))
					It("A", rt.T("A", func() { panic("bam") }))
				})
			})

			It("fires", func() {
				Ω(rt).Should(HaveTracked("A", "OF"))
				Ω(reporter.Did.Find("A")).Should(HavePanicked("bam"))
			})
		})

		Context("a SpecTimeout", func() {
			var callbackContextWasLive bool
			BeforeEach(func() {
				RunFixture("spec timeout", func() {
					BeforeEach(func() {
						DeferOnFailure(func(ctx SpecContext) {
							callbackContextWasLive = ctx.Err() == nil
							rt.Run("OF")
						})
					})
					It("A", rt.TSC("A", func(ctx SpecContext) { <-ctx.Done() }), SpecTimeout(50*time.Millisecond))
				})
			})

			It("fires, and the callback is not bound by the spec's (expired) SpecTimeout", func() {
				Ω(rt).Should(HaveTracked("A", "OF"))
				Ω(reporter.Did.Find("A")).Should(HaveTimedOut())
				Ω(callbackContextWasLive).Should(BeTrue())
			})
		})

		Context("a NodeTimeout", func() {
			BeforeEach(func() {
				RunFixture("node timeout", func() {
					BeforeEach(rt.TSC("BE", func(ctx SpecContext) {
						OF("OF")()
						<-ctx.Done()
					}), NodeTimeout(50*time.Millisecond))
					It("A", rt.T("A"))
				})
			})

			It("fires", func() {
				Ω(rt).Should(HaveTracked("BE", "OF"))
				Ω(reporter.Did.Find("A")).Should(HaveTimedOut())
			})
		})

		Context("the suite timeout", func() {
			BeforeEach(func() {
				conf.Timeout = 100 * time.Millisecond
				RunFixture("suite timeout", func() {
					BeforeEach(OF("OF"))
					It("A", rt.TSC("A", func(ctx SpecContext) { <-ctx.Done() }))
				})
			})

			It("fires", func() {
				Ω(rt).Should(HaveTracked("A", "OF"))
				Ω(reporter.Did.Find("A")).Should(HaveTimedOut("A suite timeout occurred"))
			})
		})

		Context("an AbortSuite", func() {
			BeforeEach(func() {
				RunFixture("abort", func() {
					BeforeEach(OF("OF"))
					It("A", rt.T("A", func() { AbortSuite("abort") }))
				})
			})

			It("fires", func() {
				Ω(rt).Should(HaveTracked("A", "OF"))
				Ω(reporter.Did.Find("A")).Should(HaveAborted("abort"))
			})
		})

		Context("the user's first interrupt", func() {
			BeforeEach(func() {
				RunFixture("interrupt", func() {
					BeforeEach(OF("OF"))
					It("A", rt.TSC("A", func(ctx SpecContext) {
						interruptHandler.Interrupt(interrupt_handler.InterruptCauseSignal)
						<-ctx.Done()
					}))
					AfterEach(rt.T("AE"))
				})
			})

			It("fires, before cleanup runs", func() {
				Ω(rt).Should(HaveTracked("A", "OF", "AE"))
				Ω(reporter.Did.Find("A")).Should(HaveBeenInterrupted(interrupt_handler.InterruptCauseSignal))
			})
		})

		Context("an abort triggered by another process", func() {
			BeforeEach(func() {
				RunFixture("abort by other process", func() {
					BeforeEach(OF("OF"))
					It("A", rt.TSC("A", func(ctx SpecContext) {
						interruptHandler.Interrupt(interrupt_handler.InterruptCauseAbortByOtherProcess)
						<-ctx.Done()
					}))
					AfterEach(rt.T("AE"))
				})
			})

			It("does not fire - this spec did not fail", func() {
				Ω(rt).Should(HaveTracked("A", "AE"))
				Ω(reporter.Did.Find("A")).Should(HaveBeenInterrupted(interrupt_handler.InterruptCauseAbortByOtherProcess))
			})
		})
	})

	Describe("retries", func() {
		Context("with FlakeAttempts", func() {
			BeforeEach(func() {
				attempt := 0
				success, _ := RunFixture("flake attempts", func() {
					BeforeEach(rt.T("BE", OF("OF")))
					It("A", rt.T("A", func() {
						attempt++
						if attempt < 3 {
							F("flake")
						}
					}), FlakeAttempts(3))
				})
				Ω(success).Should(BeTrue())
			})

			It("fires once on each failing attempt, never accumulating registrations", func() {
				Ω(rt).Should(HaveTracked("BE", "A", "OF", "BE", "A", "OF", "BE", "A"))
			})
		})

		Context("with FlakeAttempts in an ordered container whose BeforeAll registered the callback", func() {
			BeforeEach(func() {
				attempt := 0
				success, _ := RunFixture("flake attempts ordered", func() {
					Describe("ordered", Ordered, func() {
						BeforeAll(rt.T("BA", OF("OF-BA")))
						It("A", rt.T("A"))
						It("B", rt.T("B", func() {
							attempt++
							if attempt < 2 {
								F("flake")
							}
						}), FlakeAttempts(2))
					})
				})
				Ω(success).Should(BeTrue())
			})

			It("keeps the BeforeAll's registration across the retry", func() {
				Ω(rt).Should(HaveTracked("BA", "A", "B", "OF-BA", "B"))
			})
		})

		Context("when a BeforeAll's cleanup fails on the last spec and triggers a retry", func() {
			BeforeEach(func() {
				attempt := 0
				cleanupCalls := 0
				success, _ := RunFixture("cleanup failure triggers retry", func() {
					Describe("ordered", Ordered, func() {
						BeforeAll(rt.T("BA", both(OF("OF-BA"), func() {
							DeferCleanup(func() {
								rt.Run("C-BA")
								cleanupCalls++
								if cleanupCalls == 1 {
									F("cleanup boom")
								}
							})
						})))
						It("A", rt.T("A", func() {
							attempt++
							if attempt == 2 {
								F("boom on retry")
							}
						}), FlakeAttempts(2))
					})
				})
				Ω(success).Should(BeFalse())
			})

			It("drops the registration along with the cleanup, so it doesn't fire against torn-down state", func() {
				Ω(rt).Should(HaveTracked("BA", "A", "C-BA", "OF-BA", "A"))
			})
		})

		Context("with MustPassRepeatedly", func() {
			BeforeEach(func() {
				attempt := 0
				success, _ := RunFixture("must pass repeatedly", func() {
					BeforeEach(rt.T("BE", OF("OF")))
					It("A", rt.T("A", func() {
						attempt++
						if attempt == 2 {
							F("boom")
						}
					}), MustPassRepeatedly(3))
				})
				Ω(success).Should(BeFalse())
			})

			It("fires once, for the failing attempt", func() {
				Ω(rt).Should(HaveTracked("BE", "A", "BE", "A", "OF"))
			})
		})
	})

	Describe("bounding the callback", func() {
		Context("when the callback outlives its NodeTimeout", func() {
			BeforeEach(func() {
				RunFixture("callback times out", func() {
					BeforeEach(func() {
						DeferOnFailure(func(ctx SpecContext) {
							<-ctx.Done()
							rt.Run("OF-cancelled")
						}, NodeTimeout(50*time.Millisecond))
					})
					It("A", rt.T("A", func() { F("boom") }))
					AfterEach(rt.T("AE"))
				})
			})

			It("cancels it, records a timeout as an additional failure, and proceeds with teardown", func() {
				Ω(rt).Should(HaveTracked("A", "OF-cancelled", "AE"))
				report := reporter.Did.Find("A")
				Ω(report).Should(HaveFailed("boom"))
				Ω(report.AdditionalFailures).Should(HaveLen(1))
				Ω(report.AdditionalFailures[0].State).Should(Equal(types.SpecStateTimedout))
				Ω(report.AdditionalFailures[0].Failure.FailureNodeType).Should(Equal(types.NodeTypeOnFailureEach))
			})
		})

		Context("when the callback has no NodeTimeout", func() {
			var elapsed time.Duration
			BeforeEach(func() {
				conf.GracePeriod = 100 * time.Millisecond
				RunFixture("default bound", func() {
					BeforeEach(func() {
						DeferOnFailure(func(ctx SpecContext) {
							t := time.Now()
							<-ctx.Done()
							elapsed = time.Since(t)
							rt.Run("OF-cancelled")
						})
					})
					It("A", rt.T("A", func() { F("boom") }))
				})
			})

			It("bounds it by one GracePeriod", func() {
				Ω(rt).Should(HaveTracked("A", "OF-cancelled"))
				Ω(elapsed).Should(BeNumerically("~", 100*time.Millisecond, 80*time.Millisecond))
				Ω(reporter.Did.Find("A").AdditionalFailures[0].State).Should(Equal(types.SpecStateTimedout))
			})
		})
	})

	Describe("a misbehaving callback", func() {
		Context("when a callback fails", func() {
			BeforeEach(func() {
				success, _ := RunFixture("callback fails", func() {
					BeforeEach(rt.T("BE", both(OF("OF-1"), func() {
						DeferOnFailure(func(_ SpecContext) {
							rt.Run("OF-2")
							F("callback boom")
						})
					})))
					It("A", rt.T("A", func() { F("boom") }))
					AfterEach(rt.T("AE"))
				})
				Ω(success).Should(BeFalse())
			})

			It("records an additional failure, keeps the original failure, and runs the remaining callbacks and teardown", func() {
				Ω(rt).Should(HaveTracked("BE", "A", "OF-2", "OF-1", "AE"))
				report := reporter.Did.Find("A")
				Ω(report).Should(HaveFailed("boom"))
				Ω(report.AdditionalFailures).Should(HaveLen(1))
				Ω(report.AdditionalFailures[0].Failure.Message).Should(Equal("callback boom"))
				Ω(report.AdditionalFailures[0].Failure.FailureNodeType).Should(Equal(types.NodeTypeOnFailureEach))
			})
		})

		Context("when a callback calls AbortSuite", func() {
			BeforeEach(func() {
				success, _ := RunFixture("callback aborts", func() {
					Describe("top", func() {
						BeforeEach(func() {
							DeferOnFailure(func(_ SpecContext) { AbortSuite("abort from callback") })
						})
						It("A", rt.T("A", func() { F("boom") }))
						It("B", rt.T("B"))
					})
				})
				Ω(success).Should(BeFalse())
			})

			It("aborts the suite, keeping the original failure as an additional failure", func() {
				report := reporter.Did.Find("A")
				Ω(report).Should(HaveAborted("abort from callback"))
				Ω(report.AdditionalFailures).Should(ContainElement(HaveField("Failure.Message", "boom")))
				Ω(rt).Should(HaveTracked("A"))
			})
		})

		Context("when a callback calls AbortSuite after it has timed out", func() {
			BeforeEach(func() {
				success, _ := RunFixture("callback aborts late", func() {
					Describe("top", func() {
						BeforeEach(func() {
							DeferOnFailure(func(ctx SpecContext) {
								<-ctx.Done()
								AbortSuite("late abort")
							}, NodeTimeout(50*time.Millisecond))
						})
						It("A", rt.T("A", func() { F("boom") }))
						It("B", rt.T("B"))
					})
				})
				Ω(success).Should(BeFalse())
			})

			It("still aborts the suite", func() {
				Ω(reporter.Did.Find("A").State).Should(Equal(types.SpecStateAborted))
				Ω(rt).Should(HaveTracked("A"))
			})
		})
	})

	Describe("calling into Ginkgo from the callback", func() {
		var failureMessage string
		var failed, tFailed bool
		BeforeEach(func() {
			RunFixture("calling into ginkgo", func() {
				BeforeEach(func() {
					DeferOnFailure(func(ctx SpecContext) {
						failureMessage = CurrentSpecReport().Failure.Message
						failed = ctx.SpecReport().Failed()
						tFailed = GinkgoT().Failed()
						writer.Println("captured from the callback") // the fixture's GinkgoWriter
						AddReportEntry("screenshot", ReportEntryVisibilityFailureOrVerbose)
					})
				})
				It("A", rt.T("A", func() { F("boom") }))
			})
		})

		It("sees the failure, and can write output and report entries that land in the spec's report", func() {
			Ω(failureMessage).Should(Equal("boom"))
			Ω(failed).Should(BeTrue())
			Ω(tFailed).Should(BeTrue())
			report := reporter.Did.Find("A")
			Ω(report.CapturedGinkgoWriterOutput).Should(ContainSubstring("captured from the callback"))
			Ω(report.ReportEntries).Should(HaveLen(1))
			Ω(report.ReportEntries[0].Name).Should(Equal("screenshot"))
			Ω(report.ReportEntries[0].Visibility).Should(Equal(types.ReportEntryVisibilityFailureOrVerbose))
		})
	})

	Describe("accepting what DeferCleanup accepts", func() {
		var receivedArgs []any
		var ownContextWasUsed bool
		type ctxKey struct{}
		BeforeEach(func() {
			receivedArgs = nil
			RunFixture("defer cleanup parity", func() {
				BeforeEach(func(ctx SpecContext) {
					DeferOnFailure(func(a string, b int) { receivedArgs = []any{a, b} }, "a", 1)
					DeferOnFailure(func() error { return fmt.Errorf("returned an error") })
					DeferOnFailure(func(ctx context.Context) { ownContextWasUsed = ctx.Value(ctxKey{}) == "mine" }, context.WithValue(context.Background(), ctxKey{}, "mine"))
					DeferOnFailure(func(ctx context.Context, label string) {
						<-ctx.Done()
						rt.Run(label)
					}, "interruptible", NodeTimeout(50*time.Millisecond))
				})
				It("A", rt.T("A", func() { F("boom") }))
			})
		})

		It("passes arguments, fails on a returned error, honors a provided context, and makes context-accepting functions interruptible", func() {
			Ω(receivedArgs).Should(Equal([]any{"a", 1}))
			Ω(ownContextWasUsed).Should(BeTrue())
			Ω(rt).Should(HaveTracked("A", "interruptible"))
			report := reporter.Did.Find("A")
			Ω(report).Should(HaveFailed("boom"))
			Ω(report.AdditionalFailures).Should(ContainElement(HaveField("Failure.Message", "DeferOnFailure callback returned error: returned an error")))
			Ω(report.AdditionalFailures).Should(ContainElement(HaveField("State", types.SpecStateTimedout)))
		})
	})

	Describe("registering through GinkgoT()", func() {
		BeforeEach(func() {
			RunFixture("ginkgo t", func() {
				BeforeEach(func() {
					GinkgoT().DeferOnFailure(func(_ SpecContext) { rt.Run("OF") })
				})
				It("A", rt.T("A", func() { F("boom") }))
			})
		})

		It("works", func() {
			Ω(rt).Should(HaveTracked("A", "OF"))
		})
	})

	Describe("--sleep-on-failure", func() {
		Context("when there are callbacks", func() {
			var progressReportsWhenCallbackRan int
			BeforeEach(func() {
				progressReportsWhenCallbackRan = -1
				conf.SleepOnFailure = 50 * time.Millisecond
				RunFixture("sleep on failure with callbacks", func() {
					BeforeSuite(func() {
						DeferOnFailure(func(_ SpecContext) {
							progressReportsWhenCallbackRan = len(reporter.ProgressReports)
							rt.Run("OF-BS")
						})
					})
					BeforeEach(OF("OF"))
					It("A", rt.T("A", func() { F("boom") }))
					AfterEach(rt.T("AE"))
				})
			})

			It("pauses after all callbacks have run, then proceeds to teardown", func() {
				Ω(rt).Should(HaveTracked("A", "OF", "OF-BS", "AE"))
				Ω(progressReportsWhenCallbackRan).Should(Equal(0))
				Ω(reporter.ProgressReports).Should(HaveLen(1))
				Ω(reporter.ProgressReports[0].Message).Should(ContainSubstring("Paused on failure"))
			})
		})

		Context("when a callback calls AbortSuite", func() {
			BeforeEach(func() {
				conf.SleepOnFailure = 50 * time.Millisecond
				RunFixture("sleep on failure after a callback aborts", func() {
					BeforeEach(func() {
						DeferOnFailure(func() { AbortSuite("abort from callback") })
					})
					It("A", rt.T("A", func() { F("boom") }))
				})
			})

			It("still pauses, because the triggering failure was in the It", func() {
				Ω(reporter.Did.Find("A").State).Should(Equal(types.SpecStateAborted))
				Ω(reporter.ProgressReports).Should(ContainElement(HaveField("Message", ContainSubstring("Paused on failure"))))
			})
		})

		Context("when the user interrupts while the callbacks run", func() {
			var elapsed time.Duration
			BeforeEach(func() {
				conf.SleepOnFailure = time.Hour // would hang if the pause ran after the interrupt
				start := time.Now()
				RunFixture("sleep on failure, interrupted during callbacks", func() {
					BeforeEach(func() {
						DeferOnFailure(func(ctx SpecContext) {
							rt.Run("OF")
							interruptHandler.Interrupt(interrupt_handler.InterruptCauseSignal)
							<-ctx.Done()
						})
					})
					It("A", rt.T("A", func() { F("boom") }))
					AfterEach(rt.T("AE"))
				})
				elapsed = time.Since(start)
			})

			It("does not pause, and proceeds to cleanup", func() {
				Ω(elapsed).Should(BeNumerically("<", 5*time.Second))
				Ω(rt).Should(HaveTracked("A", "OF", "AE"))
				Ω(reporter.ProgressReports).ShouldNot(ContainElement(HaveField("Message", ContainSubstring("Paused on failure"))))
			})
		})

		Context("when the suite's deadline elapses during the pause", func() {
			var elapsed time.Duration
			BeforeEach(func() {
				conf.Timeout = 50 * time.Millisecond
				conf.SleepOnFailure = 300 * time.Millisecond
				start := time.Now()
				RunFixture("sleep on failure past the suite deadline", func() {
					It("A", rt.T("A", func() { F("boom") }))
				})
				elapsed = time.Since(start)
			})

			It("pauses for the full duration", func() {
				Ω(elapsed).Should(BeNumerically(">=", 300*time.Millisecond))
			})
		})

		Context("when the spec times out", func() {
			BeforeEach(func() {
				conf.SleepOnFailure = 50 * time.Millisecond
				RunFixture("sleep on failure with timeout", func() {
					It("A", rt.TSC("A", func(ctx SpecContext) { <-ctx.Done() }), SpecTimeout(50*time.Millisecond))
				})
			})

			It("pauses", func() {
				Ω(reporter.Did.Find("A")).Should(HaveTimedOut())
				Ω(reporter.ProgressReports).Should(ContainElement(HaveField("Message", ContainSubstring("Paused on failure"))))
			})
		})
	})
})
