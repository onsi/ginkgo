package ginkgo

import (
	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/reporters"
	"github.com/onsi/ginkgo/types"
	. "github.com/onsi/gomega"

	"math/rand"
	"sort"
	"time"
)

func init() {
	Describe("Example Collection", func() {
		var (
			fakeT *fakeTestingT
			fakeR *reporters.FakeReporter

			examplesThatWereRun []string

			collection *exampleCollection
		)

		exampleWithItFunc := func(itText string, flag flagType, fail bool) *example {
			return newExample(newItNode(itText, func() {
				examplesThatWereRun = append(examplesThatWereRun, itText)
				time.Sleep(time.Duration(0.001 * float64(time.Second)))
				if fail {
					collection.fail(failureData{
						message: itText + " Failed",
					})
				}
			}, flag, types.GenerateCodeLocation(0), 0))
		}

		BeforeEach(func() {
			fakeT = &fakeTestingT{}
			fakeR = reporters.NewFakeReporter()
			examplesThatWereRun = make([]string, 0)
		})

		Describe("enumerating and assigning example indices", func() {
			var examples []*example
			BeforeEach(func() {
				examples = []*example{
					exampleWithItFunc("C", flagTypeNone, false),
					exampleWithItFunc("A", flagTypeNone, false),
					exampleWithItFunc("B", flagTypeNone, false),
				}
				collection = newExampleCollection(fakeT, "collection description", examples, []Reporter{fakeR}, config.GinkgoConfigType{})
			})

			It("should enumerate and assign example indices", func() {
				Ω(examples[0].summary("suite-id").ExampleIndex).Should(Equal(0))
				Ω(examples[1].summary("suite-id").ExampleIndex).Should(Equal(1))
				Ω(examples[2].summary("suite-id").ExampleIndex).Should(Equal(2))
			})
		})

		Describe("shuffling the collection", func() {
			BeforeEach(func() {
				collection = newExampleCollection(fakeT, "collection description", []*example{
					exampleWithItFunc("C", flagTypeNone, false),
					exampleWithItFunc("A", flagTypeNone, false),
					exampleWithItFunc("B", flagTypeNone, false),
				}, []Reporter{fakeR}, config.GinkgoConfigType{})
			})

			It("should be sortable", func() {
				sort.Sort(collection)
				collection.run()
				Ω(examplesThatWereRun).Should(Equal([]string{"A", "B", "C"}))
			})

			It("shuffles all the examples after sorting them", func() {
				collection.shuffle(rand.New(rand.NewSource(17)))
				collection.run()
				Ω(examplesThatWereRun).Should(Equal(shuffleStrings([]string{"A", "B", "C"}, 17)), "The permutation should be the same across test runs")
			})
		})

		Describe("reporting to multiple reporter", func() {
			var otherFakeR *reporters.FakeReporter
			BeforeEach(func() {
				otherFakeR = reporters.NewFakeReporter()

				collection = newExampleCollection(fakeT, "collection description", []*example{
					exampleWithItFunc("C", flagTypeNone, false),
					exampleWithItFunc("A", flagTypeNone, false),
					exampleWithItFunc("B", flagTypeNone, false),
				}, []Reporter{fakeR, otherFakeR}, config.GinkgoConfigType{})
				collection.run()
			})

			It("reports to both reporters", func() {
				Ω(otherFakeR.BeginSummary).Should(Equal(fakeR.BeginSummary))
				Ω(otherFakeR.EndSummary).Should(Equal(fakeR.EndSummary))
				Ω(otherFakeR.ExampleSummaries).Should(Equal(fakeR.ExampleSummaries))
			})
		})

		Describe("running an example collection", func() {
			var (
				example1  *example
				example2  *example
				example3  *example
				conf      config.GinkgoConfigType
				runResult bool
			)

			BeforeEach(func() {
				conf = config.GinkgoConfigType{FocusString: "", ParallelTotal: 1, ParallelNode: 1}

				example1 = exampleWithItFunc("it 1", flagTypeNone, false)
				example2 = exampleWithItFunc("it 2", flagTypeNone, false)
				example3 = exampleWithItFunc("it 3", flagTypeNone, false)
			})

			JustBeforeEach(func() {
				collection = newExampleCollection(fakeT, "collection description", []*example{example1, example2, example3}, []Reporter{fakeR}, conf)
				runResult = collection.run()
			})

			Context("when all the examples pass", func() {
				It("should return true", func() {
					Ω(runResult).Should(BeTrue())
				})

				It("runs all the tests", func() {
					Ω(examplesThatWereRun).Should(Equal([]string{"it 1", "it 2", "it 3"}))
				})

				It("marks the suite as passed", func() {
					Ω(fakeT.didFail).Should(BeFalse())
				})

				It("publishes the correct starting suite summary", func() {
					summary := fakeR.BeginSummary
					Ω(summary.SuiteDescription).Should(Equal("collection description"))
					Ω(summary.SuiteSucceeded).Should(BeTrue())
					Ω(summary.NumberOfExamplesBeforeParallelization).Should(Equal(3))
					Ω(summary.NumberOfTotalExamples).Should(Equal(3))
					Ω(summary.NumberOfExamplesThatWillBeRun).Should(Equal(3))
					Ω(summary.NumberOfPendingExamples).Should(Equal(0))
					Ω(summary.NumberOfSkippedExamples).Should(Equal(0))
					Ω(summary.NumberOfPassedExamples).Should(Equal(0))
					Ω(summary.NumberOfFailedExamples).Should(Equal(0))
					Ω(summary.RunTime).Should(Equal(time.Duration(0)))
				})

				It("publishes the correct example summaries", func() {
					Ω(fakeR.ExampleWillRunSummaries).Should(HaveLen(3))
					Ω(fakeR.ExampleSummaries).Should(HaveLen(3))
					Ω(fakeR.ExampleSummaries[0]).Should(Equal(example1.summary(fakeR.BeginSummary.SuiteID)))
					Ω(fakeR.ExampleSummaries[1]).Should(Equal(example2.summary(fakeR.BeginSummary.SuiteID)))
					Ω(fakeR.ExampleSummaries[2]).Should(Equal(example3.summary(fakeR.BeginSummary.SuiteID)))
				})

				It("publishes the correct ending suite summary", func() {
					summary := fakeR.EndSummary
					Ω(summary.SuiteDescription).Should(Equal("collection description"))
					Ω(summary.SuiteSucceeded).Should(BeTrue())
					Ω(summary.NumberOfExamplesBeforeParallelization).Should(Equal(3))
					Ω(summary.NumberOfTotalExamples).Should(Equal(3))
					Ω(summary.NumberOfExamplesThatWillBeRun).Should(Equal(3))
					Ω(summary.NumberOfPendingExamples).Should(Equal(0))
					Ω(summary.NumberOfSkippedExamples).Should(Equal(0))
					Ω(summary.NumberOfPassedExamples).Should(Equal(3))
					Ω(summary.NumberOfFailedExamples).Should(Equal(0))
					Ω(summary.RunTime.Seconds()).Should(BeNumerically("~", 3*0.001, 0.01))
				})

				It("should publish a consistent suite ID across all summaries", func() {
					suiteId := fakeR.BeginSummary.SuiteID
					Ω(suiteId).ShouldNot(BeEmpty())
					Ω(fakeR.EndSummary.SuiteID).Should(Equal(suiteId))
					for _, exampleSummary := range fakeR.ExampleSummaries {
						Ω(exampleSummary.SuiteID).Should(Equal(suiteId))
					}
				})
			})

			Context("when examples fail", func() {
				BeforeEach(func() {
					example2 = exampleWithItFunc("failing it 2", flagTypeNone, true)
					example3 = exampleWithItFunc("failing it 3", flagTypeNone, true)
				})

				It("should return false", func() {
					Ω(runResult).Should(BeFalse())
				})

				It("runs all the tests", func() {
					Ω(examplesThatWereRun).Should(Equal([]string{"it 1", "failing it 2", "failing it 3"}))
				})

				It("marks the suite as failed", func() {
					Ω(fakeT.didFail).Should(BeTrue())
				})

				It("publishes the correct starting suite summary", func() {
					summary := fakeR.BeginSummary
					Ω(summary.SuiteDescription).Should(Equal("collection description"))
					Ω(summary.SuiteSucceeded).Should(BeTrue())
					Ω(summary.NumberOfExamplesBeforeParallelization).Should(Equal(3))
					Ω(summary.NumberOfTotalExamples).Should(Equal(3))
					Ω(summary.NumberOfExamplesThatWillBeRun).Should(Equal(3))
					Ω(summary.NumberOfPendingExamples).Should(Equal(0))
					Ω(summary.NumberOfSkippedExamples).Should(Equal(0))
					Ω(summary.NumberOfPassedExamples).Should(Equal(0))
					Ω(summary.NumberOfFailedExamples).Should(Equal(0))
					Ω(summary.RunTime).Should(Equal(time.Duration(0)))
				})

				It("publishes the correct example summaries", func() {
					Ω(fakeR.ExampleSummaries).Should(HaveLen(3))
					Ω(fakeR.ExampleSummaries[0]).Should(Equal(example1.summary(fakeR.BeginSummary.SuiteID)))
					Ω(fakeR.ExampleSummaries[1]).Should(Equal(example2.summary(fakeR.BeginSummary.SuiteID)))
					Ω(fakeR.ExampleSummaries[2]).Should(Equal(example3.summary(fakeR.BeginSummary.SuiteID)))
				})

				It("publishes the correct ending suite summary", func() {
					summary := fakeR.EndSummary
					Ω(summary.SuiteDescription).Should(Equal("collection description"))
					Ω(summary.SuiteSucceeded).Should(BeFalse())
					Ω(summary.NumberOfExamplesBeforeParallelization).Should(Equal(3))
					Ω(summary.NumberOfTotalExamples).Should(Equal(3))
					Ω(summary.NumberOfExamplesThatWillBeRun).Should(Equal(3))
					Ω(summary.NumberOfPendingExamples).Should(Equal(0))
					Ω(summary.NumberOfSkippedExamples).Should(Equal(0))
					Ω(summary.NumberOfPassedExamples).Should(Equal(1))
					Ω(summary.NumberOfFailedExamples).Should(Equal(2))
					Ω(summary.RunTime.Seconds()).Should(BeNumerically("~", 3*0.001, 0.01))
				})
			})

			Context("when examples are pending", func() {
				BeforeEach(func() {
					example1 = exampleWithItFunc("pending it 1", flagTypePending, false)
				})

				It("skips the pending examples", func() {
					Ω(examplesThatWereRun).Should(Equal([]string{"it 2", "it 3"}))
				})

				It("marks the suite as passed", func() {
					Ω(fakeT.didFail).Should(BeFalse())
				})

				It("publishes the correct starting suite summary", func() {
					summary := fakeR.BeginSummary
					Ω(summary.SuiteDescription).Should(Equal("collection description"))
					Ω(summary.SuiteSucceeded).Should(BeTrue())
					Ω(summary.NumberOfExamplesBeforeParallelization).Should(Equal(3))
					Ω(summary.NumberOfTotalExamples).Should(Equal(3))
					Ω(summary.NumberOfExamplesThatWillBeRun).Should(Equal(2))
					Ω(summary.NumberOfPendingExamples).Should(Equal(1))
					Ω(summary.NumberOfSkippedExamples).Should(Equal(0))
					Ω(summary.NumberOfPassedExamples).Should(Equal(0))
					Ω(summary.NumberOfFailedExamples).Should(Equal(0))
					Ω(summary.RunTime).Should(Equal(time.Duration(0)))
				})

				It("publishes the correct example summaries", func() {
					Ω(fakeR.ExampleSummaries).Should(HaveLen(3))
					Ω(fakeR.ExampleSummaries[0]).Should(Equal(example1.summary(fakeR.BeginSummary.SuiteID)))
					Ω(fakeR.ExampleSummaries[1]).Should(Equal(example2.summary(fakeR.BeginSummary.SuiteID)))
					Ω(fakeR.ExampleSummaries[2]).Should(Equal(example3.summary(fakeR.BeginSummary.SuiteID)))
				})

				It("publishes the correct ending suite summary", func() {
					summary := fakeR.EndSummary
					Ω(summary.SuiteDescription).Should(Equal("collection description"))
					Ω(summary.SuiteSucceeded).Should(BeTrue())
					Ω(summary.NumberOfExamplesBeforeParallelization).Should(Equal(3))
					Ω(summary.NumberOfTotalExamples).Should(Equal(3))
					Ω(summary.NumberOfExamplesThatWillBeRun).Should(Equal(2))
					Ω(summary.NumberOfPendingExamples).Should(Equal(1))
					Ω(summary.NumberOfSkippedExamples).Should(Equal(0))
					Ω(summary.NumberOfPassedExamples).Should(Equal(2))
					Ω(summary.NumberOfFailedExamples).Should(Equal(0))
					Ω(summary.RunTime.Seconds()).Should(BeNumerically("~", 3*0.001, 0.01))
				})

				Context("and --failOnPending is set", func() {
					BeforeEach(func() {
						conf.FailOnPending = true
					})

					It("should mark the suite as failed", func() {
						Ω(fakeT.didFail).Should(BeTrue())
						summary := fakeR.EndSummary
						Ω(summary.SuiteSucceeded).Should(BeFalse())
					})
				})
			})

			Context("when examples are focused", func() {
				BeforeEach(func() {
					example1 = exampleWithItFunc("focused it 1", flagTypeFocused, false)
					example3 = exampleWithItFunc("focused it 3", flagTypeFocused, false)
				})

				It("skips the non-focused examples", func() {
					Ω(examplesThatWereRun).Should(Equal([]string{"focused it 1", "focused it 3"}))
				})

				It("marks the suite as passed", func() {
					Ω(fakeT.didFail).Should(BeFalse())
				})

				It("publishes the correct starting suite summary", func() {
					summary := fakeR.BeginSummary
					Ω(summary.SuiteDescription).Should(Equal("collection description"))
					Ω(summary.SuiteSucceeded).Should(BeTrue())
					Ω(summary.NumberOfExamplesBeforeParallelization).Should(Equal(3))
					Ω(summary.NumberOfTotalExamples).Should(Equal(3))
					Ω(summary.NumberOfExamplesThatWillBeRun).Should(Equal(2))
					Ω(summary.NumberOfPendingExamples).Should(Equal(0))
					Ω(summary.NumberOfSkippedExamples).Should(Equal(1))
					Ω(summary.NumberOfPassedExamples).Should(Equal(0))
					Ω(summary.NumberOfFailedExamples).Should(Equal(0))
					Ω(summary.RunTime).Should(Equal(time.Duration(0)))
				})

				It("publishes the correct example summaries", func() {
					Ω(fakeR.ExampleSummaries).Should(HaveLen(3))
					Ω(fakeR.ExampleSummaries[0]).Should(Equal(example1.summary(fakeR.BeginSummary.SuiteID)))
					Ω(fakeR.ExampleSummaries[1]).Should(Equal(example2.summary(fakeR.BeginSummary.SuiteID)))
					Ω(fakeR.ExampleSummaries[2]).Should(Equal(example3.summary(fakeR.BeginSummary.SuiteID)))
				})

				It("publishes the correct ending suite summary", func() {
					summary := fakeR.EndSummary
					Ω(summary.SuiteDescription).Should(Equal("collection description"))
					Ω(summary.SuiteSucceeded).Should(BeTrue())
					Ω(summary.NumberOfExamplesBeforeParallelization).Should(Equal(3))
					Ω(summary.NumberOfTotalExamples).Should(Equal(3))
					Ω(summary.NumberOfExamplesThatWillBeRun).Should(Equal(2))
					Ω(summary.NumberOfPendingExamples).Should(Equal(0))
					Ω(summary.NumberOfSkippedExamples).Should(Equal(1))
					Ω(summary.NumberOfPassedExamples).Should(Equal(2))
					Ω(summary.NumberOfFailedExamples).Should(Equal(0))
					Ω(summary.RunTime.Seconds()).Should(BeNumerically("~", 3*0.001, 0.01))
				})
			})

			Context("when a regexp focusString is provided", func() {
				BeforeEach(func() {
					conf.FocusString = `collection description.*pickles\d$`
					example1 = exampleWithItFunc("focused it 1", flagTypeFocused, false)
					example2 = exampleWithItFunc("another it pickles2", flagTypeNone, false)
					example3 = exampleWithItFunc("focused it pickles3", flagTypeFocused, false)
				})

				It("ignores the programmatic focus and applies the regexp focusString", func() {
					Ω(examplesThatWereRun).Should(Equal([]string{"another it pickles2", "focused it pickles3"}))
				})

				It("publishes the correct starting suite summary", func() {
					summary := fakeR.BeginSummary
					Ω(summary.SuiteDescription).Should(Equal("collection description"))
					Ω(summary.SuiteSucceeded).Should(BeTrue())
					Ω(summary.NumberOfExamplesBeforeParallelization).Should(Equal(3))
					Ω(summary.NumberOfTotalExamples).Should(Equal(3))
					Ω(summary.NumberOfExamplesThatWillBeRun).Should(Equal(2))
					Ω(summary.NumberOfPendingExamples).Should(Equal(0))
					Ω(summary.NumberOfSkippedExamples).Should(Equal(1))
					Ω(summary.NumberOfPassedExamples).Should(Equal(0))
					Ω(summary.NumberOfFailedExamples).Should(Equal(0))
					Ω(summary.RunTime).Should(Equal(time.Duration(0)))
				})

				It("publishes the correct example summaries", func() {
					Ω(fakeR.ExampleSummaries).Should(HaveLen(3))
					Ω(fakeR.ExampleSummaries[0]).Should(Equal(example1.summary(fakeR.BeginSummary.SuiteID)))
					Ω(fakeR.ExampleSummaries[1]).Should(Equal(example2.summary(fakeR.BeginSummary.SuiteID)))
					Ω(fakeR.ExampleSummaries[2]).Should(Equal(example3.summary(fakeR.BeginSummary.SuiteID)))
				})

				It("publishes the correct ending suite summary", func() {
					summary := fakeR.EndSummary
					Ω(summary.SuiteDescription).Should(Equal("collection description"))
					Ω(summary.SuiteSucceeded).Should(BeTrue())
					Ω(summary.NumberOfExamplesBeforeParallelization).Should(Equal(3))
					Ω(summary.NumberOfTotalExamples).Should(Equal(3))
					Ω(summary.NumberOfExamplesThatWillBeRun).Should(Equal(2))
					Ω(summary.NumberOfPendingExamples).Should(Equal(0))
					Ω(summary.NumberOfSkippedExamples).Should(Equal(1))
					Ω(summary.NumberOfPassedExamples).Should(Equal(2))
					Ω(summary.NumberOfFailedExamples).Should(Equal(0))
					Ω(summary.RunTime.Seconds()).Should(BeNumerically("~", 3*0.001, 0.01))
				})
			})

			Context("when a regexp skipString is provided", func() {
				BeforeEach(func() {
					conf.SkipString = `collection description.*pickles\d$`
					example1 = exampleWithItFunc("focused it 1", flagTypeFocused, false)
					example2 = exampleWithItFunc("another it pickles2", flagTypeNone, false)
					example3 = exampleWithItFunc("focused it pickles3", flagTypeFocused, false)
				})

				It("ignores the programmatic focus and applies the regexp skipString", func() {
					Ω(examplesThatWereRun).Should(Equal([]string{"focused it 1"}))
				})

				It("publishes the correct starting suite summary", func() {
					summary := fakeR.BeginSummary
					Ω(summary.SuiteDescription).Should(Equal("collection description"))
					Ω(summary.SuiteSucceeded).Should(BeTrue())
					Ω(summary.NumberOfExamplesBeforeParallelization).Should(Equal(3))
					Ω(summary.NumberOfTotalExamples).Should(Equal(3))
					Ω(summary.NumberOfExamplesThatWillBeRun).Should(Equal(1))
					Ω(summary.NumberOfPendingExamples).Should(Equal(0))
					Ω(summary.NumberOfSkippedExamples).Should(Equal(2))
					Ω(summary.NumberOfPassedExamples).Should(Equal(0))
					Ω(summary.NumberOfFailedExamples).Should(Equal(0))
					Ω(summary.RunTime).Should(Equal(time.Duration(0)))
				})

				It("publishes the correct example summaries", func() {
					Ω(fakeR.ExampleSummaries).Should(HaveLen(3))
					Ω(fakeR.ExampleSummaries[0]).Should(Equal(example1.summary(fakeR.BeginSummary.SuiteID)))
					Ω(fakeR.ExampleSummaries[1]).Should(Equal(example2.summary(fakeR.BeginSummary.SuiteID)))
					Ω(fakeR.ExampleSummaries[2]).Should(Equal(example3.summary(fakeR.BeginSummary.SuiteID)))
				})

				It("publishes the correct ending suite summary", func() {
					summary := fakeR.EndSummary
					Ω(summary.SuiteDescription).Should(Equal("collection description"))
					Ω(summary.SuiteSucceeded).Should(BeTrue())
					Ω(summary.NumberOfExamplesBeforeParallelization).Should(Equal(3))
					Ω(summary.NumberOfTotalExamples).Should(Equal(3))
					Ω(summary.NumberOfExamplesThatWillBeRun).Should(Equal(1))
					Ω(summary.NumberOfPendingExamples).Should(Equal(0))
					Ω(summary.NumberOfSkippedExamples).Should(Equal(2))
					Ω(summary.NumberOfPassedExamples).Should(Equal(1))
					Ω(summary.NumberOfFailedExamples).Should(Equal(0))
					Ω(summary.RunTime.Seconds()).Should(BeNumerically("~", 3*0.001, 0.01))
				})
			})

			Context("when both a regexp skipString and focusString are provided", func() {
				BeforeEach(func() {
					conf.SkipString = `collection description.*2`
					conf.FocusString = `collection description.*A`
					example1 = exampleWithItFunc("A1", flagTypeFocused, false)
					example2 = exampleWithItFunc("A2", flagTypeNone, false)
					example3 = exampleWithItFunc("B1", flagTypeFocused, false)
				})

				It("ignores the programmatic focus and ANDs the focusString and skipString", func() {
					Ω(examplesThatWereRun).Should(Equal([]string{"A1"}))
				})

				It("publishes the correct starting suite summary", func() {
					summary := fakeR.BeginSummary
					Ω(summary.SuiteDescription).Should(Equal("collection description"))
					Ω(summary.SuiteSucceeded).Should(BeTrue())
					Ω(summary.NumberOfExamplesBeforeParallelization).Should(Equal(3))
					Ω(summary.NumberOfTotalExamples).Should(Equal(3))
					Ω(summary.NumberOfExamplesThatWillBeRun).Should(Equal(1))
					Ω(summary.NumberOfPendingExamples).Should(Equal(0))
					Ω(summary.NumberOfSkippedExamples).Should(Equal(2))
					Ω(summary.NumberOfPassedExamples).Should(Equal(0))
					Ω(summary.NumberOfFailedExamples).Should(Equal(0))
					Ω(summary.RunTime).Should(Equal(time.Duration(0)))
				})

				It("publishes the correct example summaries", func() {
					Ω(fakeR.ExampleSummaries).Should(HaveLen(3))
					Ω(fakeR.ExampleSummaries[0]).Should(Equal(example1.summary(fakeR.BeginSummary.SuiteID)))
					Ω(fakeR.ExampleSummaries[1]).Should(Equal(example2.summary(fakeR.BeginSummary.SuiteID)))
					Ω(fakeR.ExampleSummaries[2]).Should(Equal(example3.summary(fakeR.BeginSummary.SuiteID)))
				})

				It("publishes the correct ending suite summary", func() {
					summary := fakeR.EndSummary
					Ω(summary.SuiteDescription).Should(Equal("collection description"))
					Ω(summary.SuiteSucceeded).Should(BeTrue())
					Ω(summary.NumberOfExamplesBeforeParallelization).Should(Equal(3))
					Ω(summary.NumberOfTotalExamples).Should(Equal(3))
					Ω(summary.NumberOfExamplesThatWillBeRun).Should(Equal(1))
					Ω(summary.NumberOfPendingExamples).Should(Equal(0))
					Ω(summary.NumberOfSkippedExamples).Should(Equal(2))
					Ω(summary.NumberOfPassedExamples).Should(Equal(1))
					Ω(summary.NumberOfFailedExamples).Should(Equal(0))
					Ω(summary.RunTime.Seconds()).Should(BeNumerically("~", 3*0.001, 0.01))
				})
			})

			Context("when a examples are run in parallel", func() {
				BeforeEach(func() {
					conf.ParallelTotal = 2
					conf.ParallelNode = 2
				})

				It("trims the example set before running them", func() {
					Ω(examplesThatWereRun).Should(Equal([]string{"it 2", "it 3"}))
				})

				It("publishes the correct starting suite summary", func() {
					summary := fakeR.BeginSummary
					Ω(summary.SuiteDescription).Should(Equal("collection description"))
					Ω(summary.SuiteSucceeded).Should(BeTrue())
					Ω(summary.NumberOfExamplesBeforeParallelization).Should(Equal(3))
					Ω(summary.NumberOfTotalExamples).Should(Equal(2))
					Ω(summary.NumberOfExamplesThatWillBeRun).Should(Equal(2))
					Ω(summary.NumberOfPendingExamples).Should(Equal(0))
					Ω(summary.NumberOfSkippedExamples).Should(Equal(0))
					Ω(summary.NumberOfPassedExamples).Should(Equal(0))
					Ω(summary.NumberOfFailedExamples).Should(Equal(0))
					Ω(summary.RunTime).Should(Equal(time.Duration(0)))
				})

				It("publishes the correct example summaries", func() {
					Ω(fakeR.ExampleSummaries).Should(HaveLen(2))
					Ω(fakeR.ExampleSummaries[0]).Should(Equal(example2.summary(fakeR.BeginSummary.SuiteID)))
					Ω(fakeR.ExampleSummaries[1]).Should(Equal(example3.summary(fakeR.BeginSummary.SuiteID)))
				})

				It("publishes the correct ending suite summary", func() {
					summary := fakeR.EndSummary
					Ω(summary.SuiteDescription).Should(Equal("collection description"))
					Ω(summary.SuiteSucceeded).Should(BeTrue())
					Ω(summary.NumberOfExamplesBeforeParallelization).Should(Equal(3))
					Ω(summary.NumberOfTotalExamples).Should(Equal(2))
					Ω(summary.NumberOfExamplesThatWillBeRun).Should(Equal(2))
					Ω(summary.NumberOfPendingExamples).Should(Equal(0))
					Ω(summary.NumberOfSkippedExamples).Should(Equal(0))
					Ω(summary.NumberOfPassedExamples).Should(Equal(2))
					Ω(summary.NumberOfFailedExamples).Should(Equal(0))
					Ω(summary.RunTime.Seconds()).Should(BeNumerically("~", 3*0.001, 0.01))
				})
			})
		})

		Describe("measurements", func() {
			exampleWithMeasure := func(text string) *example {
				return newExample(newMeasureNode(text, func(b Benchmarker) {
					examplesThatWereRun = append(examplesThatWereRun, text)
				}, flagTypeNone, types.GenerateCodeLocation(0), 1))
			}

			var conf config.GinkgoConfigType

			BeforeEach(func() {
				conf = config.GinkgoConfigType{}
			})

			JustBeforeEach(func() {
				collection = newExampleCollection(fakeT, "collection description", []*example{
					exampleWithItFunc("C", flagTypeNone, false),
					exampleWithItFunc("A", flagTypeNone, false),
					exampleWithItFunc("B", flagTypeNone, false),
					exampleWithMeasure("measure"),
				}, []Reporter{fakeR}, conf)

				collection.run()
			})

			It("runs the measurement", func() {
				Ω(examplesThatWereRun).Should(ContainElement("A"))
				Ω(examplesThatWereRun).Should(ContainElement("measure"))
			})

			Context("when instructed to skip measurements", func() {
				BeforeEach(func() {
					conf = config.GinkgoConfigType{
						SkipMeasurements: true,
					}
				})

				It("skips the measurements", func() {
					Ω(examplesThatWereRun).Should(ContainElement("A"))
					Ω(examplesThatWereRun).ShouldNot(ContainElement("measure"))
				})
			})
		})
	})
}
