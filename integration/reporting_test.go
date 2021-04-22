package integration_test

import (
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/internal/test_helpers"
	"github.com/onsi/ginkgo/reporters"
	"github.com/onsi/ginkgo/types"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
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

	Describe("JSON and JUnit reporting", func() {
		loadJSONReports := func(pkg string, target string) []types.Report {
			data := []byte(fm.ContentOf(pkg, target))
			reports := []types.Report{}
			Ω(json.Unmarshal(data, &reports)).Should(Succeed())
			return reports
		}

		checkJSONReport := func(report types.Report) {
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

		checkJSONSubpackageReport := func(report types.Report) {
			Ω(report.SuitePath).Should(Equal(fm.AbsPathTo("reporting", "reporting_sub_package")))
			Ω(report.SuiteDescription).Should(Equal("Reporting SubPackage Suite"))
			Ω(report.SuiteConfig.ParallelTotal).Should(Equal(2))
			Ω(report.SpecReports).Should(HaveLen(3))

			specReports := Reports(report.SpecReports)
			Ω(specReports.Find("passes here too")).Should(HavePassed())
			Ω(specReports.Find("fails here too")).Should(HaveFailed("fail!", types.FailureNodeIsLeafNode, CapturedStdOutput("some std output")))
			Ω(specReports.Find("panics here too")).Should(HavePanicked("bam!"))
		}

		checkJSONFailedCompilationReport := func(report types.Report) {
			Ω(report.SuitePath).Should(Equal(fm.AbsPathTo("reporting", "malformed_sub_package")))
			Ω(report.SuiteDescription).Should(Equal(""))
			Ω(report.SuiteConfig.RandomSeed).Should(Equal(int64(17)))
			Ω(report.SpecialSuiteFailureReason).Should(ContainSubstring("Failed to compile malformed_sub_package:"))
			Ω(report.SpecReports).Should(HaveLen(0))
		}

		loadJUnitReport := func(pkg string, target string) reporters.JUnitTestSuites {
			data := []byte(fm.ContentOf(pkg, target))
			reports := reporters.JUnitTestSuites{}
			Ω(xml.Unmarshal(data, &reports)).Should(Succeed())
			return reports
		}

		getTestCase := func(name string, tests []reporters.JUnitTestCase) reporters.JUnitTestCase {
			for _, test := range tests {
				if test.Name == name {
					return test
				}
			}

			return reporters.JUnitTestCase{}
		}

		checkJUnitReport := func(suite reporters.JUnitTestSuite) {
			Ω(suite.Name).Should(Equal("ReportingFixture Suite"))
			Ω(suite.Package).Should(Equal(fm.AbsPathTo("reporting")))
			Ω(suite.Tests).Should(Equal(10))
			Ω(suite.Disabled).Should(Equal(1))
			Ω(suite.Skipped).Should(Equal(1))
			Ω(suite.Errors).Should(Equal(1))
			Ω(suite.Failures).Should(Equal(2))
			Ω(suite.Properties.WithName("SuiteSucceeded")).Should(Equal("false"))
			Ω(suite.Properties.WithName("RandomSeed")).Should(Equal("17"))
			Ω(suite.Properties.WithName("ParallelTotal")).Should(Equal("2"))
			Ω(getTestCase("[BeforeSuite]", suite.TestCases).Status).Should(Equal("passed"))
			Ω(getTestCase("[It] reporting test passes", suite.TestCases).Classname).Should(Equal("ReportingFixture Suite"))
			Ω(getTestCase("[It] reporting test passes", suite.TestCases).Status).Should(Equal("passed"))
			Ω(getTestCase("[It] reporting test passes", suite.TestCases).Failure).Should(BeNil())
			Ω(getTestCase("[It] reporting test passes", suite.TestCases).Error).Should(BeNil())
			Ω(getTestCase("[It] reporting test passes", suite.TestCases).Skipped).Should(BeNil())

			Ω(getTestCase("[It] reporting test panics", suite.TestCases).Status).Should(Equal("panicked"))
			Ω(getTestCase("[It] reporting test panics", suite.TestCases).Error.Message).Should(Equal("boom"))
			Ω(getTestCase("[It] reporting test panics", suite.TestCases).Error.Type).Should(Equal("panicked"))

			Ω(getTestCase("[It] reporting test fails", suite.TestCases).Failure.Message).Should(Equal("fail!"))
			Ω(getTestCase("[It] reporting test fails", suite.TestCases).Status).Should(Equal("failed"))
			Ω(getTestCase("[It] reporting test fails", suite.TestCases).SystemErr).Should(Equal("some ginkgo-writer output"))

			Ω(getTestCase("[It] reporting test is pending", suite.TestCases).Status).Should(Equal("pending"))
			Ω(getTestCase("[It] reporting test is pending", suite.TestCases).Skipped.Message).Should(Equal("pending"))

			Ω(getTestCase("[ReportAfterSuite] my report", suite.TestCases).Status).Should(Equal("failed"))

			Ω(getTestCase("[It] reporting test is skipped", suite.TestCases).Status).Should(Equal("skipped"))
			Ω(getTestCase("[It] reporting test is skipped", suite.TestCases).Skipped.Message).Should(Equal("skipped - skip"))
		}

		checkJUnitSubpackageReport := func(suite reporters.JUnitTestSuite) {
			Ω(suite.Name).Should(Equal("Reporting SubPackage Suite"))
			Ω(suite.Package).Should(Equal(fm.AbsPathTo("reporting", "reporting_sub_package")))
			Ω(suite.Tests).Should(Equal(3))
			Ω(suite.Errors).Should(Equal(1))
			Ω(suite.Failures).Should(Equal(1))
		}

		checkJUnitFailedCompilationReport := func(suite reporters.JUnitTestSuite) {
			Ω(suite.Name).Should(Equal(""))
			Ω(suite.Package).Should(Equal(fm.AbsPathTo("reporting", "malformed_sub_package")))
			Ω(suite.Properties.WithName("SpecialSuiteFailureReason")).Should(ContainSubstring("Failed to compile malformed_sub_package:"))
		}

		checkUnifiedJUnitReport := func(report reporters.JUnitTestSuites) {
			Ω(report.TestSuites).Should(HaveLen(3))
			Ω(report.Tests).Should(Equal(13))
			Ω(report.Disabled).Should(Equal(2))
			Ω(report.Errors).Should(Equal(2))
			Ω(report.Failures).Should(Equal(3))

			checkJUnitReport(report.TestSuites[0])
			checkJUnitFailedCompilationReport(report.TestSuites[1])
			checkJUnitSubpackageReport(report.TestSuites[2])
		}

		Context("the default behavior", func() {
			BeforeEach(func() {
				session := startGinkgo(fm.PathTo("reporting"), "--no-color", "-r", "--keep-going", "--nodes=2", "--json-report=out.json", "--junit-report=out.xml", "-seed=17")
				Eventually(session).Should(gexec.Exit(1))
				Ω(session).ShouldNot(gbytes.Say("Could not open"))
			})

			It("generates single unified json and junit reports", func() {
				reports := loadJSONReports("reporting", "out.json")
				Ω(reports).Should(HaveLen(3))
				checkJSONReport(reports[0])
				checkJSONFailedCompilationReport(reports[1])
				checkJSONSubpackageReport(reports[2])

				junitReport := loadJUnitReport("reporting", "out.xml")
				checkUnifiedJUnitReport(junitReport)
			})
		})

		Context("with -output-dir", func() {
			BeforeEach(func() {
				session := startGinkgo(fm.PathTo("reporting"), "--no-color", "-r", "--keep-going", "--nodes=2", "--json-report=out.json", "--junit-report=out.xml", "--output-dir=./reports", "-seed=17")
				Eventually(session).Should(gexec.Exit(1))
				Ω(session).ShouldNot(gbytes.Say("Could not open"))
			})

			It("places the single unified json and junit reports in output-dir", func() {
				reports := loadJSONReports("reporting", "reports/out.json")
				Ω(reports).Should(HaveLen(3))
				checkJSONReport(reports[0])
				checkJSONFailedCompilationReport(reports[1])
				checkJSONSubpackageReport(reports[2])

				junitReport := loadJUnitReport("reporting", "reports/out.xml")
				checkUnifiedJUnitReport(junitReport)
			})
		})

		Context("with -keep-separate-reports", func() {
			BeforeEach(func() {
				session := startGinkgo(fm.PathTo("reporting"), "--no-color", "-r", "--keep-going", "--nodes=2", "--json-report=out.json", "--junit-report=out.xml", "--keep-separate-reports", "-seed=17")
				Eventually(session).Should(gexec.Exit(1))
				Ω(session).ShouldNot(gbytes.Say("Could not open"))
			})

			It("keeps the separate reports in their respective packages", func() {
				reports := loadJSONReports("reporting", "out.json")
				Ω(reports).Should(HaveLen(1))
				checkJSONReport(reports[0])
				checkJUnitReport(loadJUnitReport("reporting", "out.xml").TestSuites[0])

				reports = loadJSONReports("reporting", "reporting_sub_package/out.json")
				Ω(reports).Should(HaveLen(1))
				checkJSONSubpackageReport(reports[0])
				checkJUnitSubpackageReport(loadJUnitReport("reporting", "reporting_sub_package/out.xml").TestSuites[0])

				reports = loadJSONReports("reporting", "malformed_sub_package/out.json")
				Ω(reports).Should(HaveLen(1))
				checkJSONFailedCompilationReport(reports[0])
				checkJUnitFailedCompilationReport(loadJUnitReport("reporting", "malformed_sub_package/out.xml").TestSuites[0])

				Ω(fm.PathTo("reporting", "nonginkgo_sub_package/out.json")).ShouldNot(BeAnExistingFile())
				Ω(fm.PathTo("reporting", "nonginkgo_sub_package/out.xml")).ShouldNot(BeAnExistingFile())

			})
		})

		Context("with -keep-separate-reports and -output-dir", func() {
			BeforeEach(func() {
				session := startGinkgo(fm.PathTo("reporting"), "--no-color", "-r", "--keep-going", "--nodes=2", "--json-report=out.json", "--junit-report=out.xml", "--keep-separate-reports", "--output-dir=./reports", "-seed=17")
				Eventually(session).Should(gexec.Exit(1))
				Ω(session).ShouldNot(gbytes.Say("Could not open"))
			})

			It("places the separate reports in the -output-dir", func() {
				reports := loadJSONReports("reporting", "reports/reporting_out.json")
				Ω(reports).Should(HaveLen(1))
				checkJSONReport(reports[0])
				checkJUnitReport(loadJUnitReport("reporting", "reports/reporting_out.xml").TestSuites[0])

				reports = loadJSONReports("reporting", "reports/reporting_sub_package_out.json")
				Ω(reports).Should(HaveLen(1))
				checkJSONSubpackageReport(reports[0])
				checkJUnitSubpackageReport(loadJUnitReport("reporting", "reports/reporting_sub_package_out.xml").TestSuites[0])

				reports = loadJSONReports("reporting", "reports/malformed_sub_package_out.json")
				Ω(reports).Should(HaveLen(1))
				checkJSONFailedCompilationReport(reports[0])
				checkJUnitFailedCompilationReport(loadJUnitReport("reporting", "reports/malformed_sub_package_out.xml").TestSuites[0])

				Ω(fm.PathTo("reporting", "reports/nonginkgo_sub_package_out.json")).ShouldNot(BeAnExistingFile())
			})
		})
	})
})
