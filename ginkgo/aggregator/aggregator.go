/*

Aggregator is a reporter used by the Ginkgo CLI to aggregate and present parallel test output
as one coherent stream.  You shouldn't need to use this in your code.  To run tests in parallel:

	ginkgo -nodes=N

where N is the number of nodes you desire.

To disable streaming mode and, instead, have the test output blobbed onto screen when all the parallel nodes complete:

	ginkgo -nodes=N -stream=false

*/
package aggregator

import (
	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/stenographer"
	"github.com/onsi/ginkgo/types"
	"time"
)

type configAndSuite struct {
	config  config.GinkgoConfigType
	summary *types.SuiteSummary
}

type Aggregator struct {
	nodeCount    int
	config       config.DefaultReporterConfigType
	stenographer stenographer.Stenographer
	result       chan bool

	suiteBeginnings           chan configAndSuite
	aggregatedSuiteBeginnings []configAndSuite

	exampleCompletions chan *types.ExampleSummary
	completedExamples  []*types.ExampleSummary

	suiteEndings           chan *types.SuiteSummary
	aggregatedSuiteEndings []*types.SuiteSummary

	startTime time.Time
}

func NewAggregator(nodeCount int, result chan bool, config config.DefaultReporterConfigType, stenographer stenographer.Stenographer) *Aggregator {
	aggregator := &Aggregator{
		nodeCount:    nodeCount,
		result:       result,
		config:       config,
		stenographer: stenographer,

		suiteBeginnings:           make(chan configAndSuite, 0),
		aggregatedSuiteBeginnings: []configAndSuite{},

		exampleCompletions: make(chan *types.ExampleSummary, 0),
		completedExamples:  []*types.ExampleSummary{},

		suiteEndings:           make(chan *types.SuiteSummary, 0),
		aggregatedSuiteEndings: []*types.SuiteSummary{},
	}

	go aggregator.mux()

	return aggregator
}

func (aggregator *Aggregator) SpecSuiteWillBegin(config config.GinkgoConfigType, summary *types.SuiteSummary) {
	aggregator.suiteBeginnings <- configAndSuite{config, summary}
}

func (aggregator *Aggregator) ExampleWillRun(exampleSummary *types.ExampleSummary) {
	//noop
}

func (aggregator *Aggregator) ExampleDidComplete(exampleSummary *types.ExampleSummary) {
	aggregator.exampleCompletions <- exampleSummary
}

func (aggregator *Aggregator) SpecSuiteDidEnd(summary *types.SuiteSummary) {
	aggregator.suiteEndings <- summary
}

func (aggregator *Aggregator) mux() {
loop:
	for {
		select {
		case configAndSuite := <-aggregator.suiteBeginnings:
			aggregator.registerSuiteBeginning(configAndSuite)
		case exampleSummary := <-aggregator.exampleCompletions:
			aggregator.registerExampleCompletion(exampleSummary)
		case suite := <-aggregator.suiteEndings:
			finished, passed := aggregator.registerSuiteEnding(suite)
			if finished {
				aggregator.result <- passed
				break loop
			}
		}
	}
}

func (aggregator *Aggregator) registerSuiteBeginning(configAndSuite configAndSuite) {
	aggregator.aggregatedSuiteBeginnings = append(aggregator.aggregatedSuiteBeginnings, configAndSuite)

	if len(aggregator.aggregatedSuiteBeginnings) == 1 {
		aggregator.startTime = time.Now()
	}

	if len(aggregator.aggregatedSuiteBeginnings) != aggregator.nodeCount {
		return
	}

	aggregator.stenographer.AnnounceSuite(configAndSuite.summary.SuiteDescription, configAndSuite.config.RandomSeed, configAndSuite.config.RandomizeAllSpecs)

	numberOfSpecsToRun := 0
	totalNumberOfSpecs := 0
	for _, configAndSuite := range aggregator.aggregatedSuiteBeginnings {
		numberOfSpecsToRun += configAndSuite.summary.NumberOfExamplesThatWillBeRun
		totalNumberOfSpecs += configAndSuite.summary.NumberOfTotalExamples
	}

	aggregator.stenographer.AnnounceNumberOfSpecs(numberOfSpecsToRun, totalNumberOfSpecs)
	aggregator.stenographer.AnnounceAggregatedParallelRun(aggregator.nodeCount)
	aggregator.flushCompletedExamples()
}

func (aggregator *Aggregator) registerExampleCompletion(exampleSummary *types.ExampleSummary) {
	aggregator.completedExamples = append(aggregator.completedExamples, exampleSummary)
	aggregator.flushCompletedExamples()
}

func (aggregator *Aggregator) flushCompletedExamples() {
	if len(aggregator.aggregatedSuiteBeginnings) != aggregator.nodeCount {
		return
	}

	for _, exampleSummary := range aggregator.completedExamples {
		aggregator.announceExample(exampleSummary)
	}

	aggregator.completedExamples = []*types.ExampleSummary{}
}

func (aggregator *Aggregator) announceExample(exampleSummary *types.ExampleSummary) {
	if aggregator.config.Verbose && exampleSummary.State != types.ExampleStatePending && exampleSummary.State != types.ExampleStateSkipped {
		aggregator.stenographer.AnnounceExampleWillRun(exampleSummary)
	}

	aggregator.stenographer.AnnounceCapturedOutput(exampleSummary)

	switch exampleSummary.State {
	case types.ExampleStatePassed:
		if exampleSummary.IsMeasurement {
			aggregator.stenographer.AnnounceSuccesfulMeasurement(exampleSummary, aggregator.config.Succinct)
		} else if exampleSummary.RunTime.Seconds() >= aggregator.config.SlowSpecThreshold {
			aggregator.stenographer.AnnounceSuccesfulSlowExample(exampleSummary, aggregator.config.Succinct)
		} else {
			aggregator.stenographer.AnnounceSuccesfulExample(exampleSummary)
		}
	case types.ExampleStatePending:
		aggregator.stenographer.AnnouncePendingExample(exampleSummary, aggregator.config.NoisyPendings, aggregator.config.Succinct)
	case types.ExampleStateSkipped:
		aggregator.stenographer.AnnounceSkippedExample(exampleSummary)
	case types.ExampleStateTimedOut:
		aggregator.stenographer.AnnounceExampleTimedOut(exampleSummary, aggregator.config.Succinct)
	case types.ExampleStatePanicked:
		aggregator.stenographer.AnnounceExamplePanicked(exampleSummary, aggregator.config.Succinct)
	case types.ExampleStateFailed:
		aggregator.stenographer.AnnounceExampleFailed(exampleSummary, aggregator.config.Succinct)
	}
}

func (aggregator *Aggregator) registerSuiteEnding(suite *types.SuiteSummary) (finished bool, passed bool) {
	aggregator.aggregatedSuiteEndings = append(aggregator.aggregatedSuiteEndings, suite)
	if len(aggregator.aggregatedSuiteEndings) < aggregator.nodeCount {
		return false, false
	}

	aggregatedSuiteSummary := &types.SuiteSummary{}
	aggregatedSuiteSummary.SuiteSucceeded = true

	for _, suiteSummary := range aggregator.aggregatedSuiteEndings {
		if suiteSummary.SuiteSucceeded == false {
			aggregatedSuiteSummary.SuiteSucceeded = false
		}

		aggregatedSuiteSummary.NumberOfExamplesThatWillBeRun += suiteSummary.NumberOfExamplesThatWillBeRun
		aggregatedSuiteSummary.NumberOfTotalExamples += suiteSummary.NumberOfTotalExamples
		aggregatedSuiteSummary.NumberOfPassedExamples += suiteSummary.NumberOfPassedExamples
		aggregatedSuiteSummary.NumberOfFailedExamples += suiteSummary.NumberOfFailedExamples
		aggregatedSuiteSummary.NumberOfPendingExamples += suiteSummary.NumberOfPendingExamples
		aggregatedSuiteSummary.NumberOfSkippedExamples += suiteSummary.NumberOfSkippedExamples
	}

	aggregatedSuiteSummary.RunTime = time.Since(aggregator.startTime)
	aggregator.stenographer.AnnounceSpecRunCompletion(aggregatedSuiteSummary)

	return true, aggregatedSuiteSummary.SuiteSucceeded
}
