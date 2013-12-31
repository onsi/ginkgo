package aggregator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/ginkgo/config"
	. "github.com/onsi/ginkgo/ginkgo/aggregator"
	st "github.com/onsi/ginkgo/stenographer"
	"github.com/onsi/ginkgo/types"
	"runtime"
	"time"
)

var _ = Describe("Aggregator", func() {
	var (
		aggregator     *Aggregator
		reporterConfig config.DefaultReporterConfigType
		stenographer   *st.FakeStenographer
		result         chan bool

		ginkgoConfig1 config.GinkgoConfigType
		ginkgoConfig2 config.GinkgoConfigType

		suiteSummary1 *types.SuiteSummary
		suiteSummary2 *types.SuiteSummary

		exampleSummary *types.ExampleSummary

		suiteDescription string
	)

	BeforeEach(func() {
		reporterConfig = config.DefaultReporterConfigType{
			NoColor:           false,
			SlowSpecThreshold: 0.1,
			NoisyPendings:     true,
			Succinct:          false,
			Verbose:           true,
		}
		stenographer = st.NewFakeStenographer()
		result = make(chan bool, 1)
		aggregator = NewAggregator(2, result, reporterConfig, stenographer)

		//
		// now set up some fixture data
		//

		ginkgoConfig1 = config.GinkgoConfigType{
			RandomSeed:        1138,
			RandomizeAllSpecs: true,
			ParallelNode:      1,
			ParallelTotal:     2,
		}

		ginkgoConfig2 = config.GinkgoConfigType{
			RandomSeed:        1138,
			RandomizeAllSpecs: true,
			ParallelNode:      2,
			ParallelTotal:     2,
		}

		suiteDescription = "My Parallel Suite"

		suiteSummary1 = &types.SuiteSummary{
			SuiteDescription: suiteDescription,

			NumberOfExamplesBeforeParallelization: 30,
			NumberOfTotalExamples:                 17,
			NumberOfExamplesThatWillBeRun:         15,
			NumberOfPendingExamples:               1,
			NumberOfSkippedExamples:               1,
		}

		suiteSummary2 = &types.SuiteSummary{
			SuiteDescription: suiteDescription,

			NumberOfExamplesBeforeParallelization: 30,
			NumberOfTotalExamples:                 13,
			NumberOfExamplesThatWillBeRun:         8,
			NumberOfPendingExamples:               2,
			NumberOfSkippedExamples:               3,
		}

		exampleSummary = &types.ExampleSummary{
			State: types.ExampleStatePassed,
		}
	})

	call := func(method string, args ...interface{}) st.FakeStenographerCall {
		return st.NewFakeStenographerCall(method, args...)
	}

	beginSuite := func() {
		stenographer.Reset()
		aggregator.SpecSuiteWillBegin(ginkgoConfig2, suiteSummary2)
		aggregator.SpecSuiteWillBegin(ginkgoConfig1, suiteSummary1)
		Eventually(func() interface{} {
			return len(stenographer.Calls)
		}).Should(BeNumerically(">=", 3))
	}

	Describe("Announcing the beginning of the suite", func() {
		Context("When one of the parallel-suites starts", func() {
			BeforeEach(func() {
				aggregator.SpecSuiteWillBegin(ginkgoConfig2, suiteSummary2)
				runtime.Gosched()
			})

			It("should be silent", func() {
				Ω(stenographer.Calls).Should(BeEmpty())
			})
		})

		Context("once all of the parallel-suites have started", func() {
			BeforeEach(func() {
				aggregator.SpecSuiteWillBegin(ginkgoConfig2, suiteSummary2)
				aggregator.SpecSuiteWillBegin(ginkgoConfig1, suiteSummary1)
				Eventually(func() interface{} {
					return stenographer.Calls
				}).Should(HaveLen(3))
			})

			It("should announce the beginning of the suite", func() {
				Ω(stenographer.Calls).Should(HaveLen(3))
				Ω(stenographer.Calls[0]).Should(Equal(call("AnnounceSuite", suiteDescription, ginkgoConfig1.RandomSeed, true)))
				Ω(stenographer.Calls[1]).Should(Equal(call("AnnounceNumberOfSpecs", 23, 30)))
				Ω(stenographer.Calls[2]).Should(Equal(call("AnnounceAggregatedParallelRun", 2)))
			})
		})
	})

	Describe("Announcing examples", func() {
		Context("when the parallel-suites have not all started", func() {
			BeforeEach(func() {
				aggregator.ExampleDidComplete(exampleSummary)
				runtime.Gosched()
			})

			It("should not announce any examples", func() {
				Ω(stenographer.Calls).Should(BeEmpty())
			})

			Context("when the parallel-suites subsequently start", func() {
				BeforeEach(func() {
					beginSuite()
				})

				It("should announce the examples", func() {
					Eventually(func() interface{} {
						lastCall := stenographer.Calls[len(stenographer.Calls)-1]
						return lastCall
					}).Should(Equal(call("AnnounceSuccesfulExample", exampleSummary)))
				})
			})
		})

		Context("When the parallel-suites have all started", func() {
			BeforeEach(func() {
				beginSuite()
				stenographer.Reset()
			})

			Context("When an example completes", func() {
				BeforeEach(func() {
					aggregator.ExampleDidComplete(exampleSummary)
					Eventually(func() interface{} {
						return stenographer.Calls
					}).Should(HaveLen(3))
				})

				It("should announce that the example will run (when in verbose mode)", func() {
					Ω(stenographer.Calls[0]).Should(Equal(call("AnnounceExampleWillRun", exampleSummary)))
				})

				It("should announce the captured stdout of the example", func() {
					Ω(stenographer.Calls[1]).Should(Equal(call("AnnounceCapturedOutput", exampleSummary)))
				})

				It("should announce completion", func() {
					Ω(stenographer.Calls[2]).Should(Equal(call("AnnounceSuccesfulExample", exampleSummary)))
				})
			})
		})
	})

	Describe("Announcing the end of the suite", func() {
		BeforeEach(func() {
			beginSuite()
			stenographer.Reset()
		})

		Context("When one of the parallel-suites ends", func() {
			BeforeEach(func() {
				aggregator.SpecSuiteDidEnd(suiteSummary2)
				runtime.Gosched()
			})

			It("should be silent", func() {
				Ω(stenographer.Calls).Should(BeEmpty())
			})

			It("should not notify the channel", func() {
				Ω(result).Should(BeEmpty())
			})
		})

		Context("once all of the parallel-suites end", func() {
			BeforeEach(func() {
				time.Sleep(200 * time.Millisecond)

				suiteSummary1.SuiteSucceeded = true
				suiteSummary1.NumberOfPassedExamples = 15
				suiteSummary1.NumberOfFailedExamples = 0
				suiteSummary2.SuiteSucceeded = false
				suiteSummary2.NumberOfPassedExamples = 5
				suiteSummary2.NumberOfFailedExamples = 3

				aggregator.SpecSuiteDidEnd(suiteSummary2)
				aggregator.SpecSuiteDidEnd(suiteSummary1)
				Eventually(func() interface{} {
					return stenographer.Calls
				}).Should(HaveLen(1))
			})

			It("should announce the end of the suite", func() {
				compositeSummary := stenographer.Calls[0].Args[0].(*types.SuiteSummary)

				Ω(compositeSummary.SuiteSucceeded).Should(BeFalse())
				Ω(compositeSummary.NumberOfExamplesThatWillBeRun).Should(Equal(23))
				Ω(compositeSummary.NumberOfTotalExamples).Should(Equal(30))
				Ω(compositeSummary.NumberOfPassedExamples).Should(Equal(20))
				Ω(compositeSummary.NumberOfFailedExamples).Should(Equal(3))
				Ω(compositeSummary.NumberOfPendingExamples).Should(Equal(3))
				Ω(compositeSummary.NumberOfSkippedExamples).Should(Equal(4))
				Ω(compositeSummary.RunTime.Seconds()).Should(BeNumerically(">", 0.2))
			})
		})

		Context("when all the parallel-suites pass", func() {
			BeforeEach(func() {
				suiteSummary1.SuiteSucceeded = true
				suiteSummary2.SuiteSucceeded = true

				aggregator.SpecSuiteDidEnd(suiteSummary2)
				aggregator.SpecSuiteDidEnd(suiteSummary1)
				Eventually(func() interface{} {
					return stenographer.Calls
				}).Should(HaveLen(1))
			})

			It("should report success", func() {
				compositeSummary := stenographer.Calls[0].Args[0].(*types.SuiteSummary)

				Ω(compositeSummary.SuiteSucceeded).Should(BeTrue())
			})

			It("should notify the channel that it succeded", func(done Done) {
				Ω(<-result).Should(BeTrue())
				close(done)
			})
		})

		Context("when one of the parallel-suites fails", func() {
			BeforeEach(func() {
				suiteSummary1.SuiteSucceeded = true
				suiteSummary2.SuiteSucceeded = false

				aggregator.SpecSuiteDidEnd(suiteSummary2)
				aggregator.SpecSuiteDidEnd(suiteSummary1)
				Eventually(func() interface{} {
					return stenographer.Calls
				}).Should(HaveLen(1))
			})

			It("should report failure", func() {
				compositeSummary := stenographer.Calls[0].Args[0].(*types.SuiteSummary)

				Ω(compositeSummary.SuiteSucceeded).Should(BeFalse())
			})

			It("should notify the channel that it failed", func(done Done) {
				Ω(<-result).Should(BeFalse())
				close(done)
			})
		})
	})
})
