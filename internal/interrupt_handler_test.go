package internal_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/internal"
	. "github.com/onsi/gomega"
)

var _ = Describe("InterruptHandler", func() {
	var interruptHandler *internal.InterruptHandler
	Describe("Timeout interrupts", func() {
		BeforeEach(func() {
			interruptHandler = internal.NewInterruptHandler(500 * time.Millisecond)
		})

		AfterEach(func() {
			interruptHandler.Stop()
		})

		It("eventually closes the interrupt channel to signal an interrupt has occurred", func() {
			status := interruptHandler.Status()
			Ω(status.Interrupted).Should(BeFalse())
			Eventually(status.Channel).Should(BeClosed())

			Ω(interruptHandler.Status().Interrupted).Should(BeTrue())
		})

		It("notes the cause as 'Interrupted By Timeout'", func() {
			status := interruptHandler.Status()
			Eventually(status.Channel).Should(BeClosed())
			cause := interruptHandler.Status().Cause
			Ω(cause).Should(Equal("Interrupted by Timeout"))
			Ω(interruptHandler.InterruptMessageWithStackTraces()).Should(HavePrefix("Interrupted by Timeout\n\n"))
		})

		It("repeatedly triggers an interrupt every 1/10th of the registered timeout", func() {
			status := interruptHandler.Status()
			Ω(status.Interrupted).Should(BeFalse())
			Eventually(status.Channel).Should(BeClosed())

			status = interruptHandler.Status()
			Ω(status.Channel).ShouldNot(BeClosed())
			Eventually(status.Channel).Should(BeClosed())
		})
	})
})
