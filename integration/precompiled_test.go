package integration_test

import (
	"os"
	"os/exec"
	"path"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo/v2"
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

	It("should have the symbols in the compiled binary", func() {
		cmd := exec.Command("go", "tool", "nm", fm.PathTo("passing_ginkgo_tests", "passing_ginkgo_tests.test"))
		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Ω(err).ShouldNot(HaveOccurred())
		Eventually(session).Should(gexec.Exit(0))
		Ω(session).Should(gbytes.Say("github.com/onsi/ginkgo/v2.It")) // a symbol from ginkgo
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

var _ = Describe("ginkgo build with multiple suites", Label("build"), func() {
	It("should correctly report multiple test binaries", func() {
		fm.MountFixture("build_reporting")
		session := startGinkgo(fm.PathTo("build_reporting"), "build", "-r")
		Eventually(session).Should(gexec.Exit(0))
		output := string(session.Out.Contents())
		Ω(output).Should(ContainSubstring("Compiled suite1/suite1.test"))
		Ω(output).Should(ContainSubstring("Compiled suite2/suite2.test"))
		Ω(fm.PathTo("build_reporting", "suite1", "suite1.test")).Should(BeAnExistingFile())
		Ω(fm.PathTo("build_reporting", "suite2", "suite2.test")).Should(BeAnExistingFile())
	})
})

var _ = Describe("ginkgo build with custom output", Label("build"), func() {
	const customPath = "mycustomdir"
	var fullPath string

	BeforeEach(func() {
		fm.MountFixture("passing_ginkgo_tests")
		fullPath = fm.PathTo("passing_ginkgo_tests", customPath)
		Ω(os.Mkdir(fullPath, 0777)).To(Succeed())

		DeferCleanup(func() {
			Ω(os.RemoveAll(fullPath)).Should(Succeed())
		})
	})

	It("should build with custom path", func() {
		session := startGinkgo(fm.PathTo("passing_ginkgo_tests"), "build", "-o", customPath+"/mytestapp")
		Eventually(session).Should(gexec.Exit(0))
		output := string(session.Out.Contents())
		Ω(output).Should(And(ContainSubstring("Compiled"), ContainSubstring(customPath+"/mytestapp")))
		Ω(path.Join(fullPath, "/mytestapp")).Should(BeAnExistingFile())
	})

	It("should build with custom directory", func() {
		session := startGinkgo(fm.PathTo("passing_ginkgo_tests"), "build", "-o", customPath)
		Eventually(session).Should(gexec.Exit(0))
		output := string(session.Out.Contents())
		Ω(output).Should(And(ContainSubstring("Compiled"), ContainSubstring(customPath+"/passing_ginkgo_tests.test")))
		Ω(path.Join(fullPath, "/passing_ginkgo_tests.test")).Should(BeAnExistingFile())
	})
})
