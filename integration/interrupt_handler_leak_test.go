package integration_test

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("InterruptHandler lifecycle", func() {
	BeforeEach(func() {
		fm.MountFixture("interrupt_handler_leak")
	})

	It("does not leak the signal-handling goroutine after RunSpecs returns", func() {
		session := startGinkgo(fm.PathTo("interrupt_handler_leak"))
		Eventually(session).Should(gexec.Exit(0))
		Ω(session).ShouldNot(gbytes.Say("outlived the suite entry point"))
	})

	It("does not leak the goroutine after PreviewSpecs returns", func() {
		os.Setenv("PREVIEW", "true")
		DeferCleanup(os.Unsetenv, "PREVIEW")
		session := startGinkgo(fm.PathTo("interrupt_handler_leak"))
		Eventually(session).Should(gexec.Exit(0))
		Ω(session).ShouldNot(gbytes.Say("outlived the suite entry point"))
	})

	It("stops a fresh handler for each suite when RunSpecs is called several times", func() {
		os.Setenv("RERUN", "true")
		DeferCleanup(os.Unsetenv, "RERUN")
		session := startGinkgo(fm.PathTo("interrupt_handler_leak"))
		Eventually(session).Should(gexec.Exit(0))
		Ω(session).ShouldNot(gbytes.Say("outlived the suite entry point"))
	})

	It("does not leak the abort-polling goroutine when running in parallel", func() {
		session := startGinkgo(fm.PathTo("interrupt_handler_leak"), "-p")
		Eventually(session).Should(gexec.Exit(0))
		Ω(session).ShouldNot(gbytes.Say("outlived the suite entry point"))
	})
})
