package integration_test

import (
	"os"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/ginkgo/v2/internal/test_helpers"
	"github.com/onsi/ginkgo/v2/reporters"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Reporting", func() {
	BeforeEach(func() {
		fm.MountFixture("reporting")
	})

	Describe("in-suite reporting with ReportBeforeEach, ReportAfterEach and ReportAfterSuite", func() {
		It("reports on each test via ReportBeforeEach", func() {
			session := startGinkgo(fm.PathTo("reporting"), "--no-color")
			Eventually(session).Should(gexec.Exit(1))

			report, err := os.ReadFile(fm.PathTo("reporting", "report-before-each.out"))
			Ω(err).ShouldNot(HaveOccurred())
			lines := strings.Split(string(report), "\n")
			Ω(lines).Should(ConsistOf(
				"passes - INVALID SPEC STATE",
				"is labelled - INVALID SPEC STATE",
				"fails - INVALID SPEC STATE",
				"panics - INVALID SPEC STATE",
				"is pending - pending",
				"is skipped - INVALID SPEC STATE",
				"",
			))
		})

		It("reports on each test via ReportAfterEach", func() {
			session := startGinkgo(fm.PathTo("reporting"), "--no-color")
			Eventually(session).Should(gexec.Exit(1))

			report, err := os.ReadFile(fm.PathTo("reporting", "report-after-each.out"))
			Ω(err).ShouldNot(HaveOccurred())
			lines := strings.Split(string(report), "\n")
			Ω(lines).Should(ConsistOf(
				"passes - passed",
				"is labelled - passed",
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

			report, err := os.ReadFile(fm.PathTo("reporting", "report-after-suite.out"))
			Ω(err).ShouldNot(HaveOccurred())
			lines := strings.Split(string(report), "\n")
			Ω(lines).Should(ConsistOf(
				"ReportingFixture Suite - 17",
				"1: [BeforeSuite] - passed",
				"passes - passed",
				"is labelled - passed",
				"fails - failed",
				"panics - panicked",
				"is pending - pending",
				"is skipped - skipped",
				"1: [DeferCleanup (Suite)] - passed",
				"1: [DeferCleanup (Suite)] - passed",
				"",
			))
		})

		Context("when running in parallel", func() {
			It("reports on all the tests via ReportAfterSuite", func() {
				session := startGinkgo(fm.PathTo("reporting"), "--no-color", "--seed=17", "--procs=2")
				Eventually(session).Should(gexec.Exit(1))

				report, err := os.ReadFile(fm.PathTo("reporting", "report-after-suite.out"))
				Ω(err).ShouldNot(HaveOccurred())
				lines := strings.Split(string(report), "\n")
				Ω(lines).Should(ConsistOf(
					"ReportingFixture Suite - 17",
					"1: [BeforeSuite] - passed",
					"2: [BeforeSuite] - passed",
					"passes - passed",
					"is labelled - passed",
					"fails - failed",
					"panics - panicked",
					"is pending - pending",
					"is skipped - skipped",
					"1: [DeferCleanup (Suite)] - passed",
					"1: [DeferCleanup (Suite)] - passed",
					"2: [DeferCleanup (Suite)] - passed",
					"2: [DeferCleanup (Suite)] - passed",
					"",
				))
			})
		})

		Context("when a ReportAfterSuite node fails", func() {
			It("reports on it", func() {
				session := startGinkgo(fm.PathTo("reporting"), "--no-color", "--seed=17", "--procs=2")
				Eventually(session).Should(gexec.Exit(1))

				Ω(string(session.Out.Contents())).Should(ContainSubstring("[ReportAfterSuite] my report"))
				Ω(string(session.Out.Contents())).Should(ContainSubstring("fail!\n  In [ReportAfterSuite] at:"))
			})
		})
	})

	Describe("JSON and JUnit reporting", func() {
		checkJSONReport := func(report types.Report) {
			Ω(report.SuitePath).Should(Equal(fm.AbsPathTo("reporting")))
			Ω(report.SuiteDescription).Should(Equal("ReportingFixture Suite"))
			Ω(report.SuiteConfig.ParallelTotal).Should(Equal(2))
			Ω(report.SpecReports).Should(HaveLen(13)) //6 tests + (1 before-suite + 2 defercleanup after-suite)*2(nodes) + 1 report-after-suite

			specReports := Reports(report.SpecReports)
			Ω(specReports.WithLeafNodeType(types.NodeTypeIt)).Should(HaveLen(6))
			Ω(specReports.Find("passes")).Should(HavePassed())
			Ω(specReports.Find("is labelled")).Should(HavePassed())
			Ω(specReports.Find("is labelled").Labels()).Should(Equal([]string{"dog", "cat"}))
			Ω(specReports.Find("fails")).Should(HaveFailed("fail!", types.FailureNodeIsLeafNode, CapturedGinkgoWriterOutput("some ginkgo-writer output")))
			Ω(specReports.Find("panics")).Should(HavePanicked("boom"))
			Ω(specReports.Find("is pending")).Should(BePending())
			Ω(specReports.Find("is skipped").State).Should(Equal(types.SpecStateSkipped))
			Ω(specReports.Find("my report")).Should(HaveFailed("fail!", types.FailureNodeIsLeafNode, types.NodeTypeReportAfterSuite))
			Ω(specReports.FindByLeafNodeType(types.NodeTypeBeforeSuite)).Should(HavePassed())
			Ω(specReports.FindByLeafNodeType(types.NodeTypeCleanupAfterSuite)).Should(HavePassed())
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
			Ω(report.SpecialSuiteFailureReasons).Should(ContainElement(ContainSubstring("Failed to compile malformed_sub_package:")))
			Ω(report.SpecReports).Should(HaveLen(0))
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
			Ω(suite.Tests).Should(Equal(13))
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

			Ω(getTestCase("[It] reporting test labelled tests is labelled [dog, cat]", suite.TestCases).Status).Should(Equal("passed"))

			Ω(getTestCase("[It] reporting test panics", suite.TestCases).Status).Should(Equal("panicked"))
			Ω(getTestCase("[It] reporting test panics", suite.TestCases).Error.Message).Should(Equal("boom"))
			Ω(getTestCase("[It] reporting test panics", suite.TestCases).Error.Type).Should(Equal("panicked"))

			Ω(getTestCase("[It] reporting test fails", suite.TestCases).Failure.Message).Should(Equal("fail!"))
			Ω(getTestCase("[It] reporting test fails", suite.TestCases).Status).Should(Equal("failed"))
			Ω(getTestCase("[It] reporting test fails", suite.TestCases).SystemErr).Should(Equal("some ginkgo-writer output"))

			Ω(getTestCase("[It] reporting test is pending", suite.TestCases).Status).Should(Equal("pending"))
			Ω(getTestCase("[It] reporting test is pending", suite.TestCases).Skipped.Message).Should(Equal("pending"))

			Ω(getTestCase("[DeferCleanup (Suite)]", suite.TestCases).Status).Should(Equal("passed"))
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
			Ω(report.Tests).Should(Equal(16))
			Ω(report.Disabled).Should(Equal(2))
			Ω(report.Errors).Should(Equal(2))
			Ω(report.Failures).Should(Equal(3))

			checkJUnitReport(report.TestSuites[0])
			checkJUnitFailedCompilationReport(report.TestSuites[1])
			checkJUnitSubpackageReport(report.TestSuites[2])
		}

		checkTeamcityReport := func(data string) {
			lines := strings.Split(data, "\n")
			Ω(lines).Should(ContainElement("##teamcity[testSuiteStarted name='ReportingFixture Suite']"))

			Ω(lines).Should(ContainElement("##teamcity[testStarted name='|[BeforeSuite|]']"))
			Ω(lines).Should(ContainElement(HavePrefix("##teamcity[testFinished name='|[BeforeSuite|]'")))

			Ω(lines).Should(ContainElement("##teamcity[testStarted name='|[It|] reporting test passes']"))
			Ω(lines).Should(ContainElement(HavePrefix("##teamcity[testFinished name='|[It|] reporting test passes'")))

			Ω(lines).Should(ContainElement("##teamcity[testStarted name='|[It|] reporting test panics']"))
			Ω(lines).Should(ContainElement(HavePrefix("##teamcity[testFailed name='|[It|] reporting test panics' message='panicked - boom'")))
			Ω(lines).Should(ContainElement(HavePrefix("##teamcity[testFinished name='|[It|] reporting test panics'")))

			Ω(lines).Should(ContainElement("##teamcity[testStarted name='|[It|] reporting test fails']"))
			Ω(lines).Should(ContainElement(HavePrefix("##teamcity[testFailed name='|[It|] reporting test fails' message='failed - fail!'")))
			Ω(lines).Should(ContainElement(HavePrefix("##teamcity[testStdErr name='|[It|] reporting test fails' out='some ginkgo-writer output")))
			Ω(lines).Should(ContainElement(HavePrefix("##teamcity[testFinished name='|[It|] reporting test fails'")))

			Ω(lines).Should(ContainElement("##teamcity[testStarted name='|[It|] reporting test is pending']"))
			Ω(lines).Should(ContainElement("##teamcity[testIgnored name='|[It|] reporting test is pending' message='pending']"))
			Ω(lines).Should(ContainElement(HavePrefix("##teamcity[testFinished name='|[It|] reporting test is pending'")))

			Ω(lines).Should(ContainElement("##teamcity[testStarted name='|[It|] reporting test is skipped']"))
			Ω(lines).Should(ContainElement("##teamcity[testIgnored name='|[It|] reporting test is skipped' message='skipped - skip']"))
			Ω(lines).Should(ContainElement(HavePrefix("##teamcity[testFinished name='|[It|] reporting test is skipped'")))

			Ω(lines).Should(ContainElement("##teamcity[testSuiteFinished name='ReportingFixture Suite']"))
		}

		checkTeamcitySubpackageReport := func(data string) {
			lines := strings.Split(data, "\n")
			Ω(lines).Should(ContainElement("##teamcity[testSuiteStarted name='Reporting SubPackage Suite']"))
			Ω(lines).Should(ContainElement("##teamcity[testSuiteFinished name='Reporting SubPackage Suite']"))
		}

		checkTeamcityFailedCompilationReport := func(data string) {
			lines := strings.Split(data, "\n")
			Ω(lines).Should(ContainElement("##teamcity[testSuiteStarted name='']"))
			Ω(lines).Should(ContainElement("##teamcity[testSuiteFinished name='']"))
		}

		Context("the default behavior", func() {
			BeforeEach(func() {
				session := startGinkgo(fm.PathTo("reporting"), "--no-color", "-r", "--keep-going", "--procs=2", "--json-report=out.json", "--junit-report=out.xml", "--teamcity-report=out.tc", "-seed=17")
				Eventually(session).Should(gexec.Exit(1))
				Ω(session).ShouldNot(gbytes.Say("Could not open"))
			})

			It("generates single unified json and junit reports", func() {
				reports := fm.LoadJSONReports("reporting", "out.json")
				Ω(reports).Should(HaveLen(3))
				checkJSONReport(reports[0])
				checkJSONFailedCompilationReport(reports[1])
				checkJSONSubpackageReport(reports[2])

				junitReport := fm.LoadJUnitReport("reporting", "out.xml")
				checkUnifiedJUnitReport(junitReport)

				checkTeamcityReport(fm.ContentOf("reporting", "out.tc"))
				checkTeamcitySubpackageReport(fm.ContentOf("reporting", "out.tc"))
				checkTeamcityFailedCompilationReport(fm.ContentOf("reporting", "out.tc"))
			})
		})

		Context("with -output-dir", func() {
			BeforeEach(func() {
				session := startGinkgo(fm.PathTo("reporting"), "--no-color", "-r", "--keep-going", "--procs=2", "--json-report=out.json", "--junit-report=out.xml", "--teamcity-report=out.tc", "--output-dir=./reports", "-seed=17")
				Eventually(session).Should(gexec.Exit(1))
				Ω(session).ShouldNot(gbytes.Say("Could not open"))
			})

			It("places the single unified json and junit reports in output-dir", func() {
				reports := fm.LoadJSONReports("reporting", "reports/out.json")
				Ω(reports).Should(HaveLen(3))
				checkJSONReport(reports[0])
				checkJSONFailedCompilationReport(reports[1])
				checkJSONSubpackageReport(reports[2])

				junitReport := fm.LoadJUnitReport("reporting", "reports/out.xml")
				checkUnifiedJUnitReport(junitReport)

				checkTeamcityReport(fm.ContentOf("reporting", "reports/out.tc"))
				checkTeamcitySubpackageReport(fm.ContentOf("reporting", "reports/out.tc"))
				checkTeamcityFailedCompilationReport(fm.ContentOf("reporting", "reports/out.tc"))
			})
		})

		Context("with -keep-separate-reports", func() {
			BeforeEach(func() {
				session := startGinkgo(fm.PathTo("reporting"), "--no-color", "-r", "--keep-going", "--procs=2", "--json-report=out.json", "--junit-report=out.xml", "--teamcity-report=out.tc", "--keep-separate-reports", "-seed=17")
				Eventually(session).Should(gexec.Exit(1))
				Ω(session).ShouldNot(gbytes.Say("Could not open"))
			})

			It("keeps the separate reports in their respective packages", func() {
				reports := fm.LoadJSONReports("reporting", "out.json")
				Ω(reports).Should(HaveLen(1))
				checkJSONReport(reports[0])
				checkJUnitReport(fm.LoadJUnitReport("reporting", "out.xml").TestSuites[0])
				checkTeamcityReport(fm.ContentOf("reporting", "out.tc"))

				reports = fm.LoadJSONReports("reporting", "reporting_sub_package/out.json")
				Ω(reports).Should(HaveLen(1))
				checkJSONSubpackageReport(reports[0])
				checkJUnitSubpackageReport(fm.LoadJUnitReport("reporting", "reporting_sub_package/out.xml").TestSuites[0])
				checkTeamcitySubpackageReport(fm.ContentOf("reporting", "reporting_sub_package/out.tc"))

				reports = fm.LoadJSONReports("reporting", "malformed_sub_package/out.json")
				Ω(reports).Should(HaveLen(1))
				checkJSONFailedCompilationReport(reports[0])
				checkJUnitFailedCompilationReport(fm.LoadJUnitReport("reporting", "malformed_sub_package/out.xml").TestSuites[0])
				checkTeamcityFailedCompilationReport(fm.ContentOf("reporting", "malformed_sub_package/out.tc"))

				Ω(fm.PathTo("reporting", "nonginkgo_sub_package/out.json")).ShouldNot(BeAnExistingFile())
				Ω(fm.PathTo("reporting", "nonginkgo_sub_package/out.xml")).ShouldNot(BeAnExistingFile())
				Ω(fm.PathTo("reporting", "nonginkgo_sub_package/out.tc")).ShouldNot(BeAnExistingFile())
			})
		})

		Context("with -keep-separate-reports and -output-dir", func() {
			BeforeEach(func() {
				session := startGinkgo(fm.PathTo("reporting"), "--no-color", "-r", "--keep-going", "--procs=2", "--json-report=out.json", "--junit-report=out.xml", "--teamcity-report=out.tc", "--keep-separate-reports", "--output-dir=./reports", "-seed=17")
				Eventually(session).Should(gexec.Exit(1))
				Ω(session).ShouldNot(gbytes.Say("Could not open"))
			})

			It("places the separate reports in the -output-dir", func() {
				reports := fm.LoadJSONReports("reporting", "reports/reporting_out.json")
				Ω(reports).Should(HaveLen(1))
				checkJSONReport(reports[0])
				checkJUnitReport(fm.LoadJUnitReport("reporting", "reports/reporting_out.xml").TestSuites[0])
				checkTeamcityReport(fm.ContentOf("reporting", "reports/reporting_out.tc"))

				reports = fm.LoadJSONReports("reporting", "reports/reporting_sub_package_out.json")
				Ω(reports).Should(HaveLen(1))
				checkJSONSubpackageReport(reports[0])
				checkJUnitSubpackageReport(fm.LoadJUnitReport("reporting", "reports/reporting_sub_package_out.xml").TestSuites[0])
				checkTeamcitySubpackageReport(fm.ContentOf("reporting", "reports/reporting_sub_package_out.tc"))

				reports = fm.LoadJSONReports("reporting", "reports/malformed_sub_package_out.json")
				Ω(reports).Should(HaveLen(1))
				checkJSONFailedCompilationReport(reports[0])
				checkJUnitFailedCompilationReport(fm.LoadJUnitReport("reporting", "reports/malformed_sub_package_out.xml").TestSuites[0])
				checkTeamcityFailedCompilationReport(fm.ContentOf("reporting", "reports/malformed_sub_package_out.tc"))

				Ω(fm.PathTo("reporting", "reports/nonginkgo_sub_package_out.json")).ShouldNot(BeAnExistingFile())
			})
		})

		Context("when keep-going is not set and a suite fails", func() {
			BeforeEach(func() {
				session := startGinkgo(fm.PathTo("reporting"), "--no-color", "-r", "--procs=2", "--json-report=out.json", "--junit-report=out.xml", "--teamcity-report=out.tc", "-coverprofile=cover.out", "-cpuprofile=cpu.out", "-seed=17", "--output-dir=./reports")
				Eventually(session).Should(gexec.Exit(1))
				Ω(session).ShouldNot(gbytes.Say("Could not open"))
			})

			It("reports about the suites that did not run", func() {
				reports := fm.LoadJSONReports("reporting", "reports/out.json")
				Ω(reports).Should(HaveLen(3))
				checkJSONReport(reports[0])
				Ω(reports[1].SpecialSuiteFailureReasons).Should(ContainElement(ContainSubstring("Failed to compile malformed_sub_package")))
				Ω(reports[2].SpecialSuiteFailureReasons).Should(ContainElement("Suite did not run because prior suites failed and --keep-going is not set"))
			})
		})
	})
})
