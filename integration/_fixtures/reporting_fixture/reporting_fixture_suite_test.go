package reporting_fixture_test

import (
	"fmt"
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
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
