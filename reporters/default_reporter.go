/*
Ginkgo's Default Reporter

A number of command line flags are available to tweak Ginkgo's default output.

These are documented [here](http://onsi.github.io/ginkgo/#running_tests)
*/
package reporters

import (
	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/stenographer"
	"github.com/onsi/ginkgo/types"
)

type DefaultReporter struct {
	config       config.DefaultReporterConfigType
	stenographer stenographer.Stenographer
}

func NewDefaultReporter(config config.DefaultReporterConfigType, stenographer stenographer.Stenographer) *DefaultReporter {
	return &DefaultReporter{
		config:       config,
		stenographer: stenographer,
	}
}

func (reporter *DefaultReporter) SpecSuiteWillBegin(config config.GinkgoConfigType, summary *types.SuiteSummary) {
	reporter.stenographer.AnnounceSuite(summary.SuiteDescription, config.RandomSeed, config.RandomizeAllSpecs)
	if config.ParallelTotal > 1 {
		reporter.stenographer.AnnounceParallelRun(config.ParallelNode, config.ParallelTotal, summary.NumberOfTotalExamples, summary.NumberOfExamplesBeforeParallelization)
	}
	reporter.stenographer.AnnounceNumberOfSpecs(summary.NumberOfExamplesThatWillBeRun, summary.NumberOfTotalExamples)
}

func (reporter *DefaultReporter) ExampleWillRun(exampleSummary *types.ExampleSummary) {
	if reporter.config.Verbose && exampleSummary.State != types.ExampleStatePending && exampleSummary.State != types.ExampleStateSkipped {
		reporter.stenographer.AnnounceExampleWillRun(exampleSummary)
	}
}

func (reporter *DefaultReporter) ExampleDidComplete(exampleSummary *types.ExampleSummary) {
	switch exampleSummary.State {
	case types.ExampleStatePassed:
		if exampleSummary.IsMeasurement {
			reporter.stenographer.AnnounceSuccesfulMeasurement(exampleSummary, reporter.config.Succinct)
		} else if exampleSummary.RunTime.Seconds() >= reporter.config.SlowSpecThreshold {
			reporter.stenographer.AnnounceSuccesfulSlowExample(exampleSummary, reporter.config.Succinct)
		} else {
			reporter.stenographer.AnnounceSuccesfulExample(exampleSummary)
		}
	case types.ExampleStatePending:
		reporter.stenographer.AnnouncePendingExample(exampleSummary, reporter.config.NoisyPendings, reporter.config.Succinct)
	case types.ExampleStateSkipped:
		reporter.stenographer.AnnounceSkippedExample(exampleSummary)
	case types.ExampleStateTimedOut:
		reporter.stenographer.AnnounceExampleTimedOut(exampleSummary, reporter.config.Succinct)
	case types.ExampleStatePanicked:
		reporter.stenographer.AnnounceExamplePanicked(exampleSummary, reporter.config.Succinct)
	case types.ExampleStateFailed:
		reporter.stenographer.AnnounceExampleFailed(exampleSummary, reporter.config.Succinct)
	}
}

func (reporter *DefaultReporter) SpecSuiteDidEnd(summary *types.SuiteSummary) {
	reporter.stenographer.AnnounceSpecRunCompletion(summary)
}
