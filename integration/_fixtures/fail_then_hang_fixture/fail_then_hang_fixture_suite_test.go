package fail_then_hang_fixture_test

import (
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestFailThenHangFixture(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "FailThenHangFixture Suite")
}

var _ = Describe("Failing then hanging", func() {
	It("fails", func() {
		time.Sleep(time.Second)
		Fail("boom")
	})

	It("hangs", func() {
		time.Sleep(time.Hour)
	})

	It("hangs", func() {
		time.Sleep(time.Hour)
	})
})
