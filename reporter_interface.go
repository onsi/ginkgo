package godescribe

type Reporter interface {
	SpecSuiteWillBegin(summary *SuiteSummary)
	ExampleDidComplete(exampleSummary *ExampleSummary, summary *SuiteSummary)
	SpecSuiteDidEnd(summary *SuiteSummary)
}

type SuiteSummary struct {
	SuiteDescription string
	ExampleSummaries []*ExampleSummary

	NumberOfTotalExamples   int
	NumberOfPendingExamples int
	NumberOfSkippedExamples int
	NumberOfRunExamples     int
	RunTime                 float64

	RandomSeed           int
	RandomizeAllExamples bool
}

type ExampleSummary struct {
	String   string
	Location CodeLocation

	State                ExampleState
	Components           []*ExampleComponent
	FailedComponentIndex uint
	Runtime              float64
}

type CodeLocation struct {
	FileName   string
	LineNumber int
	Column     int
}

type ExampleState uint

const (
	ExampleStateInvalid ExampleState = iota

	ExampleStatePending //Example has been marked as pending
	ExampleStateSkip    //Example will be skipped because other examples are focused

	ExampleStateWillRun //Example will be part of this spec suite

	ExampleStatePass   // Example passed
	ExampleStateFail   // Example failed
	ExampleStatePanick // Example panicked
)

type ExampleComponent struct {
	Name          string
	ComponentType ExampleComponentType
	Location      CodeLocation
	Runtime       float64
}

type ExampleComponentType uint

const (
	ExampleComponentTypeInvalid ExampleComponentType = iota

	ExampleComponentTypeDescribe
	ExampleComponentTypeContext
	ExampleComponentTypeBeforeEach
	ExampleComponentTypeJustBeforeEach
	ExampleComponentTypeAfterEach
	ExampleComponentTypeIt
)
