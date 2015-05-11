package external_coverage_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/integration/_fixtures/coverage_fixture/external_coverage_fixture"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing suite command", func() {
	It("it should succeed", func() {
		Ω(suite_command.Tested()).Should(Equal("tested"))
	})
})
