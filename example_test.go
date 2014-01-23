package ginkgo

import (
	"fmt"
	"github.com/onsi/ginkgo/types"
	. "github.com/onsi/gomega"
	"time"
)

func init() {
	Describe("Example", func() {
		var it *itNode

		BeforeEach(func() {
			it = newItNode("It", func() {}, flagTypeNone, types.GenerateCodeLocation(0), 0)
		})

		Describe("creating examples and adding container nodes", func() {
			var (
				containerA *containerNode
				containerB *containerNode
				ex         *example
			)

			BeforeEach(func() {
				containerA = newContainerNode("A", flagTypeNone, types.GenerateCodeLocation(0))
				containerB = newContainerNode("B", flagTypeNone, types.GenerateCodeLocation(0))
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
					Ω(ex.state).Should(Equal(types.ExampleStatePending))
				})
			})

			Context("when one of the containers is pending", func() {
				BeforeEach(func() {
					containerB.flag = flagTypePending
				})

				It("should be in the pending state", func() {
					Ω(ex.state).Should(Equal(types.ExampleStatePending))
				})
			})

			Context("when one container is pending and another container is focused", func() {
				BeforeEach(func() {
					containerA.flag = flagTypeFocused
					containerB.flag = flagTypePending
				})

				It("should be focused and have the pending state", func() {
					Ω(ex.focused).Should(BeTrue())
					Ω(ex.state).Should(Equal(types.ExampleStatePending))
				})
			})
		})

		Describe("Skipping an example", func() {
			It("should mark the example as skipped", func() {
				ex := newExample(it)
				ex.skip()
				Ω(ex.state).Should(Equal(types.ExampleStateSkipped))
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
				}, types.GenerateCodeLocation(0), 0)
			}

			BeforeEach(func() {
				orderedList = make([]string, 0)
				it = newItNode("it", func() {
					orderedList = append(orderedList, "IT")
					time.Sleep(time.Duration(0.01 * float64(time.Second)))
				}, flagTypeNone, types.GenerateCodeLocation(0), 0)
				ex = newExample(it)

				innerContainer = newContainerNode("inner", flagTypeNone, types.GenerateCodeLocation(0))
				innerContainer.pushBeforeEachNode(newNode("INNER_BEFORE_A"))
				innerContainer.pushBeforeEachNode(newNode("INNER_BEFORE_B"))
				innerContainer.pushJustBeforeEachNode(newNode("INNER_JUST_BEFORE_A"))
				innerContainer.pushJustBeforeEachNode(newNode("INNER_JUST_BEFORE_B"))
				innerContainer.pushAfterEachNode(newNode("INNER_AFTER_A"))
				innerContainer.pushAfterEachNode(newNode("INNER_AFTER_B"))

				ex.addContainerNode(innerContainer)

				outerContainer = newContainerNode("outer", flagTypeNone, types.GenerateCodeLocation(0))
				outerContainer.pushBeforeEachNode(newNode("OUTER_BEFORE_A"))
				outerContainer.pushBeforeEachNode(newNode("OUTER_BEFORE_B"))
				outerContainer.pushJustBeforeEachNode(newNode("OUTER_JUST_BEFORE_A"))
				outerContainer.pushJustBeforeEachNode(newNode("OUTER_JUST_BEFORE_B"))
				outerContainer.pushAfterEachNode(newNode("OUTER_AFTER_A"))
				outerContainer.pushAfterEachNode(newNode("OUTER_AFTER_B"))

				ex.addContainerNode(outerContainer)
			})

			It("should report that it has an it node", func() {
				Ω(ex.subjectComponentType()).Should(Equal(types.ExampleComponentTypeIt))
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
					"INNER_AFTER_A",
					"INNER_AFTER_B",
					"OUTER_AFTER_A",
					"OUTER_AFTER_B",
				}))
			})

			Describe("the summary", func() {
				It("has the texts and code locations for the container nodes and the it node", func() {
					ex.run()
					summary := ex.summary("suite-id")
					Ω(summary.ComponentTexts).Should(Equal([]string{
						"outer", "inner", "it",
					}))
					Ω(summary.ComponentCodeLocations).Should(Equal([]types.CodeLocation{
						outerContainer.codeLocation, innerContainer.codeLocation, it.codeLocation,
					}))
				})

				It("should have the passed in SuiteID", func() {
					ex.run()
					summary := ex.summary("suite-id")
					Ω(summary.SuiteID).Should(Equal("suite-id"))
				})

				It("should include the example's index", func() {
					ex.exampleIndex = 17
					ex.run()
					summary := ex.summary("suite-id")
					Ω(summary.ExampleIndex).Should(Equal(17))
				})
			})

			Describe("the GinkgoTestDescription", func() {
				It("should have the GinkgoTestDescription", func() {
					ginkgoTestDescription := ex.ginkgoTestDescription()
					Ω(ginkgoTestDescription.ComponentTexts).Should(Equal([]string{
						"inner", "it",
					}))

					Ω(ginkgoTestDescription.FullTestText).Should(Equal("inner it"))
					Ω(ginkgoTestDescription.TestText).Should(Equal("it"))
					Ω(ginkgoTestDescription.IsMeasurement).Should(BeFalse())
					Ω(ginkgoTestDescription.FileName).Should(Equal(it.codeLocation.FileName))
					Ω(ginkgoTestDescription.LineNumber).Should(Equal(it.codeLocation.LineNumber))
				})
			})

			Context("when none of the runnable nodes fail", func() {
				It("has a summary reporting no failure", func() {
					ex.run()
					summary := ex.summary("suite-id")
					Ω(summary.State).Should(Equal(types.ExampleStatePassed))
					Ω(summary.RunTime.Seconds()).Should(BeNumerically(">", 0.01))
					Ω(summary.IsMeasurement).Should(BeFalse())
				})
			})

			componentTypes := []string{"BeforeEach", "JustBeforeEach", "AfterEach"}
			expectedComponentTypes := []types.ExampleComponentType{types.ExampleComponentTypeBeforeEach, types.ExampleComponentTypeJustBeforeEach, types.ExampleComponentTypeAfterEach}
			pushFuncs := []func(container *containerNode, node *runnableNode){(*containerNode).pushBeforeEachNode, (*containerNode).pushJustBeforeEachNode, (*containerNode).pushAfterEachNode}

			for i := range componentTypes {
				Context(fmt.Sprintf("when a %s node fails", componentTypes[i]), func() {
					var componentCodeLocation types.CodeLocation

					BeforeEach(func() {
						componentCodeLocation = types.GenerateCodeLocation(0)
					})

					Context("because an expectation failed", func() {
						var failure failureData

						BeforeEach(func() {
							failure = failureData{
								message:      fmt.Sprintf("%s failed", componentTypes[i]),
								codeLocation: types.GenerateCodeLocation(0),
							}
							node := newRunnableNode(func() {
								ex.fail(failure)
								ex.fail(failureData{message: "IGNORE ME!"})
							}, componentCodeLocation, 0)

							pushFuncs[i](innerContainer, node)
						})

						It("has a summary with the correct failure report", func() {
							ex.run()
							summary := ex.summary("suite-id")

							Ω(summary.State).Should(Equal(types.ExampleStateFailed))
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
						var panicCodeLocation types.CodeLocation

						BeforeEach(func() {
							node := newRunnableNode(func() {
								panicCodeLocation = types.GenerateCodeLocation(0)
								panic("kaboom!")
							}, componentCodeLocation, 0)

							pushFuncs[i](innerContainer, node)
						})

						It("has a summary with the correct failure report", func() {
							ex.run()
							summary := ex.summary("suite-id")

							Ω(summary.State).Should(Equal(types.ExampleStatePanicked))
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
							summary := ex.summary("suite-id")

							Ω(summary.State).Should(Equal(types.ExampleStateTimedOut))
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
				var componentCodeLocation types.CodeLocation

				BeforeEach(func() {
					componentCodeLocation = types.GenerateCodeLocation(0)
				})

				Context("because an expectation failed", func() {
					var failure failureData

					BeforeEach(func() {
						failure = failureData{
							message:      "it failed",
							codeLocation: types.GenerateCodeLocation(0),
						}
						ex.subject = newItNode("it", func() {
							ex.fail(failure)
							ex.fail(failureData{message: "IGNORE ME!"})
						}, flagTypeNone, componentCodeLocation, 0)
					})

					It("has a summary with the correct failure report", func() {
						ex.run()
						summary := ex.summary("suite-id")

						Ω(summary.State).Should(Equal(types.ExampleStateFailed))
						Ω(summary.Failure.Message).Should(Equal(failure.message))
						Ω(summary.Failure.Location).Should(Equal(failure.codeLocation))
						Ω(summary.Failure.ForwardedPanic).Should(BeNil())
						Ω(summary.Failure.ComponentIndex).Should(Equal(2), "Should be the it node that failed")
						Ω(summary.Failure.ComponentType).Should(Equal(types.ExampleComponentTypeIt))
						Ω(summary.Failure.ComponentCodeLocation).Should(Equal(componentCodeLocation))

						Ω(ex.failed()).Should(BeTrue())
					})
				})

				Context("because the function panicked", func() {
					var panicCodeLocation types.CodeLocation

					BeforeEach(func() {
						ex.subject = newItNode("it", func() {
							panicCodeLocation = types.GenerateCodeLocation(0)
							panic("kaboom!")
						}, flagTypeNone, componentCodeLocation, 0)
					})

					It("has a summary with the correct failure report", func() {
						ex.run()
						summary := ex.summary("suite-id")

						Ω(summary.State).Should(Equal(types.ExampleStatePanicked))
						Ω(summary.Failure.Message).Should(Equal("Test Panicked"))
						Ω(summary.Failure.Location.FileName).Should(Equal(panicCodeLocation.FileName))
						Ω(summary.Failure.Location.LineNumber).Should(Equal(panicCodeLocation.LineNumber+1), "Expect panic code location to be correct")
						Ω(summary.Failure.ForwardedPanic).Should(Equal("kaboom!"))
						Ω(summary.Failure.ComponentIndex).Should(Equal(2), "Should be the it node that failed")
						Ω(summary.Failure.ComponentType).Should(Equal(types.ExampleComponentTypeIt))
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
						summary := ex.summary("suite-id")

						Ω(summary.State).Should(Equal(types.ExampleStateTimedOut))
						Ω(summary.Failure.Message).Should(Equal("Timed out"))
						Ω(summary.Failure.Location).Should(Equal(componentCodeLocation))
						Ω(summary.Failure.ForwardedPanic).Should(BeNil())
						Ω(summary.Failure.ComponentIndex).Should(Equal(2), "Should be the it node that failed")
						Ω(summary.Failure.ComponentType).Should(Equal(types.ExampleComponentTypeIt))
						Ω(summary.Failure.ComponentCodeLocation).Should(Equal(componentCodeLocation))

						Ω(ex.failed()).Should(BeTrue())
					})
				})
			})
		})

		Describe("running measurement examples and getting summaries", func() {
			var (
				runs                  int
				componentCodeLocation types.CodeLocation
				ex                    *example
			)

			BeforeEach(func() {
				runs = 0
				componentCodeLocation = types.GenerateCodeLocation(0)
			})

			It("should report that it has a measurement", func() {
				ex = newExample(newMeasureNode("measure", func(b Benchmarker) {}, flagTypeNone, componentCodeLocation, 1))
				Ω(ex.subjectComponentType()).Should(Equal(types.ExampleComponentTypeMeasure))
			})

			Context("when the measurement does not fail", func() {
				BeforeEach(func() {
					ex = newExample(newMeasureNode("measure", func(b Benchmarker) {
						b.RecordValue("foo", float64(runs))
						runs++
					}, flagTypeNone, componentCodeLocation, 5))
				})

				It("runs the measurement samples number of times and returns statistics", func() {
					ex.run()
					summary := ex.summary("suite-id")

					Ω(runs).Should(Equal(5))

					Ω(summary.State).Should(Equal(types.ExampleStatePassed))
					Ω(summary.IsMeasurement).Should(BeTrue())
					Ω(summary.NumberOfSamples).Should(Equal(5))
					Ω(summary.Measurements).Should(HaveLen(1))
					Ω(summary.Measurements["foo"].Name).Should(Equal("foo"))
					Ω(summary.Measurements["foo"].Results).Should(Equal([]float64{0, 1, 2, 3, 4}))
				})
			})

			Context("when one of the measurement samples fails", func() {
				BeforeEach(func() {
					ex = newExample(newMeasureNode("measure", func(b Benchmarker) {
						b.RecordValue("foo", float64(runs))
						runs++
						if runs == 3 {
							ex.fail(failureData{})
						}
					}, flagTypeNone, componentCodeLocation, 5))
				})

				It("marks the measurement as failed and doesn't run any more samples", func() {
					ex.run()
					summary := ex.summary("suite-id")

					Ω(runs).Should(Equal(3))

					Ω(summary.State).Should(Equal(types.ExampleStateFailed))
					Ω(summary.IsMeasurement).Should(BeTrue())
					Ω(summary.NumberOfSamples).Should(Equal(5))
					Ω(summary.Measurements).Should(BeEmpty())
				})
			})
		})

		Describe("running AfterEach nodes when other nodes fail", func() {
			var (
				orderedList []string
				ex          *example
			)

			newNode := func(identifier string, fail bool) *runnableNode {
				return newRunnableNode(func() {
					orderedList = append(orderedList, identifier)
					if fail {
						ex.fail(failureData{
							message: identifier + " failed",
						})
					}
				}, types.GenerateCodeLocation(0), 0)
			}

			newIt := func(identifier string, fail bool) *itNode {
				return newItNode(identifier, func() {
					orderedList = append(orderedList, identifier)
					if fail {
						ex.fail(failureData{})
					}
				}, flagTypeNone, types.GenerateCodeLocation(0), 0)
			}

			BeforeEach(func() {
				orderedList = make([]string, 0)
			})

			Context("when the it node fails", func() {
				BeforeEach(func() {
					ex = newExample(newIt("it", true))

					innerContainer := newContainerNode("inner", flagTypeNone, types.GenerateCodeLocation(0))
					innerContainer.pushBeforeEachNode(newNode("INNER_BEFORE_A", false))
					innerContainer.pushBeforeEachNode(newNode("INNER_BEFORE_B", false))
					innerContainer.pushJustBeforeEachNode(newNode("INNER_JUST_BEFORE_A", false))
					innerContainer.pushJustBeforeEachNode(newNode("INNER_JUST_BEFORE_B", false))
					innerContainer.pushAfterEachNode(newNode("INNER_AFTER_A", false))
					innerContainer.pushAfterEachNode(newNode("INNER_AFTER_B", false))

					ex.addContainerNode(innerContainer)

					outerContainer := newContainerNode("outer", flagTypeNone, types.GenerateCodeLocation(0))
					outerContainer.pushBeforeEachNode(newNode("OUTER_BEFORE_A", false))
					outerContainer.pushBeforeEachNode(newNode("OUTER_BEFORE_B", false))
					outerContainer.pushJustBeforeEachNode(newNode("OUTER_JUST_BEFORE_A", false))
					outerContainer.pushJustBeforeEachNode(newNode("OUTER_JUST_BEFORE_B", false))
					outerContainer.pushAfterEachNode(newNode("OUTER_AFTER_A", false))
					outerContainer.pushAfterEachNode(newNode("OUTER_AFTER_B", false))

					ex.addContainerNode(outerContainer)
					ex.run()
				})

				It("should run all the AfterEach nodes", func() {
					Ω(orderedList).Should(Equal([]string{
						"OUTER_BEFORE_A", "OUTER_BEFORE_B", "INNER_BEFORE_A", "INNER_BEFORE_B",
						"OUTER_JUST_BEFORE_A", "OUTER_JUST_BEFORE_B", "INNER_JUST_BEFORE_A", "INNER_JUST_BEFORE_B",
						"it",
						"INNER_AFTER_A", "INNER_AFTER_B", "OUTER_AFTER_A", "OUTER_AFTER_B",
					}))

				})
			})

			Context("when an inner BeforeEach node fails", func() {
				BeforeEach(func() {
					ex = newExample(newIt("it", true))

					innerContainer := newContainerNode("inner", flagTypeNone, types.GenerateCodeLocation(0))
					innerContainer.pushBeforeEachNode(newNode("INNER_BEFORE_A", true))
					innerContainer.pushBeforeEachNode(newNode("INNER_BEFORE_B", false))
					innerContainer.pushJustBeforeEachNode(newNode("INNER_JUST_BEFORE_A", false))
					innerContainer.pushJustBeforeEachNode(newNode("INNER_JUST_BEFORE_B", false))
					innerContainer.pushAfterEachNode(newNode("INNER_AFTER_A", false))
					innerContainer.pushAfterEachNode(newNode("INNER_AFTER_B", false))

					ex.addContainerNode(innerContainer)

					outerContainer := newContainerNode("outer", flagTypeNone, types.GenerateCodeLocation(0))
					outerContainer.pushBeforeEachNode(newNode("OUTER_BEFORE_A", false))
					outerContainer.pushBeforeEachNode(newNode("OUTER_BEFORE_B", false))
					outerContainer.pushJustBeforeEachNode(newNode("OUTER_JUST_BEFORE_A", false))
					outerContainer.pushJustBeforeEachNode(newNode("OUTER_JUST_BEFORE_B", false))
					outerContainer.pushAfterEachNode(newNode("OUTER_AFTER_A", false))
					outerContainer.pushAfterEachNode(newNode("OUTER_AFTER_B", false))

					ex.addContainerNode(outerContainer)
					ex.run()
				})

				It("should run all the AfterEach nodes at nesting levels equal to or lower than the failed BeforeEach block", func() {
					Ω(orderedList).Should(Equal([]string{
						"OUTER_BEFORE_A", "OUTER_BEFORE_B", "INNER_BEFORE_A",
						"INNER_AFTER_A", "INNER_AFTER_B", "OUTER_AFTER_A", "OUTER_AFTER_B",
					}))
				})
			})

			Context("when an outer BeforeEach node fails", func() {
				BeforeEach(func() {
					ex = newExample(newIt("it", true))

					innerContainer := newContainerNode("inner", flagTypeNone, types.GenerateCodeLocation(0))
					innerContainer.pushBeforeEachNode(newNode("INNER_BEFORE_A", false))
					innerContainer.pushBeforeEachNode(newNode("INNER_BEFORE_B", false))
					innerContainer.pushJustBeforeEachNode(newNode("INNER_JUST_BEFORE_A", false))
					innerContainer.pushJustBeforeEachNode(newNode("INNER_JUST_BEFORE_B", false))
					innerContainer.pushAfterEachNode(newNode("INNER_AFTER_A", false))
					innerContainer.pushAfterEachNode(newNode("INNER_AFTER_B", false))

					ex.addContainerNode(innerContainer)

					outerContainer := newContainerNode("outer", flagTypeNone, types.GenerateCodeLocation(0))
					outerContainer.pushBeforeEachNode(newNode("OUTER_BEFORE_A", false))
					outerContainer.pushBeforeEachNode(newNode("OUTER_BEFORE_B", true))
					outerContainer.pushJustBeforeEachNode(newNode("OUTER_JUST_BEFORE_A", false))
					outerContainer.pushJustBeforeEachNode(newNode("OUTER_JUST_BEFORE_B", false))
					outerContainer.pushAfterEachNode(newNode("OUTER_AFTER_A", false))
					outerContainer.pushAfterEachNode(newNode("OUTER_AFTER_B", false))

					ex.addContainerNode(outerContainer)
					ex.run()
				})

				It("should run all the AfterEach nodes at nesting levels equal to or lower than the failed BeforeEach block", func() {
					Ω(orderedList).Should(Equal([]string{
						"OUTER_BEFORE_A", "OUTER_BEFORE_B",
						"OUTER_AFTER_A", "OUTER_AFTER_B",
					}))
				})
			})

			Context("when a JustBeforeEach node fails", func() {
				BeforeEach(func() {
					ex = newExample(newIt("it", true))

					innerContainer := newContainerNode("inner", flagTypeNone, types.GenerateCodeLocation(0))
					innerContainer.pushBeforeEachNode(newNode("INNER_BEFORE_A", false))
					innerContainer.pushBeforeEachNode(newNode("INNER_BEFORE_B", false))
					innerContainer.pushJustBeforeEachNode(newNode("INNER_JUST_BEFORE_A", false))
					innerContainer.pushJustBeforeEachNode(newNode("INNER_JUST_BEFORE_B", false))
					innerContainer.pushAfterEachNode(newNode("INNER_AFTER_A", false))
					innerContainer.pushAfterEachNode(newNode("INNER_AFTER_B", false))

					ex.addContainerNode(innerContainer)

					outerContainer := newContainerNode("outer", flagTypeNone, types.GenerateCodeLocation(0))
					outerContainer.pushBeforeEachNode(newNode("OUTER_BEFORE_A", false))
					outerContainer.pushBeforeEachNode(newNode("OUTER_BEFORE_B", false))
					outerContainer.pushJustBeforeEachNode(newNode("OUTER_JUST_BEFORE_A", true))
					outerContainer.pushJustBeforeEachNode(newNode("OUTER_JUST_BEFORE_B", false))
					outerContainer.pushAfterEachNode(newNode("OUTER_AFTER_A", false))
					outerContainer.pushAfterEachNode(newNode("OUTER_AFTER_B", false))

					ex.addContainerNode(outerContainer)
					ex.run()
				})

				It("should run all the AfterEach nodes", func() {
					Ω(orderedList).Should(Equal([]string{
						"OUTER_BEFORE_A", "OUTER_BEFORE_B", "INNER_BEFORE_A", "INNER_BEFORE_B",
						"OUTER_JUST_BEFORE_A",
						"INNER_AFTER_A", "INNER_AFTER_B", "OUTER_AFTER_A", "OUTER_AFTER_B",
					}))
				})
			})

			Context("when an AfterEach node fails", func() {
				BeforeEach(func() {
					ex = newExample(newIt("it", true))

					innerContainer := newContainerNode("inner", flagTypeNone, types.GenerateCodeLocation(0))
					innerContainer.pushBeforeEachNode(newNode("INNER_BEFORE_A", false))
					innerContainer.pushBeforeEachNode(newNode("INNER_BEFORE_B", false))
					innerContainer.pushJustBeforeEachNode(newNode("INNER_JUST_BEFORE_A", true))
					innerContainer.pushJustBeforeEachNode(newNode("INNER_JUST_BEFORE_B", false))
					innerContainer.pushAfterEachNode(newNode("INNER_AFTER_A", false))
					innerContainer.pushAfterEachNode(newNode("INNER_AFTER_B", true))

					ex.addContainerNode(innerContainer)

					outerContainer := newContainerNode("outer", flagTypeNone, types.GenerateCodeLocation(0))
					outerContainer.pushBeforeEachNode(newNode("OUTER_BEFORE_A", false))
					outerContainer.pushBeforeEachNode(newNode("OUTER_BEFORE_B", false))
					outerContainer.pushJustBeforeEachNode(newNode("OUTER_JUST_BEFORE_A", false))
					outerContainer.pushJustBeforeEachNode(newNode("OUTER_JUST_BEFORE_B", false))
					outerContainer.pushAfterEachNode(newNode("OUTER_AFTER_A", true))
					outerContainer.pushAfterEachNode(newNode("OUTER_AFTER_B", false))

					ex.addContainerNode(outerContainer)
					ex.run()
				})

				It("should nonetheless continue to run subsequent after each nodes", func() {
					Ω(orderedList).Should(Equal([]string{
						"OUTER_BEFORE_A", "OUTER_BEFORE_B", "INNER_BEFORE_A", "INNER_BEFORE_B",
						"OUTER_JUST_BEFORE_A", "OUTER_JUST_BEFORE_B", "INNER_JUST_BEFORE_A",
						"INNER_AFTER_A", "INNER_AFTER_B", "OUTER_AFTER_A", "OUTER_AFTER_B",
					}))
				})

				It("should not override the failure data of the earliest failure", func() {
					Ω(ex.summary("suite-id").Failure.Message).Should(Equal("INNER_JUST_BEFORE_A failed"))
				})
			})
		})
	})
}
