package ginkgo

import (
	"time"
)

type Reporter interface {
	SpecSuiteWillBegin(config GinkgoConfigType, summary *SuiteSummary)
	ExampleDidComplete(exampleSummary *ExampleSummary)
	SpecSuiteDidEnd(summary *SuiteSummary)
}

type SuiteSummary struct {
	SuiteDescription string

	NumberOfExamplesBeforeParallelization int
	NumberOfTotalExamples                 int
	NumberOfExamplesThatWillBeRun         int
	NumberOfPendingExamples               int
	NumberOfSkippedExamples               int
	NumberOfPassedExamples                int
	NumberOfFailedExamples                int
	RunTime                               time.Duration
}

type ExampleSummary struct {
	ComponentTexts         []string
	ComponentCodeLocations []CodeLocation

	State   ExampleState
	RunTime time.Duration
	Failure ExampleFailure
}

type ExampleFailure struct {
	Message        string
	Location       CodeLocation
	ForwardedPanic interface{}

	ComponentIndex        int
	ComponentType         ExampleComponentType
	ComponentCodeLocation CodeLocation
}
