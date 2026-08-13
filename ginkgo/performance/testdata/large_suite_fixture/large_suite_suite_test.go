package large_fixture_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestLargeSuiteFixture(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Large Suite")
}

var _ = Describe("Large Suite", func() {
	for i := 0; i < 2048; i++ {
		It("is a large suite", func() {})
	}
})
