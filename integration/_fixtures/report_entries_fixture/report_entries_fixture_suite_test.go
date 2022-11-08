package report_entries_fixture_test

import (
	"fmt"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestReportEntriesFixture(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ReportEntriesFixture Suite")
}

type StringerStruct struct {
	Label string
	Count int
}

func (s StringerStruct) ColorableString() string {
	return fmt.Sprintf("{{red}}%s {{green}}%d{{/}}", s.Label, s.Count)
}

var _ = Describe("top-level container", func() {
	var s *StringerStruct
	BeforeEach(func() {
		s = &StringerStruct{
			Label: "placeholder",
			Count: 10,
		}
	})

	It("passes", func() {
		AddReportEntry("passes-first-report", StringerStruct{"pass-bob", 1})
		AddReportEntry("passes-second-report")
		AddReportEntry("passes-third-report", 3)
		AddReportEntry("passes-pointer-report", s)
		AddReportEntry("passes-failure-report", 5, ReportEntryVisibilityFailureOrVerbose)
		AddReportEntry("passes-never-see-report", 6, ReportEntryVisibilityNever)
	})

	It("fails", func() {
		AddReportEntry("fails-first-report", StringerStruct{"fail-bob", 1})
		AddReportEntry("fails-second-report")
		AddReportEntry("fails-third-report", 3)
		AddReportEntry("fails-pointer-report", s)
		AddReportEntry("fails-failure-report", 5, ReportEntryVisibilityFailureOrVerbose)
		AddReportEntry("fails-never-see-report", 6, ReportEntryVisibilityNever)
		Fail("boom")
	})

	It("has By entries", func() {
		By("registers a By event")
		By("includes durations", func() {
			time.Sleep(time.Millisecond * 100)
		})
	})

	AfterEach(func() {
		s.Label = CurrentSpecReport().State.String()
		s.Count = 4
	})
})
