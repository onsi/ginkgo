package integration_test

import (
	"regexp"
	"runtime"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Verbose And Succinct Mode", func() {
	denoter := "•"

	if runtime.GOOS == "windows" {
		denoter = "+"
	}

	Context("when running one package", func() {
		BeforeEach(func() {
			fm.MountFixture("passing_ginkgo_tests")
		})

		It("should default to non-succinct mode", func() {
			session := startGinkgo(fm.PathTo("passing_ginkgo_tests"), "--no-color")
			Eventually(session).Should(gexec.Exit(0))
			output := session.Out.Contents()

			Ω(output).Should(ContainSubstring("Running Suite: Passing_ginkgo_tests Suite"))
		})
	})

	Context("when running more than one package", func() {
		BeforeEach(func() {
			fm.MountFixture("passing_ginkgo_tests")
			fm.MountFixture("more_ginkgo_tests")
		})

		Context("with no flags set", func() {
			It("should default to succinct mode", func() {
				session := startGinkgo(fm.TmpDir, "--no-color", "passing_ginkgo_tests", "more_ginkgo_tests")
				Eventually(session).Should(gexec.Exit(0))
				output := session.Out.Contents()

				Ω(output).Should(MatchRegexp(`\] Passing_ginkgo_tests Suite - 4/4 specs [%s]{4} SUCCESS!`, regexp.QuoteMeta(denoter)))
				Ω(output).Should(MatchRegexp(`\] More_ginkgo_tests Suite - 2/2 specs [%s]{2} SUCCESS!`, regexp.QuoteMeta(denoter)))
			})
		})

		Context("with --succinct=false", func() {
			It("should not be in succinct mode", func() {
				session := startGinkgo(fm.TmpDir, "--no-color", "--succinct=false", "passing_ginkgo_tests", "more_ginkgo_tests")
				Eventually(session).Should(gexec.Exit(0))
				output := session.Out.Contents()

				Ω(output).Should(ContainSubstring("Running Suite: Passing_ginkgo_tests Suite"))
				Ω(output).Should(ContainSubstring("Running Suite: More_ginkgo_tests Suite"))
			})
		})

		Context("with -v", func() {
			It("should not be in succinct mode, but should be verbose", func() {
				session := startGinkgo(fm.TmpDir, "--no-color", "-v", "passing_ginkgo_tests", "more_ginkgo_tests")
				Eventually(session).Should(gexec.Exit(0))
				output := session.Out.Contents()

				Ω(output).Should(ContainSubstring("Running Suite: Passing_ginkgo_tests Suite"))
				Ω(output).Should(ContainSubstring("Running Suite: More_ginkgo_tests Suite"))
				Ω(output).Should(ContainSubstring("should proxy strings"))
				Ω(output).Should(ContainSubstring("should always pass"))
			})

			It("should emit output from Bys", func() {
				session := startGinkgo(fm.PathTo("passing_ginkgo_tests"), "--no-color", "-v")
				Eventually(session).Should(gexec.Exit(0))
				output := session.Out.Contents()

				Ω(output).Should(ContainSubstring("emitting one By"))
				Ω(output).Should(ContainSubstring("emitting another By"))
			})
		})

		Context("with -vv", func() {
			It("should not be in succinct mode, but should be verbose", func() {
				session := startGinkgo(fm.TmpDir, "--no-color", "-vv", "passing_ginkgo_tests", "more_ginkgo_tests")
				Eventually(session).Should(gexec.Exit(0))
				output := session.Out.Contents()

				Ω(output).Should(ContainSubstring("Running Suite: Passing_ginkgo_tests Suite"))
				Ω(output).Should(ContainSubstring("Running Suite: More_ginkgo_tests Suite"))
				Ω(output).Should(ContainSubstring("should proxy strings"))
				Ω(output).Should(ContainSubstring("should always pass"))
			})

			It("should emit output from Bys", func() {
				session := startGinkgo(fm.PathTo("passing_ginkgo_tests"), "--no-color", "-vv")
				Eventually(session).Should(gexec.Exit(0))
				output := session.Out.Contents()

				Ω(output).Should(ContainSubstring("emitting one By"))
				Ω(output).Should(ContainSubstring("emitting another By"))
			})
		})
	})
})
