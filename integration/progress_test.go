package integration_test

import (
	"fmt"
	"os"
	"path/filepath"
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
			fm.MountFixture("progress_report")
		})

		It("emits progress when a singal is sent and when tests take too long", func() {
			session := startGinkgo(fm.PathTo("progress_report"), "--poll-progress-after=1500ms", "--poll-progress-interval=200ms", "--no-color")
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

			Eventually(session).Should(gbytes.Say(`Begin Captured GinkgoWriter Output`))
			Eventually(session).Should(gbytes.Say(`\.\.\.`))
			for i := 3; i <= 12; i++ {
				Eventually(session).Should(gbytes.Say(fmt.Sprintf("ginkgo-writer-output-%d", i)))
			}

			Eventually(session).Should(gbytes.Say(`|\s*fmt\.Println\("READY"\)`))
			Eventually(session).Should(gbytes.Say(`>\s*time\.Sleep\(time\.Second\)`))

			//first poll
			Eventually(session).Should(gbytes.Say(`--poll-progress-after tracks things that take too long \(Spec Runtime: 1\.5\d*s\)`))
			Eventually(session).Should(gbytes.Say(`>\s*time.Sleep\(2 \* time\.Second\)`))
			Eventually(session).Should(gbytes.Say(`Begin Additional Progress Reports >>`))
			Eventually(session).Should(gbytes.Say(`Some global information: 1`))
			Eventually(session).Should(gbytes.Say(`<< End Additional Progress Reports`))

			//second poll
			Eventually(session).Should(gbytes.Say(`--poll-progress-after tracks things that take too long \(Spec Runtime: 1\.7\d*s\)`))
			Eventually(session).Should(gbytes.Say(`>\s*time.Sleep\(2 \* time\.Second\)`))

			//decorator poll
			Eventually(session).Should(gbytes.Say(`decorator tracks things that take too long \(Spec Runtime: 5[\.\d]*ms\)`))
			Eventually(session).Should(gbytes.Say(`>\s*time\.Sleep\(1 \* time\.Second\)`))

			Eventually(session).Should(gexec.Exit(0))
		})

		It("allows the user to specify a source-root to find source code files", func() {
			// first we build the test with -gcflags=all=-trimpath=<PATH TO SPEC> to ensure
			// that stack traces do not contain absolute paths
			path, err := filepath.Abs(fm.PathTo("progress_report"))
			Ω(err).ShouldNot(HaveOccurred())
			session := startGinkgo(fm.PathTo("progress_report"), "build", `-gcflags=-trimpath=`+path+``)
			Eventually(session).Should(gexec.Exit(0))

			// now we move the compiled test binary to a separate directory
			fm.MkEmpty("progress_report/suite")
			os.Rename(fm.PathTo("progress_report", "progress_report.test"), fm.PathTo("progress_report", "suite", "progress_report.test"))

			//and we run and confirm that we don't see the expected source code
			session = startGinkgo(fm.PathTo("progress_report", "suite"), "--poll-progress-after=1500ms", "--poll-progress-interval=200ms", "--no-color", "-label-filter=one-second", "./progress_report.test")
			Eventually(session).Should(gexec.Exit(0))
			Ω(session).ShouldNot(gbytes.Say(`>\s*time.Sleep\(1 \* time\.Second\)`))

			// now we run, but configured with source-root and see that we have the file
			// note that multipel source-roots can be passed in
			session = startGinkgo(fm.PathTo("progress_report", "suite"), "--poll-progress-after=1500ms", "--poll-progress-interval=200ms", "--no-color", "-label-filter=one-second", "--source-root=/tmp", "--source-root="+path, "./progress_report.test")
			Eventually(session).Should(gbytes.Say(`>\s*time\.Sleep\(1 \* time\.Second\)`))
			Eventually(session).Should(gexec.Exit(0))
		})

		It("emits progress immediately and includes process information when running in parallel", func() {
			session := startGinkgo(fm.PathTo("progress_report"), "--poll-progress-after=1500ms", "--poll-progress-interval=200ms", "--no-color", "-procs=2", "-label-filter=parallel")
			Eventually(session).Should(gexec.Exit(0))

			Eventually(session.Out.Contents()).Should(ContainSubstring(`Progress Report for Ginkgo Process #1`))
			Eventually(session.Out.Contents()).Should(ContainSubstring(`Progress Report for Ginkgo Process #2`))
			Eventually(session.Out.Contents()).Should(ContainSubstring(`Some global information: 1`))
			Eventually(session.Out.Contents()).Should(ContainSubstring(`Some global information: 2`))

		})

	})
})
