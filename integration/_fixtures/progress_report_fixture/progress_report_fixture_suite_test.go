package progress_report_fixture_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestProgressReportFixture(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ProgressReportFixture Suite")
}
