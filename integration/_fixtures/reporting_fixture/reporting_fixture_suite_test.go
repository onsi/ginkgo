package reporting_fixture_test

import (
	"fmt"
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/types"
	. "github.com/onsi/gomega"
)

func TestReportingFixture(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ReportingFixture Suite")
}

var f *os.File

var _ = BeforeSuite(func() {
	var err error
	f, err = os.Create("report-after-each.out")
	Ω(err).ShouldNot(HaveOccurred())
})

var _ = AfterSuite(func() {
	f.Close()
})

var _ = ReportAfterEach(func(report SpecReport) {
	fmt.Fprintf(f, "%s - %s\n", report.SpecText(), report.State)
})

var _ = ReportAfterSuite("my report", func(report Report) {
	f, err := os.Create("report-after-suite.out")
	Ω(err).ShouldNot(HaveOccurred())

	fmt.Fprintf(f, "%s - %d\n", report.SuiteDescription, report.SuiteConfig.RandomSeed)
	for _, specReport := range report.SpecReports {
		if specReport.LeafNodeType.Is(types.NodeTypesForSuiteLevelNodes...) {
			fmt.Fprintf(f, "%d: [%s] - %s\n", specReport.GinkgoParallelNode, specReport.LeafNodeType, specReport.State)
		} else {
			fmt.Fprintf(f, "%s - %s\n", specReport.SpecText(), specReport.State)
		}
	}

	f.Close()

	Fail("fail!")
})
