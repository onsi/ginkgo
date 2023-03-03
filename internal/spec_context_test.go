package internal_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/internal"
	. "github.com/onsi/gomega"
)

var _ = Describe("SpecContext", func() {
	It("allows access to the current spec report", func(c SpecContext) {
		Ω(c.SpecReport().LeafNodeText).Should(Equal("allows access to the current spec report"))
	})

	It("can be wrapped and still retrieved", func(c SpecContext) {
		Ω(c.Value("GINKGO_SPEC_CONTEXT")).Should(Equal(c))

		wrappedC := context.WithValue(c, "foo", "bar")

		_, ok := wrappedC.(SpecContext)
		Ω(ok).Should(BeFalse())
		Ω(wrappedC.Value("GINKGO_SPEC_CONTEXT").(SpecContext).SpecReport().LeafNodeText).Should(Equal("can be wrapped and still retrieved"))
	})

	It("can attach and detach progress reporters", func(c SpecContext) {
		type CompleteSpecContext interface {
			AttachProgressReporter(func() string) func()
			QueryProgressReporters(ctx context.Context, failer *internal.Failer) []string
		}

		wrappedC := context.WithValue(c, "foo", "bar")
		ctx := wrappedC.Value("GINKGO_SPEC_CONTEXT").(CompleteSpecContext)

		Ω(ctx.QueryProgressReporters(context.Background(), nil)).Should(BeEmpty())

		cancelA := ctx.AttachProgressReporter(func() string { return "A" })
		Ω(ctx.QueryProgressReporters(context.Background(), nil)).Should(Equal([]string{"A"}))
		cancelB := ctx.AttachProgressReporter(func() string { return "B" })
		Ω(ctx.QueryProgressReporters(context.Background(), nil)).Should(Equal([]string{"A", "B"}))
		cancelC := ctx.AttachProgressReporter(func() string { return "C" })
		Ω(ctx.QueryProgressReporters(context.Background(), nil)).Should(Equal([]string{"A", "B", "C"}))
		cancelB()
		Ω(ctx.QueryProgressReporters(context.Background(), nil)).Should(Equal([]string{"A", "C"}))
		cancelA()
		Ω(ctx.QueryProgressReporters(context.Background(), nil)).Should(Equal([]string{"C"}))
		cancelC()
		Ω(ctx.QueryProgressReporters(context.Background(), nil)).Should(BeEmpty())
	})
})
