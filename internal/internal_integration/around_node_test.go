package internal_integration_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/ginkgo/v2/internal/test_helpers"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
)

func AN(run string) types.AroundNodeDecorator {
	return AroundNode(func() {
		rt.Run(run)
	})
}

var _ = Describe("The AroundNode decorator", func() {
	Context("when applied to individual nodes", func() {
		It("is scoped to run just around that node", func() {
			success, hPF := RunFixture("around node test", func() {
				BeforeSuite(rt.T("before-suite"), AN("before-suite-around"))
				Describe("container", func() {
					BeforeEach(rt.T("before-each"), AN("before-each-around"))
					It("runs", rt.T("it"), AN("it-around"), AroundNode(func(ctx context.Context, body func(ctx context.Context)) {
						rt.Run("it-around-2-before")
						body(ctx)
						rt.Run("it-around-2-after")
					}))
				})
			})

			Ω(success).Should(BeTrue())
			Ω(rt).Should(HaveTracked(
				"before-suite-around", "before-suite",
				"before-each-around", "before-each",
				"it-around", "it-around-2-before", "it", "it-around-2-after",
			))
			Ω(hPF).Should(BeFalse())
		})
	})

	Context("when DeferCleanup is called", func() {
		It("attaches the caller's AroundNode", func() {
			success, _ := RunFixture("around node test with defer cleanup", func() {
				It("A", AN("A-around"), func() {
					rt.Run("A")
					DeferCleanup(AN("DC-around"), func() {
						rt.Run("DC")
					})
				})
			})
			Ω(success).Should(BeTrue())
			Ω(rt).Should(HaveTracked(
				"A-around", "A",
				"A-around", "DC-around", "DC",
			))
		})
	})

	Context("when included in a hierarchy", func() {
		It("correctly tracks nodes in the hierarchy", func() {
			success, _ := RunFixture("around node test with hierarchy", func() {
				Describe("outer", AN("outer-around"), AN("outer-around-2"), func() {
					It("A", AN("A-around"), rt.T("A"))
					Describe("inner", AN("inner-around"), func() {
						It("B", AN("B-around"), rt.T("B", func() {
							DeferCleanup(AN("DC-around"), rt.T("DC"))
						}))
					})
					BeforeEach(AN("before-each-around"), rt.T("before-each"))
				})
				AfterEach(AN("after-each-around"), rt.T("after-each"))
			})

			Ω(success).Should(BeTrue())
			Ω(rt).Should(HaveTracked(
				"outer-around", "outer-around-2", "before-each-around", "before-each",
				"outer-around", "outer-around-2", "A-around", "A",
				"after-each-around", "after-each",
				"outer-around", "outer-around-2", "before-each-around", "before-each",
				"outer-around", "outer-around-2", "inner-around", "B-around", "B",
				"after-each-around", "after-each",
				"outer-around", "outer-around-2", "inner-around", "B-around", "DC-around", "DC",
			))
		})
	})

	Context("when applied to RunSpecs", func() {
		It("runs for all nodes, including the suite-level nodes", func() {
			success, hPF := RunFixture("around node test with suite-level around", func() {
				BeforeSuite(rt.T("before-suite"), AN("before-suite-around"))
				Describe("container", AN("container-around"), func() {
					BeforeEach(rt.T("before-each"), AN("before-each-around"))
					It("runs", rt.T("it"), AN("it-around"), AroundNode(func(ctx context.Context, body func(ctx context.Context)) {
						rt.Run("it-around-2-before")
						body(ctx)
						rt.Run("it-around-2-after")
					}))
				})
			}, AN("suite-around-1"), AN("suite-around-2"))

			Ω(success).Should(BeTrue())
			Ω(rt).Should(HaveTracked(
				"suite-around-1", "suite-around-2", "before-suite-around", "before-suite",
				"suite-around-1", "suite-around-2", "container-around", "before-each-around", "before-each",
				"suite-around-1", "suite-around-2", "container-around", "it-around", "it-around-2-before", "it", "it-around-2-after",
			))
			Ω(hPF).Should(BeFalse())
		})
	})

	Context("when it modifies the context", func() {
		var newCtx context.Context
		It("provides the node with a wrapped version of the context that can, nonetheless, be accessed and unwrapped", AroundNode(func(ctx context.Context, body func(ctx context.Context)) {
			newCtx = context.WithValue(ctx, "wrapped", "value")
			body(newCtx)
		}), func(ctx SpecContext) {
			Ω(ctx.Value("wrapped")).Should(Equal("value"))
			Ω(ctx).ShouldNot(Equal(newCtx))
			Ω(ctx.WrappedContext()).Should(Equal(newCtx))
		})

		It("works with the more complex suite setup nodes too", func() {
			succes, _ := RunFixture("around node test with complex context", func() {
				SynchronizedBeforeSuite(func(ctx SpecContext) []byte {
					rt.Run("SBS-primary")
					Ω(ctx.Value("wrapped")).Should(Equal("value"))
					return []byte("data")
				}, func(ctx SpecContext, data []byte) {
					rt.Run("SBS-all")
					Ω(ctx.Value("wrapped")).Should(Equal("value"))
					Ω(data).Should(Equal([]byte("data")))
				}, AroundNode(func(ctx context.Context) context.Context {
					rt.Run("SBS-around")
					return context.WithValue(ctx, "wrapped", "value")
				}))
				It("runs", rt.T("A"))
			}, AN("suite-around-1"))
			Ω(succes).Should(BeTrue())
			Ω(rt).Should(HaveTracked(
				"suite-around-1", "SBS-around", "SBS-primary",
				"suite-around-1", "SBS-around", "SBS-all",
				"suite-around-1", "A",
			))
			Ω(reporter.Did).Should(HaveEach(HavePassed()))
		})

		It("should still allow timeouts", func(ctx SpecContext) {
			c := make(chan struct{})
			success, _ := RunFixture("around node test with timeout", func() {
				It("A",
					AroundNode(func(ctx context.Context, body func(ctx context.Context)) {
						ctx, cancel := context.WithTimeout(ctx, 100*time.Minute)
						defer cancel()
						ctx = context.WithValue(ctx, "timeout", "value")
						body(ctx)
					}),
					func(ctx SpecContext) {
						Ω(ctx.Value("timeout")).Should(Equal("value"))
						<-ctx.Done()
						close(c)
					},
					NodeTimeout(100*time.Millisecond))
			})

			Ω(success).Should(BeFalse())
			Ω(reporter.Did.Find("A")).Should(HaveTimedOut())
			Eventually(c, ctx).Should(BeClosed())
		}, NodeTimeout(time.Second))

	})

	Context("when the user fails to call the body function", func() {
		It("fails", func() {
			success, _ := RunFixture("around node test that fails to call the body function", func() {
				It("A", AroundNode(func(ctx context.Context, body func(ctx context.Context)) {}), func() {
					rt.Run("A")
				})
			})
			Ω(success).Should(BeFalse())
			Ω(reporter.Did.Find("A")).Should(HaveFailed("An AroundNode failed to call the passed in function."))
			Ω(rt).Should(HaveTrackedNothing())
		})
	})

	Context("when the user passes in a nil context", func() {
		It("fails", func() {
			success, _ := RunFixture("around node test that passes in a nil context", func() {
				It("A", AroundNode(func(ctx context.Context, body func(ctx context.Context)) {
					body(nil)
				}), func() {
					rt.Run("A")
				})
			})
			Ω(success).Should(BeFalse())
			Ω(reporter.Did.Find("A")).Should(HaveFailed("An AroundNode failed to pass a valid Ginkgo SpecContext in.  You must always pass in a context derived from the context passed to you."))
			Ω(rt).Should(HaveTrackedNothing())
		})
	})

	Context("when the user passes in a context that does not inherit from the chain", func() {
		It("fails", func() {
			success, _ := RunFixture("around node test that passes in a context that does not inherit from the chain", func() {
				It("A", AroundNode(func(ctx context.Context, body func(ctx context.Context)) {
					body(context.Background())
				}), func() {
					rt.Run("A")
				})
			})
			Ω(success).Should(BeFalse())
			Ω(reporter.Did.Find("A")).Should(HaveFailed("An AroundNode failed to pass a valid Ginkgo SpecContext in.  You must always pass in a context derived from the context passed to you."))
			Ω(rt).Should(HaveTrackedNothing())
		})
	})
})
