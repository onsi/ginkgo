package progress_reporter_fixture_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestProgressReporterFixture(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ProgressReporterFixture Suite")
}
