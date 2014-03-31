package specrunner_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/internal/specrunner"
	"github.com/onsi/ginkgo/types"
	. "github.com/onsi/gomega"

	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/internal/codelocation"
	"github.com/onsi/ginkgo/internal/containernode"
	"github.com/onsi/ginkgo/internal/example"
	Failer "github.com/onsi/ginkgo/internal/failer"
	"github.com/onsi/ginkgo/internal/leafnodes"
	"github.com/onsi/ginkgo/internal/types"
	Writer "github.com/onsi/ginkgo/internal/writer"
	"github.com/onsi/ginkgo/reporters"
)

var noneFlag = internaltypes.FlagTypeNone
var focusedFlag = internaltypes.FlagTypeFocused
var pendingFlag = internaltypes.FlagTypePending

var _ = Describe("Example Collection", func() {
	var (
		T         *fakeTestingT
		reporter1 *reporters.FakeReporter
		reporter2 *reporters.FakeReporter
		failer    *Failer.Failer
		writer    *Writer.FakeGinkgoWriter

		examplesThatRan []string

		runner *SpecRunner
	)

	newExample := func(text string, flag internaltypes.FlagType, fail bool) *example.Example {
		subject := leafnodes.NewItNode(text, func() {
			writer.AddEvent(text)
			examplesThatRan = append(examplesThatRan, text)
			if fail {
				failer.Fail(text, codelocation.New(0))
			}
		}, flag, codelocation.New(0), 0, failer, 0)

		return example.New(subject, []*containernode.ContainerNode{})
	}

	newExampleWithBody := func(text string, body interface{}) *example.Example {
		subject := leafnodes.NewItNode(text, body, noneFlag, codelocation.New(0), 0, failer, 0)

		return example.New(subject, []*containernode.ContainerNode{})
	}

	newRunner := func(config config.GinkgoConfigType, examples ...*example.Example) *SpecRunner {
		return New(T, "description", example.NewExamples(examples), []reporters.Reporter{reporter1, reporter2}, writer, config)
	}

	BeforeEach(func() {
		T = &fakeTestingT{}
		reporter1 = reporters.NewFakeReporter()
		reporter2 = reporters.NewFakeReporter()
		writer = Writer.NewFake()
		failer = Failer.New()

		examplesThatRan = []string{}
	})

	Describe("Running and Reporting", func() {
		var exampleA, pendingExample, anotherPendingExample, failedExample, exampleB, skippedExample *example.Example
		BeforeEach(func() {
			exampleA = newExample("example A", noneFlag, false)
			pendingExample = newExample("pending example", pendingFlag, false)
			anotherPendingExample = newExample("another pending example", pendingFlag, false)
			failedExample = newExample("failed example", noneFlag, true)
			exampleB = newExample("example B", noneFlag, false)
			skippedExample = newExample("skipped example", noneFlag, false)
			skippedExample.Skip()

			runner = newRunner(config.GinkgoConfigType{RandomSeed: 17}, exampleA, pendingExample, anotherPendingExample, failedExample, exampleB, skippedExample)
			runner.Run()
		})

		It("should skip skipped/pending tests", func() {
			Ω(examplesThatRan).Should(Equal([]string{"example A", "failed example", "example B"}))
		})

		It("should report to any attached reporters", func() {
			Ω(reporter1.Config).Should(Equal(reporter2.Config))
			Ω(reporter1.BeginSummary).Should(Equal(reporter2.BeginSummary))
			Ω(reporter1.ExampleWillRunSummaries).Should(Equal(reporter2.ExampleWillRunSummaries))
			Ω(reporter1.ExampleSummaries).Should(Equal(reporter2.ExampleSummaries))
			Ω(reporter1.EndSummary).Should(Equal(reporter2.EndSummary))
		})

		It("should report the passed in config", func() {
			Ω(reporter1.Config.RandomSeed).Should(BeNumerically("==", 17))
		})

		It("should report the beginning of the suite", func() {
			Ω(reporter1.BeginSummary.SuiteDescription).Should(Equal("description"))
			Ω(reporter1.BeginSummary.SuiteID).Should(MatchRegexp("[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}"))
			Ω(reporter1.BeginSummary.NumberOfExamplesBeforeParallelization).Should(Equal(6))
			Ω(reporter1.BeginSummary.NumberOfTotalExamples).Should(Equal(6))
			Ω(reporter1.BeginSummary.NumberOfExamplesThatWillBeRun).Should(Equal(3))
			Ω(reporter1.BeginSummary.NumberOfPendingExamples).Should(Equal(2))
			Ω(reporter1.BeginSummary.NumberOfSkippedExamples).Should(Equal(1))
		})

		It("should report the end of the suite", func() {
			Ω(reporter1.EndSummary.SuiteDescription).Should(Equal("description"))
			Ω(reporter1.EndSummary.SuiteSucceeded).Should(BeFalse())
			Ω(reporter1.EndSummary.SuiteID).Should(MatchRegexp("[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}"))
			Ω(reporter1.EndSummary.NumberOfExamplesBeforeParallelization).Should(Equal(6))
			Ω(reporter1.EndSummary.NumberOfTotalExamples).Should(Equal(6))
			Ω(reporter1.EndSummary.NumberOfExamplesThatWillBeRun).Should(Equal(3))
			Ω(reporter1.EndSummary.NumberOfPendingExamples).Should(Equal(2))
			Ω(reporter1.EndSummary.NumberOfSkippedExamples).Should(Equal(1))
			Ω(reporter1.EndSummary.NumberOfPassedExamples).Should(Equal(2))
			Ω(reporter1.EndSummary.NumberOfFailedExamples).Should(Equal(1))
		})
	})

	Describe("reporting on examples", func() {
		var proceed chan bool
		BeforeEach(func() {
			proceed = make(chan bool)
			skippedExample := newExample("SKIP", noneFlag, false)
			skippedExample.Skip()

			runner = newRunner(
				config.GinkgoConfigType{},
				skippedExample,
				newExample("PENDING", pendingFlag, false),
				newExampleWithBody("RUN", func() {
					<-proceed
				}),
			)
			go runner.Run()
		})

		It("should report about pending/skipped examples", func() {
			Eventually(func() interface{} {
				return reporter1.ExampleWillRunSummaries
			}).Should(HaveLen(3))

			Ω(reporter1.ExampleWillRunSummaries[0].ComponentTexts[0]).Should(Equal("SKIP"))
			Ω(reporter1.ExampleWillRunSummaries[1].ComponentTexts[0]).Should(Equal("PENDING"))
			Ω(reporter1.ExampleWillRunSummaries[2].ComponentTexts[0]).Should(Equal("RUN"))

			Ω(reporter1.ExampleSummaries[0].ComponentTexts[0]).Should(Equal("SKIP"))
			Ω(reporter1.ExampleSummaries[1].ComponentTexts[0]).Should(Equal("PENDING"))
			Ω(reporter1.ExampleSummaries).Should(HaveLen(2))

			close(proceed)

			Eventually(func() interface{} {
				return reporter1.ExampleSummaries
			}).Should(HaveLen(3))
			Ω(reporter1.ExampleSummaries[2].ComponentTexts[0]).Should(Equal("RUN"))
		})
	})

	Describe("Marking failure and success", func() {
		Context("when all tests pass", func() {
			BeforeEach(func() {
				runner = newRunner(config.GinkgoConfigType{}, newExample("passing", noneFlag, false), newExample("pending", pendingFlag, false))
			})

			It("should return true and report success", func() {
				Ω(runner.Run()).Should(BeTrue())
				Ω(reporter1.EndSummary.SuiteSucceeded).Should(BeTrue())
				Ω(T.didFail).Should(BeFalse())
			})
		})

		Context("when a test fails", func() {
			BeforeEach(func() {
				runner = newRunner(config.GinkgoConfigType{}, newExample("failing", noneFlag, true), newExample("pending", pendingFlag, false))
			})

			It("should return false and report failure", func() {
				Ω(runner.Run()).Should(BeFalse())
				Ω(reporter1.EndSummary.SuiteSucceeded).Should(BeFalse())
				Ω(T.didFail).Should(BeTrue())
			})
		})

		Context("when there is a pending test, but pendings count as failures", func() {
			BeforeEach(func() {
				runner = newRunner(config.GinkgoConfigType{FailOnPending: true}, newExample("passing", noneFlag, false), newExample("pending", pendingFlag, false))
			})

			It("should return false and report failure", func() {
				Ω(runner.Run()).Should(BeFalse())
				Ω(reporter1.EndSummary.SuiteSucceeded).Should(BeFalse())
				Ω(T.didFail).Should(BeTrue())
			})
		})
	})

	Describe("Managing the writer", func() {
		BeforeEach(func() {
			runner = newRunner(
				config.GinkgoConfigType{},
				newExample("A", noneFlag, false),
				newExample("B", noneFlag, true),
				newExample("C", noneFlag, false),
			)
			runner.Run()
		})

		It("should truncate between tests, but only dump if a test fails", func() {
			Ω(writer.EventStream).Should(Equal([]string{"TRUNCATE", "A", "TRUNCATE", "B", "DUMP", "TRUNCATE", "C"}))
		})
	})

	Describe("CurrentExampleSummary", func() {
		It("should return the example summary for the currently running example", func() {
			var summary *types.ExampleSummary
			runner = newRunner(
				config.GinkgoConfigType{},
				newExample("A", noneFlag, false),
				newExampleWithBody("B", func() {
					var ok bool
					summary, ok = runner.CurrentExampleSummary()
					Ω(ok).Should(BeTrue())
				}),
				newExample("C", noneFlag, false),
			)
			runner.Run()

			Ω(summary.ComponentTexts).Should(Equal([]string{"B"}))

			summary, ok := runner.CurrentExampleSummary()
			Ω(summary).Should(BeNil())
			Ω(ok).Should(BeFalse())
		})
	})

	Context("When running tests in parallel", func() {
		It("reports the correct number of examples before parallelization", func() {
			examples := example.NewExamples([]*example.Example{
				newExample("A", noneFlag, false),
				newExample("B", pendingFlag, false),
				newExample("C", noneFlag, false),
			})
			examples.TrimForParallelization(2, 1)
			runner = New(T, "description", examples, []reporters.Reporter{reporter1, reporter2}, writer, config.GinkgoConfigType{})
			runner.Run()

			Ω(reporter1.EndSummary.NumberOfExamplesBeforeParallelization).Should(Equal(3))
			Ω(reporter1.EndSummary.NumberOfTotalExamples).Should(Equal(2))
			Ω(reporter1.EndSummary.NumberOfExamplesThatWillBeRun).Should(Equal(1))
			Ω(reporter1.EndSummary.NumberOfPendingExamples).Should(Equal(1))
		})
	})

	Describe("generating a suite id", func() {
		It("should generate an id randomly", func() {
			runnerA := newRunner(config.GinkgoConfigType{})
			runnerA.Run()
			IDA := reporter1.BeginSummary.SuiteID

			runnerB := newRunner(config.GinkgoConfigType{})
			runnerB.Run()
			IDB := reporter1.BeginSummary.SuiteID

			IDRegexp := "[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}"
			Ω(IDA).Should(MatchRegexp(IDRegexp))
			Ω(IDB).Should(MatchRegexp(IDRegexp))

			Ω(IDA).ShouldNot(Equal(IDB))
		})
	})
})
