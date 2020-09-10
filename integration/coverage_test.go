package integration_test

import (
	"os/exec"
	"regexp"

	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Coverage Specs", func() {
	BeforeEach(func() {
		fm.MountFixture("coverage")
	})

	processCoverageProfile := func(path string) string {
		profileOutput, err := exec.Command("go", "tool", "cover", fmt.Sprintf("-func=%s", path)).CombinedOutput()
		ExpectWithOffset(1, err).ShouldNot(HaveOccurred())
		return string(profileOutput)
	}

	Context("when running a single package in series or in parallel with -cover", func() {
		It("emits the coverage pecentage and generates a cover profile", func() {
			seriesSession := startGinkgo(fm.PathTo("coverage"), "--no-color", "-cover")
			Eventually(seriesSession).Should(gexec.Exit(0))
			Ω(seriesSession.Out).Should(gbytes.Say(`coverage: 80\.0% of statements`))
			seriesCoverage := processCoverageProfile(fm.PathTo("coverage", "coverprofile.out"))
			fm.RemoveFile("coverage", "coverprofile.out")

			parallelSession := startGinkgo(fm.PathTo("coverage"), "--no-color", "-nodes=2", "-cover")
			Eventually(parallelSession).Should(gexec.Exit(0))
			parallelCoverage := processCoverageProfile(fm.PathTo("coverage", "coverprofile.out"))

			Ω(parallelCoverage).Should(Equal(seriesCoverage))
		})
	})

	Context("with -coverpkg", func() {
		It("computes coverage of the passed-in additional packages", func() {
			coverPkgFlag := fmt.Sprintf("-coverpkg=%s,%s", fm.PackageNameFor("coverage"), fm.PackageNameFor("coverage/external_coverage"))
			seriesSession := startGinkgo(fm.PathTo("coverage"), coverPkgFlag)
			Eventually(seriesSession).Should(gexec.Exit(0))
			Ω(seriesSession.Out).Should(gbytes.Say("coverage: 71.4% of statements in"))
			seriesCoverage := processCoverageProfile(fm.PathTo("coverage", "coverprofile.out"))
			fm.RemoveFile("coverage", "coverprofile.out")

			parallelSession := startGinkgo(fm.PathTo("coverage"), "--no-color", "-nodes=2", coverPkgFlag)
			Eventually(parallelSession).Should(gexec.Exit(0))
			parallelCoverage := processCoverageProfile(fm.PathTo("coverage", "coverprofile.out"))

			Ω(parallelCoverage).Should(Equal(seriesCoverage))
		})
	})

	Context("with a custom profile name", func() {
		It("generates cover profiles with the specified name", func() {
			session := startGinkgo(fm.PathTo("coverage"), "--no-color", "-coverprofile=myprofile.out")
			Eventually(session).Should(gexec.Exit(0))
			Ω(session.Out).Should(gbytes.Say(`coverage: 80\.0% of statements`))
			Ω(fm.PathTo("coverage", "myprofile.out")).Should(BeAnExistingFile())
			Ω(fm.PathTo("coverage", "coverprofile.out")).ShouldNot(BeAnExistingFile())
		})
	})

	Context("when multiple suites are tested", func() {
		BeforeEach(func() {
			fm.MountFixture("combined_coverage")
		})

		It("generates a single cover profile", func() {
			session := startGinkgo(fm.PathTo("combined_coverage"), "--no-color", "--cover", "-r", "--covermode=atomic")
			Eventually(session).Should(gexec.Exit(0))
			Ω(fm.PathTo("combined_coverage", "coverprofile.out")).Should(BeAnExistingFile())
			Ω(fm.PathTo("combined_coverage", "first_package/coverprofile.out")).ShouldNot(BeAnExistingFile())
			Ω(fm.PathTo("combined_coverage", "second_package/coverprofile.out")).ShouldNot(BeAnExistingFile())

			By("ensuring there is only one 'mode:' line")
			re := regexp.MustCompile(`mode: atomic`)
			content := fm.ContentOf("combined_coverage", "coverprofile.out")
			matches := re.FindAllStringIndex(content, -1)
			Ω(len(matches)).Should(Equal(1))
		})

		Context("when -keep-separate-coverprofiles is set", func() {
			It("generates separate coverprofiles", func() {
				session := startGinkgo(fm.PathTo("combined_coverage"), "--no-color", "--cover", "-r", "--keep-separate-coverprofiles")
				Eventually(session).Should(gexec.Exit(0))
				Ω(fm.PathTo("combined_coverage", "coverprofile.out")).ShouldNot(BeAnExistingFile())
				Ω(fm.PathTo("combined_coverage", "first_package/coverprofile.out")).Should(BeAnExistingFile())
				Ω(fm.PathTo("combined_coverage", "second_package/coverprofile.out")).Should(BeAnExistingFile())
			})
		})
	})

	Context("when -output-dir is set", func() {
		BeforeEach(func() {
			fm.MountFixture("combined_coverage")
		})

		It("puts the cover profile in -output-dir", func() {
			session := startGinkgo(fm.PathTo("combined_coverage"), "--no-color", "--cover", "-r", "--output-dir=./output")
			Eventually(session).Should(gexec.Exit(0))
			Ω(fm.PathTo("combined_coverage", "output/coverprofile.out")).Should(BeAnExistingFile())
		})

		Context("when -keep-separate-coverprofiles is set", func() {
			It("puts namespaced coverprofiels in the -output-dir", func() {
				session := startGinkgo(fm.PathTo("combined_coverage"), "--no-color", "--cover", "-r", "--output-dir=./output", "--keep-separate-coverprofiles")
				Eventually(session).Should(gexec.Exit(0))
				Ω(fm.PathTo("combined_coverage", "output/coverprofile.out")).ShouldNot(BeAnExistingFile())
				Ω(fm.PathTo("combined_coverage", "output/first_package_coverprofile.out")).Should(BeAnExistingFile())
				Ω(fm.PathTo("combined_coverage", "output/second_package_coverprofile.out")).Should(BeAnExistingFile())
			})
		})
	})
})
