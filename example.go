package ginkgo

import (
	"time"
)

type example struct {
	it      *itNode
	focused bool

	containers []*containerNode

	state               ExampleState
	runTime             time.Duration
	failure             ExampleFailure
	didInterceptFailure bool
	interceptedFailure  failureData
}

func newExample(it *itNode) *example {
	ex := &example{
		it:      it,
		focused: it.flag == flagTypeFocused,
	}

	if it.flag == flagTypePending {
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

	outcome, failure := ex.it.run()
	if ex.handleOutcomeAndFailure(len(ex.containers), ExampleComponentTypeIt, ex.it.codeLocation, outcome, failure) {
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

func (ex *example) summary() *ExampleSummary {
	componentTexts := make([]string, len(ex.containers)+1)
	componentCodeLocations := make([]CodeLocation, len(ex.containers)+1)

	for i, container := range ex.containers {
		componentTexts[i] = container.text
		componentCodeLocations[i] = container.codeLocation
	}

	componentTexts[len(ex.containers)] = ex.it.text
	componentCodeLocations[len(ex.containers)] = ex.it.codeLocation

	return &ExampleSummary{
		ComponentTexts:         componentTexts,
		ComponentCodeLocations: componentCodeLocations,
		State:   ex.state,
		RunTime: ex.runTime,
		Failure: ex.failure,
	}
}

func (ex *example) concatenatedString() string {
	s := ""
	for _, container := range ex.containers {
		s += container.text + " "
	}

	return s + ex.it.text
}
