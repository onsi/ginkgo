package reporters

import (
	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/stenographer"
	"github.com/onsi/ginkgo/types"
)

type DefaultReporter struct {
	config       config.DefaultReporterConfigType
	stenographer *stenographer.Stenographer
}

func NewDefaultReporter(config config.DefaultReporterConfigType) *DefaultReporter {
	return &DefaultReporter{
		config:       config,
		stenographer: stenographer.New(!config.NoColor),
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
			reporter.stenographer.AnnounceSuccesfulMeasurement(exampleSummary)
		} else if exampleSummary.RunTime.Seconds() >= reporter.config.SlowSpecThreshold {
			reporter.stenographer.AnnounceSuccesfulSlowExample(exampleSummary)
		} else {
			reporter.stenographer.AnnounceSuccesfulExample(exampleSummary)
		}
	case types.ExampleStatePending:
		reporter.stenographer.AnnouncePendingExample(exampleSummary, reporter.config.NoisyPendings)
	case types.ExampleStateSkipped:
		reporter.stenographer.AnnounceSkippedExample(exampleSummary)
	case types.ExampleStateTimedOut:
		reporter.stenographer.AnnounceExampleTimedOut(exampleSummary)
	case types.ExampleStatePanicked:
		reporter.stenographer.AnnounceExamplePanicked(exampleSummary)
	case types.ExampleStateFailed:
		reporter.stenographer.AnnounceExampleFailed(exampleSummary)
	}
}

func (reporter *DefaultReporter) SpecSuiteDidEnd(summary *types.SuiteSummary) {
	reporter.stenographer.AnnounceSpecRunCompletion(summary)
}
