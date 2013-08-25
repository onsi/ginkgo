package ginkgo

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
