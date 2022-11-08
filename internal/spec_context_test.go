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

		wrappedC := context.WithValue(c, "foo", "bar")

		_, ok := wrappedC.(SpecContext)
		Ω(ok).Should(BeFalse())
		Ω(wrappedC.Value("GINKGO_SPEC_CONTEXT").(SpecContext).SpecReport().LeafNodeText).Should(Equal("can be wrapped and still retreived"))
	})

	It("can attach and detach progress reporters", func(c SpecContext) {
		type CompleteSpecContext interface {
			AttachProgressReporter(func() string) func()
			QueryProgressReporters() []string
		}

		wrappedC := context.WithValue(c, "foo", "bar")
		ctx := wrappedC.Value("GINKGO_SPEC_CONTEXT").(CompleteSpecContext)

		Ω(ctx.QueryProgressReporters()).Should(BeEmpty())

		cancelA := ctx.AttachProgressReporter(func() string { return "A" })
		Ω(ctx.QueryProgressReporters()).Should(Equal([]string{"A"}))
		cancelB := ctx.AttachProgressReporter(func() string { return "B" })
		Ω(ctx.QueryProgressReporters()).Should(Equal([]string{"A", "B"}))
		cancelC := ctx.AttachProgressReporter(func() string { return "C" })
		Ω(ctx.QueryProgressReporters()).Should(Equal([]string{"A", "B", "C"}))
		cancelB()
		Ω(ctx.QueryProgressReporters()).Should(Equal([]string{"A", "C"}))
		cancelA()
		Ω(ctx.QueryProgressReporters()).Should(Equal([]string{"C"}))
		cancelC()
		Ω(ctx.QueryProgressReporters()).Should(BeEmpty())
	})
})
