package ginkgo

import (
	"github.com/onsi/ginkgo/config"

	"time"
)

type Reporter interface {
	SpecSuiteWillBegin(config config.GinkgoConfigType, summary *SuiteSummary)
	ExampleDidComplete(exampleSummary *ExampleSummary)
	SpecSuiteDidEnd(summary *SuiteSummary)
}

type SuiteSummary struct {
	SuiteDescription string
	SuiteSucceeded   bool

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

	State     ExampleState
	RunTime   time.Duration
	Failure   ExampleFailure
	Benchmark ExampleBenchmark
}

type ExampleFailure struct {
	Message        string
	Location       CodeLocation
	ForwardedPanic interface{}

	ComponentIndex        int
	ComponentType         ExampleComponentType
	ComponentCodeLocation CodeLocation
}

type ExampleBenchmark struct {
	IsBenchmark     bool
	NumberOfSamples int
	FastestTime     time.Duration
	SlowestTime     time.Duration
	AverageTime     time.Duration
	StdDeviation    time.Duration
}
