//go:build freebsd || openbsd || netbsd || dragonfly || darwin || linux || solaris
// +build freebsd openbsd netbsd dragonfly darwin linux solaris

package interrupt_handler_test

import (
	"syscall"
	"time"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/internal/interrupt_handler"
	"github.com/onsi/ginkgo/v2/internal/parallel_support"
	. "github.com/onsi/ginkgo/v2/internal/test_helpers"
	. "github.com/onsi/gomega"
)

var _ = Describe("InterruptHandler", func() {
	var trigger func()
	var interruptHandler *interrupt_handler.InterruptHandler
	BeforeEach(func() {
		trigger = func() {
			syscall.Kill(syscall.Getpid(), syscall.SIGUSR2)
		}
	})

	Describe("Signal interrupts", func() {
		BeforeEach(func() {
			interruptHandler = interrupt_handler.NewInterruptHandler(nil, syscall.SIGUSR2)
			DeferCleanup(interruptHandler.Stop)
		}, OncePerOrdered)

		It("starts off uninterrupted", func() {
			status := interruptHandler.Status()
			Ω(status.Interrupted()).Should(BeFalse())
			Ω(status.Cause).Should(Equal(interrupt_handler.InterruptCauseInvalid))
			Ω(status.Level).Should(Equal(interrupt_handler.InterruptLevelUninterrupted))
			Consistently(status.Channel).ShouldNot(BeClosed())
		})

		When("interrupted repeatedly", Ordered, func() {
			var status interrupt_handler.InterruptStatus
			BeforeAll(func() {
				status = interruptHandler.Status()
				Ω(status.Interrupted()).Should(BeFalse())
			})

			Specify("when first interrupted, it closes the channel and goes to the next Cleanup and Report level", func() {
				trigger()
				Eventually(status.Channel).Should(BeClosed())

				status = interruptHandler.Status()
				Ω(status.Interrupted()).Should(BeTrue())
				Ω(status.Cause).Should(Equal(interrupt_handler.InterruptCauseSignal))
				Ω(status.Message()).Should(Equal("Interrupted by User"))
				Ω(status.Level).Should(Equal(interrupt_handler.InterruptLevelCleanupAndReport))
				Ω(status.ShouldIncludeProgressReport()).Should(BeTrue())
			})

			Specify("when interrupted a second time, it closes the next channel and goes to the Report Only level", func() {
				trigger()
				Eventually(status.Channel).Should(BeClosed())

				status = interruptHandler.Status()
				Ω(status.Interrupted()).Should(BeTrue())
				Ω(status.Cause).Should(Equal(interrupt_handler.InterruptCauseSignal))
				Ω(status.Message()).Should(Equal("Interrupted by User"))
				Ω(status.Level).Should(Equal(interrupt_handler.InterruptLevelReportOnly))
				Ω(status.ShouldIncludeProgressReport()).Should(BeTrue())
			})

			Specify("when interrupted a third time, it closes the next channel and goes to the Bail-Out level", func() {
				trigger()
				Eventually(status.Channel).Should(BeClosed())

				status = interruptHandler.Status()
				Ω(status.Interrupted()).Should(BeTrue())
				Ω(status.Cause).Should(Equal(interrupt_handler.InterruptCauseSignal))
				Ω(status.Message()).Should(Equal("Interrupted by User"))
				Ω(status.Level).Should(Equal(interrupt_handler.InterruptLevelBailOut))
				Ω(status.ShouldIncludeProgressReport()).Should(BeTrue())
			})

			Specify("when interrupted again and again, it no longer interrupts once the last level has been reached", func() {
				trigger()
				Consistently(status.Channel).ShouldNot(BeClosed())
				trigger()
				Consistently(status.Channel).ShouldNot(BeClosed())

				status = interruptHandler.Status()
				Ω(status.Interrupted()).Should(BeTrue())
				Ω(status.Cause).Should(Equal(interrupt_handler.InterruptCauseSignal))
				Ω(status.Message()).Should(Equal("Interrupted by User"))
				Ω(status.Level).Should(Equal(interrupt_handler.InterruptLevelBailOut))
				Ω(status.ShouldIncludeProgressReport()).Should(BeTrue())
			})
		})
	})

	Describe("Interrupting when another Ginkgo process has aborted", func() {
		var client parallel_support.Client
		BeforeEach(func() {
			_, client, _ = SetUpServerAndClient(2)
			interruptHandler = interrupt_handler.NewInterruptHandler(client, syscall.SIGUSR2)
			DeferCleanup(interruptHandler.Stop)
		})

		It("interrupts when the server is told to abort", func() {
			status := interruptHandler.Status()
			Consistently(status.Channel).ShouldNot(BeClosed())

			client.PostAbort()

			Eventually(status.Channel).Should(BeClosed())
		})

		It("notes the correct cause and returns an interrupt message that does not include a progress report", func() {
			status := interruptHandler.Status()
			Ω(status.Interrupted()).Should(BeFalse())

			client.PostAbort()
			Eventually(status.Channel).Should(BeClosed())

			status = interruptHandler.Status()
			Ω(status.Cause).Should(Equal(interrupt_handler.InterruptCauseAbortByOtherProcess))
			Ω(status.Level).Should(Equal(interrupt_handler.InterruptLevelCleanupAndReport))
			Ω(status.Interrupted()).Should(BeTrue())
			Ω(status.Message()).Should(Equal("Interrupted by Other Ginkgo Process"))
			Ω(status.ShouldIncludeProgressReport()).Should(BeFalse())
		})

		It("does not retrigger on subsequent aborts", func() {
			status := interruptHandler.Status()
			Ω(status.Interrupted()).Should(BeFalse())

			client.PostAbort()
			Eventually(status.Channel).Should(BeClosed())

			status = interruptHandler.Status()
			client.PostAbort()
			Consistently(status.Channel, interrupt_handler.ABORT_POLLING_INTERVAL+100*time.Millisecond).ShouldNot(BeClosed())

			status = interruptHandler.Status()
			Ω(status.Cause).Should(Equal(interrupt_handler.InterruptCauseAbortByOtherProcess))
			Ω(status.Level).Should(Equal(interrupt_handler.InterruptLevelCleanupAndReport))
		})

		It("does not trigger if the suite has already been interrupted", func() {
			status := interruptHandler.Status()
			Ω(status.Interrupted()).Should(BeFalse())

			trigger()
			Eventually(status.Channel).Should(BeClosed())

			status = interruptHandler.Status()
			Ω(status.Cause).Should(Equal(interrupt_handler.InterruptCauseSignal))
			Ω(status.Level).Should(Equal(interrupt_handler.InterruptLevelCleanupAndReport))

			status = interruptHandler.Status()
			client.PostAbort()
			Consistently(status.Channel, interrupt_handler.ABORT_POLLING_INTERVAL+100*time.Millisecond).ShouldNot(BeClosed())

			status = interruptHandler.Status()
			Ω(status.Cause).Should(Equal(interrupt_handler.InterruptCauseSignal))
			Ω(status.Level).Should(Equal(interrupt_handler.InterruptLevelCleanupAndReport))
		})

		It("doesn't just rely on the ABORT_POLLING_INTERVAL timer to report that the interrupt has happened", func() {
			client.PostAbort()
			Ω(interruptHandler.Status().Cause).Should(Equal(interrupt_handler.InterruptCauseAbortByOtherProcess))
		}, MustPassRepeatedly(10))
	})
})
