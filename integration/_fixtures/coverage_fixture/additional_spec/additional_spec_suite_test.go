package additional_spec_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/onsi/ginkgo/v2/integration/_fixtures/coverage_fixture"
	. "github.com/onsi/ginkgo/v2/integration/_fixtures/coverage_fixture/external_coverage"

	"testing"
)

func TestAdditionalSpecSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "AdditionalSpec Suite")
}

var _ = Describe("CoverageFixture", func() {
	It("should test E", func() {
		Ω(E()).Should(Equal("tested by additional"))
	})

	It("should test external package", func() {
		Ω(TestedByAdditional()).Should(Equal("tested by additional"))
	})
})
