package internal_integration_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/ginkgo/v2/internal/test_helpers"
	. "github.com/onsi/gomega"
)

var _ = Describe("when config.DryRun is enabled", func() {
	BeforeEach(func() {
		conf.DryRun = true
		conf.SkipStrings = []string{"E"}

		RunFixture("dry run", func() {
			ReportBeforeSuite(func(report Report) { rt.RunWithData("report-before-suite", "report", report) })
			BeforeSuite(rt.T("before-suite"))
			BeforeEach(rt.T("bef"))
			ReportBeforeEach(func(_ SpecReport) { rt.Run("report-before-each") })
			Describe("container", func() {
				It("A", rt.T("A"))
				It("B", rt.T("B", func() { F() }))
				PIt("C", rt.T("C", func() { F() }))
				It("D", rt.T("D"))
				It("E", rt.T("E"))
			})
			AfterEach(rt.T("aft"))
			AfterSuite(rt.T("after-suite"))
			ReportAfterEach(func(_ SpecReport) { rt.Run("report-after-each") })
			ReportAfterSuite("", func(_ Report) { rt.Run("report-after-suite") })
		})
	})

	It("does not run any tests but does invoke reporters", func() {
		Ω(rt).Should(HaveTracked(
			"report-before-suite",                     //BeforeSuite
			"report-before-each", "report-after-each", //A
			"report-before-each", "report-after-each", //B
			"report-before-each", "report-after-each", //C
			"report-before-each", "report-after-each", //D
			"report-before-each", "report-after-each", //E
			"report-after-suite", //AfterSuite
		))
	})

	It("correctly calculates the number of specs that will run", func() {
		report := rt.DataFor("report-before-suite")["report"].(Report)
		Ω(report.PreRunStats.SpecsThatWillRun).Should(Equal(3))
		Ω(report.PreRunStats.TotalSpecs).Should(Equal(5))
	})

	It("reports on the tests (both that they will run and that they did run) and honors skip state", func() {
		Ω(reporter.Will.Names()).Should(Equal([]string{"A", "B", "C", "D", "E"}))
		Ω(reporter.Will.Find("C")).Should(BePending())
		Ω(reporter.Will.Find("E")).Should(HaveBeenSkipped())

		Ω(reporter.Did.Names()).Should(Equal([]string{"A", "B", "C", "D", "E"}))
		Ω(reporter.Did.Find("A")).Should(HavePassed())
		Ω(reporter.Did.Find("B")).Should(HavePassed())
		Ω(reporter.Did.Find("C")).Should(BePending())
		Ω(reporter.Did.Find("D")).Should(HavePassed())
		Ω(reporter.Did.Find("E")).Should(HaveBeenSkipped())
	})

	It("reports the correct statistics", func() {
		Ω(reporter.End).Should(BeASuiteSummary(NSpecs(5), NPassed(3), NPending(1), NSkipped(1)))
	})
})
