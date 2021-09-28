package types

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const GINKGO_FOCUS_EXIT_CODE = 197
const GINKGO_TIME_FORMAT = "01/02/06 15:04:05.999"

// Report captures information about a Ginkgo test run
type Report struct {
	//SuitePath captures the absolute path to the test suite
	SuitePath string

	//SuiteDescription cpatures the description string passed to the DSL's RunSpecs() command
	SuiteDescription string

	//SuiteSucceeded captures the success or failure status of the test run
	//If true, the test run is considered successful.
	//If false, the test run is considered unsucccessful
	SuiteSucceeded bool

	//SuiteHasProgrammaticFocus captures whether the test suite has a test or set of tests that are programatically focused
	//(i.e an `FIt` or an `FDescribe`
	SuiteHasProgrammaticFocus bool

	//SpecialSuiteFailureReasons may contain special failure reasons
	//For example, a test suite might be considered "failed" even if none of the individual specs
	//have a failure state.  For example, if the user has configured --fail-on-pending the test suite
	//will have failed if there are pending tests even though all non-pending tests may have passed.  In such
	//cases, Ginkgo populates SpecialSuiteFailureReasons with a clear message indicating the reason for the failure.
	//SpecialSuiteFailureReasons is also populated if the test suite is interrupted by the user.
	//Since multiple special failure reasons can occur, this field is a slice.
	SpecialSuiteFailureReasons []string

	//PreRunStats contains a set of stats captured before the test run begins.  This is primarily used
	//by Ginkgo's reporter to tell the user how many specs are in the current suite (PreRunStats.TotalSpecs)
	//and how many it intends to run (PreRunStats.SpecsThatWillRun) after applying any relevant focus or skip filters.
	PreRunStats PreRunStats

	//StartTime and EndTime capture the start and end time of the test run
	StartTime time.Time
	EndTime   time.Time

	//RunTime captures the duration of the test run
	RunTime time.Duration

	//SuiteConfig captures the Ginkgo configuration governing this test run
	//SuiteConfig includes information necessary for reproducing an identical test run,
	//such as the random seed and any filters applied during the test run
	SuiteConfig SuiteConfig

	//SpecReports is a list of all SpecReports generated by this test run
	SpecReports SpecReports
}

//PreRunStats contains a set of stats captured before the test run begins.  This is primarily used
//by Ginkgo's reporter to tell the user how many specs are in the current suite (PreRunStats.TotalSpecs)
//and how many it intends to run (PreRunStats.SpecsThatWillRun) after applying any relevant focus or skip filters.
type PreRunStats struct {
	TotalSpecs       int
	SpecsThatWillRun int
}

//Add is ued by Ginkgo's parallel aggregation mechanisms to combine test run reports form individual parallel processes
//to form a complete final report.
func (report Report) Add(other Report) Report {
	report.SuiteSucceeded = report.SuiteSucceeded && other.SuiteSucceeded

	if other.StartTime.Before(report.StartTime) {
		report.StartTime = other.StartTime
	}

	if other.EndTime.After(report.EndTime) {
		report.EndTime = other.EndTime
	}

	specialSuiteFailureReasons := []string{}
	reasonsLookup := map[string]bool{}
	for _, reasons := range [][]string{report.SpecialSuiteFailureReasons, other.SpecialSuiteFailureReasons} {
		for _, reason := range reasons {
			if !reasonsLookup[reason] {
				reasonsLookup[reason] = true
				specialSuiteFailureReasons = append(specialSuiteFailureReasons, reason)
			}
		}
	}
	report.SpecialSuiteFailureReasons = specialSuiteFailureReasons
	report.RunTime = report.EndTime.Sub(report.StartTime)

	reports := make(SpecReports, len(report.SpecReports)+len(other.SpecReports))
	for i := range report.SpecReports {
		reports[i] = report.SpecReports[i]
	}
	offset := len(report.SpecReports)
	for i := range other.SpecReports {
		reports[i+offset] = other.SpecReports[i]
	}

	report.SpecReports = reports
	return report
}

// SpecReport captures information about a Ginkgo spec.
type SpecReport struct {
	// ContainerHierarchyTexts is a slice containing the text strings of
	// all Describe/Context/When containers in this spec's hierarchy.
	ContainerHierarchyTexts []string

	// ContainerHierarchyLocations is a slice containing the CodeLocations of
	// all Describe/Context/When containers in this spec's hierarchy.
	ContainerHierarchyLocations []CodeLocation

	// ContainerHierarchyLabels is a slice containing the labels of
	// all Describe/Context/When containers in this spec's hierarchy
	ContainerHierarchyLabels [][]string

	// LeafNodeType, LeadNodeLocation, LeafNodeLabels and LeafNodeText capture the NodeType, CodeLocation, and text
	// of the Ginkgo node being tested (typically an NodeTypeIt node, though this can also be
	// one of the NodeTypesForSuiteLevelNodes node types)
	LeafNodeType     NodeType
	LeafNodeLocation CodeLocation
	LeafNodeLabels   []string
	LeafNodeText     string

	// State captures whether the spec has passed, failed, etc.
	State SpecState

	// StartTime and EndTime capture the start and end time of the spec
	StartTime time.Time
	EndTime   time.Time

	// RunTime captures the duration of the spec
	RunTime time.Duration

	// GinkgoParallelNode captures the parallel node that this spec ran on
	GinkgoParallelNode int

	//Failure is populated if a spec has failed, panicked, been interrupted, or skipped by the user (e.g. calling Skip())
	//It includes detailed information about the Failure
	Failure Failure

	// NumAttempts captures the number of times this Spec was run.  Flakey specs can be retried with
	// ginkgo --flake-attempts=N
	NumAttempts int

	// CapturedGinkgoWriterOutput contains text printed to the GinkgoWriter
	CapturedGinkgoWriterOutput string

	// CapturedStdOutErr contains text printed to stdout/stderr (when running in parallel)
	// This is always empty when running in series or calling CurrentSpecReport()
	// It is used internally by Ginkgo's reporter
	CapturedStdOutErr string

	// ReportEntries contains any reports added via `AddReportEntry`
	ReportEntries ReportEntries
}

func (report SpecReport) MarshalJSON() ([]byte, error) {
	//All this to avoid emitting an empty Failure struct in the JSON
	out := struct {
		ContainerHierarchyTexts     []string
		ContainerHierarchyLocations []CodeLocation
		ContainerHierarchyLabels    [][]string
		LeafNodeType                NodeType
		LeafNodeLocation            CodeLocation
		LeafNodeLabels              []string
		LeafNodeText                string
		State                       SpecState
		StartTime                   time.Time
		EndTime                     time.Time
		RunTime                     time.Duration
		GinkgoParallelNode          int
		Failure                     *Failure `json:",omitempty"`
		NumAttempts                 int
		CapturedGinkgoWriterOutput  string        `json:",omitempty"`
		CapturedStdOutErr           string        `json:",omitempty"`
		ReportEntries               ReportEntries `json:",omitempty"`
	}{
		ContainerHierarchyTexts:     report.ContainerHierarchyTexts,
		ContainerHierarchyLocations: report.ContainerHierarchyLocations,
		ContainerHierarchyLabels:    report.ContainerHierarchyLabels,
		LeafNodeType:                report.LeafNodeType,
		LeafNodeLocation:            report.LeafNodeLocation,
		LeafNodeLabels:              report.LeafNodeLabels,
		LeafNodeText:                report.LeafNodeText,
		State:                       report.State,
		StartTime:                   report.StartTime,
		EndTime:                     report.EndTime,
		RunTime:                     report.RunTime,
		GinkgoParallelNode:          report.GinkgoParallelNode,
		Failure:                     nil,
		ReportEntries:               nil,
		NumAttempts:                 report.NumAttempts,
		CapturedGinkgoWriterOutput:  report.CapturedGinkgoWriterOutput,
		CapturedStdOutErr:           report.CapturedStdOutErr,
	}

	if !report.Failure.IsZero() {
		out.Failure = &(report.Failure)
	}
	if len(report.ReportEntries) > 0 {
		out.ReportEntries = report.ReportEntries
	}

	return json.Marshal(out)
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
// (SpecStateFailed, SpecStatePanicked, SpecStateinterrupted)
func (report SpecReport) Failed() bool {
	return report.State.Is(SpecStateFailureStates...)
}

//FullText returns a concatenation of all the report.ContainerHierarchyTexts and report.LeafNodeText
func (report SpecReport) FullText() string {
	texts := []string{}
	texts = append(texts, report.ContainerHierarchyTexts...)
	if report.LeafNodeText != "" {
		texts = append(texts, report.LeafNodeText)
	}
	return strings.Join(texts, " ")
}

//Labels returns a deduped set of all the spec's Labels.
func (report SpecReport) Labels() []string {
	out := []string{}
	seen := map[string]bool{}
	for _, labels := range report.ContainerHierarchyLabels {
		for _, label := range labels {
			if !seen[label] {
				seen[label] = true
				out = append(out, label)
			}
		}
	}
	for _, label := range report.LeafNodeLabels {
		if !seen[label] {
			seen[label] = true
			out = append(out, label)
		}
	}

	return out
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

type SpecReports []SpecReport

//WithLeafNodeType returns the subset of SpecReports with LeafNodeType matching one of the requested NodeTypes
func (reports SpecReports) WithLeafNodeType(nodeType ...NodeType) SpecReports {
	out := SpecReports{}
	for _, report := range reports {
		if report.LeafNodeType.Is(nodeType...) {
			out = append(out, report)
		}
	}
	return out
}

//WithState returns the subset of SpecReports with State matching one of the requested SpecStates
func (reports SpecReports) WithState(states ...SpecState) SpecReports {
	out := SpecReports{}
	for _, report := range reports {
		if report.State.Is(states...) {
			out = append(out, report)
		}
	}
	return out
}

//CountWithState returns the number of SpecReports with State matching one of the requested SpecStates
func (reports SpecReports) CountWithState(states ...SpecState) int {
	n := 0
	for _, report := range reports {
		if report.State.Is(states...) {
			n += 1
		}
	}
	return n
}

//CountWithState returns the number of SpecReports that passed after multiple attempts
func (reports SpecReports) CountOfFlakedSpecs() int {
	n := 0
	for _, report := range reports.WithState(SpecStatePassed) {
		if report.NumAttempts > 1 {
			n += 1
		}
	}
	return n
}

// Failure captures failure information for an individual test
type Failure struct {
	// Message - the failure message passed into Fail(...).  When using a matcher library
	// like Gomega, this will contain the failure message generated by Gomega.
	Message string

	// Location - the CodeLocation where the failure occurred
	// This CodeLocation will include a fully-populated StackTrace
	Location CodeLocation

	// ForwardedPanic - if the failure represents a captured panic (i.e. Summary.State == SpecStatePanicked)
	// then ForwardedPanic will be populated with a string representation of the captured panic.
	ForwardedPanic string `json:",omitempty"`

	// FailureNodeContext - one of three contexts describing the node in which the failure occured:
	// FailureNodeIsLeafNode means the failure occured in the leaf node of the associated SpecReport. None of the other FailureNode fields will be populated
	// FailureNodeAtTopLevel means the failure occured in a non-leaf node that is defined at the top-level of the spec (i.e. not in a container). FailureNodeType and FailureNodeLocation will be populated.
	// FailureNodeInContainer means the failure occured in a non-leaf node that is defined within a container.  FailureNodeType, FailureNodeLocaiton, and FailureNodeContainerIndex will be populated.
	//
	//  FailureNodeType will contain the NodeType of the node in which the failure occurred.
	//  FailureNodeLocation will contain the CodeLocation of the node in which the failure occurred.
	// If populated, FailureNodeContainerIndex will be the index into SpecReport.ContainerHierarchyTexts and SpecReport.ContainerHierarchyLocations that represents the parent container of the node in which the failure occurred.
	FailureNodeContext        FailureNodeContext
	FailureNodeType           NodeType
	FailureNodeLocation       CodeLocation
	FailureNodeContainerIndex int
}

func (f Failure) IsZero() bool {
	return f == Failure{}
}

// FailureNodeContext captures the location context for the node containing the failing line of code
type FailureNodeContext uint

const (
	FailureNodeContextInvalid FailureNodeContext = iota

	FailureNodeIsLeafNode
	FailureNodeAtTopLevel
	FailureNodeInContainer
)

var fncEnumSupport = NewEnumSupport(map[uint]string{
	uint(FailureNodeContextInvalid): "INVALID FAILURE NODE CONTEXT",
	uint(FailureNodeIsLeafNode):     "leaf-node",
	uint(FailureNodeAtTopLevel):     "top-level",
	uint(FailureNodeInContainer):    "in-container",
})

func (fnc FailureNodeContext) String() string {
	return fncEnumSupport.String(uint(fnc))
}
func (fnc *FailureNodeContext) UnmarshalJSON(b []byte) error {
	out, err := fncEnumSupport.UnmarshJSON(b)
	*fnc = FailureNodeContext(out)
	return err
}
func (fnc FailureNodeContext) MarshalJSON() ([]byte, error) {
	return fncEnumSupport.MarshJSON(uint(fnc))
}

// ReportEntry captures information attached to `SpecReport` via `AddReportEntry`
type ReportEntry struct {
	// Visibility captures the visibility policy for this ReportEntry
	Visibility ReportEntryVisibility
	// Time captures the time the AddReportEntry was called
	Time time.Time
	// Location captures the location of the AddReportEntry call
	Location CodeLocation
	// Name captures the name of this report
	Name string
	// Value captures the (optional) object passed into AddReportEntry - this can be
	// anything the user wants.  If Value is a `fmt.Stringer` then Value.String() is called to produce a textual representation
	Value interface{}
	// ReportEntry needs to be encoded when generating json-formatted reports.  The concrete type of
	// Value is lost during the encode/decode round-trip - if Value is a `fmt.Stringer()` that means that we'd
	// lose the correct string representation of the value.  To circument this, we capture `Value.String()`
	// in `Representation` when encoding `ReportEntry`.  This is done just before encoding to ensure
	// we have the latest state of `Value`.
	Representation string
}

// ColorableStringer is an interface that ReportEntry values can satisfy.  If they do then ColorableStirng() is used to generate their representation.
type ColorableStringer interface {
	ColorableString() string
}

// StringRepresentation() returns the string representation of the ReportEntry --
// if Value is nil, empty string is returned
// if Value is a `ColorableStringer` then `Value.ColorableString()` is returned
// if Value is a `fmt.Stringer` then `Value.String()` is returned
// otherwise the Value is formatted with "%+v"
func (entry ReportEntry) StringRepresentation() string {
	if entry.Value == nil {
		return ""
	}

	if colorableStringer, ok := entry.Value.(ColorableStringer); ok {
		return colorableStringer.ColorableString()
	}

	if stringer, ok := entry.Value.(fmt.Stringer); ok {
		return stringer.String()
	}
	if entry.Representation != "" {
		return entry.Representation
	}
	return fmt.Sprintf("%+v", entry.Value)
}

func (entry ReportEntry) MarshalJSON() ([]byte, error) {
	//All this to capture the representaiton at encoding-time, not creating time
	//This way users can Report on pointers and get their final values at reporting-time
	out := struct {
		Visibility     ReportEntryVisibility
		Time           time.Time
		Location       CodeLocation
		Name           string
		Value          interface{}
		Representation string `json:",omitempty"`
	}{
		Visibility:     entry.Visibility,
		Time:           entry.Time,
		Location:       entry.Location,
		Name:           entry.Name,
		Value:          entry.Value,
		Representation: entry.Representation,
	}

	if entry.Value != nil {
		if colorableStringer, ok := entry.Value.(ColorableStringer); ok {
			out.Representation = colorableStringer.ColorableString()
		} else if stringer, ok := entry.Value.(fmt.Stringer); ok {
			out.Representation = stringer.String()
		}
	}

	return json.Marshal(out)
}

type ReportEntries []ReportEntry

func (re ReportEntries) HasVisibility(visibilities ...ReportEntryVisibility) bool {
	for _, entry := range re {
		if entry.Visibility.Is(visibilities...) {
			return true
		}
	}
	return false
}

func (re ReportEntries) WithVisibility(visibilities ...ReportEntryVisibility) ReportEntries {
	out := ReportEntries{}

	for _, entry := range re {
		if entry.Visibility.Is(visibilities...) {
			out = append(out, entry)
		}
	}

	return out
}

// ReportEntryVisibility governs the visibility of ReportEntries in Ginkgo's console reporter
type ReportEntryVisibility uint

const (
	// Always print out this ReportEntry
	ReportEntryVisibilityAlways ReportEntryVisibility = iota
	// Only print out this ReportEntry if the spec fails or if the test is run with -v
	ReportEntryVisibilityFailureOrVerbose
	// Never print out this ReportEntry (note that ReportEntrys are always encoded in machine readable reports (e.g. JSON, JUnit, etc.))
	ReportEntryVisibilityNever
)

var revEnumSupport = NewEnumSupport(map[uint]string{
	uint(ReportEntryVisibilityAlways):           "always",
	uint(ReportEntryVisibilityFailureOrVerbose): "failure-or-verbose",
	uint(ReportEntryVisibilityNever):            "never",
})

func (rev ReportEntryVisibility) String() string {
	return revEnumSupport.String(uint(rev))
}
func (rev *ReportEntryVisibility) UnmarshalJSON(b []byte) error {
	out, err := revEnumSupport.UnmarshJSON(b)
	*rev = ReportEntryVisibility(out)
	return err
}
func (rev ReportEntryVisibility) MarshalJSON() ([]byte, error) {
	return revEnumSupport.MarshJSON(uint(rev))
}

func (v ReportEntryVisibility) Is(visibilities ...ReportEntryVisibility) bool {
	for _, visibility := range visibilities {
		if v == visibility {
			return true
		}
	}

	return false
}

// SpecState captures the state of a spec
// To determine if a given `state` represents a failure state, use `state.Is(SpecStateFailureStates...)`
type SpecState uint

const (
	SpecStateInvalid SpecState = iota

	SpecStatePending
	SpecStateSkipped
	SpecStatePassed
	SpecStateFailed
	SpecStateAborted
	SpecStatePanicked
	SpecStateInterrupted
)

var ssEnumSupport = NewEnumSupport(map[uint]string{
	uint(SpecStateInvalid):     "INVALID SPEC STATE",
	uint(SpecStatePending):     "pending",
	uint(SpecStateSkipped):     "skipped",
	uint(SpecStatePassed):      "passed",
	uint(SpecStateFailed):      "failed",
	uint(SpecStateAborted):     "aborted",
	uint(SpecStatePanicked):    "panicked",
	uint(SpecStateInterrupted): "interrupted",
})

func (ss SpecState) String() string {
	return ssEnumSupport.String(uint(ss))
}
func (ss *SpecState) UnmarshalJSON(b []byte) error {
	out, err := ssEnumSupport.UnmarshJSON(b)
	*ss = SpecState(out)
	return err
}
func (ss SpecState) MarshalJSON() ([]byte, error) {
	return ssEnumSupport.MarshJSON(uint(ss))
}

var SpecStateFailureStates = []SpecState{SpecStateFailed, SpecStateAborted, SpecStatePanicked, SpecStateInterrupted}

func (ss SpecState) Is(states ...SpecState) bool {
	for _, testState := range states {
		if testState == ss {
			return true
		}
	}

	return false
}

// NodeType captures the type of a given Ginkgo Node
type NodeType uint

const (
	NodeTypeInvalid NodeType = iota

	NodeTypeContainer
	NodeTypeIt

	NodeTypeBeforeEach
	NodeTypeJustBeforeEach
	NodeTypeAfterEach
	NodeTypeJustAfterEach

	NodeTypeBeforeAll
	NodeTypeAfterAll

	NodeTypeBeforeSuite
	NodeTypeSynchronizedBeforeSuite
	NodeTypeAfterSuite
	NodeTypeSynchronizedAfterSuite

	NodeTypeReportBeforeEach
	NodeTypeReportAfterEach
	NodeTypeReportAfterSuite

	NodeTypeCleanupInvalid
	NodeTypeCleanupAfterEach
	NodeTypeCleanupAfterAll
	NodeTypeCleanupAfterSuite
)

var NodeTypesForContainerAndIt = []NodeType{NodeTypeContainer, NodeTypeIt}
var NodeTypesForSuiteLevelNodes = []NodeType{NodeTypeBeforeSuite, NodeTypeSynchronizedBeforeSuite, NodeTypeAfterSuite, NodeTypeSynchronizedAfterSuite, NodeTypeReportAfterSuite}

var ntEnumSupport = NewEnumSupport(map[uint]string{
	uint(NodeTypeInvalid):                 "INVALID NODE TYPE",
	uint(NodeTypeContainer):               "Container",
	uint(NodeTypeIt):                      "It",
	uint(NodeTypeBeforeEach):              "BeforeEach",
	uint(NodeTypeJustBeforeEach):          "JustBeforeEach",
	uint(NodeTypeAfterEach):               "AfterEach",
	uint(NodeTypeJustAfterEach):           "JustAfterEach",
	uint(NodeTypeBeforeAll):               "BeforeAll",
	uint(NodeTypeAfterAll):                "AfterAll",
	uint(NodeTypeBeforeSuite):             "BeforeSuite",
	uint(NodeTypeSynchronizedBeforeSuite): "SynchronizedBeforeSuite",
	uint(NodeTypeAfterSuite):              "AfterSuite",
	uint(NodeTypeSynchronizedAfterSuite):  "SynchronizedAfterSuite",
	uint(NodeTypeReportBeforeEach):        "ReportBeforeEach",
	uint(NodeTypeReportAfterEach):         "ReportAfterEach",
	uint(NodeTypeReportAfterSuite):        "ReportAfterSuite",
	uint(NodeTypeCleanupInvalid):          "INVALID CLEANUP NODE",
	uint(NodeTypeCleanupAfterEach):        "DeferCleanup (AfterEach)",
	uint(NodeTypeCleanupAfterAll):         "DeferCleanup (AfterAll)",
	uint(NodeTypeCleanupAfterSuite):       "DeferCleanup (AfterSuite)",
})

func (nt NodeType) String() string {
	return ntEnumSupport.String(uint(nt))
}
func (nt *NodeType) UnmarshalJSON(b []byte) error {
	out, err := ntEnumSupport.UnmarshJSON(b)
	*nt = NodeType(out)
	return err
}
func (nt NodeType) MarshalJSON() ([]byte, error) {
	return ntEnumSupport.MarshJSON(uint(nt))
}

func (nt NodeType) Is(nodeTypes ...NodeType) bool {
	for _, nodeType := range nodeTypes {
		if nt == nodeType {
			return true
		}
	}

	return false
}

type NodeTypes []NodeType

func (nt NodeTypes) Contains(expected NodeType) bool {
	for _, nodeType := range nt {
		if nodeType == expected {
			return true
		}
	}
	return false
}
