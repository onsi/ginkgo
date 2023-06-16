package internal_integration_test

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
)

var _ = Describe("ReportEntries", func() {
	Context("happy path", func() {
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

		It("also emits report", func() {
			Ω(reporter.ReportEntries).Should(HaveLen(3))
			Ω(reporter.ReportEntries[0].Name).Should(Equal("bridge"))
			Ω(reporter.ReportEntries[1].Name).Should(Equal("medical"))
			Ω(reporter.ReportEntries[2].Name).Should(Equal("engineering"))
		})
	})

	Context("avoiding races", func() {
		BeforeEach(func() {
			success, _ := RunFixture("Report Entries - but no races", func() {
				BeforeEach(func() {
					stop := make(chan any)
					done := make(chan any)
					ticker := time.NewTicker(10 * time.Millisecond)
					i := 0
					go func() {
						for {
							select {
							case <-ticker.C:
								AddReportEntry(fmt.Sprintf("report-%d", i))
								i++
							case <-stop:
								ticker.Stop()
								close(done)
								return
							}
						}
					}()
					DeferCleanup(func() {
						close(stop)
						<-done
					})
				})

				It("reporter", func() {
					for i := 0; i < 5; i++ {
						time.Sleep(20 * time.Millisecond)
						AddReportEntry(fmt.Sprintf("waiting... %d", i))
						Ω(len(CurrentSpecReport().ReportEntries)).Should(BeNumerically("<", (i+1)*10))
					}
				})

				ReportAfterEach(func(report SpecReport) {
					//no races here, either
					Ω(len(report.ReportEntries)).Should(BeNumerically(">", 5))
				})

			})
			Ω(success).Should(BeTrue())
		})

		It("attaches entries without racing", func() {
			Ω(reporter.Did.Find("reporter").ReportEntries).Should(ContainElement(HaveField("Name", "report-0")))
			Ω(reporter.Did.Find("reporter").ReportEntries).Should(ContainElement(HaveField("Name", "report-2")))
			Ω(reporter.Did.Find("reporter").ReportEntries).Should(ContainElement(HaveField("Name", "waiting... 1")))
			Ω(reporter.Did.Find("reporter").ReportEntries).Should(ContainElement(HaveField("Name", "waiting... 3")))
		})
	})
})
