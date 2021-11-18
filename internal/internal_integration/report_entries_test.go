package internal_integration_test

import (
	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
)

var _ = Describe("ReportEntries", func() {
	BeforeEach(func() {
		success, _ := RunFixture("Report Entries", func() {
			BeforeSuite(func() {
				AddReportEntry("bridge", "engaged")
			})

			It("adds-entries", func() {
				AddReportEntry("medical", "healthy")
				AddReportEntry("engineering", "on fire")
			})

			It("adds-no-entries", func() {})
		})
		Ω(success).Should(BeTrue())
	})

	It("attaches entries to the report", func() {
		Ω(reporter.Did.Find("adds-entries").ReportEntries[0].Name).Should(Equal("medical"))
		Ω(reporter.Did.Find("adds-entries").ReportEntries[0].Value.String()).Should(Equal("healthy"))
		Ω(reporter.Did.Find("adds-entries").ReportEntries[1].Name).Should(Equal("engineering"))
		Ω(reporter.Did.Find("adds-entries").ReportEntries[1].Value.String()).Should(Equal("on fire"))
		Ω(reporter.Did.Find("adds-no-entries").ReportEntries).Should(BeEmpty())
		Ω(reporter.Did.FindByLeafNodeType(types.NodeTypeBeforeSuite).ReportEntries[0].Name).Should(Equal("bridge"))
		Ω(reporter.Did.FindByLeafNodeType(types.NodeTypeBeforeSuite).ReportEntries[0].Value.String()).Should(Equal("engaged"))
	})
})
