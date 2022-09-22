package internal_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("SpecContext", func() {
	It("allows access to the current spec report", func(c SpecContext) {
		Ω(c.SpecReport().LeafNodeText).Should(Equal("allows access to the current spec report"))
	})

	It("can be wrapped and still retreived", func(c SpecContext) {
		Ω(c.Value("GINKGO_SPEC_CONTEXT")).Should(Equal(c))

		wrappedC, _ := context.WithCancel(c)
		wrappedC = context.WithValue(wrappedC, "foo", "bar")

		_, ok := wrappedC.(SpecContext)
		Ω(ok).Should(BeFalse())
		Ω(wrappedC.Value("GINKGO_SPEC_CONTEXT").(SpecContext).SpecReport().LeafNodeText).Should(Equal("can be wrapped and still retreived"))
	})
})
