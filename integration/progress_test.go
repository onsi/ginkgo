package integration_test

import (
	"regexp"
	"strconv"
	"syscall"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Emitting progress", func() {
	Describe("progress reports", func() {
		BeforeEach(func() {
			fm.MountFixture("progress_reporter")
		})

		It("emits progress when a singal is sent and when tests take too long", func() {
			session := startGinkgo(fm.PathTo("progress_reporter"), "--poll-progress-after=1500ms", "--poll-progress-interval=200ms", "--no-color")
			Eventually(session).Should(gbytes.Say(`READY `))
			buf := make([]byte, 128)
			_, err := session.Out.Read(buf)
			Ω(err).ShouldNot(HaveOccurred())
			pid, err := strconv.Atoi(regexp.MustCompile(`\d*`).FindString(string(buf)))
			Ω(err).ShouldNot(HaveOccurred())

			syscall.Kill(pid, syscall.SIGUSR1)
			Eventually(session).Should(gbytes.Say(`can track on demand \(Spec Runtime:`))
			Eventually(session).Should(gbytes.Say(`In \[It\] \(Node Runtime:`))
			Eventually(session).Should(gbytes.Say(`\[By Step\] Step B \(Step Runtime:`))
			Eventually(session).Should(gbytes.Say(`|\s*fmt\.Println\("READY"\)`))
			Eventually(session).Should(gbytes.Say(`>\s*time\.Sleep\(time\.Second\)`))

			//first poll
			Eventually(session).Should(gbytes.Say(`--poll-progress-after tracks things that take too long \(Spec Runtime: 1\.5\d*s\)`))
			Eventually(session).Should(gbytes.Say(`>\s*time.Sleep\(2 \* time\.Second\)`))

			//second poll
			Eventually(session).Should(gbytes.Say(`--poll-progress-after tracks things that take too long \(Spec Runtime: 1\.7\d*s\)`))
			Eventually(session).Should(gbytes.Say(`>\s*time.Sleep\(2 \* time\.Second\)`))

			//decorator poll
			Eventually(session).Should(gbytes.Say(`decorator tracks things that take too long \(Spec Runtime: 5[\.\d]*ms\)`))
			Eventually(session).Should(gbytes.Say(`>\s*time\.Sleep\(1 \* time\.Second\)`))

			Eventually(session).Should(gexec.Exit(0))
		})

		It("emits progress immediately and includes process information when running in parallel", func() {
			session := startGinkgo(fm.PathTo("progress_reporter"), "--poll-progress-after=1500ms", "--poll-progress-interval=200ms", "--no-color", "-procs=2", "-label-filter=parallel")
			Eventually(session).Should(gexec.Exit(0))

			Eventually(session.Err.Contents()).Should(ContainSubstring(`Progress Report for Ginkgo Process #1`))
			Eventually(session.Err.Contents()).Should(ContainSubstring(`Progress Report for Ginkgo Process #2`))

		})

	})

	Describe("the --progress flag", func() {
		var session *gexec.Session
		var args []string
		BeforeEach(func() {
			args = []string{"--no-color"}
			fm.MountFixture("progress")
		})

		JustBeforeEach(func() {
			session = startGinkgo(fm.PathTo("progress"), args...)
			Eventually(session).Should(gexec.Exit(0))
		})

		Context("with the -progress flag, but no -v flag", func() {
			BeforeEach(func() {
				args = append(args, "-progress")
			})

			It("should not emit progress", func() {
				Ω(session).ShouldNot(gbytes.Say("[bB]efore"))
			})
		})

		Context("with the -v flag", func() {
			BeforeEach(func() {
				args = append(args, "-v")
			})

			It("should not emit progress", func() {
				Ω(session).ShouldNot(gbytes.Say(`\[BeforeEach\]`))
				Ω(session).Should(gbytes.Say(`>outer before<`))
			})
		})

		Context("with the -progress flag and the -v flag", func() {
			BeforeEach(func() {
				args = append(args, "-progress", "-v")
			})

			It("should emit progress (by writing to the GinkgoWriter)", func() {
				// First spec

				Ω(session).Should(gbytes.Say(`\[BeforeEach\] ProgressFixture`))
				Ω(session).Should(gbytes.Say(`>outer before<`))

				Ω(session).Should(gbytes.Say(`\[BeforeEach\] Inner Context`))
				Ω(session).Should(gbytes.Say(`>inner before<`))

				Ω(session).Should(gbytes.Say(`\[BeforeEach\] when Inner When`))
				Ω(session).Should(gbytes.Say(`>inner before<`))

				Ω(session).Should(gbytes.Say(`\[JustBeforeEach\] ProgressFixture`))
				Ω(session).Should(gbytes.Say(`>outer just before<`))

				Ω(session).Should(gbytes.Say(`\[JustBeforeEach\] Inner Context`))
				Ω(session).Should(gbytes.Say(`>inner just before<`))

				Ω(session).Should(gbytes.Say(`\[It\] should emit progress as it goes`))
				Ω(session).Should(gbytes.Say(`>it<`))

				Ω(session).Should(gbytes.Say(`\[AfterEach\] Inner Context`))
				Ω(session).Should(gbytes.Say(`>inner after<`))

				Ω(session).Should(gbytes.Say(`\[AfterEach\] ProgressFixture`))
				Ω(session).Should(gbytes.Say(`>outer after<`))

				// Second spec

				Ω(session).Should(gbytes.Say(`\[BeforeEach\] ProgressFixture`))
				Ω(session).Should(gbytes.Say(`>outer before<`))

				Ω(session).Should(gbytes.Say(`\[JustBeforeEach\] ProgressFixture`))
				Ω(session).Should(gbytes.Say(`>outer just before<`))

				Ω(session).Should(gbytes.Say(`\[It\] should emit progress as it goes`))
				Ω(session).Should(gbytes.Say(`>specify<`))

				Ω(session).Should(gbytes.Say(`\[AfterEach\] ProgressFixture`))
				Ω(session).Should(gbytes.Say(`>outer after<`))
			})
		})
	})
})
