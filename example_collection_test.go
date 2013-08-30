package ginkgo

import (
	. "github.com/onsi/gomega"
	"math/rand"
	"time"
)

func init() {
	Describe("Example Collection", func() {
		var (
			fakeT    *fakeTestingT
			fakeR    *fakeReporter
			example1 *example
			example2 *example
			example3 *example

			itsThatWereRun []string

			collection *exampleCollection
		)

		exampleWithItFunc := func(itText string, flag flagType, fail bool) *example {
			return newExample(newItNode(itText, func() {
				itsThatWereRun = append(itsThatWereRun, itText)
				time.Sleep(time.Duration(0.001 * float64(time.Second)))
				if fail {
					collection.fail(failureData{
						message: itText + " Failed",
					})
				}
			}, flag, generateCodeLocation(0), 0))
		}

		BeforeEach(func() {
			fakeT = &fakeTestingT{}
			fakeR = &fakeReporter{}
			itsThatWereRun = make([]string, 0)

			example1 = exampleWithItFunc("it 1", flagTypeNone, false)
			example2 = exampleWithItFunc("it 2", flagTypeNone, false)
			example3 = exampleWithItFunc("it 3", flagTypeNone, false)
		})

		Describe("shuffling the collection", func() {
			It("shuffles all the examples", func() {
				collection = newExampleCollection(fakeT, "collection description", []*example{example1, example2, example3}, fakeR)
				collection.shuffle(rand.New(rand.NewSource(10)))
				collection.run()
				Ω(itsThatWereRun).Should(Equal([]string{"it 2", "it 3", "it 1"}), "The permutation should be the same across test runs")
			})
		})

		Describe("running an example collection", func() {
			JustBeforeEach(func() {
				collection = newExampleCollection(fakeT, "collection description", []*example{example1, example2, example3}, fakeR)
				collection.run()
			})

			Context("when all the examples pass", func() {
				It("runs all the tests", func() {
					Ω(itsThatWereRun).Should(Equal([]string{"it 1", "it 2", "it 3"}))
				})

				It("marks the suite as passed", func() {
					Ω(fakeT.didFail).Should(BeFalse())
				})

				It("publishes the correct starting suite summary", func() {
					summary := fakeR.beginSummary
					Ω(summary.SuiteDescription).Should(Equal("collection description"))
					Ω(summary.NumberOfTotalExamples).Should(Equal(3))
					Ω(summary.NumberOfExamplesThatWillBeRun).Should(Equal(3))
					Ω(summary.NumberOfPendingExamples).Should(Equal(0))
					Ω(summary.NumberOfSkippedExamples).Should(Equal(0))
					Ω(summary.NumberOfPassedExamples).Should(Equal(0))
					Ω(summary.NumberOfFailedExamples).Should(Equal(0))
					Ω(summary.RunTime).Should(Equal(time.Duration(0)))
				})

				It("publishes the correct example summaries", func() {
					Ω(fakeR.exampleSummaries).Should(HaveLen(3))
					Ω(fakeR.exampleSummaries[0]).Should(Equal(example1.summary()))
					Ω(fakeR.exampleSummaries[1]).Should(Equal(example2.summary()))
					Ω(fakeR.exampleSummaries[2]).Should(Equal(example3.summary()))
				})

				It("publishes the correct ending suite summary", func() {
					summary := fakeR.endSummary
					Ω(summary.SuiteDescription).Should(Equal("collection description"))
					Ω(summary.NumberOfTotalExamples).Should(Equal(3))
					Ω(summary.NumberOfExamplesThatWillBeRun).Should(Equal(3))
					Ω(summary.NumberOfPendingExamples).Should(Equal(0))
					Ω(summary.NumberOfSkippedExamples).Should(Equal(0))
					Ω(summary.NumberOfPassedExamples).Should(Equal(3))
					Ω(summary.NumberOfFailedExamples).Should(Equal(0))
					Ω(summary.RunTime.Seconds()).Should(BeNumerically("~", 3*0.001, 0.01))
				})
			})

			Context("when examples fail", func() {
				BeforeEach(func() {
					example2 = exampleWithItFunc("failing it 2", flagTypeNone, true)
					example3 = exampleWithItFunc("failing it 3", flagTypeNone, true)
				})

				It("runs all the tests", func() {
					Ω(itsThatWereRun).Should(Equal([]string{"it 1", "failing it 2", "failing it 3"}))
				})

				It("marks the suite as failed", func() {
					Ω(fakeT.didFail).Should(BeTrue())
				})

				It("publishes the correct starting suite summary", func() {
					summary := fakeR.beginSummary
					Ω(summary.SuiteDescription).Should(Equal("collection description"))
					Ω(summary.NumberOfTotalExamples).Should(Equal(3))
					Ω(summary.NumberOfExamplesThatWillBeRun).Should(Equal(3))
					Ω(summary.NumberOfPendingExamples).Should(Equal(0))
					Ω(summary.NumberOfSkippedExamples).Should(Equal(0))
					Ω(summary.NumberOfPassedExamples).Should(Equal(0))
					Ω(summary.NumberOfFailedExamples).Should(Equal(0))
					Ω(summary.RunTime).Should(Equal(time.Duration(0)))
				})

				It("publishes the correct example summaries", func() {
					Ω(fakeR.exampleSummaries).Should(HaveLen(3))
					Ω(fakeR.exampleSummaries[0]).Should(Equal(example1.summary()))
					Ω(fakeR.exampleSummaries[1]).Should(Equal(example2.summary()))
					Ω(fakeR.exampleSummaries[2]).Should(Equal(example3.summary()))
				})

				It("publishes the correct ending suite summary", func() {
					summary := fakeR.endSummary
					Ω(summary.SuiteDescription).Should(Equal("collection description"))
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
					Ω(itsThatWereRun).Should(Equal([]string{"it 2", "it 3"}))
				})

				It("marks the suite as passed", func() {
					Ω(fakeT.didFail).Should(BeFalse())
				})

				It("publishes the correct starting suite summary", func() {
					summary := fakeR.beginSummary
					Ω(summary.SuiteDescription).Should(Equal("collection description"))
					Ω(summary.NumberOfTotalExamples).Should(Equal(3))
					Ω(summary.NumberOfExamplesThatWillBeRun).Should(Equal(2))
					Ω(summary.NumberOfPendingExamples).Should(Equal(1))
					Ω(summary.NumberOfSkippedExamples).Should(Equal(0))
					Ω(summary.NumberOfPassedExamples).Should(Equal(0))
					Ω(summary.NumberOfFailedExamples).Should(Equal(0))
					Ω(summary.RunTime).Should(Equal(time.Duration(0)))
				})

				It("publishes the correct example summaries", func() {
					Ω(fakeR.exampleSummaries).Should(HaveLen(3))
					Ω(fakeR.exampleSummaries[0]).Should(Equal(example1.summary()))
					Ω(fakeR.exampleSummaries[1]).Should(Equal(example2.summary()))
					Ω(fakeR.exampleSummaries[2]).Should(Equal(example3.summary()))
				})

				It("publishes the correct ending suite summary", func() {
					summary := fakeR.endSummary
					Ω(summary.SuiteDescription).Should(Equal("collection description"))
					Ω(summary.NumberOfTotalExamples).Should(Equal(3))
					Ω(summary.NumberOfExamplesThatWillBeRun).Should(Equal(2))
					Ω(summary.NumberOfPendingExamples).Should(Equal(1))
					Ω(summary.NumberOfSkippedExamples).Should(Equal(0))
					Ω(summary.NumberOfPassedExamples).Should(Equal(2))
					Ω(summary.NumberOfFailedExamples).Should(Equal(0))
					Ω(summary.RunTime.Seconds()).Should(BeNumerically("~", 3*0.001, 0.01))
				})
			})

			Context("when examples are focused", func() {
				BeforeEach(func() {
					example1 = exampleWithItFunc("focused it 1", flagTypeFocused, false)
					example3 = exampleWithItFunc("focused it 3", flagTypeFocused, false)
				})

				It("skips the non-focused examples", func() {
					Ω(itsThatWereRun).Should(Equal([]string{"focused it 1", "focused it 3"}))
				})

				It("marks the suite as passed", func() {
					Ω(fakeT.didFail).Should(BeFalse())
				})

				It("publishes the correct starting suite summary", func() {
					summary := fakeR.beginSummary
					Ω(summary.SuiteDescription).Should(Equal("collection description"))
					Ω(summary.NumberOfTotalExamples).Should(Equal(3))
					Ω(summary.NumberOfExamplesThatWillBeRun).Should(Equal(2))
					Ω(summary.NumberOfPendingExamples).Should(Equal(0))
					Ω(summary.NumberOfSkippedExamples).Should(Equal(1))
					Ω(summary.NumberOfPassedExamples).Should(Equal(0))
					Ω(summary.NumberOfFailedExamples).Should(Equal(0))
					Ω(summary.RunTime).Should(Equal(time.Duration(0)))
				})

				It("publishes the correct example summaries", func() {
					Ω(fakeR.exampleSummaries).Should(HaveLen(3))
					Ω(fakeR.exampleSummaries[0]).Should(Equal(example1.summary()))
					Ω(fakeR.exampleSummaries[1]).Should(Equal(example2.summary()))
					Ω(fakeR.exampleSummaries[2]).Should(Equal(example3.summary()))
				})

				It("publishes the correct ending suite summary", func() {
					summary := fakeR.endSummary
					Ω(summary.SuiteDescription).Should(Equal("collection description"))
					Ω(summary.NumberOfTotalExamples).Should(Equal(3))
					Ω(summary.NumberOfExamplesThatWillBeRun).Should(Equal(2))
					Ω(summary.NumberOfPendingExamples).Should(Equal(0))
					Ω(summary.NumberOfSkippedExamples).Should(Equal(1))
					Ω(summary.NumberOfPassedExamples).Should(Equal(2))
					Ω(summary.NumberOfFailedExamples).Should(Equal(0))
					Ω(summary.RunTime.Seconds()).Should(BeNumerically("~", 3*0.001, 0.01))
				})
			})
		})
	})
}
