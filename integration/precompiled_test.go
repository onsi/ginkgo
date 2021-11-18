package integration_test

import (
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("ginkgo build", func() {
	BeforeEach(func() {
		fm.MountFixture("passing_ginkgo_tests")
		session := startGinkgo(fm.PathTo("passing_ginkgo_tests"), "build")
		Eventually(session).Should(gexec.Exit(0))
		output := string(session.Out.Contents())
		Ω(output).Should(ContainSubstring("Compiled passing_ginkgo_tests.test"))
	})

	It("should build a test binary", func() {
		Ω(fm.PathTo("passing_ginkgo_tests", "passing_ginkgo_tests.test")).Should(BeAnExistingFile())
	})

	It("should be possible to run the test binary directly", func() {
		cmd := exec.Command("./passing_ginkgo_tests.test")
		cmd.Dir = fm.PathTo("passing_ginkgo_tests")
		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Ω(err).ShouldNot(HaveOccurred())
		Eventually(session).Should(gexec.Exit(0))
		Ω(session).Should(gbytes.Say("Running Suite: Passing_ginkgo_tests Suite"))
	})

	It("should be possible to run the test binary via ginkgo", func() {
		session := startGinkgo(fm.PathTo("passing_ginkgo_tests"), "./passing_ginkgo_tests.test")
		Eventually(session).Should(gexec.Exit(0))
		Ω(session).Should(gbytes.Say("Running Suite: Passing_ginkgo_tests Suite"))
	})

	It("should be possible to run the test binary in parallel", func() {
		session := startGinkgo(fm.PathTo("passing_ginkgo_tests"), "--procs=2", "--no-color", "./passing_ginkgo_tests.test")
		Eventually(session).Should(gexec.Exit(0))
		Ω(session).Should(gbytes.Say("Running Suite: Passing_ginkgo_tests Suite"))
		Ω(session).Should(gbytes.Say("Running in parallel across 2 processes"))
	})
})
