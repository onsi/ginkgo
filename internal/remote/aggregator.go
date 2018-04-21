/*

Aggregator is a reporter used by the Ginkgo CLI to aggregate and present parallel test output
coherently as tests complete.  You shouldn't need to use this in your code.  To run tests in parallel:

	ginkgo -nodes=N

where N is the number of nodes you desire.
*/
package remote

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/reporters/stenographer"
	"github.com/onsi/ginkgo/types"
)

type configAndSuite struct {
	config  config.GinkgoConfigType
	summary *types.SuiteSummary
}

type currentRunningSpec struct {
	Spec               *types.SpecSummary `json:"spec"`
	StartTime          time.Time          `json:"startTime"`
	CapturedOutput     string             `json:"output"`
	GinkgoWriterOutput string             `json:"ginkgoWriterOutput"`
}

type nodeStatus struct {
	Node      int                 `json:"node"`
	IsRunning bool                `json:"isRunning"`
	StartTime time.Time           `json:"startTime"`
	Spec      *currentRunningSpec `json:"spec"`
}

func (node *nodeStatus) startSpec(specSummary *types.SpecSummary) {
	node.Spec = &currentRunningSpec{
		Spec:      specSummary,
		StartTime: time.Now(),
	}
}

func (node *nodeStatus) updateSpecDebugOutput(debugOutput *types.ParallelSpecDebugOutput) {
	if node.Spec != nil {
		node.Spec.CapturedOutput = debugOutput.CapturedOutput
		node.Spec.GinkgoWriterOutput = debugOutput.GinkgoWriterOutput
	}
}

func (node *nodeStatus) lastSpecIsDone() {
	node.Spec = nil
}

type Aggregator struct {
	nodeCount    int
	config       config.DefaultReporterConfigType
	stenographer stenographer.Stenographer
	result       chan bool

	aggregatedSuiteBeginnings []configAndSuite
	aggregatedBeforeSuites    []*types.SetupSummary
	aggregatedAfterSuites     []*types.SetupSummary
	completedSpecs            []*types.SpecSummary
	aggregatedSuiteEndings    []*types.SuiteSummary
	specs                     []*types.SpecSummary
	nodeStatus                map[int]*nodeStatus

	lock *sync.Mutex

	startTime time.Time
}

func NewAggregator(nodeCount int, result chan bool, config config.DefaultReporterConfigType, stenographer stenographer.Stenographer) *Aggregator {
	aggregator := &Aggregator{
		nodeCount:    nodeCount,
		result:       result,
		config:       config,
		stenographer: stenographer,

		nodeStatus: map[int]*nodeStatus{},
		lock:       &sync.Mutex{},
	}

	return aggregator
}

func (aggregator *Aggregator) SpecSuiteWillBegin(config config.GinkgoConfigType, summary *types.SuiteSummary) {
	aggregator.lock.Lock()
	defer aggregator.lock.Unlock()

	aggregator.nodeStatus[summary.GinkgoNode] = &nodeStatus{
		Node:      summary.GinkgoNode,
		IsRunning: true,
		StartTime: time.Now(),
	}

	cs := configAndSuite{config, summary}

	aggregator.aggregatedSuiteBeginnings = append(aggregator.aggregatedSuiteBeginnings, cs)

	if len(aggregator.aggregatedSuiteBeginnings) == 1 {
		aggregator.startTime = time.Now()
	}

	if len(aggregator.aggregatedSuiteBeginnings) != aggregator.nodeCount {
		return
	}

	aggregator.stenographer.AnnounceSuite(cs.summary.SuiteDescription, cs.config.RandomSeed, cs.config.RandomizeAllSpecs, aggregator.config.Succinct)

	totalNumberOfSpecs := 0
	if len(aggregator.aggregatedSuiteBeginnings) > 0 {
		totalNumberOfSpecs = cs.summary.NumberOfSpecsBeforeParallelization
	}

	aggregator.stenographer.AnnounceTotalNumberOfSpecs(totalNumberOfSpecs, aggregator.config.Succinct)
	aggregator.stenographer.AnnounceAggregatedParallelRun(aggregator.nodeCount, aggregator.config.Succinct)
	aggregator.flushCompletedSpecs()

}

func (aggregator *Aggregator) BeforeSuiteDidRun(setupSummary *types.SetupSummary) {
	aggregator.lock.Lock()
	defer aggregator.lock.Unlock()

	aggregator.aggregatedBeforeSuites = append(aggregator.aggregatedBeforeSuites, setupSummary)
	aggregator.flushCompletedSpecs()
}

func (aggregator *Aggregator) AfterSuiteDidRun(setupSummary *types.SetupSummary) {
	aggregator.lock.Lock()
	defer aggregator.lock.Unlock()

	aggregator.aggregatedAfterSuites = append(aggregator.aggregatedAfterSuites, setupSummary)
	aggregator.flushCompletedSpecs()
}

func (aggregator *Aggregator) SpecWillRun(specSummary *types.SpecSummary) {
	aggregator.lock.Lock()
	defer aggregator.lock.Unlock()

	aggregator.nodeStatus[specSummary.GinkgoNode].startSpec(specSummary)
}

func (aggregator *Aggregator) UpdateSpecDebugOutput(debugOutput *types.ParallelSpecDebugOutput) {
	aggregator.lock.Lock()
	defer aggregator.lock.Unlock()

	aggregator.nodeStatus[debugOutput.GinkgoNode].updateSpecDebugOutput(debugOutput)
}

func (aggregator *Aggregator) SpecDidComplete(specSummary *types.SpecSummary) {
	aggregator.lock.Lock()
	defer aggregator.lock.Unlock()

	aggregator.nodeStatus[specSummary.GinkgoNode].lastSpecIsDone()

	aggregator.completedSpecs = append(aggregator.completedSpecs, specSummary)
	aggregator.specs = append(aggregator.specs, specSummary)
	aggregator.flushCompletedSpecs()
}

func (aggregator *Aggregator) SpecSuiteDidEnd(summary *types.SuiteSummary) {
	aggregator.lock.Lock()
	defer aggregator.lock.Unlock()

	aggregator.nodeStatus[summary.GinkgoNode].IsRunning = false

	aggregator.aggregatedSuiteEndings = append(aggregator.aggregatedSuiteEndings, summary)
	if len(aggregator.aggregatedSuiteEndings) < aggregator.nodeCount {
		return
	}

	aggregatedSuiteSummary := &types.SuiteSummary{}
	aggregatedSuiteSummary.SuiteSucceeded = true

	for _, suiteSummary := range aggregator.aggregatedSuiteEndings {
		if suiteSummary.SuiteSucceeded == false {
			aggregatedSuiteSummary.SuiteSucceeded = false
		}

		aggregatedSuiteSummary.NumberOfSpecsThatWillBeRun += suiteSummary.NumberOfSpecsThatWillBeRun
		aggregatedSuiteSummary.NumberOfTotalSpecs += suiteSummary.NumberOfTotalSpecs
		aggregatedSuiteSummary.NumberOfPassedSpecs += suiteSummary.NumberOfPassedSpecs
		aggregatedSuiteSummary.NumberOfFailedSpecs += suiteSummary.NumberOfFailedSpecs
		aggregatedSuiteSummary.NumberOfPendingSpecs += suiteSummary.NumberOfPendingSpecs
		aggregatedSuiteSummary.NumberOfSkippedSpecs += suiteSummary.NumberOfSkippedSpecs
		aggregatedSuiteSummary.NumberOfFlakedSpecs += suiteSummary.NumberOfFlakedSpecs
	}

	aggregatedSuiteSummary.RunTime = time.Since(aggregator.startTime)

	aggregator.stenographer.SummarizeFailures(aggregator.specs)
	aggregator.stenographer.AnnounceSpecRunCompletion(aggregatedSuiteSummary, aggregator.config.Succinct)

	aggregator.result <- aggregatedSuiteSummary.SuiteSucceeded
}

func (aggregator *Aggregator) DebugReport() []byte {
	aggregator.lock.Lock()
	defer aggregator.lock.Unlock()

	encoded, _ := json.Marshal(aggregator.nodeStatus)
	return encoded
}

func (aggregator *Aggregator) flushCompletedSpecs() {
	if len(aggregator.aggregatedSuiteBeginnings) != aggregator.nodeCount {
		return
	}

	for _, setupSummary := range aggregator.aggregatedBeforeSuites {
		aggregator.announceBeforeSuite(setupSummary)
	}

	for _, specSummary := range aggregator.completedSpecs {
		aggregator.announceSpec(specSummary)
	}

	for _, setupSummary := range aggregator.aggregatedAfterSuites {
		aggregator.announceAfterSuite(setupSummary)
	}

	aggregator.aggregatedBeforeSuites = []*types.SetupSummary{}
	aggregator.completedSpecs = []*types.SpecSummary{}
	aggregator.aggregatedAfterSuites = []*types.SetupSummary{}
}

func (aggregator *Aggregator) announceBeforeSuite(setupSummary *types.SetupSummary) {
	aggregator.stenographer.AnnounceCapturedOutput(setupSummary.CapturedOutput)
	if setupSummary.State != types.SpecStatePassed {
		aggregator.stenographer.AnnounceBeforeSuiteFailure(setupSummary, aggregator.config.Succinct, aggregator.config.FullTrace)
	}
}

func (aggregator *Aggregator) announceAfterSuite(setupSummary *types.SetupSummary) {
	aggregator.stenographer.AnnounceCapturedOutput(setupSummary.CapturedOutput)
	if setupSummary.State != types.SpecStatePassed {
		aggregator.stenographer.AnnounceAfterSuiteFailure(setupSummary, aggregator.config.Succinct, aggregator.config.FullTrace)
	}
}

func (aggregator *Aggregator) announceSpec(specSummary *types.SpecSummary) {
	if aggregator.config.Verbose && specSummary.State != types.SpecStatePending && specSummary.State != types.SpecStateSkipped {
		aggregator.stenographer.AnnounceSpecWillRun(specSummary)
	}

	aggregator.stenographer.AnnounceCapturedOutput(specSummary.CapturedOutput)

	switch specSummary.State {
	case types.SpecStatePassed:
		if specSummary.IsMeasurement {
			aggregator.stenographer.AnnounceSuccesfulMeasurement(specSummary, aggregator.config.Succinct)
		} else if specSummary.RunTime.Seconds() >= aggregator.config.SlowSpecThreshold {
			aggregator.stenographer.AnnounceSuccesfulSlowSpec(specSummary, aggregator.config.Succinct)
		} else {
			aggregator.stenographer.AnnounceSuccesfulSpec(specSummary)
		}

	case types.SpecStatePending:
		aggregator.stenographer.AnnouncePendingSpec(specSummary, aggregator.config.NoisyPendings && !aggregator.config.Succinct)
	case types.SpecStateSkipped:
		aggregator.stenographer.AnnounceSkippedSpec(specSummary, aggregator.config.Succinct || !aggregator.config.NoisySkippings, aggregator.config.FullTrace)
	case types.SpecStateTimedOut:
		aggregator.stenographer.AnnounceSpecTimedOut(specSummary, aggregator.config.Succinct, aggregator.config.FullTrace)
	case types.SpecStatePanicked:
		aggregator.stenographer.AnnounceSpecPanicked(specSummary, aggregator.config.Succinct, aggregator.config.FullTrace)
	case types.SpecStateFailed:
		aggregator.stenographer.AnnounceSpecFailed(specSummary, aggregator.config.Succinct, aggregator.config.FullTrace)
	}
}
