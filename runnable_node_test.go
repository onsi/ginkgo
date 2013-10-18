package ginkgo

import (
	"github.com/onsi/ginkgo/types"
	. "github.com/onsi/gomega"
	"runtime"
	"time"
)

func init() {
	Describe("RunnableNode", func() {
		Describe("basic construction parameters", func() {
			It("should store off the passed in code location", func() {
				codeLocation := types.GenerateCodeLocation(0)
				Ω(newRunnableNode(func() {}, codeLocation, 0).codeLocation).Should(Equal(codeLocation))
			})
		})

		Describe("running the passed in function", func() {
			Context("when the function is synchronous and does not panic", func() {
				It("should run the function and report a runOutcomeCompleted", func() {
					didRun := false
					runnableNode := newRunnableNode(func() {
						didRun = true
					}, types.GenerateCodeLocation(0), 0)

					outcome, failure := runnableNode.run()

					Ω(didRun).Should(BeTrue())
					Ω(outcome).Should(Equal(runOutcomeCompleted))
					Ω(failure).Should(BeZero())
				})
			})

			Context("when the function is synchronous and *does* panic", func() {
				var (
					codeLocation types.CodeLocation
					outcome      runOutcome
					failure      failureData
				)

				BeforeEach(func() {
					node := newRunnableNode(func() {
						codeLocation = types.GenerateCodeLocation(0)
						panic("ack!")
					}, types.GenerateCodeLocation(0), 0)

					outcome, failure = node.run()
				})

				It("should run the function and report a runOutcomePanicked", func() {
					Ω(outcome).Should(Equal(runOutcomePanicked))
					Ω(failure.message).Should(Equal("Test Panicked"))
				})

				It("should include the code location of the panic itself", func() {
					Ω(failure.codeLocation.FileName).Should(Equal(codeLocation.FileName))
					Ω(failure.codeLocation.LineNumber).Should(Equal(codeLocation.LineNumber + 1))
				})

				It("should include the panic data", func() {
					Ω(failure.forwardedPanic).Should(Equal("ack!"))
				})
			})

			Context("when the function is asynchronous", func() {
				var (
					node               *runnableNode
					sleepDuration      time.Duration
					timeoutDuration    time.Duration
					numberOfGoRoutines int
				)

				BeforeEach(func() {
					sleepDuration = time.Duration(0.001 * float64(time.Second))
					timeoutDuration = time.Duration(1 * float64(time.Second))
				})

				JustBeforeEach(func() {
					node = newRunnableNode(func(done Done) {
						numberOfGoRoutines = runtime.NumGoroutine()
						time.Sleep(sleepDuration)
						done <- true
					}, types.GenerateCodeLocation(0), timeoutDuration)
				})

				It("should run the function as a goroutine", func() {
					initialNumberOfGoRoutines := runtime.NumGoroutine()
					outcome, failure := node.run()

					Ω(outcome).Should(Equal(runOutcomeCompleted))
					Ω(failure).Should(BeZero())

					Ω(numberOfGoRoutines).Should(Equal(initialNumberOfGoRoutines + 1))
				})

				Context("when the function takes longer than the timeout", func() {
					BeforeEach(func() {
						sleepDuration = time.Duration(0.002 * float64(time.Second))
						timeoutDuration = time.Duration(0.001 * float64(time.Second))
					})

					It("should timeout", func() {
						outcome, failure := node.run()
						Ω(outcome).Should(Equal(runOutcomeTimedOut))
						Ω(failure.message).Should(Equal("Timed out"))
						Ω(failure.codeLocation).Should(Equal(node.codeLocation))
					})
				})
			})

			Context("when the function takes the wrong kind of argument", func() {
				It("should panic", func() {
					Ω(func() {
						newRunnableNode(func(oops string) {
						}, types.GenerateCodeLocation(0), 0)
					}).Should(Panic())
				})
			})

			Context("when the function takes more than one argument", func() {
				It("should panic", func() {
					Ω(func() {
						newRunnableNode(func(done Done, oops string) {
						}, types.GenerateCodeLocation(0), 0)
					}).Should(Panic())
				})
			})
		})
	})

	Describe("ItNodes", func() {
		It("should save off the text and flags", func() {
			codeLocation := types.GenerateCodeLocation(0)
			it := newItNode("my it node", func() {}, flagTypeFocused, codeLocation, 0)
			Ω(it.flag).Should(Equal(flagTypeFocused))
			Ω(it.text).Should(Equal("my it node"))
			Ω(it.codeLocation).Should(Equal(codeLocation))
		})
	})
}
