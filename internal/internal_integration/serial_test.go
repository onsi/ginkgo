package internal_integration_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/ginkgo/v2/internal/test_helpers"
	. "github.com/onsi/gomega"
)

var _ = Describe("Serial", func() {
	var fixture func()
	BeforeEach(func() {
		fixture = func() {
			Context("container", func() {
				It("A", rt.T("A", func() { time.Sleep(10 * time.Millisecond) }))
				It("B", rt.T("B", func() { time.Sleep(10 * time.Millisecond) }))
				It("C", Serial, rt.T("C", func() { time.Sleep(10 * time.Millisecond) }))
				It("D", rt.T("D", func() { time.Sleep(10 * time.Millisecond) }))
				It("E", rt.T("E", func() { time.Sleep(10 * time.Millisecond) }))
				It("F", Serial, rt.T("F", func() { time.Sleep(10 * time.Millisecond) }))
				It("G", rt.T("G", func() { time.Sleep(10 * time.Millisecond) }))
				It("H", Serial, rt.T("H", func() { time.Sleep(10 * time.Millisecond) }))
			})
		}
	})

	Context("when running in series", func() {
		BeforeEach(func() {
			conf.ParallelTotal = 1
			conf.ParallelProcess = 1
			success, _ := RunFixture("in-series", fixture)
			Ω(success).Should(BeTrue())
		})

		It("runs and reports on all the tests", func() {
			Ω(rt).Should(HaveTracked("A", "B", "C", "D", "E", "F", "G", "H"))
			Ω(reporter.Did.Names()).Should(Equal([]string{"A", "B", "C", "D", "E", "F", "G", "H"}))
		})
	})

	Context("when running in parallel", func() {
		BeforeEach(func() {
			SetUpForParallel(2)
		})

		Describe("when running as proc 1", func() {
			BeforeEach(func() {
				conf.ParallelProcess = 1
			})

			It("participates in running parallel tests, then runs the serial tests after all other procs have finished", func() {
				done := make(chan any)
				go func() {
					defer GinkgoRecover()
					success, _ := RunFixture("happy-path", fixture)
					Ω(success).Should(BeTrue())
					close(done)
				}()
				Eventually(rt).Should(HaveTracked("A", "B", "D", "E", "G"))
				Consistently(rt, 100*time.Millisecond).Should(HaveTracked("A", "B", "D", "E", "G"))
				close(exitChannels[2])
				Eventually(rt).Should(HaveTracked("A", "B", "D", "E", "G", "C", "F", "H"))
				Eventually(done).Should(BeClosed())
			})
		})

		Describe("when running as a non-primary proc", func() {
			BeforeEach(func() {
				conf.ParallelProcess = 2
			})

			It("participates in running parallel tests, but never runs the serial tests", func() {
				close(exitChannels[1])
				success, _ := RunFixture("happy-path", fixture)
				Ω(success).Should(BeTrue())
				Ω(rt).Should(HaveTracked("A", "B", "D", "E", "G"))
			})
		})
	})
})
