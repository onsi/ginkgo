package integration_test

import (
	"encoding/json"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Interrupt and Timeout", MarkSlow, func() {
	Context("when interrupting a suite", func() {
		It("gives the user feedback as the session is interrupted", func() {
			fm.MountFixture("hanging")

			//we need to signal the actual process, so we must compile the test first
			session := startGinkgo(fm.PathTo("hanging"), "build")
			Eventually(session).Should(gexec.Exit(0))

			//then run the compiled test directly
			cmd := exec.Command("./hanging.test", "--test.v", "--ginkgo.no-color", "--ginkgo.grace-period=2s")
			cmd.Dir = fm.PathTo("hanging")
			var err error
			session, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Ω(err).ShouldNot(HaveOccurred())

			Eventually(session).Should(gbytes.Say("Sleeping..."))
			session.Interrupt()
			Eventually(session).Should(gbytes.Say(`First interrupt received`))
			Eventually(session).Should(gbytes.Say("Begin Captured GinkgoWriter Output"))
			Eventually(session).Should(gbytes.Say("Just beginning"))
			Eventually(session).Should(gbytes.Say("Almost there..."))
			Eventually(session).Should(gbytes.Say("Hanging Out"))
			Eventually(session).Should(gbytes.Say(`goroutine \d+ \[select\]`))
			Eventually(session).Should(gbytes.Say(`>\s*select {`), "The actual source code gets emitted")

			Eventually(session, time.Second*5).Should(gbytes.Say("Cleaning up once..."), "The two second grace period should move on past the napping node")
			Eventually(session).Should(gbytes.Say("Cleaning up twice..."))
			Eventually(session).Should(gbytes.Say("Sleeping again..."))
			session.Interrupt()
			Eventually(session).Should(gbytes.Say(`Second interrupt received`))
			Eventually(session).Should(gbytes.Say(`goroutine \d+ \[sleep\]`))
			Eventually(session).Should(gbytes.Say(`>\s*time.Sleep\(time.Hour\)`), "The actual source code gets emitted now")

			Eventually(session).Should(gbytes.Say(`\[INTERRUPTED\]`))
			Eventually(session).Should(gbytes.Say(`Interrupted by User`))
			Eventually(session).Should(gbytes.Say(`Spec Goroutine`))
			Eventually(session).Should(gbytes.Say(`goroutine \d+ \[select\]`))
			Eventually(session).Should(gbytes.Say(`>\s*select {`), "The actual source code gets emitted now")
			Eventually(session).Should(gbytes.Say(`Other Goroutines`))
			Eventually(session).Should(gbytes.Say(`main\.main\(\)`))

			Eventually(session).Should(gbytes.Say("Reporting at the end"))

			Eventually(session).Should(gbytes.Say(`FAIL! - Interrupted by User`))
			Eventually(session, time.Second*10).Should(gexec.Exit(1))

			// the last AfterEach and the AfterSuite don't actually run
			Ω(string(session.Out.Contents())).ShouldNot(ContainSubstring("Cleaning up thrice"))
			Ω(string(session.Out.Contents())).ShouldNot(ContainSubstring("Heading Out After Suite"))
		})
	})

	Context("when the suite times out", func() {
		It("interrupts the suite and gives the user feedback as it does so", func() {
			fm.MountFixture("hanging")

			session := startGinkgo(fm.PathTo("hanging"), "--no-color", "--timeout=5s", "--grace-period=1s")
			Eventually(session, time.Second*10).Should(gexec.Exit(1))

			Ω(session).Should(gbytes.Say("Sleeping..."))
			Ω(session).Should(gbytes.Say("Got your signal, but still taking a nap"), "the timeout has signaled the it to stop, but it's napping...")
			Ω(session).Should(gbytes.Say("A running node failed to exit in time"), "so we forcibly casue it to exit after the grace-period elapses")
			Ω(session).Should(gbytes.Say("Cleaning up once..."))
			Ω(session).Should(gbytes.Say("Cleaning up twice..."))
			Ω(session).Should(gbytes.Say("Cleaning up thrice..."), "we manage to get here even though the second after-each gets stuck.  that's thanks to the GracePeriod configuration.")

			Ω(session).Should(gbytes.Say(`\[TIMEDOUT\]`))
			Ω(session).Should(gbytes.Say(`Spec Goroutine`))
			Ω(session).Should(gbytes.Say(`goroutine \d+ \[select\]`))
			Ω(session).Should(gbytes.Say(`>\s*select {`), "The actual source code gets emitted now")
			Ω(session).ShouldNot(gbytes.Say(`Other Goroutines`))

			Ω(session).Should(gbytes.Say("FAIL! - Suite Timeout Elapsed"))
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
