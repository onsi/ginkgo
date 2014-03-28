package example

import (
	"github.com/onsi/ginkgo/internal/containernode"
	"github.com/onsi/ginkgo/internal/leafnodes"
	"github.com/onsi/ginkgo/internal/types"
	"github.com/onsi/ginkgo/types"
	"time"
)

type Example struct {
	subject internaltypes.SubjectNode
	focused bool

	containers []*containernode.ContainerNode

	state   types.ExampleState
	runTime time.Duration
	failure types.ExampleFailure
}

func New(subject internaltypes.SubjectNode, containers []*containernode.ContainerNode) *Example {
	ex := &Example{
		subject:    subject,
		containers: containers,
		focused:    subject.Flag() == internaltypes.FlagTypeFocused,
	}

	ex.processFlag(subject.Flag())
	for i := len(containers) - 1; i >= 0; i-- {
		ex.processFlag(containers[i].Flag())
	}

	return ex
}

func (ex *Example) processFlag(flag internaltypes.FlagType) {
	if flag == internaltypes.FlagTypeFocused {
		ex.focused = true
	} else if flag == internaltypes.FlagTypePending {
		ex.state = types.ExampleStatePending
	}
}

func (ex *Example) Skip() {
	ex.state = types.ExampleStateSkipped
}

func (ex *Example) Failed() bool {
	return ex.state == types.ExampleStateFailed || ex.state == types.ExampleStatePanicked || ex.state == types.ExampleStateTimedOut
}

func (ex *Example) Passed() bool {
	return ex.state == types.ExampleStatePassed
}

func (ex *Example) Pending() bool {
	return ex.state == types.ExampleStatePending
}

func (ex *Example) Skipped() bool {
	return ex.state == types.ExampleStateSkipped
}

func (ex *Example) Focused() bool {
	return ex.focused
}

func (ex *Example) Run() {
	startTime := time.Now()
	defer func() {
		ex.runTime = time.Since(startTime)
	}()

	for sample := 0; sample < ex.subject.Samples(); sample++ {
		ex.state, ex.failure = ex.runSample(sample)

		if ex.state != types.ExampleStatePassed {
			return
		}
	}
}

func (ex *Example) runSample(sample int) (exampleState types.ExampleState, exampleFailure types.ExampleFailure) {
	exampleState = types.ExampleStatePassed
	exampleFailure = types.ExampleFailure{}
	innerMostContainerIndexToUnwind := -1

	defer func() {
		for i := innerMostContainerIndexToUnwind; i >= 0; i-- {
			container := ex.containers[i]
			for _, afterEach := range container.AfterEachNodes() {
				afterEachState, afterEachFailure := afterEach.Run()
				if afterEachState != types.ExampleStatePassed && exampleState == types.ExampleStatePassed {
					exampleState = afterEachState
					exampleFailure = afterEachFailure
				}
			}
		}
	}()

	for i, container := range ex.containers {
		innerMostContainerIndexToUnwind = i
		for _, beforeEach := range container.BeforeEachNodes() {
			exampleState, exampleFailure = beforeEach.Run()
			if exampleState != types.ExampleStatePassed {
				return
			}
		}
	}

	for _, container := range ex.containers {
		for _, justBeforeEach := range container.JustBeforeEachNodes() {
			exampleState, exampleFailure = justBeforeEach.Run()
			if exampleState != types.ExampleStatePassed {
				return
			}
		}
	}

	exampleState, exampleFailure = ex.subject.Run()

	if exampleState != types.ExampleStatePassed {
		return
	}

	return
}

func (ex *Example) IsMeasurement() bool {
	return ex.subject.Type() == types.ExampleComponentTypeMeasure
}

func (ex *Example) Summary(suiteID string) *types.ExampleSummary {
	componentTexts := make([]string, len(ex.containers)+1)
	componentCodeLocations := make([]types.CodeLocation, len(ex.containers)+1)

	for i, container := range ex.containers {
		componentTexts[i] = container.Text()
		componentCodeLocations[i] = container.CodeLocation()
	}

	componentTexts[len(ex.containers)] = ex.subject.Text()
	componentCodeLocations[len(ex.containers)] = ex.subject.CodeLocation()

	return &types.ExampleSummary{
		IsMeasurement:          ex.IsMeasurement(),
		NumberOfSamples:        ex.subject.Samples(),
		ComponentTexts:         componentTexts,
		ComponentCodeLocations: componentCodeLocations,
		State:        ex.state,
		RunTime:      ex.runTime,
		Failure:      ex.failure,
		Measurements: ex.measurementsReport(),
		SuiteID:      suiteID,
	}
}

func (ex *Example) measurementsReport() map[string]*types.ExampleMeasurement {
	if !ex.IsMeasurement() || ex.Failed() {
		return map[string]*types.ExampleMeasurement{}
	}

	return ex.subject.(*leafnodes.MeasureNode).MeasurementsReport()
}

func (ex *Example) ConcatenatedString() string {
	s := ""
	for _, container := range ex.containers {
		s += container.Text() + " "
	}

	return s + ex.subject.Text()
}
