package reporters_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/config"
	"github.com/onsi/ginkgo/v2/reporters"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
)

type deprecatedReporter struct {
	config      config.GinkgoConfigType
	begin       types.SuiteSummary
	beforeSuite types.SetupSummary
	will        []types.SpecSummary
	did         []types.SpecSummary
	afterSuite  types.SetupSummary
	end         types.SuiteSummary
}

func (dr *deprecatedReporter) SuiteWillBegin(config config.GinkgoConfigType, summary *types.SuiteSummary) {
	dr.config = config
	dr.begin = *summary
}
func (dr *deprecatedReporter) BeforeSuiteDidRun(setupSummary *types.SetupSummary) {
	dr.beforeSuite = *setupSummary
}
func (dr *deprecatedReporter) SpecWillRun(specSummary *types.SpecSummary) {
	dr.will = append(dr.will, *specSummary)
}
func (dr *deprecatedReporter) SpecDidComplete(specSummary *types.SpecSummary) {
	dr.did = append(dr.did, *specSummary)
}
func (dr *deprecatedReporter) AfterSuiteDidRun(setupSummary *types.SetupSummary) {
	dr.afterSuite = *setupSummary
}
func (dr *deprecatedReporter) SuiteDidEnd(summary *types.SuiteSummary) {
	dr.end = *summary
}

var _ = Describe("DeprecatedReporter", func() {
	var report types.Report
	var reporter *deprecatedReporter

	BeforeEach(func() {
		reporter = &deprecatedReporter{}
		report = types.Report{
			SuiteDescription: "suite-description",
			SuitePath:        "suite-path",
			SuiteSucceeded:   false,
			PreRunStats: types.PreRunStats{
				TotalSpecs:       10,
				SpecsThatWillRun: 9,
			},
			RunTime: time.Minute,
			SuiteConfig: types.SuiteConfig{
				RandomSeed: 17,
			},
			SpecReports: types.SpecReports{
				types.SpecReport{
					LeafNodeType:               types.NodeTypeBeforeSuite,
					LeafNodeLocation:           cl0,
					RunTime:                    time.Second,
					CapturedGinkgoWriterOutput: "gw",
					CapturedStdOutErr:          "std",
					State:                      types.SpecStatePassed,
				},

				types.SpecReport{
					ContainerHierarchyTexts:     []string{"A", "B"},
					ContainerHierarchyLocations: []types.CodeLocation{cl0, cl1},
					LeafNodeType:                types.NodeTypeIt,
					LeafNodeLocation:            cl2,
					LeafNodeText:                "it",
					RunTime:                     time.Second,
					CapturedGinkgoWriterOutput:  "gw",
					CapturedStdOutErr:           "std",
					State:                       types.SpecStatePassed,
					NumAttempts:                 1,
				},

				types.SpecReport{
					ContainerHierarchyTexts:     []string{"A", "B"},
					ContainerHierarchyLocations: []types.CodeLocation{cl0, cl1},
					LeafNodeType:                types.NodeTypeIt,
					LeafNodeLocation:            cl2,
					LeafNodeText:                "it",
					RunTime:                     time.Second,
					CapturedGinkgoWriterOutput:  "gw",
					CapturedStdOutErr:           "std",
					State:                       types.SpecStatePassed,
					NumAttempts:                 2,
				},

				types.SpecReport{
					ContainerHierarchyTexts:     []string{"A", "B"},
					ContainerHierarchyLocations: []types.CodeLocation{cl0, cl1},
					LeafNodeType:                types.NodeTypeIt,
					LeafNodeLocation:            cl2,
					LeafNodeText:                "it",
					RunTime:                     time.Second,
					CapturedGinkgoWriterOutput:  "gw",
					CapturedStdOutErr:           "std",
					State:                       types.SpecStatePending,
					NumAttempts:                 0,
				},

				types.SpecReport{
					ContainerHierarchyTexts:     []string{"A", "B"},
					ContainerHierarchyLocations: []types.CodeLocation{cl0, cl1},
					LeafNodeType:                types.NodeTypeIt,
					LeafNodeLocation:            cl2,
					LeafNodeText:                "it",
					RunTime:                     time.Second,
					CapturedGinkgoWriterOutput:  "gw",
					CapturedStdOutErr:           "std",
					State:                       types.SpecStateSkipped,
					NumAttempts:                 1,
					Failure: types.Failure{
						Message:                   "skipped by user in a before each",
						Location:                  cl3,
						FailureNodeContext:        types.FailureNodeInContainer,
						FailureNodeContainerIndex: 1,
						FailureNodeLocation:       cl4,
						FailureNodeType:           types.NodeTypeBeforeEach,
					},
				},

				types.SpecReport{
					ContainerHierarchyTexts:     []string{"A", "B"},
					ContainerHierarchyLocations: []types.CodeLocation{cl0, cl1},
					LeafNodeType:                types.NodeTypeIt,
					LeafNodeLocation:            cl2,
					LeafNodeText:                "it",
					RunTime:                     time.Second,
					CapturedGinkgoWriterOutput:  "gw",
					CapturedStdOutErr:           "std",
					NumAttempts:                 3,
					State:                       types.SpecStateFailed,
					Failure: types.Failure{
						Message:             "failed in the it",
						Location:            cl3,
						FailureNodeContext:  types.FailureNodeIsLeafNode,
						FailureNodeLocation: cl2,
						FailureNodeType:     types.NodeTypeIt,
					},
				},

				types.SpecReport{
					ContainerHierarchyTexts:     []string{"A", "B"},
					ContainerHierarchyLocations: []types.CodeLocation{cl0, cl1},
					LeafNodeType:                types.NodeTypeIt,
					LeafNodeLocation:            cl2,
					LeafNodeText:                "it",
					RunTime:                     time.Second,
					CapturedGinkgoWriterOutput:  "gw",
					CapturedStdOutErr:           "std",
					NumAttempts:                 3,
					State:                       types.SpecStatePanicked,
					Failure: types.Failure{
						Message:             "panicked in a top-level just before each",
						Location:            cl3,
						ForwardedPanic:      "bam!",
						FailureNodeContext:  types.FailureNodeAtTopLevel,
						FailureNodeLocation: cl4,
						FailureNodeType:     types.NodeTypeJustBeforeEach,
					},
				},

				types.SpecReport{
					LeafNodeType:     types.NodeTypeAfterSuite,
					LeafNodeLocation: cl0,
					RunTime:          time.Second,
					State:            types.SpecStateFailed,
					Failure: types.Failure{
						Message:             "failure-message",
						Location:            cl1,
						FailureNodeContext:  types.FailureNodeIsLeafNode,
						FailureNodeType:     types.NodeTypeAfterSuite,
						FailureNodeLocation: cl0,
					},
					CapturedGinkgoWriterOutput: "gw",
					CapturedStdOutErr:          "std",
				},

				types.SpecReport{
					LeafNodeText:               "report",
					LeafNodeType:               types.NodeTypeReportAfterSuite,
					LeafNodeLocation:           cl0,
					RunTime:                    time.Second,
					CapturedGinkgoWriterOutput: "gw",
					CapturedStdOutErr:          "std",
				},
			},
		}

		reporters.ReportViaDeprecatedReporter(reporter, report)
	})

	It("submits a SuiteWillBegin report with config and suite", func() {
		Ω(reporter.config.RandomSeed).Should(Equal(int64(17)))
		Ω(reporter.begin.NumberOfTotalSpecs).Should(Equal(10))
		Ω(reporter.begin.NumberOfSpecsBeforeParallelization).Should(Equal(10))
		Ω(reporter.begin.NumberOfSpecsThatWillBeRun).Should(Equal(9))
		Ω(reporter.begin.SuiteID).Should(Equal("suite-path"))
		Ω(reporter.begin.SuiteDescription).Should(Equal("suite-description"))
	})

	It("submits reports for BeforeSuite", func() {
		Ω(reporter.beforeSuite).Should(Equal(types.DeprecatedSetupSummary{
			ComponentType:  types.SpecComponentTypeBeforeSuite,
			CodeLocation:   cl0,
			State:          types.SpecStatePassed,
			RunTime:        time.Second,
			Failure:        types.DeprecatedSpecFailure{},
			CapturedOutput: "std\ngw",
			SuiteID:        "suite-path",
		}))
	})

	It("submits reports for each spec", func() {
		Ω(reporter.will).Should(HaveLen(6))
		Ω(reporter.did).Should(HaveLen(6))

		Ω(reporter.did[0]).Should(Equal(types.DeprecatedSpecSummary{
			ComponentTexts:         []string{"A", "B", "it"},
			ComponentCodeLocations: []types.CodeLocation{cl0, cl1, cl2},
			State:                  types.SpecStatePassed,
			RunTime:                time.Second,
			Failure:                types.DeprecatedSpecFailure{},
			NumberOfSamples:        1,
			CapturedOutput:         "std\ngw",
			SuiteID:                "suite-path",
		}))

		Ω(reporter.did[1]).Should(Equal(types.DeprecatedSpecSummary{
			ComponentTexts:         []string{"A", "B", "it"},
			ComponentCodeLocations: []types.CodeLocation{cl0, cl1, cl2},
			State:                  types.SpecStatePassed,
			RunTime:                time.Second,
			Failure:                types.DeprecatedSpecFailure{},
			NumberOfSamples:        2,
			CapturedOutput:         "std\ngw",
			SuiteID:                "suite-path",
		}))

		Ω(reporter.did[2]).Should(Equal(types.DeprecatedSpecSummary{
			ComponentTexts:         []string{"A", "B", "it"},
			ComponentCodeLocations: []types.CodeLocation{cl0, cl1, cl2},
			State:                  types.SpecStatePending,
			RunTime:                time.Second,
			Failure:                types.DeprecatedSpecFailure{},
			NumberOfSamples:        0,
			CapturedOutput:         "std\ngw",
			SuiteID:                "suite-path",
		}))

		Ω(reporter.did[3]).Should(Equal(types.DeprecatedSpecSummary{
			ComponentTexts:         []string{"A", "B", "it"},
			ComponentCodeLocations: []types.CodeLocation{cl0, cl1, cl2},

			State:   types.SpecStateSkipped,
			RunTime: time.Second,
			Failure: types.DeprecatedSpecFailure{
				Message:               "skipped by user in a before each",
				Location:              cl3,
				ForwardedPanic:        "",
				ComponentIndex:        1,
				ComponentCodeLocation: cl4,
				ComponentType:         types.SpecComponentTypeBeforeEach,
			},
			NumberOfSamples: 1,
			CapturedOutput:  "std\ngw",
			SuiteID:         "suite-path",
		}))

		Ω(reporter.did[4]).Should(Equal(types.DeprecatedSpecSummary{
			ComponentTexts:         []string{"A", "B", "it"},
			ComponentCodeLocations: []types.CodeLocation{cl0, cl1, cl2},

			State:   types.SpecStateFailed,
			RunTime: time.Second,
			Failure: types.DeprecatedSpecFailure{
				Message:               "failed in the it",
				Location:              cl3,
				ForwardedPanic:        "",
				ComponentIndex:        2,
				ComponentCodeLocation: cl2,
				ComponentType:         types.SpecComponentTypeIt,
			},
			NumberOfSamples: 3,
			CapturedOutput:  "std\ngw",
			SuiteID:         "suite-path",
		}))

		Ω(reporter.did[5]).Should(Equal(types.DeprecatedSpecSummary{
			ComponentTexts:         []string{"A", "B", "it"},
			ComponentCodeLocations: []types.CodeLocation{cl0, cl1, cl2},

			State:   types.SpecStatePanicked,
			RunTime: time.Second,
			Failure: types.DeprecatedSpecFailure{
				Message:               "panicked in a top-level just before each",
				Location:              cl3,
				ForwardedPanic:        "bam!",
				ComponentIndex:        -1,
				ComponentCodeLocation: cl4,
				ComponentType:         types.SpecComponentTypeJustBeforeEach,
			},
			NumberOfSamples: 3,
			CapturedOutput:  "std\ngw",
			SuiteID:         "suite-path",
		}))
	})

	It("submits reports for AfterSuite", func() {
		Ω(reporter.afterSuite).Should(Equal(types.DeprecatedSetupSummary{
			ComponentType: types.SpecComponentTypeAfterSuite,
			CodeLocation:  cl0,
			State:         types.SpecStateFailed,
			RunTime:       time.Second,
			Failure: types.DeprecatedSpecFailure{
				Message:               "failure-message",
				Location:              cl1,
				ComponentIndex:        -1,
				ComponentType:         types.SpecComponentTypeAfterSuite,
				ComponentCodeLocation: cl0,
			},
			CapturedOutput: "std\ngw",
			SuiteID:        "suite-path",
		}))
	})

	It("reports the end of the suite", func() {
		Ω(reporter.end.RunTime).Should(Equal(time.Minute))
		Ω(reporter.end.SuiteSucceeded).Should(BeFalse())
	})
})
