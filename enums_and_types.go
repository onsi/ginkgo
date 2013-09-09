package ginkgo

type node interface {
	nodeType() nodeType
	getText() string
}

type exampleSubject interface {
	node

	getFlag() flagType
	run() (runOutcome, failureData)
	getCodeLocation() CodeLocation
}

type flagType uint

const (
	flagTypeNone flagType = iota
	flagTypeFocused
	flagTypePending
)

type runOutcome uint

const (
	runOutcomeInvalid runOutcome = iota
	runOutcomePanicked
	runOutcomeTimedOut
	runOutcomeCompleted
)

type nodeType uint

const (
	nodeTypeInvalid nodeType = iota
	nodeTypeContainer
	nodeTypeIt
	nodeTypeBenchmark
)

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

type ExampleComponentType uint

const (
	ExampleComponentTypeInvalid ExampleComponentType = iota

	ExampleComponentTypeBeforeEach
	ExampleComponentTypeJustBeforeEach
	ExampleComponentTypeAfterEach
	ExampleComponentTypeIt
)
