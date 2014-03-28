package leafnode_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/onsi/ginkgo/internal/leafnode"

	"github.com/onsi/ginkgo/internal/codelocation"
	"github.com/onsi/ginkgo/internal/types"
	"github.com/onsi/ginkgo/types"
	"runtime"
	"time"
)

type runnable interface {
	Run() (outcome types.ExampleState, failure types.ExampleFailure)
	CodeLocation() types.CodeLocation
}

func SharedRunnableNodeBehaviors(build func(body interface{}, timeout time.Duration) runnable) {
	var (
		outcome      types.ExampleState
		failure      types.ExampleFailure
		codeLocation types.CodeLocation
	)

	Describe("running the passed in function", func() {
		Context("when the function is synchronous and does not panic", func() {
			It("should run the function and report a types.ExampleStatePassed", func() {
				didRun := false
				outcome, failure := build(func() {
					didRun = true
				}, 0).Run()

				Ω(didRun).Should(BeTrue())
				Ω(outcome).Should(Equal(types.ExampleStatePassed))
				Ω(failure).Should(BeZero())
			})
		})

		Context("when the function is synchronous and *does* panic", func() {
			BeforeEach(func() {
				outcome, failure = build(func() {
					codeLocation = codelocation.New(0)
					panic("ack!")
				}, 0).Run()
			})

			It("should run the function and report a types.ExampleStatePanicked", func() {
				Ω(outcome).Should(Equal(types.ExampleStatePanicked))
				Ω(failure.Message).Should(Equal("Test Panicked"))
			})

			It("should include the code location of the panic itself", func() {
				Ω(failure.Location.FileName).Should(Equal(codeLocation.FileName))
				Ω(failure.Location.LineNumber).Should(Equal(codeLocation.LineNumber + 1))
			})

			It("should include the panic data", func() {
				Ω(failure.ForwardedPanic).Should(Equal("ack!"))
			})
		})

		Context("when the function is asynchronous", func() {
			var (
				node               runnable
				sleepDuration      time.Duration
				timeoutDuration    time.Duration
				numberOfGoRoutines int
				doPanic            bool
			)

			JustBeforeEach(func() {
				node = build(func(done Done) {
					if doPanic {
						codeLocation = codelocation.New(0)
						panic("ack!")
					}
					numberOfGoRoutines = runtime.NumGoroutine()
					time.Sleep(sleepDuration)
					close(done)
				}, timeoutDuration)
			})

			BeforeEach(func() {
				sleepDuration = time.Duration(0.001 * float64(time.Second))
				timeoutDuration = time.Duration(1 * float64(time.Second))
				doPanic = false
			})

			It("should run the function as a goroutine", func() {
				initialNumberOfGoRoutines := runtime.NumGoroutine()
				node.Run()
				Ω(numberOfGoRoutines).Should(Equal(initialNumberOfGoRoutines + 1))
			})

			Context("when the function takes less time than the timeout", func() {
				It("should pass", func() {
					outcome, failure := node.Run()

					Ω(outcome).Should(Equal(types.ExampleStatePassed))
					Ω(failure).Should(BeZero())
				})
			})

			Context("when the function takes longer than the timeout", func() {
				BeforeEach(func() {
					sleepDuration = time.Duration(0.002 * float64(time.Second))
					timeoutDuration = time.Duration(0.001 * float64(time.Second))
				})

				It("should timeout", func() {
					outcome, failure := node.Run()
					Ω(outcome).Should(Equal(types.ExampleStateTimedOut))
					Ω(failure.Message).Should(Equal("Timed out"))
					Ω(failure.Location).Should(Equal(node.CodeLocation()))
				})
			})

			Context("when the function panics", func() {
				BeforeEach(func() {
					doPanic = true
				})

				JustBeforeEach(func() {
					outcome, failure = node.Run()
				})

				It("should run the function and report a types.ExampleStatePanicked", func() {
					Ω(outcome).Should(Equal(types.ExampleStatePanicked))
					Ω(failure.Message).Should(Equal("Test Panicked"))
				})

				It("should include the code location of the panic itself", func() {
					Ω(failure.Location.FileName).Should(Equal(codeLocation.FileName))
					Ω(failure.Location.LineNumber).Should(Equal(codeLocation.LineNumber + 1))
				})

				It("should include the panic data", func() {
					Ω(failure.ForwardedPanic).Should(Equal("ack!"))
				})
			})
		})

		Context("when the function takes the wrong kind of argument", func() {
			It("should panic", func() {
				Ω(func() {
					build(func(oops string) {}, 0)
				}).Should(Panic())
			})
		})

		Context("when the function takes more than one argument", func() {
			It("should panic", func() {
				Ω(func() {
					build(func(done Done, oops string) {}, 0)
				}).Should(Panic())
			})
		})
	})
}

var _ = Describe("Shared RunnableNode behavior", func() {
	Describe("It Nodes", func() {
		SharedRunnableNodeBehaviors(func(body interface{}, timeout time.Duration) runnable {
			return NewItNode("", body, internaltypes.FlagTypeFocused, codelocation.New(0), timeout)
		})
	})

	Describe("BeforeEach Nodes", func() {
		SharedRunnableNodeBehaviors(func(body interface{}, timeout time.Duration) runnable {
			return NewBeforeEachNode(body, codelocation.New(0), timeout)
		})
	})

	Describe("AfterEach Nodes", func() {
		SharedRunnableNodeBehaviors(func(body interface{}, timeout time.Duration) runnable {
			return NewAfterEachNode(body, codelocation.New(0), timeout)
		})
	})

	Describe("JustBeforeEach Nodes", func() {
		SharedRunnableNodeBehaviors(func(body interface{}, timeout time.Duration) runnable {
			return NewJustBeforeEachNode(body, codelocation.New(0), timeout)
		})
	})
})
