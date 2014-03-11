package failing_ginkgo_tests_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/ginkgo/integration/_fixtures/run_fixtures/failing_ginkgo_tests"
	. "github.com/onsi/gomega"
)

var _ = Describe("FailingGinkgoTests", func() {
	It("should pass", func() {
		Ω(AlwaysFalse()).Should(BeTrue())
	})

	It("should fail", func() {
		Ω(AlwaysFalse()).Should(BeFalse())
	})
})
