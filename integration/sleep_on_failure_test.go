package integration_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("--sleep-on-failure", func() {
	BeforeEach(func() {
		fm.MountFixture("sleep_on_failure")
	})

	Context("when running serially with --sleep-on-failure set", func() {
		It("pauses a failed spec before teardown, then proceeds to teardown", func() {
			start := time.Now()
			session := startGinkgo(fm.PathTo("sleep_on_failure"), "--no-color", "--sleep-on-failure=2s")
			Eventually(session).Should(gexec.Exit(1))
			elapsed := time.Since(start)
			output := string(session.Out.Contents())

			// it actually waited for (about) the configured duration
			Ω(elapsed).Should(BeNumerically(">=", 1500*time.Millisecond), "should have paused for ~2s")

			// it announced the pause and what to do
			Ω(output).Should(ContainSubstring("Paused on failure"))

			// it paused before teardown: the failing body runs, then the pause, then teardown
			Ω(output).Should(MatchRegexp(`(?s)FAILING-SPEC-BODY-RAN.*Paused on failure.*TEARDOWN-RAN`),
				"the pause must occur after the failing spec body and before its teardown")

			// teardown still ran after the pause
			Ω(output).Should(ContainSubstring("TEARDOWN-RAN"))
		})

		It("does not pause specs that pass", func() {
			session := startGinkgo(fm.PathTo("sleep_on_failure"), "--no-color", "--sleep-on-failure=1h", "--focus=passes and is never paused")
			// if the passing spec were paused, this would hang for an hour; the suite timeout/Eventually guards us
			Eventually(session, 30*time.Second).Should(gexec.Exit(0))
			output := string(session.Out.Contents())

			Ω(output).Should(ContainSubstring("PASSING-SPEC-RAN"))
			Ω(output).ShouldNot(ContainSubstring("Paused on failure"))
		})
	})

	Context("when running in parallel with --sleep-on-failure set", func() {
		It("exits with a helpful error instead of pausing", func() {
			start := time.Now()
			session := startGinkgo(fm.PathTo("sleep_on_failure"), "--no-color", "--procs=2", "--sleep-on-failure=1h")
			Eventually(session).Should(gexec.Exit(1))
			elapsed := time.Since(start)
			output := string(session.Out.Contents()) + string(session.Err.Contents())

			// it must not have paused (would have hung for an hour)
			Ω(elapsed).Should(BeNumerically("<", time.Minute))

			// it explains the serial-only restriction
			Ω(output).Should(ContainSubstring("Ginkgo only supports --sleep-on-failure in serial mode"))
			Ω(output).ShouldNot(ContainSubstring("Paused on failure"))
		})
	})
})
