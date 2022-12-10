package reporting_fixture_test

import (
	"fmt"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
)

func TestReportingFixture(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ReportingFixture Suite")
}

var beforeEachReport, afterEachReport *os.File

var _ = ReportBeforeSuite(func(report Report) {
	f, err := os.Create(fmt.Sprintf("report-before-suite-%d.out", GinkgoParallelProcess()))
	Ω(err).ShouldNot(HaveOccurred())

	fmt.Fprintf(f, "%s - %d\n", report.SuiteDescription, report.SuiteConfig.RandomSeed)
	fmt.Fprintf(f, "%d of %d", report.PreRunStats.SpecsThatWillRun, report.PreRunStats.TotalSpecs)

	f.Close()
})

var _ = BeforeSuite(func() {
	var err error
	beforeEachReport, err = os.Create("report-before-each.out")
	Ω(err).ShouldNot(HaveOccurred())
	DeferCleanup(beforeEachReport.Close)

	afterEachReport, err = os.Create("report-after-each.out")
	Ω(err).ShouldNot(HaveOccurred())
	DeferCleanup(afterEachReport.Close)
})

var _ = ReportBeforeEach(func(report SpecReport) {
	fmt.Fprintf(beforeEachReport, "%s - %s\n", report.LeafNodeText, report.State)
})

var _ = ReportAfterEach(func(report SpecReport) {
	fmt.Fprintf(afterEachReport, "%s - %s\n", report.LeafNodeText, report.State)
})

var _ = ReportAfterSuite("my report", func(report Report) {
	f, err := os.Create("report-after-suite.out")
	Ω(err).ShouldNot(HaveOccurred())

	fmt.Fprintf(f, "%s - %d\n", report.SuiteDescription, report.SuiteConfig.RandomSeed)
	for _, specReport := range report.SpecReports {
		if specReport.LeafNodeType.Is(types.NodeTypesForSuiteLevelNodes) || specReport.LeafNodeType.Is(types.NodeTypeCleanupAfterSuite) {
			fmt.Fprintf(f, "%d: [%s] - %s\n", specReport.ParallelProcess, specReport.LeafNodeType, specReport.State)
		} else {
			fmt.Fprintf(f, "%s - %s\n", specReport.LeafNodeText, specReport.State)
		}
	}

	f.Close()

	Fail("fail!")
})
