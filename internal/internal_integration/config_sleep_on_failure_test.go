package internal_integration_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/onsi/ginkgo/v2/internal/interrupt_handler"
	. "github.com/onsi/ginkgo/v2/internal/test_helpers"
)

var _ = Describe("--sleep-on-failure", func() {
	Describe("when a spec fails", func() {
		BeforeEach(func() {
			conf.SleepOnFailure = 50 * time.Millisecond
			success, _ := RunFixture("sleep on failure - failing spec", func() {
				Context("container", func() {
					BeforeEach(rt.T("bef"))
					It("A", rt.T("A", func() {
						DeferCleanup(rt.T("cleanup"))
						F("boom", cl)
					}))
					JustAfterEach(rt.T("just-after"))
					AfterEach(rt.T("aft"))
				})
			})
			Ω(success).Should(BeFalse())
		})

		It("pauses at the moment of failure, before any teardown runs", func() {
			// The pause is hooked into runNode's failure path, so it happens the instant
			// the It fails - before any JustAfterEach/AfterEach/DeferCleanup nodes run.
			// Teardown then proceeds in the normal order once the pause elapses.
			Ω(rt).Should(HaveTracked("bef", "A", "just-after", "aft", "cleanup"))
		})

		It("still runs teardown and cleanup after the pause", func() {
			Ω(rt).Should(HaveRun("aft"))
			Ω(rt).Should(HaveRun("cleanup"))
		})

		It("emits the failure and then a progress report announcing the pause", func() {
			Ω(reporter.Did.Find("A")).Should(HaveFailed("boom", cl))
			Ω(reporter.ProgressReports).ShouldNot(BeEmpty())
			Ω(reporter.ProgressReports[0].Message).Should(ContainSubstring("Paused on failure"))
		})
	})

	Describe("when a failure occurs in a BeforeEach", func() {
		BeforeEach(func() {
			conf.SleepOnFailure = 50 * time.Millisecond
			success, _ := RunFixture("sleep on failure - failing beforeeach", func() {
				Context("container", func() {
					BeforeEach(rt.T("bef", func() {
						F("boom", cl)
					}))
					It("A", rt.T("A"))
					AfterEach(rt.T("aft"))
				})
			})
			Ω(success).Should(BeFalse())
		})

		It("pauses on the setup failure before teardown, and skips the It", func() {
			Ω(rt).Should(HaveTracked("bef", "aft"))
			Ω(reporter.ProgressReports).ShouldNot(BeEmpty())
			Ω(reporter.ProgressReports[0].Message).Should(ContainSubstring("Paused on failure"))
		})
	})

	Describe("when a spec passes", func() {
		var start time.Time
		BeforeEach(func() {
			conf.SleepOnFailure = time.Hour // would hang if it ever fired on success
			start = time.Now()
			success, _ := RunFixture("sleep on failure - passing spec", func() {
				It("A", rt.T("A"))
				AfterEach(rt.T("aft"))
			})
			Ω(success).Should(BeTrue())
		})

		It("does not pause and emits no pause progress report", func() {
			Ω(time.Since(start)).Should(BeNumerically("<", time.Second))
			Ω(rt).Should(HaveTracked("A", "aft"))
			Ω(reporter.ProgressReports).Should(BeEmpty())
		})
	})

	Describe("when a failure occurs in teardown", func() {
		var start time.Time
		BeforeEach(func() {
			conf.SleepOnFailure = time.Hour // would hang if the teardown failure paused
			start = time.Now()
			success, _ := RunFixture("sleep on failure - failing teardown", func() {
				It("A", rt.T("A"))
				AfterEach(rt.T("aft", func() {
					F("boom", cl)
				}))
			})
			Ω(success).Should(BeFalse())
		})

		It("does not pause (teardown is already running, nothing to inspect)", func() {
			Ω(time.Since(start)).Should(BeNumerically("<", time.Second))
			Ω(rt).Should(HaveTracked("A", "aft"))
			Ω(reporter.ProgressReports).Should(BeEmpty())
		})
	})

	Describe("when the feature is disabled (duration is zero)", func() {
		BeforeEach(func() {
			conf.SleepOnFailure = 0
			success, _ := RunFixture("sleep on failure - disabled", func() {
				It("A", rt.T("A", func() {
					F("boom", cl)
				}))
				AfterEach(rt.T("aft"))
			})
			Ω(success).Should(BeFalse())
		})

		It("does not pause and emits no pause progress report", func() {
			Ω(rt).Should(HaveTracked("A", "aft"))
			Ω(reporter.ProgressReports).Should(BeEmpty())
		})
	})

	Describe("when the user interrupts during the pause", func() {
		var start time.Time
		BeforeEach(func() {
			conf.SleepOnFailure = time.Hour // long enough that only the interrupt can end it
			start = time.Now()
			success, _ := RunFixture("sleep on failure - interrupted", func() {
				Context("container", func() {
					It("A", rt.T("A", func() {
						// fire the interrupt shortly after this node's body returns, so it
						// lands while the suite is paused waiting on the failure
						go func() {
							time.Sleep(100 * time.Millisecond)
							interruptHandler.Interrupt(interrupt_handler.InterruptCauseSignal)
						}()
						F("boom", cl)
					}))
					AfterEach(rt.T("aft"))
				})
			})
			Ω(success).Should(BeFalse())
		})

		It("ends the pause early and proceeds to run teardown", func() {
			Ω(time.Since(start)).Should(BeNumerically("<", time.Minute))
			Ω(rt).Should(HaveTracked("A", "aft"))
		})
	})
})
