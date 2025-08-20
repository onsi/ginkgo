package types_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/onsi/ginkgo/v2/types"
)

var _ = Describe("AroundNodes", func() {
	cl := types.NewCodeLocation(0)
	Describe("building an AroundNode", func() {
		It("captures the code location", func() {
			an := types.AroundNode(func(ctx context.Context, body func(context.Context)) {
				body(context.WithValue(ctx, "key", "value"))
			}, cl)

			Ω(an).ShouldNot(BeZero())
			Ω(an.CodeLocation).Should(Equal(cl))

			c := context.Background()
			var calledContext context.Context
			an.Body(c, func(ctx context.Context) {
				calledContext = ctx
			})
			Ω(calledContext.Value("key")).Should(Equal("value"))
		})

		It("supports bare functions", func() {
			called := false
			an := types.AroundNode(func() {
				called = true
			}, cl)
			c := context.Background()
			Ω(an).ShouldNot(BeZero())
			var calledContext context.Context
			an.Body(c, func(ctx context.Context) {
				calledContext = ctx
			})
			Ω(called).Should(BeTrue())
			Ω(calledContext).Should(Equal(c))
		})

		It("supports context transformers", func() {
			an := types.AroundNode(func(c context.Context) context.Context {
				return context.WithValue(c, "key", "value")
			}, cl)
			c := context.Background()
			Ω(an).ShouldNot(BeZero())
			var calledContext context.Context
			an.Body(c, func(ctx context.Context) {
				calledContext = ctx
			})
			Ω(calledContext.Value("key")).Should(Equal("value"))
		})
	})

	Describe("Clone", func() {
		It("clones the slice", func() {
			an1 := types.AroundNode(func() {}, types.NewCodeLocation(0))
			an2 := types.AroundNode(func() {}, types.NewCodeLocation(0))
			an3 := types.AroundNode(func() {}, types.NewCodeLocation(0))
			original := types.AroundNodes{an1, an2, an3}
			clone := original.Clone()
			an4 := types.AroundNode(func() {}, types.NewCodeLocation(0))
			clone[2] = an4
			Ω(original).Should(HaveExactElements(
				HaveField("CodeLocation", an1.CodeLocation),
				HaveField("CodeLocation", an2.CodeLocation),
				HaveField("CodeLocation", an3.CodeLocation),
			))
			Ω(clone).Should(HaveExactElements(
				HaveField("CodeLocation", an1.CodeLocation),
				HaveField("CodeLocation", an2.CodeLocation),
				HaveField("CodeLocation", an4.CodeLocation),
			))
		})
	})

	Describe("Append", func() {
		It("appends the node to the slice", func() {
			an1 := types.AroundNode(func() {}, types.NewCodeLocation(0))
			an2 := types.AroundNode(func() {}, types.NewCodeLocation(0))
			an3 := types.AroundNode(func() {}, types.NewCodeLocation(0))
			nodes := types.AroundNodes{an1, an2}
			newNodes := nodes.Append(an3)
			Ω(nodes).Should(HaveExactElements(
				HaveField("CodeLocation", an1.CodeLocation),
				HaveField("CodeLocation", an2.CodeLocation),
			))
			Ω(newNodes).Should(HaveExactElements(
				HaveField("CodeLocation", an1.CodeLocation),
				HaveField("CodeLocation", an2.CodeLocation),
				HaveField("CodeLocation", an3.CodeLocation),
			))
		})
	})
})
