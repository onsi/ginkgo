package ginkgo

import (
	"time"
)

type example struct {
	it             *itNode
	hasPendingFlag bool
	hasFocusFlag   bool

	containers []*containerNode
	skipped    bool

	outcome            runOutcome
	runTime            time.Duration
	failure            ExampleFailure
	interceptedFailure failureData
}

func newExample(it *itNode) *example {
	return &example{
		it:             it,
		hasPendingFlag: it.flag == flagTypePending,
		hasFocusFlag:   it.flag == flagTypeFocused,
	}
}

func (ex *example) addContainerNode(container *containerNode) {
	ex.containers = append([]*containerNode{container}, ex.containers...)
	if container.flag == flagTypeFocused {
		ex.hasFocusFlag = true
	} else if container.flag == flagTypePending {
		ex.hasPendingFlag = true
	}
}

func (ex *example) skip() {
	ex.skipped = true
}

func (ex *example) fail(failure failureData) {
	empty := failureData{}
	if ex.interceptedFailure == empty {
		ex.interceptedFailure = failure
	}
}

func (ex *example) failed() bool {
	return ex.outcome == runOutcomeFailed || ex.outcome == runOutcomePanicked || ex.outcome == runOutcomeTimedOut
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

func (ex *example) handleOutcomeAndFailure(containerIndex int, componentType ExampleComponentType, codeLocation CodeLocation, outcome runOutcome, failure failureData) (didFail bool) {
	empty := failureData{}
	if ex.interceptedFailure != empty {
		ex.outcome = runOutcomeFailed
		ex.failure = ExampleFailure{
			Message:               ex.interceptedFailure.message,
			Location:              ex.interceptedFailure.codeLocation,
			ForwardedPanic:        ex.interceptedFailure.forwardedPanic,
			ComponentIndex:        containerIndex,
			ComponentType:         componentType,
			ComponentCodeLocation: codeLocation,
		}
		return true
	} else if outcome != runOutcomePassed {
		ex.outcome = outcome
		ex.failure = ExampleFailure{
			Message:               failure.message,
			Location:              failure.codeLocation,
			ForwardedPanic:        failure.forwardedPanic,
			ComponentIndex:        containerIndex,
			ComponentType:         componentType,
			ComponentCodeLocation: codeLocation,
		}
		return true
	} else {
		ex.outcome = runOutcomePassed
	}
	return false
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

	var state ExampleState
	if ex.skipped {
		state = ExampleStateSkipped
	} else if ex.hasPendingFlag {
		state = ExampleStatePending
	} else if ex.outcome == runOutcomeFailed {
		state = ExampleStateFailed
	} else if ex.outcome == runOutcomePassed {
		state = ExampleStatePassed
	} else if ex.outcome == runOutcomePanicked {
		state = ExampleStatePanicked
	} else if ex.outcome == runOutcomeTimedOut {
		state = ExampleStateTimedOut
	}

	return &ExampleSummary{
		ComponentTexts:         componentTexts,
		ComponentCodeLocations: componentCodeLocations,
		State:   state,
		RunTime: ex.runTime,
		Failure: ex.failure,
	}
}
