package leafnodes_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/internal/leafnodes"
	. "github.com/onsi/gomega"

	"reflect"
	"runtime"
	"time"

	"github.com/onsi/ginkgo/internal/codelocation"
	Failer "github.com/onsi/ginkgo/internal/failer"
	"github.com/onsi/ginkgo/types"
)

type runnable interface {
	Run() (outcome types.SpecState, failure types.SpecFailure)
	CodeLocation() types.CodeLocation
}

func SynchronousSharedRunnerBehaviors(build func(body interface{}, timeout time.Duration, failer *Failer.Failer, componentCodeLocation types.CodeLocation, componentIndex int) runnable, componentType types.SpecComponentType) {
	var (
		outcome types.SpecState
		failure types.SpecFailure

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

				Ω(outcome).Should(Equal(types.SpecStatePassed))
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

				Ω(outcome).Should(Equal(types.SpecStateFailed))
				Ω(failure).Should(Equal(types.SpecFailure{
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

				Ω(outcome).Should(Equal(types.SpecStatePanicked))
				innerCodeLocation.LineNumber++
				Ω(failure).Should(Equal(types.SpecFailure{
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

func AsynchronousSharedRunnerBehaviors(build func(body interface{}, timeout time.Duration, failer *Failer.Failer, componentCodeLocation types.CodeLocation, componentIndex int) runnable, componentType types.SpecComponentType) {
	var (
		outcome types.SpecState
		failure types.SpecFailure

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
				Ω(outcome).Should(Equal(types.SpecStatePassed))
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

				Ω(outcome).Should(Equal(types.SpecStateFailed))
				Ω(failure).Should(Equal(types.SpecFailure{
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

				Ω(outcome).Should(Equal(types.SpecStateTimedOut))
				Ω(failure).Should(Equal(types.SpecFailure{
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

				Ω(outcome).Should(Equal(types.SpecStatePanicked))
				innerCodeLocation.LineNumber++
				Ω(failure).Should(Equal(types.SpecFailure{
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

func InvalidSharedRunnerBehaviors(build func(body interface{}, timeout time.Duration, failer *Failer.Failer, componentCodeLocation types.CodeLocation, componentIndex int) runnable, componentType types.SpecComponentType) {
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
			return NewItNode("", body, types.FlagTypeFocused, componentCodeLocation, timeout, failer, componentIndex)
		}

		SynchronousSharedRunnerBehaviors(build, types.SpecComponentTypeIt)
		AsynchronousSharedRunnerBehaviors(build, types.SpecComponentTypeIt)
		InvalidSharedRunnerBehaviors(build, types.SpecComponentTypeIt)
	})

	Describe("Measure Nodes", func() {
		build := func(body interface{}, _ time.Duration, failer *Failer.Failer, componentCodeLocation types.CodeLocation, componentIndex int) runnable {
			return NewMeasureNode("", func(Benchmarker) {
				reflect.ValueOf(body).Call([]reflect.Value{})
			}, types.FlagTypeFocused, componentCodeLocation, 10, failer, componentIndex)
		}

		SynchronousSharedRunnerBehaviors(build, types.SpecComponentTypeMeasure)
	})

	Describe("BeforeEach Nodes", func() {
		build := func(body interface{}, timeout time.Duration, failer *Failer.Failer, componentCodeLocation types.CodeLocation, componentIndex int) runnable {
			return NewBeforeEachNode(body, componentCodeLocation, timeout, failer, componentIndex)
		}

		SynchronousSharedRunnerBehaviors(build, types.SpecComponentTypeBeforeEach)
		AsynchronousSharedRunnerBehaviors(build, types.SpecComponentTypeBeforeEach)
		InvalidSharedRunnerBehaviors(build, types.SpecComponentTypeBeforeEach)
	})

	Describe("AfterEach Nodes", func() {
		build := func(body interface{}, timeout time.Duration, failer *Failer.Failer, componentCodeLocation types.CodeLocation, componentIndex int) runnable {
			return NewAfterEachNode(body, componentCodeLocation, timeout, failer, componentIndex)
		}

		SynchronousSharedRunnerBehaviors(build, types.SpecComponentTypeAfterEach)
		AsynchronousSharedRunnerBehaviors(build, types.SpecComponentTypeAfterEach)
		InvalidSharedRunnerBehaviors(build, types.SpecComponentTypeAfterEach)
	})

	Describe("JustBeforeEach Nodes", func() {
		build := func(body interface{}, timeout time.Duration, failer *Failer.Failer, componentCodeLocation types.CodeLocation, componentIndex int) runnable {
			return NewJustBeforeEachNode(body, componentCodeLocation, timeout, failer, componentIndex)
		}

		SynchronousSharedRunnerBehaviors(build, types.SpecComponentTypeJustBeforeEach)
		AsynchronousSharedRunnerBehaviors(build, types.SpecComponentTypeJustBeforeEach)
		InvalidSharedRunnerBehaviors(build, types.SpecComponentTypeJustBeforeEach)
	})
})
