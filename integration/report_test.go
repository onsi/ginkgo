package integration_test

import (
	"io/ioutil"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Report", func() {
	BeforeEach(func() {
		fm.MountFixture("reporting")
	})

	It("reports on each test via ReportAfterEach", func() {
		session := startGinkgo(fm.PathTo("reporting"), "--no-color")
		Eventually(session).Should(gexec.Exit(1))

		report, err := ioutil.ReadFile(fm.PathTo("reporting", "report-after-each.out"))
		Ω(err).ShouldNot(HaveOccurred())
		lines := strings.Split(string(report), "\n")
		Ω(lines).Should(ConsistOf(
			"passes - passed",
			"fails - failed",
			"panics - panicked",
			"is pending - pending",
			"is skipped - skipped",
			"",
		))
	})

	It("reports on all the tests via ReportAfterSuite", func() {
		session := startGinkgo(fm.PathTo("reporting"), "--no-color", "--seed=17")
		Eventually(session).Should(gexec.Exit(1))

		report, err := ioutil.ReadFile(fm.PathTo("reporting", "report-after-suite.out"))
		Ω(err).ShouldNot(HaveOccurred())
		lines := strings.Split(string(report), "\n")
		Ω(lines).Should(ConsistOf(
			"ReportingFixture Suite - 17",
			"1: [BeforeSuite] - passed",
			"passes - passed",
			"fails - failed",
			"panics - panicked",
			"is pending - pending",
			"is skipped - skipped",
			"1: [AfterSuite] - passed",
			"",
		))
	})

	Context("when running in parallel", func() {
		It("reports on all the tests via ReportAfterSuite", func() {
			session := startGinkgo(fm.PathTo("reporting"), "--no-color", "--seed=17", "--nodes=2")
			Eventually(session).Should(gexec.Exit(1))

			report, err := ioutil.ReadFile(fm.PathTo("reporting", "report-after-suite.out"))
			Ω(err).ShouldNot(HaveOccurred())
			lines := strings.Split(string(report), "\n")
			Ω(lines).Should(ConsistOf(
				"ReportingFixture Suite - 17",
				"1: [BeforeSuite] - passed",
				"2: [BeforeSuite] - passed",
				"passes - passed",
				"fails - failed",
				"panics - panicked",
				"is pending - pending",
				"is skipped - skipped",
				"1: [AfterSuite] - passed",
				"2: [AfterSuite] - passed",
				"",
			))
		})
	})

	Context("when a ReportAfterSuite node fails", func() {
		It("reports on it", func() {
			session := startGinkgo(fm.PathTo("reporting"), "--no-color", "--seed=17", "--nodes=2")
			Eventually(session).Should(gexec.Exit(1))

			Ω(string(session.Out.Contents())).Should(ContainSubstring("[ReportAfterSuite] my report"))
			Ω(string(session.Out.Contents())).Should(ContainSubstring("fail!\n  In [ReportAfterSuite] at:"))
		})
	})
})
