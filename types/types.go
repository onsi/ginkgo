package types

import (
	"encoding/json"
	"strings"
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

// SpecReport captures information about a Ginkgo spec.
type SpecReport struct {
	// NodeTexts is a slice containing the text strings of
	// all Describe/Context/When containers in this test's hierarchy.
	// The last element in NodeTexts is the string of the It itself.
	NodeTexts []string

	// NodeLocations is a slice containing the CodeLocations of
	// all Describe/Context/When containers in this test's hirerachy.
	// The last element in NodeLcoations is the CodeLocation of the It itself
	NodeLocations []CodeLocation

	// LeafNodeType and LeadNodeLocation capture the NodeType and CodeLocation
	// of the Ginkgo node being tested (typically an NodeTypeIt node, though this can also be
	// one of the NodeTypeBeforeSuite or NodeTypeAfterSuite node types)
	LeafNodeType     NodeType
	LeafNodeLocation CodeLocation

	// State captures whether the test has passed, failed, etc.
	State SpecState

	// RunTime captures the duration of the test
	RunTime time.Duration

	//Failure is populated if a test has failed, panicked, been interrupted, or skipped by the user (e.g. calling Skip())
	//It includes detailed information about the Failure
	Failure Failure

	// NumAttempts captures the number of times this Spec was run.  Flakey tests can be retried with
	// ginkgo --flake-attempts=N
	NumAttempts int

	// CapturedGinkgoWriterOutput contains text printed to the GinkgoWriter
	CapturedGinkgoWriterOutput string

	// CapturedStdOutErr contains text printed to stdout/stderr (when running in parallel)
	// This is always empty when running in series or calling CurrentSpecReport()
	// It is used internally by Ginkgo's reporter
	CapturedStdOutErr string
}

// CombinedOutput returns a single string representation of both CapturedStdOutErr and CapturedGinkgoWriterOutput
// Note that both are empty when using CurrentSpecReport() so CurrentSpecReport().CombinedOutput() will always be empty.
// CombinedOutput() is used internally by Ginkgo's reporter.
func (report SpecReport) CombinedOutput() string {
	if report.CapturedStdOutErr == "" {
		return report.CapturedGinkgoWriterOutput
	}
	if report.CapturedGinkgoWriterOutput == "" {
		return report.CapturedStdOutErr
	}
	return report.CapturedStdOutErr + "\n" + report.CapturedGinkgoWriterOutput
}

//Failed returns true if report.State is one of the SpecStateFailureStates
// (SpecStateFAiled, SpecStatePanicked, SpecStateinterrupted)
func (report SpecReport) Failed() bool {
	return report.State.Is(SpecStateFailureStates...)
}

//FullText returns a concatenation of all the report.NodeTexts
func (report SpecReport) FullText() string {
	return strings.Join(report.NodeTexts, " ")
}

//SpecText returns the text of the spec node (i.e. the string passed into It())
func (report SpecReport) SpecText() string {
	if len(report.NodeTexts) == 0 {
		return ""
	}
	return report.NodeTexts[len(report.NodeTexts)-1]
}

//FileName() returns the name of the file containing the spec
func (report SpecReport) FileName() string {
	return report.LeafNodeLocation.FileName
}

//LineNumber() returns the line number of the leaf node
func (report SpecReport) LineNumber() int {
	return report.LeafNodeLocation.LineNumber
}

//FailureMessage() returns the failure message (or empty string if the test hasn't failed)
func (report SpecReport) FailureMessage() string {
	return report.Failure.Message
}

//FailureLocation() returns the location of the failure (or an empty CodeLocation if the test hasn't failed)
func (report SpecReport) FailureLocation() CodeLocation {
	return report.Failure.Location
}

// Failure captures failure information for an individual test
type Failure struct {
	// Message - the failure message passed into Fail(...).  When using a matcher library
	// like Gomega, this will contain the failure message generated by Gomega.
	Message string

	// Location - the CodeLocation where the failure occurred
	Location CodeLocation

	// ForwardedPanic - if the failure represents a captured panic (i.e. Summary.State == SpecStatePanicked)
	// then ForwardedPanic will be populated with a string representation of the captured panic.
	ForwardedPanic string

	// NodeIndex - the index into Summary.NodeTexts and Summary.NodeLocation that represents the
	// container node where this failure took place.  If NodeIndex is -1 then the failure occured at the top level
	// (i.e. outside of any container node).
	NodeIndex int

	// NodeType - the NodeType of the node where the failure occured.  For example, a failure occurring in a
	// BeforeEach in a Describe container nested within a top-level Describe would have a NodeType of NodeTypeBeforeEach
	// and a NodeIndex of 1.
	NodeType NodeType
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

func (s SpecState) String() string {
	switch s {
	case SpecStatePending:
		return "pending"
	case SpecStateSkipped:
		return "skipped"
	case SpecStatePassed:
		return "passed"
	case SpecStateFailed:
		return "failed"
	case SpecStatePanicked:
		return "panicked"
	case SpecStateInterrupted:
		return "interrupted"
	}

	return "INVALID SPEC STATE"
}

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

	NodeTypeReportAfterEach
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
	case NodeTypeReportAfterEach:
		return "ReportAfterEach"
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
