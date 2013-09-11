package ginkgo

import (
	"fmt"
	. "github.com/onsi/gomega"
	"math"
	"time"
)

func init() {
	Describe("Example", func() {
		var it *itNode

		BeforeEach(func() {
			it = newItNode("It", func() {}, flagTypeNone, generateCodeLocation(0), 0)
		})

		Describe("creating examples and adding container nodes", func() {
			var (
				containerA *containerNode
				containerB *containerNode
				ex         *example
			)

			BeforeEach(func() {
				containerA = newContainerNode("A", flagTypeNone, generateCodeLocation(0))
				containerB = newContainerNode("B", flagTypeNone, generateCodeLocation(0))
			})

			JustBeforeEach(func() {
				ex = newExample(it)
				ex.addContainerNode(containerB)
				ex.addContainerNode(containerA)
			})

			It("should store off the it node", func() {
				Ω(ex.subject).Should(Equal(it))
			})

			It("should store off the container nodes in reverse order", func() {
				Ω(ex.containers).Should(Equal([]*containerNode{containerA, containerB}))
			})

			It("should provide the concatenated strings", func() {
				Ω(ex.concatenatedString()).Should(Equal("A B It"))
			})

			Context("when neither the It node nor the containers is focused or pending", func() {
				It("should not be focused or pending", func() {
					Ω(ex.focused).Should(BeFalse())
					Ω(ex.state).Should(BeZero())
				})
			})

			Context("when the It node is focused", func() {
				BeforeEach(func() {
					it.flag = flagTypeFocused
				})

				It("should be focused", func() {
					Ω(ex.focused).Should(BeTrue())
				})
			})

			Context("when one of the containers is focused", func() {
				BeforeEach(func() {
					containerB.flag = flagTypeFocused
				})

				It("should be focused", func() {
					Ω(ex.focused).Should(BeTrue())
				})
			})

			Context("when the It node is pending", func() {
				BeforeEach(func() {
					it.flag = flagTypePending
				})

				It("should be in the pending state", func() {
					Ω(ex.state).Should(Equal(ExampleStatePending))
				})
			})

			Context("when one of the containers is pending", func() {
				BeforeEach(func() {
					containerB.flag = flagTypePending
				})

				It("should be in the pending state", func() {
					Ω(ex.state).Should(Equal(ExampleStatePending))
				})
			})

			Context("when one container is pending and another container is focused", func() {
				BeforeEach(func() {
					containerA.flag = flagTypeFocused
					containerB.flag = flagTypePending
				})

				It("should be focused and have the pending state", func() {
					Ω(ex.focused).Should(BeTrue())
					Ω(ex.state).Should(Equal(ExampleStatePending))
				})
			})
		})

		Describe("Skipping an example", func() {
			It("should mark the example as skipped", func() {
				ex := newExample(it)
				ex.skip()
				Ω(ex.state).Should(Equal(ExampleStateSkipped))
			})
		})

		Describe("skippedOrPending", func() {
			It("should be false if the example is neither pending nor skipped", func() {
				ex := newExample(it)
				Ω(ex.skippedOrPending()).Should(BeFalse())
			})

			It("should be true if the example is pending", func() {
				it.flag = flagTypePending
				ex := newExample(it)
				Ω(ex.skippedOrPending()).Should(BeTrue())
			})

			It("should be true if the example is skipped", func() {
				ex := newExample(it)
				ex.skip()
				Ω(ex.skippedOrPending()).Should(BeTrue())
			})
		})

		Describe("pending", func() {
			It("should be false if the example is not pending", func() {
				ex := newExample(it)
				Ω(ex.pending()).Should(BeFalse())
			})

			It("should be true if the example is pending", func() {
				it.flag = flagTypePending
				ex := newExample(it)
				Ω(ex.pending()).Should(BeTrue())
			})
		})

		Describe("running examples and getting summaries", func() {
			var (
				orderedList    []string
				it             *itNode
				innerContainer *containerNode
				outerContainer *containerNode
				ex             *example
			)

			newNode := func(identifier string) *runnableNode {
				return newRunnableNode(func() {
					orderedList = append(orderedList, identifier)
				}, generateCodeLocation(0), 0)
			}

			BeforeEach(func() {
				orderedList = make([]string, 0)
				it = newItNode("it", func() {
					orderedList = append(orderedList, "IT")
					time.Sleep(time.Duration(0.01 * float64(time.Second)))
				}, flagTypeNone, generateCodeLocation(0), 0)
				ex = newExample(it)

				innerContainer = newContainerNode("inner", flagTypeNone, generateCodeLocation(0))
				innerContainer.pushBeforeEachNode(newNode("INNER_BEFORE_A"))
				innerContainer.pushBeforeEachNode(newNode("INNER_BEFORE_B"))
				innerContainer.pushJustBeforeEachNode(newNode("INNER_JUST_BEFORE_A"))
				innerContainer.pushJustBeforeEachNode(newNode("INNER_JUST_BEFORE_B"))
				innerContainer.pushAfterEachNode(newNode("INNER_AFTER_A"))
				innerContainer.pushAfterEachNode(newNode("INNER_AFTER_B"))

				ex.addContainerNode(innerContainer)

				outerContainer = newContainerNode("outer", flagTypeNone, generateCodeLocation(0))
				outerContainer.pushBeforeEachNode(newNode("OUTER_BEFORE_A"))
				outerContainer.pushBeforeEachNode(newNode("OUTER_BEFORE_B"))
				outerContainer.pushJustBeforeEachNode(newNode("OUTER_JUST_BEFORE_A"))
				outerContainer.pushJustBeforeEachNode(newNode("OUTER_JUST_BEFORE_B"))
				outerContainer.pushAfterEachNode(newNode("OUTER_AFTER_A"))
				outerContainer.pushAfterEachNode(newNode("OUTER_AFTER_B"))

				ex.addContainerNode(outerContainer)
			})

			It("should report that it has an it node", func() {
				Ω(ex.subjectComponentType()).Should(Equal(ExampleComponentTypeIt))
			})

			It("runs the before/justBefore/after nodes in each of the containers, and the it node, in the correct order", func() {
				ex.run()
				Ω(orderedList).Should(Equal([]string{
					"OUTER_BEFORE_A",
					"OUTER_BEFORE_B",
					"INNER_BEFORE_A",
					"INNER_BEFORE_B",
					"OUTER_JUST_BEFORE_A",
					"OUTER_JUST_BEFORE_B",
					"INNER_JUST_BEFORE_A",
					"INNER_JUST_BEFORE_B",
					"IT",
					"INNER_AFTER_B",
					"INNER_AFTER_A",
					"OUTER_AFTER_B",
					"OUTER_AFTER_A",
				}))
			})

			Describe("the summary", func() {
				It("has the texts and code locations for the container nodes and the it node", func() {
					ex.run()
					summary := ex.summary()
					Ω(summary.ComponentTexts).Should(Equal([]string{
						"outer", "inner", "it",
					}))
					Ω(summary.ComponentCodeLocations).Should(Equal([]CodeLocation{
						outerContainer.codeLocation, innerContainer.codeLocation, it.codeLocation,
					}))
				})
			})

			Context("when none of the runnable nodes fail", func() {
				It("has a summary reporting no failure", func() {
					ex.run()
					summary := ex.summary()
					Ω(summary.State).Should(Equal(ExampleStatePassed))
					Ω(summary.RunTime.Seconds()).Should(BeNumerically(">", 0.01))
					Ω(summary.Benchmark.IsBenchmark).Should(BeFalse())
				})
			})

			componentTypes := []string{"BeforeEach", "JustBeforeEach", "AfterEach"}
			expectedComponentTypes := []ExampleComponentType{ExampleComponentTypeBeforeEach, ExampleComponentTypeJustBeforeEach, ExampleComponentTypeAfterEach}
			pushFuncs := []func(container *containerNode, node *runnableNode){(*containerNode).pushBeforeEachNode, (*containerNode).pushJustBeforeEachNode, (*containerNode).pushAfterEachNode}

			for i := range componentTypes {
				Context(fmt.Sprintf("when a %s node fails", componentTypes[i]), func() {
					var componentCodeLocation CodeLocation

					BeforeEach(func() {
						componentCodeLocation = generateCodeLocation(0)
					})

					Context("because an expectation failed", func() {
						var failure failureData

						BeforeEach(func() {
							failure = failureData{
								message:      fmt.Sprintf("%s failed", componentTypes[i]),
								codeLocation: generateCodeLocation(0),
							}
							node := newRunnableNode(func() {
								ex.fail(failure)
								ex.fail(failureData{message: "IGNORE ME!"})
							}, componentCodeLocation, 0)

							pushFuncs[i](innerContainer, node)
						})

						It("has a summary with the correct failure report", func() {
							ex.run()
							summary := ex.summary()

							Ω(summary.State).Should(Equal(ExampleStateFailed))
							Ω(summary.Failure.Message).Should(Equal(failure.message))
							Ω(summary.Failure.Location).Should(Equal(failure.codeLocation))
							Ω(summary.Failure.ForwardedPanic).Should(BeNil())
							Ω(summary.Failure.ComponentIndex).Should(Equal(1), "Should be the inner container that failed")
							Ω(summary.Failure.ComponentType).Should(Equal(expectedComponentTypes[i]))
							Ω(summary.Failure.ComponentCodeLocation).Should(Equal(componentCodeLocation))

							Ω(ex.failed()).Should(BeTrue())
						})
					})

					Context("because the function panicked", func() {
						var panicCodeLocation CodeLocation

						BeforeEach(func() {
							node := newRunnableNode(func() {
								panicCodeLocation = generateCodeLocation(0)
								panic("kaboom!")
							}, componentCodeLocation, 0)

							pushFuncs[i](innerContainer, node)
						})

						It("has a summary with the correct failure report", func() {
							ex.run()
							summary := ex.summary()

							Ω(summary.State).Should(Equal(ExampleStatePanicked))
							Ω(summary.Failure.Message).Should(Equal("Test Panicked"))
							Ω(summary.Failure.Location.FileName).Should(Equal(panicCodeLocation.FileName))
							Ω(summary.Failure.Location.LineNumber).Should(Equal(panicCodeLocation.LineNumber+1), "Expect panic code location to be correct")
							Ω(summary.Failure.ForwardedPanic).Should(Equal("kaboom!"))
							Ω(summary.Failure.ComponentIndex).Should(Equal(1), "Should be the inner container that failed")
							Ω(summary.Failure.ComponentType).Should(Equal(expectedComponentTypes[i]))
							Ω(summary.Failure.ComponentCodeLocation).Should(Equal(componentCodeLocation))

							Ω(ex.failed()).Should(BeTrue())
						})
					})

					Context("because the function timed out", func() {
						BeforeEach(func() {
							node := newRunnableNode(func(done Done) {
								time.Sleep(time.Duration(0.002 * float64(time.Second)))
								done <- true
							}, componentCodeLocation, time.Duration(0.001*float64(time.Second)))

							pushFuncs[i](innerContainer, node)
						})

						It("has a summary with the correct failure report", func() {
							ex.run()
							summary := ex.summary()

							Ω(summary.State).Should(Equal(ExampleStateTimedOut))
							Ω(summary.Failure.Message).Should(Equal("Timed out"))
							Ω(summary.Failure.Location).Should(Equal(componentCodeLocation))
							Ω(summary.Failure.ForwardedPanic).Should(BeNil())
							Ω(summary.Failure.ComponentIndex).Should(Equal(1), "Should be the inner container that failed")
							Ω(summary.Failure.ComponentType).Should(Equal(expectedComponentTypes[i]))
							Ω(summary.Failure.ComponentCodeLocation).Should(Equal(componentCodeLocation))

							Ω(ex.failed()).Should(BeTrue())
						})
					})
				})
			}

			Context("when the it node fails", func() {
				var componentCodeLocation CodeLocation

				BeforeEach(func() {
					componentCodeLocation = generateCodeLocation(0)
				})

				Context("because an expectation failed", func() {
					var failure failureData

					BeforeEach(func() {
						failure = failureData{
							message:      "it failed",
							codeLocation: generateCodeLocation(0),
						}
						ex.subject = newItNode("it", func() {
							ex.fail(failure)
							ex.fail(failureData{message: "IGNORE ME!"})
						}, flagTypeNone, componentCodeLocation, 0)
					})

					It("has a summary with the correct failure report", func() {
						ex.run()
						summary := ex.summary()

						Ω(summary.State).Should(Equal(ExampleStateFailed))
						Ω(summary.Failure.Message).Should(Equal(failure.message))
						Ω(summary.Failure.Location).Should(Equal(failure.codeLocation))
						Ω(summary.Failure.ForwardedPanic).Should(BeNil())
						Ω(summary.Failure.ComponentIndex).Should(Equal(2), "Should be the it node that failed")
						Ω(summary.Failure.ComponentType).Should(Equal(ExampleComponentTypeIt))
						Ω(summary.Failure.ComponentCodeLocation).Should(Equal(componentCodeLocation))

						Ω(ex.failed()).Should(BeTrue())
					})
				})

				Context("because the function panicked", func() {
					var panicCodeLocation CodeLocation

					BeforeEach(func() {
						ex.subject = newItNode("it", func() {
							panicCodeLocation = generateCodeLocation(0)
							panic("kaboom!")
						}, flagTypeNone, componentCodeLocation, 0)
					})

					It("has a summary with the correct failure report", func() {
						ex.run()
						summary := ex.summary()

						Ω(summary.State).Should(Equal(ExampleStatePanicked))
						Ω(summary.Failure.Message).Should(Equal("Test Panicked"))
						Ω(summary.Failure.Location.FileName).Should(Equal(panicCodeLocation.FileName))
						Ω(summary.Failure.Location.LineNumber).Should(Equal(panicCodeLocation.LineNumber+1), "Expect panic code location to be correct")
						Ω(summary.Failure.ForwardedPanic).Should(Equal("kaboom!"))
						Ω(summary.Failure.ComponentIndex).Should(Equal(2), "Should be the it node that failed")
						Ω(summary.Failure.ComponentType).Should(Equal(ExampleComponentTypeIt))
						Ω(summary.Failure.ComponentCodeLocation).Should(Equal(componentCodeLocation))

						Ω(ex.failed()).Should(BeTrue())
					})
				})

				Context("because the function timed out", func() {
					BeforeEach(func() {
						ex.subject = newItNode("it", func(done Done) {
							time.Sleep(time.Duration(0.002 * float64(time.Second)))
							done <- true
						}, flagTypeNone, componentCodeLocation, time.Duration(0.001*float64(time.Second)))
					})

					It("has a summary with the correct failure report", func() {
						ex.run()
						summary := ex.summary()

						Ω(summary.State).Should(Equal(ExampleStateTimedOut))
						Ω(summary.Failure.Message).Should(Equal("Timed out"))
						Ω(summary.Failure.Location).Should(Equal(componentCodeLocation))
						Ω(summary.Failure.ForwardedPanic).Should(BeNil())
						Ω(summary.Failure.ComponentIndex).Should(Equal(2), "Should be the it node that failed")
						Ω(summary.Failure.ComponentType).Should(Equal(ExampleComponentTypeIt))
						Ω(summary.Failure.ComponentCodeLocation).Should(Equal(componentCodeLocation))

						Ω(ex.failed()).Should(BeTrue())
					})
				})
			})
		})

		Describe("running benchmark examples and getting summaries", func() {
			var (
				runs                  int
				componentCodeLocation CodeLocation
				ex                    *example
			)

			BeforeEach(func() {
				runs = 0
				componentCodeLocation = generateCodeLocation(0)
			})

			It("should report that it has a benchmark", func() {
				ex = newExample(newBenchmarkNode("benchmark", func() {}, flagTypeNone, componentCodeLocation, 0, 5, time.Duration(1*float64(time.Second))))
				Ω(ex.subjectComponentType()).Should(Equal(ExampleComponentTypeBenchmark))
			})

			Context("when the benchmark does not fail", func() {
				BeforeEach(func() {
					ex = newExample(newBenchmarkNode("benchmark", func() {
						runs++
						time.Sleep(time.Duration(0.001 * float64(time.Second) * float64(runs)))
					}, flagTypeNone, componentCodeLocation, 0, 5, time.Duration(1*float64(time.Second))))
				})

				It("runs the benchmark samples number of times and returns timing statistics", func() {
					ex.run()
					summary := ex.summary()

					Ω(runs).Should(Equal(5))

					Ω(summary.State).Should(Equal(ExampleStatePassed))
					Ω(summary.RunTime.Seconds()).Should(BeNumerically(">", 0.001+0.002+0.003+0.004+0.005))
					Ω(summary.Benchmark.IsBenchmark).Should(BeTrue())
					Ω(summary.Benchmark.NumberOfSamples).Should(Equal(5))
					Ω(summary.Benchmark.FastestTime.Seconds()).Should(BeNumerically("~", 0.001, 0.0009))
					Ω(summary.Benchmark.SlowestTime.Seconds()).Should(BeNumerically("~", 0.005, 0.0009))
					Ω(summary.Benchmark.AverageTime.Seconds()).Should(BeNumerically("~", 0.003, 0.0009))
					stdDev := math.Sqrt((0.001*0.001+0.002*0.002+0.003*0.003+0.004*0.004+0.005*0.005)/5 - 0.003*0.003)
					Ω(summary.Benchmark.StdDeviation.Seconds()).Should(BeNumerically("~", stdDev, 0.0009))
				})
			})

			Context("when one of the benchmark samples fails", func() {
				BeforeEach(func() {
					ex = newExample(newBenchmarkNode("benchmark", func() {
						runs++
						if runs == 3 {
							ex.fail(failureData{})
						}
						time.Sleep(time.Duration(0.001 * float64(time.Second) * float64(runs)))
					}, flagTypeNone, componentCodeLocation, 0, 5, time.Duration(1*float64(time.Second))))
				})

				It("marks the benchmark as failed and doesn't run any more samples", func() {
					ex.run()
					summary := ex.summary()

					Ω(runs).Should(Equal(3))

					Ω(summary.State).Should(Equal(ExampleStateFailed))
					Ω(summary.Benchmark.IsBenchmark).Should(BeTrue())
					Ω(summary.Benchmark.NumberOfSamples).Should(Equal(5))
					Ω(summary.Benchmark.FastestTime.Seconds()).Should(BeNumerically("==", 0))
					Ω(summary.Benchmark.SlowestTime.Seconds()).Should(BeNumerically("==", 0))
					Ω(summary.Benchmark.AverageTime.Seconds()).Should(BeNumerically("==", 0))
					Ω(summary.Benchmark.StdDeviation.Seconds()).Should(BeNumerically("==", 0))
				})
			})

			Context("when a benchmark sample fails", func() {
				BeforeEach(func() {
					ex = newExample(newBenchmarkNode("benchmark", func() {
						runs++
						time.Sleep(time.Duration(0.001 * float64(time.Second) * float64(runs)))
					}, flagTypeNone, componentCodeLocation, 0, 5, time.Duration(0.0025*float64(time.Second))))
				})

				It("marks the benchmark as failed and doesn't run any more samples", func() {
					ex.run()
					summary := ex.summary()

					Ω(runs).Should(Equal(3))

					Ω(summary.State).Should(Equal(ExampleStateFailed))
					Ω(summary.Failure.Message).Should(ContainSubstring("Benchmark sample took"))
					Ω(summary.Failure.Location).Should(Equal(componentCodeLocation))
					Ω(summary.Failure.ComponentIndex).Should(Equal(0))
					Ω(summary.Failure.ComponentType).Should(Equal(ExampleComponentTypeBenchmark))
					Ω(summary.Failure.ForwardedPanic).Should(BeNil())

					Ω(summary.Benchmark.IsBenchmark).Should(BeTrue())
					Ω(summary.Benchmark.NumberOfSamples).Should(Equal(5))
					Ω(summary.Benchmark.FastestTime.Seconds()).Should(BeNumerically("==", 0))
					Ω(summary.Benchmark.SlowestTime.Seconds()).Should(BeNumerically("==", 0))
					Ω(summary.Benchmark.AverageTime.Seconds()).Should(BeNumerically("==", 0))
					Ω(summary.Benchmark.StdDeviation.Seconds()).Should(BeNumerically("==", 0))
				})
			})
		})
	})
}
