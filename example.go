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

func (ex *example) run() {
	startTime := time.Now()
	defer func() {
		ex.runTime = time.Since(startTime)
	}()

	desiredSamples := 1
	if ex.subject.nodeType() == nodeTypeBenchmark {
		desiredSamples = ex.subject.(*benchmarkNode).samples
	}
	ex.sampleRunTimes = make([]time.Duration, desiredSamples)

	for sample := 0; sample < desiredSamples; sample++ {
		for i, container := range ex.containers {
			for _, beforeEach := range container.beforeEachNodes {
				outcome, failure := beforeEach.run()
				if ex.handleOutcomeAndFailure(i, ExampleComponentTypeBeforeEach, beforeEach.codeLocation, outcome, failure) {
					return
				}
			}
		}

		for i, container := range ex.containers {
			for _, justBeforeEach := range container.justBeforeEachNodes {
				outcome, failure := justBeforeEach.run()
				if ex.handleOutcomeAndFailure(i, ExampleComponentTypeJustBeforeEach, justBeforeEach.codeLocation, outcome, failure) {
					return
				}
			}
		}

		sampleTime := time.Now()
		outcome, failure := ex.subject.run()
		ex.sampleRunTimes[sample] = time.Since(sampleTime)
		if ex.handleOutcomeAndFailure(len(ex.containers), ex.subjectComponentType(), ex.subject.getCodeLocation(), outcome, failure) {
			return
		}
		if ex.handleBenchmarkFailure(ex.sampleRunTimes[sample]) {
			return
		}

		for i := len(ex.containers) - 1; i >= 0; i-- {
			container := ex.containers[i]
			for j := len(container.afterEachNodes) - 1; j >= 0; j-- {
				outcome, failure := container.afterEachNodes[j].run()
				if ex.handleOutcomeAndFailure(i, ExampleComponentTypeAfterEach, container.afterEachNodes[j].codeLocation, outcome, failure) {
					return
				}
			}
		}
	}
}

func (ex *example) subjectComponentType() ExampleComponentType {
	if ex.subject.nodeType() == nodeTypeBenchmark {
		return ExampleComponentTypeBenchmark
	} else {
		return ExampleComponentTypeIt
	}
}

func (ex *example) failed() bool {
	return ex.state == ExampleStateFailed || ex.state == ExampleStatePanicked || ex.state == ExampleStateTimedOut
}

func (ex *example) skippedOrPending() bool {
	return ex.state == ExampleStateSkipped || ex.state == ExampleStatePending
}

func (ex *example) handleOutcomeAndFailure(containerIndex int, componentType ExampleComponentType, codeLocation CodeLocation, outcome runOutcome, failure failureData) (didFail bool) {
	if ex.didInterceptFailure {
		ex.state = ExampleStateFailed
		failure = ex.interceptedFailure
		didFail = true
	} else if outcome == runOutcomePanicked {
		ex.state = ExampleStatePanicked
		didFail = true
	} else if outcome == runOutcomeTimedOut {
		ex.state = ExampleStateTimedOut
		didFail = true
	} else {
		ex.state = ExampleStatePassed
		didFail = false
	}

	if didFail {
		ex.failure = ExampleFailure{
			Message:               failure.message,
			Location:              failure.codeLocation,
			ForwardedPanic:        failure.forwardedPanic,
			ComponentIndex:        containerIndex,
			ComponentType:         componentType,
			ComponentCodeLocation: codeLocation,
		}
	}

	return didFail
}

func (ex *example) handleBenchmarkFailure(sampleTime time.Duration) (didFail bool) {
	if ex.subject.nodeType() != nodeTypeBenchmark {
		return false
	}

	node := ex.subject.(*benchmarkNode)
	if sampleTime < node.maximumTime {
		return false
	}

	ex.state = ExampleStateFailed
	message := fmt.Sprintf("Benchmark sample took: %.4fs\nThis exceeds the allowed maximum: %.4fs", sampleTime.Seconds(), node.maximumTime.Seconds())

	ex.failure = ExampleFailure{
		Message:               message,
		Location:              node.getCodeLocation(),
		ComponentIndex:        len(ex.containers),
		ComponentType:         ExampleComponentTypeBenchmark,
		ComponentCodeLocation: node.getCodeLocation(),
	}

	return true
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
		Benchmark: ex.benchmark(),
	}
}

func (ex *example) benchmark() ExampleBenchmark {
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
