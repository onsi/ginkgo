package godescribe

import (
	"time"
)

type Reporter interface {
	RandomizationStrategy(randomSeed int64, randomizeAllExamples bool)
	SpecSuiteWillBegin(summary *SuiteSummary)
	ExampleDidComplete(exampleSummary *ExampleSummary)
	SpecSuiteDidEnd(summary *SuiteSummary)
}

type SuiteSummary struct {
	SuiteDescription string

	NumberOfTotalExamples         int
	NumberOfExamplesThatWillBeRun int
	NumberOfPendingExamples       int
	NumberOfSkippedExamples       int
	NumberOfPassedExamples        int
	NumberOfFailedExamples        int
	RunTime                       time.Duration
}

type ExampleSummary struct {
	ComponentTexts         []string
	ComponentCodeLocations []CodeLocation

	State   ExampleState
	RunTime time.Duration
	Failure ExampleFailure
}

type ExampleState uint

const (
	ExampleStateInvalid ExampleState = iota

	ExampleStatePending
	ExampleStateSkipped
	ExampleStatePassed
	ExampleStateFailed
	ExampleStatePanicked
	ExampleStateTimedOut
)

type ExampleFailure struct {
	Message        string
	Location       CodeLocation
	ForwardedPanic interface{}

	ComponentIndex        int
	ComponentType         ExampleComponentType
	ComponentCodeLocation CodeLocation
}

type ExampleComponentType uint

const (
	ExampleComponentTypeInvalid ExampleComponentType = iota

	ExampleComponentTypeBeforeEach
	ExampleComponentTypeJustBeforeEach
	ExampleComponentTypeAfterEach
	ExampleComponentTypeIt
)
