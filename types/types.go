package types

import (
	"encoding/json"
	"time"
)

const GINKGO_FOCUS_EXIT_CODE = 197

type SuiteSummary struct {
	SuiteDescription string
	SuiteSucceeded   bool

	NumberOfTotalSpecs         int
	NumberOfSpecsThatWillBeRun int

	NumberOfSkippedSpecs int
	NumberOfPassedSpecs  int
	NumberOfFailedSpecs  int
	NumberOfPendingSpecs int
	NumberOfFlakedSpecs  int
	RunTime              time.Duration
}

func (summary SuiteSummary) NumberOfSpecsThatRan() int {
	return summary.NumberOfPassedSpecs + summary.NumberOfFailedSpecs
}

func (summary SuiteSummary) Add(other SuiteSummary) SuiteSummary {
	out := SuiteSummary{}
	out.SuiteDescription = summary.SuiteDescription
	out.SuiteSucceeded = summary.SuiteSucceeded && other.SuiteSucceeded
	out.NumberOfTotalSpecs = summary.NumberOfTotalSpecs
	out.NumberOfSpecsThatWillBeRun = summary.NumberOfSpecsThatWillBeRun

	out.NumberOfSkippedSpecs = summary.NumberOfSkippedSpecs + other.NumberOfSkippedSpecs
	out.NumberOfPassedSpecs = summary.NumberOfPassedSpecs + other.NumberOfPassedSpecs
	out.NumberOfFailedSpecs = summary.NumberOfFailedSpecs + other.NumberOfFailedSpecs
	out.NumberOfPendingSpecs = summary.NumberOfPendingSpecs + other.NumberOfPendingSpecs
	out.NumberOfFlakedSpecs = summary.NumberOfFlakedSpecs + other.NumberOfFlakedSpecs
	if summary.RunTime > other.RunTime {
		out.RunTime = summary.RunTime
	} else {
		out.RunTime = other.RunTime
	}

	return out
}

type Summary struct {
	NodeTexts     []string
	NodeLocations []CodeLocation

	LeafNodeType     NodeType
	LeafNodeLocation CodeLocation

	State       SpecState
	RunTime     time.Duration
	Failure     Failure
	NumAttempts int

	CapturedStdOutErr          string
	CapturedGinkgoWriterOutput string
}

func (summary Summary) CombinedOutput() string {
	output := summary.CapturedStdOutErr
	if output == "" {
		output = summary.CapturedGinkgoWriterOutput
	} else {
		output = output + "\n" + summary.CapturedGinkgoWriterOutput
	}
	return output
}

type Failure struct {
	Message        string
	Location       CodeLocation
	ForwardedPanic string

	NodeIndex int
	NodeType  NodeType
}

func (f Failure) IsZero() bool {
	return f == Failure{}
}

type SpecState uint

const (
	SpecStateInvalid SpecState = iota

	SpecStatePending
	SpecStateSkipped
	SpecStatePassed
	SpecStateFailed
	SpecStatePanicked
	SpecStateInterrupted
)

var SpecStateFailureStates = []SpecState{SpecStateFailed, SpecStatePanicked, SpecStateInterrupted}

func (state SpecState) Is(states ...SpecState) bool {
	for _, testState := range states {
		if testState == state {
			return true
		}
	}

	return false
}

type NodeType uint

const (
	NodeTypeInvalid NodeType = iota

	NodeTypeContainer
	NodeTypeIt

	NodeTypeBeforeEach
	NodeTypeJustBeforeEach
	NodeTypeAfterEach
	NodeTypeJustAfterEach

	NodeTypeBeforeSuite
	NodeTypeSynchronizedBeforeSuite
	NodeTypeAfterSuite
	NodeTypeSynchronizedAfterSuite
)

var NodeTypesForContainerAndIt = []NodeType{NodeTypeContainer, NodeTypeIt}
var NodeTypesForSuiteSetup = []NodeType{NodeTypeBeforeSuite, NodeTypeSynchronizedBeforeSuite, NodeTypeAfterSuite, NodeTypeSynchronizedAfterSuite}

func (nt NodeType) Is(nodeTypes ...NodeType) bool {
	for _, nodeType := range nodeTypes {
		if nt == nodeType {
			return true
		}
	}

	return false
}

func (nt NodeType) String() string {
	switch nt {
	case NodeTypeContainer:
		return "Container"
	case NodeTypeIt:
		return "It"
	case NodeTypeBeforeEach:
		return "BeforeEach"
	case NodeTypeJustBeforeEach:
		return "JustBeforeEach"
	case NodeTypeAfterEach:
		return "AfterEach"
	case NodeTypeJustAfterEach:
		return "JustAfterEach"
	case NodeTypeBeforeSuite:
		return "BeforeSuite"
	case NodeTypeSynchronizedBeforeSuite:
		return "SynchronizedBeforeSuite"
	case NodeTypeAfterSuite:
		return "AfterSuite"
	case NodeTypeSynchronizedAfterSuite:
		return "SynchronizedAfterSuite"
	}

	return "INVALID NODE TYPE"
}

type RemoteBeforeSuiteState int

const (
	RemoteBeforeSuiteStateInvalid RemoteBeforeSuiteState = iota

	RemoteBeforeSuiteStatePending
	RemoteBeforeSuiteStatePassed
	RemoteBeforeSuiteStateFailed
	RemoteBeforeSuiteStateDisappeared
)

type RemoteBeforeSuiteData struct {
	Data  []byte
	State RemoteBeforeSuiteState
}

func (r RemoteBeforeSuiteData) ToJSON() []byte {
	data, _ := json.Marshal(r)
	return data
}

type RemoteAfterSuiteData struct {
	CanRun bool
}
