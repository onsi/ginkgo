package integration_test

import (
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Interrupt and Timeout", func() {
	BeforeEach(func() {
		fm.MountFixture("hanging")
	})

	Context("when interrupting a suite", func() {
		var session *gexec.Session
		BeforeEach(func() {
			//we need to signal the actual process, so we must compile the test first
			session = startGinkgo(fm.PathTo("hanging"), "build")
			Eventually(session).Should(gexec.Exit(0))

			//then run the compiled test directly
			cmd := exec.Command("./hanging.test", "--test.v", "--ginkgo.no-color")
			cmd.Dir = fm.PathTo("hanging")
			var err error
			session, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Ω(err).ShouldNot(HaveOccurred())

			Eventually(session).Should(gbytes.Say("Sleeping..."))
			session.Interrupt()
			Eventually(session).Should(gbytes.Say("Sleeping again..."))
			session.Interrupt()
			Eventually(session, 1000).Should(gexec.Exit(1))
		})

		It("should emit the contents of the GinkgoWriter", func() {
			Ω(session).Should(gbytes.Say("Just beginning"))
			Ω(session).Should(gbytes.Say("Almost there..."))
			Ω(session).Should(gbytes.Say("Hanging Out"))
		})

		It("should report where the suite was interrupted", func() {
			Ω(session).Should(gbytes.Say(`\[INTERRUPTED\]`))
			Ω(session).Should(gbytes.Say(`Interrupted by User`))
			Ω(session).Should(gbytes.Say(`Here's a stack trace of all running goroutines:`))
			Ω(session).Should(gbytes.Say(`\[It\] .*hanging_test.go:24`))
		})

		It("should run the AfterEach and the AfterSuite", func() {
			Ω(session).Should(gbytes.Say("Cleaning up once..."))
			Ω(session).Should(gbytes.Say("Cleaning up twice..."))
			Ω(session).Should(gbytes.Say("Cleaning up thrice..."))
			Ω(session).Should(gbytes.Say("Heading Out After Suite"))
		})
		It("should emit a special failure reason", func() {
			Ω(session).Should(gbytes.Say("FAIL! - Interrupted by User"))
		})
	})

	Context("when the suite times out", func() {
		var session *gexec.Session
		BeforeEach(func() {
			session = startGinkgo(fm.PathTo("hanging"), "--no-color", "--timeout=3s")
			Eventually(session).Should(gexec.Exit(1))
		})

		It("should report where and why the suite was interrupted", func() {
			Ω(session).Should(gbytes.Say(`\[INTERRUPTED\]`))
			Ω(session).Should(gbytes.Say(`Interrupted by Timeout`))
			Ω(session).Should(gbytes.Say(`Here's a stack trace of all running goroutines:`))
			Ω(session).Should(gbytes.Say(`\[It\] .*hanging_test.go:24`))
		})

		It("should run the AfterEach and the AfterSuite", func() {
			Ω(session).Should(gbytes.Say("Cleaning up once..."))
			Ω(session).Should(gbytes.Say("Cleaning up twice..."))
			Ω(session).Should(gbytes.Say("Cleaning up thrice..."))
			Ω(session).Should(gbytes.Say("Heading Out After Suite"))
		})

		It("should emit a special failure reason", func() {
			Ω(session).Should(gbytes.Say("FAIL! - Interrupted by Timeout"))
		})
	})
})
