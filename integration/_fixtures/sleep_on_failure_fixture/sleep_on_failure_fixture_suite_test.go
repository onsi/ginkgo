package sleep_on_failure_fixture_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSleepOnFailureFixture(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SleepOnFailure Fixture Suite")
}
