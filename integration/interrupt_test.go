package integration_test

import (
	"encoding/json"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Interrupt and Timeout", func() {
	Context("when interrupting a suite", func() {
		var session *gexec.Session
		BeforeEach(func() {
			fm.MountFixture("hanging")

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
			fm.MountFixture("hanging")

			session = startGinkgo(fm.PathTo("hanging"), "--no-color", "--timeout=5s")
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

	Describe("applying the timeout to multiple suites", func() {
		It("tracks the timeout across the suites, decrementing the available timeout for each individual suite, and reports on any suites that did not run because the timeout elapsed", Label("slow"), func() {
			fm.MountFixture("timeout")
			session := startGinkgo(fm.PathTo("timeout"), "--no-color", "-r", "--timeout=10s", "--keep-going", "--json-report=out.json")
			Eventually(session).Should(gbytes.Say("TimeoutA Suite"))
			Eventually(session, "15s").Should(gexec.Exit(1))
			Ω(session).Should(gbytes.Say(`timeout_D ./timeout_D \[Suite did not run because the timeout elapsed\]`))

			data := []byte(fm.ContentOf("timeout", "out.json"))
			reports := []types.Report{}
			Ω(json.Unmarshal(data, &reports)).Should(Succeed())
			Ω(reports[3].SpecialSuiteFailureReasons).Should(ContainElement("Suite did not run because the timeout elapsed"))
		})
	})
})
