package ginkgo

import (
	"time"
)

type example struct {
	subject exampleSubject
	focused bool

	containers []*containerNode

	state               ExampleState
	runTime             time.Duration
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
	if ex.subject.nodeType() == nodeTypeMeasure {
		return ExampleComponentTypeMeasure
	} else {
		return ExampleComponentTypeIt
	}
}

func (ex *example) desiredNumberOfSamples() int {
	if ex.subject.nodeType() == nodeTypeMeasure {
		return ex.subject.(*measureNode).samples
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

	outcome, failure := ex.subject.run()
	exampleState, exampleFailure = ex.processOutcomeAndFailure(len(ex.containers), ex.subjectComponentType(), ex.subject.getCodeLocation(), outcome, failure)

	if exampleState != ExampleStatePassed {
		return
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
		IsMeasurement:          ex.subjectComponentType() == ExampleComponentTypeMeasure,
		NumberOfSamples:        ex.desiredNumberOfSamples(),
		ComponentTexts:         componentTexts,
		ComponentCodeLocations: componentCodeLocations,
		State:        ex.state,
		RunTime:      ex.runTime,
		Failure:      ex.failure,
		Measurements: ex.measurementsReport(),
	}
}

func (ex *example) measurementsReport() (measurements map[string]*ExampleMeasurement) {
	if ex.subjectComponentType() != ExampleComponentTypeMeasure {
		return
	}
	if ex.failed() {
		return
	}

	return ex.subject.(*measureNode).measurementsReport()
}

func (ex *example) concatenatedString() string {
	s := ""
	for _, container := range ex.containers {
		s += container.text + " "
	}

	return s + ex.subject.getText()
}
