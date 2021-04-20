package integration_test

import (
	"encoding/json"
	"io/ioutil"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/internal/test_helpers"
	"github.com/onsi/ginkgo/types"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Reporting", func() {
	BeforeEach(func() {
		fm.MountFixture("reporting")
	})

	Describe("in-suite reporting with ReportAfterEach and ReportAfterSuite", func() {
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

	Describe("JSON reporting", func() {
		loadReports := func(pkg string, target string) []types.Report {
			data := []byte(fm.ContentOf(pkg, target))
			reports := []types.Report{}
			Ω(json.Unmarshal(data, &reports)).Should(Succeed())
			return reports
		}

		checkReport := func(report types.Report) {
			Ω(report.SuitePath).Should(Equal(fm.AbsPathTo("reporting")))
			Ω(report.SuiteDescription).Should(Equal("ReportingFixture Suite"))
			Ω(report.SuiteConfig.ParallelTotal).Should(Equal(2))
			Ω(report.SpecReports).Should(HaveLen(5 + 4 + 1))

			specReports := Reports(report.SpecReports)
			Ω(specReports.WithLeafNodeType(types.NodeTypeIt)).Should(HaveLen(5))
			Ω(specReports.Find("passes")).Should(HavePassed())
			Ω(specReports.Find("fails")).Should(HaveFailed("fail!", types.FailureNodeIsLeafNode, CapturedGinkgoWriterOutput("some ginkgo-writer output")))
			Ω(specReports.Find("panics")).Should(HavePanicked("boom"))
			Ω(specReports.Find("is pending")).Should(BePending())
			Ω(specReports.Find("is skipped").State).Should(Equal(types.SpecStateSkipped))
			Ω(specReports.Find("my report")).Should(HaveFailed("fail!", types.FailureNodeIsLeafNode, types.NodeTypeReportAfterSuite))
			Ω(specReports.FindByLeafNodeType(types.NodeTypeBeforeSuite)).Should(HavePassed())
			Ω(specReports.FindByLeafNodeType(types.NodeTypeAfterSuite)).Should(HavePassed())
		}

		checkSubpackageReport := func(report types.Report) {
			Ω(report.SuitePath).Should(Equal(fm.AbsPathTo("reporting", "reporting_sub_package")))
			Ω(report.SuiteDescription).Should(Equal("Reporting SubPackage Suite"))
			Ω(report.SuiteConfig.ParallelTotal).Should(Equal(2))
			Ω(report.SpecReports).Should(HaveLen(3))

			specReports := Reports(report.SpecReports)
			Ω(specReports.Find("passes here too")).Should(HavePassed())
			Ω(specReports.Find("fails here too")).Should(HaveFailed("fail!", types.FailureNodeIsLeafNode, CapturedStdOutput("some std output")))
			Ω(specReports.Find("panics here too")).Should(HavePanicked("bam!"))
		}

		Context("the default behavior", func() {
			BeforeEach(func() {
				session := startGinkgo(fm.PathTo("reporting"), "--no-color", "-r", "--keep-going", "--nodes=2", "--json-report=out.json")
				Eventually(session).Should(gexec.Exit(1))
			})

			It("generates a single unified json report", func() {
				reports := loadReports("reporting", "out.json")
				Ω(reports).Should(HaveLen(2))
				checkReport(reports[0])
				checkSubpackageReport(reports[1])
			})
		})

		Context("with -output-dir", func() {
			BeforeEach(func() {
				session := startGinkgo(fm.PathTo("reporting"), "--no-color", "-r", "--keep-going", "--nodes=2", "--json-report=out.json", "--output-dir=./reports")
				Eventually(session).Should(gexec.Exit(1))
			})

			It("places the single unified json report in output-dir", func() {
				reports := loadReports("reporting", "reports/out.json")
				Ω(reports).Should(HaveLen(2))
				checkReport(reports[0])
				checkSubpackageReport(reports[1])
			})
		})

		Context("with -keep-separate-reports", func() {
			BeforeEach(func() {
				session := startGinkgo(fm.PathTo("reporting"), "--no-color", "-r", "--keep-going", "--nodes=2", "--json-report=out.json", "--keep-separate-reports")
				Eventually(session).Should(gexec.Exit(1))
			})

			It("keeps the separate reports in their respective packages", func() {
				reports := loadReports("reporting", "out.json")
				Ω(reports).Should(HaveLen(1))
				checkReport(reports[0])

				reports = loadReports("reporting", "reporting_sub_package/out.json")
				Ω(reports).Should(HaveLen(1))
				checkSubpackageReport(reports[0])
			})
		})

		Context("with -keep-separate-reports and -output-dir", func() {
			BeforeEach(func() {
				session := startGinkgo(fm.PathTo("reporting"), "--no-color", "-r", "--keep-going", "--nodes=2", "--json-report=out.json", "--keep-separate-reports", "--output-dir=./reports")
				Eventually(session).Should(gexec.Exit(1))
			})

			It("places the separate reports in the -output-dir", func() {
				reports := loadReports("reporting", "reports/reporting_out.json")
				Ω(reports).Should(HaveLen(1))
				checkReport(reports[0])

				reports = loadReports("reporting", "reports/reporting_sub_package_out.json")
				Ω(reports).Should(HaveLen(1))
				checkSubpackageReport(reports[0])

			})
		})
	})
})
