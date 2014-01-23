package ginkgo

import (
	"github.com/onsi/ginkgo/types"
	"strings"
	"time"
)

type example struct {
	subject exampleSubject
	focused bool

	containers []*containerNode

	state               types.ExampleState
	runTime             time.Duration
	failure             types.ExampleFailure
	didInterceptFailure bool
	interceptedFailure  failureData
	exampleIndex        int
}

func newExample(subject exampleSubject) *example {
	ex := &example{
		subject: subject,
		focused: subject.getFlag() == flagTypeFocused,
	}

	if subject.getFlag() == flagTypePending {
		ex.state = types.ExampleStatePending
	}

	return ex
}

func (ex *example) addContainerNode(container *containerNode) {
	ex.containers = append([]*containerNode{container}, ex.containers...)
	if container.flag == flagTypeFocused {
		ex.focused = true
	} else if container.flag == flagTypePending {
		ex.state = types.ExampleStatePending
	}
}

func (ex *example) fail(failure failureData) {
	if !ex.didInterceptFailure {
		ex.interceptedFailure = failure
		ex.didInterceptFailure = true
	}
}

func (ex *example) skip() {
	ex.state = types.ExampleStateSkipped
}

func (ex *example) subjectComponentType() types.ExampleComponentType {
	if ex.subject.nodeType() == nodeTypeMeasure {
		return types.ExampleComponentTypeMeasure
	} else {
		return types.ExampleComponentTypeIt
	}
}

func (ex *example) desiredNumberOfSamples() int {
	if ex.subject.nodeType() == nodeTypeMeasure {
		return ex.subject.(*measureNode).samples
	}

	return 1
}

func (ex *example) failed() bool {
	return ex.state == types.ExampleStateFailed || ex.state == types.ExampleStatePanicked || ex.state == types.ExampleStateTimedOut
}

func (ex *example) skippedOrPending() bool {
	return ex.state == types.ExampleStateSkipped || ex.state == types.ExampleStatePending
}

func (ex *example) pending() bool {
	return ex.state == types.ExampleStatePending
}

func (ex *example) run() {
	startTime := time.Now()
	defer func() {
		ex.runTime = time.Since(startTime)
	}()

	for sample := 0; sample < ex.desiredNumberOfSamples(); sample++ {
		ex.state, ex.failure = ex.runSample(sample)

		if ex.state != types.ExampleStatePassed {
			return
		}
	}
}

func (ex *example) runSample(sample int) (exampleState types.ExampleState, exampleFailure types.ExampleFailure) {
	exampleState = types.ExampleStatePassed
	exampleFailure = types.ExampleFailure{}
	innerMostContainerIndexToUnwind := 0

	defer func() {
		if len(ex.containers) > 0 {
			for i := innerMostContainerIndexToUnwind; i >= 0; i-- {
				container := ex.containers[i]
				for _, afterEach := range container.afterEachNodes {
					outcome, failure := afterEach.run()
					afterEachState, afterEachFailure := ex.processOutcomeAndFailure(i, types.ExampleComponentTypeAfterEach, afterEach.codeLocation, outcome, failure)
					if afterEachState != types.ExampleStatePassed && exampleState == types.ExampleStatePassed {
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
			exampleState, exampleFailure = ex.processOutcomeAndFailure(i, types.ExampleComponentTypeBeforeEach, beforeEach.codeLocation, outcome, failure)
			if exampleState != types.ExampleStatePassed {
				return
			}
		}
	}

	for i, container := range ex.containers {
		for _, justBeforeEach := range container.justBeforeEachNodes {
			outcome, failure := justBeforeEach.run()
			exampleState, exampleFailure = ex.processOutcomeAndFailure(i, types.ExampleComponentTypeJustBeforeEach, justBeforeEach.codeLocation, outcome, failure)
			if exampleState != types.ExampleStatePassed {
				return
			}
		}
	}

	outcome, failure := ex.subject.run()
	exampleState, exampleFailure = ex.processOutcomeAndFailure(len(ex.containers), ex.subjectComponentType(), ex.subject.getCodeLocation(), outcome, failure)

	if exampleState != types.ExampleStatePassed {
		return
	}

	return
}

func (ex *example) processOutcomeAndFailure(containerIndex int, componentType types.ExampleComponentType, codeLocation types.CodeLocation, outcome runOutcome, failure failureData) (exampleState types.ExampleState, exampleFailure types.ExampleFailure) {
	exampleFailure = types.ExampleFailure{}
	exampleState = types.ExampleStatePassed

	if ex.didInterceptFailure {
		exampleState = types.ExampleStateFailed
		failure = ex.interceptedFailure
	} else if outcome == runOutcomePanicked {
		exampleState = types.ExampleStatePanicked
	} else if outcome == runOutcomeTimedOut {
		exampleState = types.ExampleStateTimedOut
	} else {
		return
	}

	exampleFailure = types.ExampleFailure{
		Message:               failure.message,
		Location:              failure.codeLocation,
		ForwardedPanic:        failure.forwardedPanic,
		ComponentIndex:        containerIndex,
		ComponentType:         componentType,
		ComponentCodeLocation: codeLocation,
	}

	return
}

func (ex *example) summary(suiteID string) *types.ExampleSummary {
	componentTexts := make([]string, len(ex.containers)+1)
	componentCodeLocations := make([]types.CodeLocation, len(ex.containers)+1)

	for i, container := range ex.containers {
		componentTexts[i] = container.text
		componentCodeLocations[i] = container.codeLocation
	}

	componentTexts[len(ex.containers)] = ex.subject.getText()
	componentCodeLocations[len(ex.containers)] = ex.subject.getCodeLocation()

	return &types.ExampleSummary{
		IsMeasurement:          ex.subjectComponentType() == types.ExampleComponentTypeMeasure,
		NumberOfSamples:        ex.desiredNumberOfSamples(),
		ComponentTexts:         componentTexts,
		ComponentCodeLocations: componentCodeLocations,
		State:        ex.state,
		RunTime:      ex.runTime,
		Failure:      ex.failure,
		Measurements: ex.measurementsReport(),
		SuiteID:      suiteID,
		ExampleIndex: ex.exampleIndex,
	}
}

func (ex *example) ginkgoTestDescription() GinkgoTestDescription {
	summary := ex.summary("")

	leafCodeLocation := summary.ComponentCodeLocations[len(summary.ComponentCodeLocations)-1]

	return GinkgoTestDescription{
		ComponentTexts: summary.ComponentTexts[1:],
		FullTestText:   strings.Join(summary.ComponentTexts[1:], " "),
		TestText:       summary.ComponentTexts[len(summary.ComponentTexts)-1],
		IsMeasurement:  summary.IsMeasurement,
		FileName:       leafCodeLocation.FileName,
		LineNumber:     leafCodeLocation.LineNumber,
	}
}

func (ex *example) measurementsReport() (measurements map[string]*types.ExampleMeasurement) {
	measurements = map[string]*types.ExampleMeasurement{}
	if ex.subjectComponentType() != types.ExampleComponentTypeMeasure {
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
