package defer_on_failure_fixture_test

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDeferOnFailureFixture(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "DeferOnFailureFixture Suite")
}

var _ = SynchronizedBeforeSuite(func() []byte {
	return nil
}, func(_ []byte) {
	// registered in the all-processes function, so it covers every spec on every process
	DeferOnFailure(func(ctx context.Context) {
		AddReportEntry("suite-dump", GinkgoParallelProcess())
	})
})

var _ = Describe("specs", func() {
	BeforeEach(func() {
		DeferOnFailure(AddReportEntry, "each-dump")
	})

	AfterEach(func() {
		AddReportEntry("after-each")
	})

	It("passes", func() {})

	It("fails", func() {
		Fail("boom")
	})
})

var _ = Describe("ordered", Ordered, func() {
	BeforeAll(func() {
		GinkgoT().DeferOnFailure(AddReportEntry, "all-dump")
	})

	It("passes first", func() {})

	It("fails second", func() {
		Fail("boom")
	})
})
