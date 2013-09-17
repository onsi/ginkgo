package ginkgo

import (
	"fmt"
	"math"
	"time"
)

type example struct {
	subject exampleSubject
	focused bool

	containers []*containerNode

	state               ExampleState
	runTime             time.Duration
	sampleRunTimes      []time.Duration
	failure             ExampleFailure
	didInterceptFailure bool
	interceptedFailure  failureData
}

func newExample(subject exampleSubject) *example {
	ex := &example{
		subject: subject,
		focused: subject.getFlag() == flagTypeFocused,
	}

	if subject.getFlag() == flagTypePending {
		ex.state = ExampleStatePending
	}

	return ex
}

func (ex *example) addContainerNode(container *containerNode) {
	ex.containers = append([]*containerNode{container}, ex.containers...)
	if container.flag == flagTypeFocused {
		ex.focused = true
	} else if container.flag == flagTypePending {
		ex.state = ExampleStatePending
	}
}

func (ex *example) fail(failure failureData) {
	if !ex.didInterceptFailure {
		ex.interceptedFailure = failure
		ex.didInterceptFailure = true
	}
}

func (ex *example) skip() {
	ex.state = ExampleStateSkipped
}

func (ex *example) subjectComponentType() ExampleComponentType {
	if ex.subject.nodeType() == nodeTypeBenchmark {
		return ExampleComponentTypeBenchmark
	} else {
		return ExampleComponentTypeIt
	}
}

func (ex *example) desiredNumberOfSamples() int {
	if ex.subject.nodeType() == nodeTypeBenchmark {
		return ex.subject.(*benchmarkNode).samples
	}

	return 1
}

func (ex *example) failed() bool {
	return ex.state == ExampleStateFailed || ex.state == ExampleStatePanicked || ex.state == ExampleStateTimedOut
}

func (ex *example) skippedOrPending() bool {
	return ex.state == ExampleStateSkipped || ex.state == ExampleStatePending
}

func (ex *example) pending() bool {
	return ex.state == ExampleStatePending
}

func (ex *example) run() {
	startTime := time.Now()
	defer func() {
		ex.runTime = time.Since(startTime)
	}()

	ex.sampleRunTimes = make([]time.Duration, ex.desiredNumberOfSamples())

	for sample := 0; sample < ex.desiredNumberOfSamples(); sample++ {
		ex.state, ex.failure = ex.runSample(sample)

		if ex.state != ExampleStatePassed {
			return
		}
	}
}

func (ex *example) runSample(sample int) (exampleState ExampleState, exampleFailure ExampleFailure) {
	exampleState = ExampleStatePassed
	exampleFailure = ExampleFailure{}
	innerMostContainerIndexToUnwind := 0

	defer func() {
		if len(ex.containers) > 0 {
			for i := innerMostContainerIndexToUnwind; i >= 0; i-- {
				container := ex.containers[i]
				for _, afterEach := range container.afterEachNodes {
					outcome, failure := afterEach.run()
					afterEachState, afterEachFailure := ex.processOutcomeAndFailure(i, ExampleComponentTypeAfterEach, afterEach.codeLocation, outcome, failure)
					if afterEachState != ExampleStatePassed && exampleState == ExampleStatePassed {
						exampleState = afterEachState
						exampleFailure = afterEachFailure
					}
				}
			}
		}
	}()

	for i, container := range ex.containers {
		innerMostContainerIndexToUnwind = i
		for _, beforeEach := range container.beforeEachNodes {
			outcome, failure := beforeEach.run()
			exampleState, exampleFailure = ex.processOutcomeAndFailure(i, ExampleComponentTypeBeforeEach, beforeEach.codeLocation, outcome, failure)
			if exampleState != ExampleStatePassed {
				return
			}
		}
	}

	for i, container := range ex.containers {
		for _, justBeforeEach := range container.justBeforeEachNodes {
			outcome, failure := justBeforeEach.run()
			exampleState, exampleFailure = ex.processOutcomeAndFailure(i, ExampleComponentTypeJustBeforeEach, justBeforeEach.codeLocation, outcome, failure)
			if exampleState != ExampleStatePassed {
				return
			}
		}
	}

	sampleTime := time.Now()
	outcome, failure := ex.subject.run()
	ex.sampleRunTimes[sample] = time.Since(sampleTime)

	exampleState, exampleFailure = ex.processOutcomeAndFailure(len(ex.containers), ex.subjectComponentType(), ex.subject.getCodeLocation(), outcome, failure)
	if exampleState != ExampleStatePassed {
		return
	}

	if ex.subject.nodeType() == nodeTypeBenchmark {
		exampleState, exampleFailure = ex.processBenchmark(ex.sampleRunTimes[sample])
	}

	return
}

func (ex *example) processOutcomeAndFailure(containerIndex int, componentType ExampleComponentType, codeLocation CodeLocation, outcome runOutcome, failure failureData) (exampleState ExampleState, exampleFailure ExampleFailure) {
	exampleFailure = ExampleFailure{}
	exampleState = ExampleStatePassed

	if ex.didInterceptFailure {
		exampleState = ExampleStateFailed
		failure = ex.interceptedFailure
	} else if outcome == runOutcomePanicked {
		exampleState = ExampleStatePanicked
	} else if outcome == runOutcomeTimedOut {
		exampleState = ExampleStateTimedOut
	} else {
		return
	}

	exampleFailure = ExampleFailure{
		Message:               failure.message,
		Location:              failure.codeLocation,
		ForwardedPanic:        failure.forwardedPanic,
		ComponentIndex:        containerIndex,
		ComponentType:         componentType,
		ComponentCodeLocation: codeLocation,
	}

	return
}

func (ex *example) processBenchmark(sampleTime time.Duration) (exampleState ExampleState, exampleFailure ExampleFailure) {
	exampleFailure = ExampleFailure{}
	exampleState = ExampleStatePassed

	node := ex.subject.(*benchmarkNode)
	if sampleTime < node.maximumTime {
		return
	}

	exampleState = ExampleStateFailed
	message := fmt.Sprintf("Benchmark sample took: %.4fs\nThis exceeds the allowed maximum: %.4fs", sampleTime.Seconds(), node.maximumTime.Seconds())

	exampleFailure = ExampleFailure{
		Message:               message,
		Location:              node.getCodeLocation(),
		ComponentIndex:        len(ex.containers),
		ComponentType:         ExampleComponentTypeBenchmark,
		ComponentCodeLocation: node.getCodeLocation(),
	}

	return
}

func (ex *example) summary() *ExampleSummary {
	componentTexts := make([]string, len(ex.containers)+1)
	componentCodeLocations := make([]CodeLocation, len(ex.containers)+1)

	for i, container := range ex.containers {
		componentTexts[i] = container.text
		componentCodeLocations[i] = container.codeLocation
	}

	componentTexts[len(ex.containers)] = ex.subject.getText()
	componentCodeLocations[len(ex.containers)] = ex.subject.getCodeLocation()

	return &ExampleSummary{
		ComponentTexts:         componentTexts,
		ComponentCodeLocations: componentCodeLocations,
		State:     ex.state,
		RunTime:   ex.runTime,
		Failure:   ex.failure,
		Benchmark: ex.benchmarkReport(),
	}
}

func (ex *example) benchmarkReport() ExampleBenchmark {
	if ex.subject.nodeType() != nodeTypeBenchmark {
		return ExampleBenchmark{}
	}

	if ex.failed() {
		return ExampleBenchmark{
			IsBenchmark:     true,
			NumberOfSamples: len(ex.sampleRunTimes),
		}
	}

	max := time.Duration(math.MinInt64)
	min := time.Duration(math.MaxInt64)
	sum := time.Duration(0)
	sumOfSquares := time.Duration(0)

	for _, sample := range ex.sampleRunTimes {
		if sample > max {
			max = sample
		}
		if sample < min {
			min = sample
		}
		sum += sample
		sumOfSquares += sample * sample
	}

	n := float64(len(ex.sampleRunTimes))
	mean := time.Duration(float64(sum) / n)
	stdDev := time.Duration(math.Sqrt(float64(sumOfSquares)/n - float64(mean*mean)))

	return ExampleBenchmark{
		IsBenchmark:     true,
		NumberOfSamples: len(ex.sampleRunTimes),
		FastestTime:     min,
		SlowestTime:     max,
		AverageTime:     mean,
		StdDeviation:    stdDev,
	}
}

func (ex *example) concatenatedString() string {
	s := ""
	for _, container := range ex.containers {
		s += container.text + " "
	}

	return s + ex.subject.getText()
}
