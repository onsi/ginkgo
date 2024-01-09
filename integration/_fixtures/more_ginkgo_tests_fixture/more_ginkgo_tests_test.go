package more_ginkgo_tests_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/ginkgo/v2/integration/_fixtures/more_ginkgo_tests_fixture"
	. "github.com/onsi/gomega"
)

var _ = Describe("MoreGinkgoTests", func() {
	It("should pass", func() {
		Ω(AlwaysTrue()).Should(BeTrue())
	})

	It("should always pass", func() {
		Ω(AlwaysTrue()).Should(BeTrue())
	})

	It("should match testing.TB", func() {
		var tbFunc = func(_ testing.TB) {
			Ω(AlwaysTrue()).Should(BeTrue())
		}

		tbFunc(GinkgoTB())
	})
})
