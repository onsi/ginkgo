package integration_test

import (
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("OutputInterceptor", func() {
	Context("excercising the edge case reported in issue #851", func() {
		BeforeEach(func() {
			fm.MountFixture("interceptor")

			cmd := exec.Command("go", "build")
			cmd.Dir = fm.PathTo("interceptor")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Ω(err).ShouldNot(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))

			Ω(fm.PathTo("interceptor", "interceptor")).Should(BeAnExistingFile())
		})

		It("exercises the edge case reported in issue #851 - by asserting that output interception does not hang indefinitely if a process is spawned with cmd.Stdout=os.Stdout", func() {
			sess := startGinkgo(fm.PathTo("interceptor"), "--no-color")
			Eventually(sess).Should(gexec.Exit(0))
		})
	})

	Context("pausing/resuming output interception", func() {
		BeforeEach(func() {
			fm.MountFixture("pause_resume_interception")
		})

		It("can pause and resume interception", func() {
			sess := startGinkgo(fm.PathTo("pause_resume_interception"), "--no-color", "--procs=2", "--json-report=report.json")
			Eventually(sess).Should(gexec.Exit(0))

			output := string(sess.Out.Contents())
			Ω(output).Should(ContainSubstring("    CAPTURED OUTPUT A\n"))
			Ω(output).Should(ContainSubstring("    CAPTURED OUTPUT B\n"))

			Ω(output).ShouldNot(ContainSubstring("OUTPUT TO CONSOLE"))

			report := fm.LoadJSONReports("pause_resume_interception", "report.json")[0]
			Ω(report.SpecReports[0].CapturedStdOutErr).Should(Equal("CAPTURED OUTPUT A\nCAPTURED OUTPUT B\n"))
		})
	})
})
