package reporters_test

import (
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/reporters"
	st "github.com/onsi/ginkgo/stenographer"
	"github.com/onsi/ginkgo/types"
	. "github.com/onsi/gomega"
	"time"
)

var _ = Describe("DefaultReporter", func() {
	var (
		reporter       *reporters.DefaultReporter
		reporterConfig config.DefaultReporterConfigType
		stenographer   *st.FakeStenographer

		ginkgoConfig config.GinkgoConfigType
		suite        *types.SuiteSummary
		example      *types.ExampleSummary
	)

	BeforeEach(func() {
		stenographer = st.NewFakeStenographer()
		reporterConfig = config.DefaultReporterConfigType{
			NoColor:           false,
			SlowSpecThreshold: 0.1,
			NoisyPendings:     true,
			Succinct:          true,
			Verbose:           true,
		}

		reporter = reporters.NewDefaultReporter(reporterConfig, stenographer)
	})

	call := func(method string, args ...interface{}) st.FakeStenographerCall {
		return st.NewFakeStenographerCall(method, args...)
	}

	Describe("SpecSuiteWillBegin", func() {
		BeforeEach(func() {
			suite = &types.SuiteSummary{
				SuiteDescription:              "A Sweet Suite",
				NumberOfTotalExamples:         10,
				NumberOfExamplesThatWillBeRun: 8,
			}

			ginkgoConfig = config.GinkgoConfigType{
				RandomSeed:        1138,
				RandomizeAllSpecs: true,
			}
		})

		Context("when a serial (non-parallel) suite begins", func() {
			BeforeEach(func() {
				ginkgoConfig.ParallelTotal = 1

				reporter.SpecSuiteWillBegin(ginkgoConfig, suite)
			})

			It("should announce the suite, then announce the number of specs", func() {
				Ω(stenographer.Calls).Should(HaveLen(2))
				Ω(stenographer.Calls[0]).Should(Equal(call("AnnounceSuite", "A Sweet Suite", ginkgoConfig.RandomSeed, true)))
				Ω(stenographer.Calls[1]).Should(Equal(call("AnnounceNumberOfSpecs", 8, 10)))
			})
		})

		Context("when a parallel suite begins", func() {
			BeforeEach(func() {
				ginkgoConfig.ParallelTotal = 2
				ginkgoConfig.ParallelNode = 1
				suite.NumberOfExamplesBeforeParallelization = 20

				reporter.SpecSuiteWillBegin(ginkgoConfig, suite)
			})

			It("should announce the suite, announce that it's a parallel run, then announce the number of specs", func() {
				Ω(stenographer.Calls).Should(HaveLen(3))
				Ω(stenographer.Calls[0]).Should(Equal(call("AnnounceSuite", "A Sweet Suite", ginkgoConfig.RandomSeed, true)))
				Ω(stenographer.Calls[1]).Should(Equal(call("AnnounceParallelRun", 1, 2, 10, 20)))
				Ω(stenographer.Calls[2]).Should(Equal(call("AnnounceNumberOfSpecs", 8, 10)))
			})
		})
	})

	Describe("ExampleWillRun", func() {
		Context("When running in verbose mode", func() {
			Context("and the example will run", func() {
				BeforeEach(func() {
					example = &types.ExampleSummary{}
					reporter.ExampleWillRun(example)
				})

				It("should announce that the example will run", func() {
					Ω(stenographer.Calls).Should(HaveLen(1))
					Ω(stenographer.Calls[0]).Should(Equal(call("AnnounceExampleWillRun", example)))
				})
			})

			Context("and the example will not run", func() {
				Context("because it is pending", func() {
					BeforeEach(func() {
						example = &types.ExampleSummary{
							State: types.ExampleStatePending,
						}
						reporter.ExampleWillRun(example)
					})

					It("should announce nothing", func() {
						Ω(stenographer.Calls).Should(BeEmpty())
					})
				})

				Context("because it is skipped", func() {
					BeforeEach(func() {
						example = &types.ExampleSummary{
							State: types.ExampleStateSkipped,
						}
						reporter.ExampleWillRun(example)
					})

					It("should announce nothing", func() {
						Ω(stenographer.Calls).Should(BeEmpty())
					})
				})
			})
		})

		Context("When not running in verbose mode", func() {
			BeforeEach(func() {
				reporterConfig.Verbose = false
				reporter = reporters.NewDefaultReporter(reporterConfig, stenographer)
				example = &types.ExampleSummary{}
				reporter.ExampleWillRun(example)
			})

			It("should announce nothing", func() {
				Ω(stenographer.Calls).Should(BeEmpty())
			})
		})
	})

	Describe("ExampleDidComplete", func() {
		JustBeforeEach(func() {
			reporter.ExampleDidComplete(example)
		})

		BeforeEach(func() {
			example = &types.ExampleSummary{}
		})

		Context("When the example passed", func() {
			BeforeEach(func() {
				example.State = types.ExampleStatePassed
			})

			Context("When the example was a measurement", func() {
				BeforeEach(func() {
					example.IsMeasurement = true
				})

				It("should announce the measurement", func() {
					Ω(stenographer.Calls[0]).Should(Equal(call("AnnounceSuccesfulMeasurement", example, true)))
				})
			})

			Context("When the example is slow", func() {
				BeforeEach(func() {
					example.RunTime = time.Second
				})

				It("should announce that it was slow", func() {
					Ω(stenographer.Calls[0]).Should(Equal(call("AnnounceSuccesfulSlowExample", example, true)))
				})
			})

			Context("Otherwise", func() {
				It("should announce the succesful example", func() {
					Ω(stenographer.Calls[0]).Should(Equal(call("AnnounceSuccesfulExample", example)))
				})
			})
		})

		Context("When the example is pending", func() {
			BeforeEach(func() {
				example.State = types.ExampleStatePending
			})

			It("should announce the pending example", func() {
				Ω(stenographer.Calls[0]).Should(Equal(call("AnnouncePendingExample", example, true, true)))
			})
		})

		Context("When the example is skipped", func() {
			BeforeEach(func() {
				example.State = types.ExampleStateSkipped
			})

			It("should announce the skipped example", func() {
				Ω(stenographer.Calls[0]).Should(Equal(call("AnnounceSkippedExample", example)))
			})
		})

		Context("When the example timed out", func() {
			BeforeEach(func() {
				example.State = types.ExampleStateTimedOut
			})

			It("should announce the timedout example", func() {
				Ω(stenographer.Calls[0]).Should(Equal(call("AnnounceExampleTimedOut", example, true)))
			})
		})

		Context("When the example panicked", func() {
			BeforeEach(func() {
				example.State = types.ExampleStatePanicked
			})

			It("should announce the panicked example", func() {
				Ω(stenographer.Calls[0]).Should(Equal(call("AnnounceExamplePanicked", example, true)))
			})
		})

		Context("When the example failed", func() {
			BeforeEach(func() {
				example.State = types.ExampleStateFailed
			})

			It("should announce the failed example", func() {
				Ω(stenographer.Calls[0]).Should(Equal(call("AnnounceExampleFailed", example, true)))
			})
		})
	})

	Describe("SpecSuiteDidEnd", func() {
		BeforeEach(func() {
			suite = &types.SuiteSummary{}
			reporter.SpecSuiteDidEnd(suite)
		})

		It("should announce the spec run's completion", func() {
			Ω(stenographer.Calls[0]).Should(Equal(call("AnnounceSpecRunCompletion", suite)))
		})
	})
})
