package failures_file_fixture_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestFailuresFileFixture(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "FailuresFileFixture Suite")
}
