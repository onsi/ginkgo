package specrunner_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/internal/specrunner"
	"github.com/onsi/ginkgo/types"
	. "github.com/onsi/gomega"

	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/internal/codelocation"
	"github.com/onsi/ginkgo/internal/containernode"
	Failer "github.com/onsi/ginkgo/internal/failer"
	"github.com/onsi/ginkgo/internal/leafnodes"
	"github.com/onsi/ginkgo/internal/spec"
	Writer "github.com/onsi/ginkgo/internal/writer"
	"github.com/onsi/ginkgo/reporters"
)

var noneFlag = types.FlagTypeNone
var focusedFlag = types.FlagTypeFocused
var pendingFlag = types.FlagTypePending

var _ = Describe("Spec Collection", func() {
	var (
		reporter1 *reporters.FakeReporter
		reporter2 *reporters.FakeReporter
		failer    *Failer.Failer
		writer    *Writer.FakeGinkgoWriter

		specsThatRan []string

		runner *SpecRunner
	)

	newSpec := func(text string, flag types.FlagType, fail bool) *spec.Spec {
		subject := leafnodes.NewItNode(text, func() {
			writer.AddEvent(text)
			specsThatRan = append(specsThatRan, text)
			if fail {
				failer.Fail(text, codelocation.New(0))
			}
		}, flag, codelocation.New(0), 0, failer, 0)

		return spec.New(subject, []*containernode.ContainerNode{})
	}

	newSpecWithBody := func(text string, body interface{}) *spec.Spec {
		subject := leafnodes.NewItNode(text, body, noneFlag, codelocation.New(0), 0, failer, 0)

		return spec.New(subject, []*containernode.ContainerNode{})
	}

	newRunner := func(config config.GinkgoConfigType, specs ...*spec.Spec) *SpecRunner {
		return New("description", spec.NewSpecs(specs), []reporters.Reporter{reporter1, reporter2}, writer, config)
	}

	BeforeEach(func() {
		reporter1 = reporters.NewFakeReporter()
		reporter2 = reporters.NewFakeReporter()
		writer = Writer.NewFake()
		failer = Failer.New()

		specsThatRan = []string{}
	})

	Describe("Running and Reporting", func() {
		var specA, pendingSpec, anotherPendingSpec, failedSpec, specB, skippedSpec *spec.Spec
		BeforeEach(func() {
			specA = newSpec("spec A", noneFlag, false)
			pendingSpec = newSpec("pending spec", pendingFlag, false)
			anotherPendingSpec = newSpec("another pending spec", pendingFlag, false)
			failedSpec = newSpec("failed spec", noneFlag, true)
			specB = newSpec("spec B", noneFlag, false)
			skippedSpec = newSpec("skipped spec", noneFlag, false)
			skippedSpec.Skip()

			runner = newRunner(config.GinkgoConfigType{RandomSeed: 17}, specA, pendingSpec, anotherPendingSpec, failedSpec, specB, skippedSpec)
			runner.Run()
		})

		It("should skip skipped/pending tests", func() {
			Ω(specsThatRan).Should(Equal([]string{"spec A", "failed spec", "spec B"}))
		})

		It("should report to any attached reporters", func() {
			Ω(reporter1.Config).Should(Equal(reporter2.Config))
			Ω(reporter1.BeginSummary).Should(Equal(reporter2.BeginSummary))
			Ω(reporter1.SpecWillRunSummaries).Should(Equal(reporter2.SpecWillRunSummaries))
			Ω(reporter1.SpecSummaries).Should(Equal(reporter2.SpecSummaries))
			Ω(reporter1.EndSummary).Should(Equal(reporter2.EndSummary))
		})

		It("should report the passed in config", func() {
			Ω(reporter1.Config.RandomSeed).Should(BeNumerically("==", 17))
		})

		It("should report the beginning of the suite", func() {
			Ω(reporter1.BeginSummary.SuiteDescription).Should(Equal("description"))
			Ω(reporter1.BeginSummary.SuiteID).Should(MatchRegexp("[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}"))
			Ω(reporter1.BeginSummary.NumberOfSpecsBeforeParallelization).Should(Equal(6))
			Ω(reporter1.BeginSummary.NumberOfTotalSpecs).Should(Equal(6))
			Ω(reporter1.BeginSummary.NumberOfSpecsThatWillBeRun).Should(Equal(3))
			Ω(reporter1.BeginSummary.NumberOfPendingSpecs).Should(Equal(2))
			Ω(reporter1.BeginSummary.NumberOfSkippedSpecs).Should(Equal(1))
		})

		It("should report the end of the suite", func() {
			Ω(reporter1.EndSummary.SuiteDescription).Should(Equal("description"))
			Ω(reporter1.EndSummary.SuiteSucceeded).Should(BeFalse())
			Ω(reporter1.EndSummary.SuiteID).Should(MatchRegexp("[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}"))
			Ω(reporter1.EndSummary.NumberOfSpecsBeforeParallelization).Should(Equal(6))
			Ω(reporter1.EndSummary.NumberOfTotalSpecs).Should(Equal(6))
			Ω(reporter1.EndSummary.NumberOfSpecsThatWillBeRun).Should(Equal(3))
			Ω(reporter1.EndSummary.NumberOfPendingSpecs).Should(Equal(2))
			Ω(reporter1.EndSummary.NumberOfSkippedSpecs).Should(Equal(1))
			Ω(reporter1.EndSummary.NumberOfPassedSpecs).Should(Equal(2))
			Ω(reporter1.EndSummary.NumberOfFailedSpecs).Should(Equal(1))
		})
	})

	Describe("reporting on specs", func() {
		var proceed chan bool
		BeforeEach(func() {
			proceed = make(chan bool)
			skippedSpec := newSpec("SKIP", noneFlag, false)
			skippedSpec.Skip()

			runner = newRunner(
				config.GinkgoConfigType{},
				skippedSpec,
				newSpec("PENDING", pendingFlag, false),
				newSpecWithBody("RUN", func() {
					<-proceed
				}),
			)
			go runner.Run()
		})

		It("should report about pending/skipped specs", func() {
			Eventually(func() interface{} {
				return reporter1.SpecWillRunSummaries
			}).Should(HaveLen(3))

			Ω(reporter1.SpecWillRunSummaries[0].ComponentTexts[0]).Should(Equal("SKIP"))
			Ω(reporter1.SpecWillRunSummaries[1].ComponentTexts[0]).Should(Equal("PENDING"))
			Ω(reporter1.SpecWillRunSummaries[2].ComponentTexts[0]).Should(Equal("RUN"))

			Ω(reporter1.SpecSummaries[0].ComponentTexts[0]).Should(Equal("SKIP"))
			Ω(reporter1.SpecSummaries[1].ComponentTexts[0]).Should(Equal("PENDING"))
			Ω(reporter1.SpecSummaries).Should(HaveLen(2))

			close(proceed)

			Eventually(func() interface{} {
				return reporter1.SpecSummaries
			}).Should(HaveLen(3))
			Ω(reporter1.SpecSummaries[2].ComponentTexts[0]).Should(Equal("RUN"))
		})
	})

	Describe("Marking failure and success", func() {
		Context("when all tests pass", func() {
			BeforeEach(func() {
				runner = newRunner(config.GinkgoConfigType{}, newSpec("passing", noneFlag, false), newSpec("pending", pendingFlag, false))
			})

			It("should return true and report success", func() {
				Ω(runner.Run()).Should(BeTrue())
				Ω(reporter1.EndSummary.SuiteSucceeded).Should(BeTrue())
			})
		})

		Context("when a test fails", func() {
			BeforeEach(func() {
				runner = newRunner(config.GinkgoConfigType{}, newSpec("failing", noneFlag, true), newSpec("pending", pendingFlag, false))
			})

			It("should return false and report failure", func() {
				Ω(runner.Run()).Should(BeFalse())
				Ω(reporter1.EndSummary.SuiteSucceeded).Should(BeFalse())
			})
		})

		Context("when there is a pending test, but pendings count as failures", func() {
			BeforeEach(func() {
				runner = newRunner(config.GinkgoConfigType{FailOnPending: true}, newSpec("passing", noneFlag, false), newSpec("pending", pendingFlag, false))
			})

			It("should return false and report failure", func() {
				Ω(runner.Run()).Should(BeFalse())
				Ω(reporter1.EndSummary.SuiteSucceeded).Should(BeFalse())
			})
		})
	})

	Describe("Managing the writer", func() {
		BeforeEach(func() {
			runner = newRunner(
				config.GinkgoConfigType{},
				newSpec("A", noneFlag, false),
				newSpec("B", noneFlag, true),
				newSpec("C", noneFlag, false),
			)
			runner.Run()
		})

		It("should truncate between tests, but only dump if a test fails", func() {
			Ω(writer.EventStream).Should(Equal([]string{"TRUNCATE", "A", "TRUNCATE", "B", "DUMP", "TRUNCATE", "C"}))
		})
	})

	Describe("CurrentSpecSummary", func() {
		It("should return the spec summary for the currently running spec", func() {
			var summary *types.SpecSummary
			runner = newRunner(
				config.GinkgoConfigType{},
				newSpec("A", noneFlag, false),
				newSpecWithBody("B", func() {
					var ok bool
					summary, ok = runner.CurrentSpecSummary()
					Ω(ok).Should(BeTrue())
				}),
				newSpec("C", noneFlag, false),
			)
			runner.Run()

			Ω(summary.ComponentTexts).Should(Equal([]string{"B"}))

			summary, ok := runner.CurrentSpecSummary()
			Ω(summary).Should(BeNil())
			Ω(ok).Should(BeFalse())
		})
	})

	Context("When running tests in parallel", func() {
		It("reports the correct number of specs before parallelization", func() {
			specs := spec.NewSpecs([]*spec.Spec{
				newSpec("A", noneFlag, false),
				newSpec("B", pendingFlag, false),
				newSpec("C", noneFlag, false),
			})
			specs.TrimForParallelization(2, 1)
			runner = New("description", specs, []reporters.Reporter{reporter1, reporter2}, writer, config.GinkgoConfigType{})
			runner.Run()

			Ω(reporter1.EndSummary.NumberOfSpecsBeforeParallelization).Should(Equal(3))
			Ω(reporter1.EndSummary.NumberOfTotalSpecs).Should(Equal(2))
			Ω(reporter1.EndSummary.NumberOfSpecsThatWillBeRun).Should(Equal(1))
			Ω(reporter1.EndSummary.NumberOfPendingSpecs).Should(Equal(1))
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
