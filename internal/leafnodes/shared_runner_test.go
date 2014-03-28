package leafnodes_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"reflect"

	. "github.com/onsi/ginkgo/internal/leafnodes"

	"github.com/onsi/ginkgo/internal/codelocation"
	Failer "github.com/onsi/ginkgo/internal/failer"
	"github.com/onsi/ginkgo/internal/types"
	"github.com/onsi/ginkgo/types"
	"runtime"
	"time"
)

type runnable interface {
	Run() (outcome types.ExampleState, failure types.ExampleFailure)
	CodeLocation() types.CodeLocation
}

func SynchronousSharedRunnerBehaviors(build func(body interface{}, timeout time.Duration, failer *Failer.Failer, componentCodeLocation types.CodeLocation, componentIndex int) runnable, componentType types.ExampleComponentType) {
	var (
		outcome types.ExampleState
		failure types.ExampleFailure

		failer *Failer.Failer

		componentIndex        int
		componentCodeLocation types.CodeLocation
		innerCodeLocation     types.CodeLocation

		didRun bool
	)

	BeforeEach(func() {
		failer = Failer.New()
		componentIndex = 3
		componentCodeLocation = codelocation.New(0)
		innerCodeLocation = codelocation.New(0)

		didRun = false
	})

	Describe("synchronous functions", func() {
		Context("when the function passes", func() {
			BeforeEach(func() {
				outcome, failure = build(func() {
					didRun = true
				}, 0, failer, componentCodeLocation, componentIndex).Run()
			})

			It("should have a succesful outcome", func() {
				Ω(didRun).Should(BeTrue())

				Ω(outcome).Should(Equal(types.ExampleStatePassed))
				Ω(failure).Should(BeZero())
			})
		})

		Context("when a failure occurs", func() {
			BeforeEach(func() {
				outcome, failure = build(func() {
					didRun = true
					failer.Fail("bam", innerCodeLocation)
					panic("should not matter")
				}, 0, failer, componentCodeLocation, componentIndex).Run()
			})

			It("should return the failure", func() {
				Ω(didRun).Should(BeTrue())

				Ω(outcome).Should(Equal(types.ExampleStateFailed))
				Ω(failure).Should(Equal(types.ExampleFailure{
					Message:               "bam",
					Location:              innerCodeLocation,
					ForwardedPanic:        nil,
					ComponentIndex:        componentIndex,
					ComponentType:         componentType,
					ComponentCodeLocation: componentCodeLocation,
				}))
			})
		})

		Context("when a panic occurs", func() {
			BeforeEach(func() {
				outcome, failure = build(func() {
					didRun = true
					innerCodeLocation = codelocation.New(0)
					panic("ack!")
				}, 0, failer, componentCodeLocation, componentIndex).Run()
			})

			It("should return the panic", func() {
				Ω(didRun).Should(BeTrue())

				Ω(outcome).Should(Equal(types.ExampleStatePanicked))
				innerCodeLocation.LineNumber++
				Ω(failure).Should(Equal(types.ExampleFailure{
					Message:               "Test Panicked",
					Location:              innerCodeLocation,
					ForwardedPanic:        "ack!",
					ComponentIndex:        componentIndex,
					ComponentType:         componentType,
					ComponentCodeLocation: componentCodeLocation,
				}))
			})
		})
	})
}

func AsynchronousSharedRunnerBehaviors(build func(body interface{}, timeout time.Duration, failer *Failer.Failer, componentCodeLocation types.CodeLocation, componentIndex int) runnable, componentType types.ExampleComponentType) {
	var (
		outcome types.ExampleState
		failure types.ExampleFailure

		failer *Failer.Failer

		componentIndex        int
		componentCodeLocation types.CodeLocation
		innerCodeLocation     types.CodeLocation

		didRun bool
	)

	BeforeEach(func() {
		failer = Failer.New()
		componentIndex = 3
		componentCodeLocation = codelocation.New(0)
		innerCodeLocation = codelocation.New(0)

		didRun = false
	})

	Describe("asynchronous functions", func() {
		var timeoutDuration time.Duration

		BeforeEach(func() {
			timeoutDuration = time.Duration(1 * float64(time.Second))
		})

		Context("when running", func() {
			It("should run the function as a goroutine, and block until it's done", func() {
				initialNumberOfGoRoutines := runtime.NumGoroutine()
				numberOfGoRoutines := 0

				build(func(done Done) {
					didRun = true
					numberOfGoRoutines = runtime.NumGoroutine()
					close(done)
				}, timeoutDuration, failer, componentCodeLocation, componentIndex).Run()

				Ω(didRun).Should(BeTrue())
				Ω(numberOfGoRoutines).Should(Equal(initialNumberOfGoRoutines + 1))
			})
		})

		Context("when the function passes", func() {
			BeforeEach(func() {
				outcome, failure = build(func(done Done) {
					didRun = true
					close(done)
				}, timeoutDuration, failer, componentCodeLocation, componentIndex).Run()
			})

			It("should have a succesful outcome", func() {
				Ω(didRun).Should(BeTrue())
				Ω(outcome).Should(Equal(types.ExampleStatePassed))
				Ω(failure).Should(BeZero())
			})
		})

		Context("when the function fails", func() {
			BeforeEach(func() {
				outcome, failure = build(func(done Done) {
					didRun = true
					failer.Fail("bam", innerCodeLocation)
					time.Sleep(20 * time.Millisecond)
					panic("doesn't matter")
					close(done)
				}, 10*time.Millisecond, failer, componentCodeLocation, componentIndex).Run()
			})

			It("should return the failure", func() {
				Ω(didRun).Should(BeTrue())

				Ω(outcome).Should(Equal(types.ExampleStateFailed))
				Ω(failure).Should(Equal(types.ExampleFailure{
					Message:               "bam",
					Location:              innerCodeLocation,
					ForwardedPanic:        nil,
					ComponentIndex:        componentIndex,
					ComponentType:         componentType,
					ComponentCodeLocation: componentCodeLocation,
				}))
			})
		})

		Context("when the function times out", func() {
			BeforeEach(func() {
				outcome, failure = build(func(done Done) {
					didRun = true
					time.Sleep(20 * time.Millisecond)
					panic("doesn't matter")
					close(done)
				}, 10*time.Millisecond, failer, componentCodeLocation, componentIndex).Run()
			})

			It("should return the timeout", func() {
				Ω(didRun).Should(BeTrue())

				Ω(outcome).Should(Equal(types.ExampleStateTimedOut))
				Ω(failure).Should(Equal(types.ExampleFailure{
					Message:               "Timed out",
					Location:              componentCodeLocation,
					ForwardedPanic:        nil,
					ComponentIndex:        componentIndex,
					ComponentType:         componentType,
					ComponentCodeLocation: componentCodeLocation,
				}))
			})
		})

		Context("when the function panics", func() {
			BeforeEach(func() {
				outcome, failure = build(func(done Done) {
					didRun = true
					innerCodeLocation = codelocation.New(0)
					panic("ack!")
				}, 10*time.Millisecond, failer, componentCodeLocation, componentIndex).Run()
			})

			It("should return the panic", func() {
				Ω(didRun).Should(BeTrue())

				Ω(outcome).Should(Equal(types.ExampleStatePanicked))
				innerCodeLocation.LineNumber++
				Ω(failure).Should(Equal(types.ExampleFailure{
					Message:               "Test Panicked",
					Location:              innerCodeLocation,
					ForwardedPanic:        "ack!",
					ComponentIndex:        componentIndex,
					ComponentType:         componentType,
					ComponentCodeLocation: componentCodeLocation,
				}))
			})
		})
	})
}

func InvalidSharedRunnerBehaviors(build func(body interface{}, timeout time.Duration, failer *Failer.Failer, componentCodeLocation types.CodeLocation, componentIndex int) runnable, componentType types.ExampleComponentType) {
	var (
		failer                *Failer.Failer
		componentIndex        int
		componentCodeLocation types.CodeLocation
		innerCodeLocation     types.CodeLocation
	)

	BeforeEach(func() {
		failer = Failer.New()
		componentIndex = 3
		componentCodeLocation = codelocation.New(0)
		innerCodeLocation = codelocation.New(0)
	})

	Describe("invalid functions", func() {
		Context("when passed something that's not a function", func() {
			It("should panic", func() {
				Ω(func() {
					build("not a function", 0, failer, componentCodeLocation, componentIndex)
				}).Should(Panic())
			})
		})

		Context("when the function takes the wrong kind of argument", func() {
			It("should panic", func() {
				Ω(func() {
					build(func(oops string) {}, 0, failer, componentCodeLocation, componentIndex)
				}).Should(Panic())
			})
		})

		Context("when the function takes more than one argument", func() {
			It("should panic", func() {
				Ω(func() {
					build(func(done Done, oops string) {}, 0, failer, componentCodeLocation, componentIndex)
				}).Should(Panic())
			})
		})
	})
}

var _ = Describe("Shared RunnableNode behavior", func() {
	Describe("It Nodes", func() {
		build := func(body interface{}, timeout time.Duration, failer *Failer.Failer, componentCodeLocation types.CodeLocation, componentIndex int) runnable {
			return NewItNode("", body, internaltypes.FlagTypeFocused, componentCodeLocation, timeout, failer, componentIndex)
		}

		SynchronousSharedRunnerBehaviors(build, types.ExampleComponentTypeIt)
		AsynchronousSharedRunnerBehaviors(build, types.ExampleComponentTypeIt)
		InvalidSharedRunnerBehaviors(build, types.ExampleComponentTypeIt)
	})

	Describe("Measure Nodes", func() {
		build := func(body interface{}, _ time.Duration, failer *Failer.Failer, componentCodeLocation types.CodeLocation, componentIndex int) runnable {
			return NewMeasureNode("", func(Benchmarker) {
				reflect.ValueOf(body).Call([]reflect.Value{})
			}, internaltypes.FlagTypeFocused, componentCodeLocation, 10, failer, componentIndex)
		}

		SynchronousSharedRunnerBehaviors(build, types.ExampleComponentTypeMeasure)
	})

	Describe("BeforeEach Nodes", func() {
		build := func(body interface{}, timeout time.Duration, failer *Failer.Failer, componentCodeLocation types.CodeLocation, componentIndex int) runnable {
			return NewBeforeEachNode(body, componentCodeLocation, timeout, failer, componentIndex)
		}

		SynchronousSharedRunnerBehaviors(build, types.ExampleComponentTypeBeforeEach)
		AsynchronousSharedRunnerBehaviors(build, types.ExampleComponentTypeBeforeEach)
		InvalidSharedRunnerBehaviors(build, types.ExampleComponentTypeBeforeEach)
	})

	Describe("AfterEach Nodes", func() {
		build := func(body interface{}, timeout time.Duration, failer *Failer.Failer, componentCodeLocation types.CodeLocation, componentIndex int) runnable {
			return NewAfterEachNode(body, componentCodeLocation, timeout, failer, componentIndex)
		}

		SynchronousSharedRunnerBehaviors(build, types.ExampleComponentTypeAfterEach)
		AsynchronousSharedRunnerBehaviors(build, types.ExampleComponentTypeAfterEach)
		InvalidSharedRunnerBehaviors(build, types.ExampleComponentTypeAfterEach)
	})

	Describe("JustBeforeEach Nodes", func() {
		build := func(body interface{}, timeout time.Duration, failer *Failer.Failer, componentCodeLocation types.CodeLocation, componentIndex int) runnable {
			return NewJustBeforeEachNode(body, componentCodeLocation, timeout, failer, componentIndex)
		}

		SynchronousSharedRunnerBehaviors(build, types.ExampleComponentTypeJustBeforeEach)
		AsynchronousSharedRunnerBehaviors(build, types.ExampleComponentTypeJustBeforeEach)
		InvalidSharedRunnerBehaviors(build, types.ExampleComponentTypeJustBeforeEach)
	})
})
