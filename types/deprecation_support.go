package types

import (
	"strconv"
	"time"

	"github.com/onsi/ginkgo/formatter"
)

type Deprecation struct {
	Message string
	DocLink string
}

type deprecations struct{}

var Deprecations = deprecations{}

func (d deprecations) CustomReporter() Deprecation {
	return Deprecation{
		Message: "Support for custom reporters has been removed in V2.  Please read the documentation linked to below for Ginkgo's new behavior and for a migration path:",
		DocLink: "removed-custom-reporters",
	}
}

func (d deprecations) Async() Deprecation {
	return Deprecation{
		Message: "You are passing a Done channel to a test node to test asynchronous behavior.  This is deprecated in Ginkgo V2.  Your test will run synchronously and the timeout will be ignored.",
		DocLink: "removed-async-testing",
	}
}

func (d deprecations) Measure() Deprecation {
	return Deprecation{
		Message: "Measure is deprecated and will be removed in Ginkgo V2.  Please migrate to gomega/gmeasure.",
		DocLink: "removed-measure",
	}
}

func (d deprecations) CurrentGinkgoTestDescription() Deprecation {
	return Deprecation{
		Message: "CurrentGinkgoTestDescription() is deprecated in Ginkgo V2.  Use CurrentSpecReport() instead.",
		DocLink: "changed-currentginkgotestdescription",
	}
}

func (d deprecations) Convert() Deprecation {
	return Deprecation{
		Message: "The convert command is deprecated in Ginkgo V2",
		DocLink: "removed-ginkgo-convert",
	}
}

func (d deprecations) Blur() Deprecation {
	return Deprecation{
		Message: "The blur command is deprecated in Ginkgo V2.  Use 'ginkgo unfocus' instead.",
	}
}

type DeprecationTracker struct {
	deprecations map[Deprecation][]CodeLocation
}

func NewDeprecationTracker() *DeprecationTracker {
	return &DeprecationTracker{
		deprecations: map[Deprecation][]CodeLocation{},
	}
}

func (d *DeprecationTracker) TrackDeprecation(deprecation Deprecation, cl ...CodeLocation) {
	if len(cl) == 1 {
		d.deprecations[deprecation] = append(d.deprecations[deprecation], cl[0])
	} else {
		d.deprecations[deprecation] = []CodeLocation{}
	}
}

func (d *DeprecationTracker) DidTrackDeprecations() bool {
	return len(d.deprecations) > 0
}

func (d *DeprecationTracker) DeprecationsReport() string {
	out := formatter.F("{{light-yellow}}You're using deprecated Ginkgo functionality:{{/}}\n")
	out += formatter.F("{{light-yellow}}============================================={{/}}\n")
	for deprecation, locations := range d.deprecations {
		out += formatter.Fi(1, "{{yellow}}"+deprecation.Message+"{{/}}\n")
		if deprecation.DocLink != "" {
			out += formatter.Fi(1, "{{bold}}Learn more at:{{/}} {{cyan}}{{underline}}https://github.com/onsi/ginkgo/blob/v2/docs/MIGRATING_TO_V2.md#%s{{/}}\n", deprecation.DocLink)
		}
		for _, location := range locations {
			out += formatter.Fi(2, "{{gray}}%s{{/}}\n", location)
		}
	}
	return out
}

/*
	A set of deprecations to make the transition from v1 to v2 easier for users who have written custom reporters.
*/

type SetupSummary = DeprecatedSetupSummary
type SpecSummary = DeprecatedSpecSummary
type SpecMeasurement = DeprecatedSpecMeasurement
type SpecComponentType = NodeType
type SpecFailure = DeprecatedSpecFailure

var (
	SpecComponentTypeInvalid                 = NodeTypeInvalid
	SpecComponentTypeContainer               = NodeTypeContainer
	SpecComponentTypeIt                      = NodeTypeIt
	SpecComponentTypeBeforeEach              = NodeTypeBeforeEach
	SpecComponentTypeJustBeforeEach          = NodeTypeJustBeforeEach
	SpecComponentTypeAfterEach               = NodeTypeAfterEach
	SpecComponentTypeJustAfterEach           = NodeTypeJustAfterEach
	SpecComponentTypeBeforeSuite             = NodeTypeBeforeSuite
	SpecComponentTypeSynchronizedBeforeSuite = NodeTypeSynchronizedBeforeSuite
	SpecComponentTypeAfterSuite              = NodeTypeAfterSuite
	SpecComponentTypeSynchronizedAfterSuite  = NodeTypeSynchronizedAfterSuite
)

type DeprecatedSetupSummary struct {
	ComponentType SpecComponentType
	CodeLocation  CodeLocation

	State   SpecState
	RunTime time.Duration
	Failure SpecFailure

	CapturedOutput string
	SuiteID        string
}

func DeprecatedSetupSummaryFromSpecReport(report SpecReport) *DeprecatedSetupSummary {
	return &DeprecatedSetupSummary{
		ComponentType:  report.LeafNodeType,
		CodeLocation:   report.LeafNodeLocation,
		State:          report.State,
		RunTime:        report.RunTime,
		Failure:        deprecatedSpecFailureFromFailure(report.Failure),
		CapturedOutput: report.CombinedOutput(),
	}
}

type DeprecatedSpecSummary struct {
	ComponentTexts         []string
	ComponentCodeLocations []CodeLocation

	State           SpecState
	RunTime         time.Duration
	Failure         SpecFailure
	IsMeasurement   bool
	NumberOfSamples int
	Measurements    map[string]*DeprecatedSpecMeasurement

	CapturedOutput string
	SuiteID        string
}

func DeprecatedSpecSummaryFromSpecReport(report SpecReport) *DeprecatedSpecSummary {
	return &DeprecatedSpecSummary{
		ComponentTexts:         report.NodeTexts,
		ComponentCodeLocations: report.NodeLocations,
		State:                  report.State,
		RunTime:                report.RunTime,
		Failure:                deprecatedSpecFailureFromFailure(report.Failure),
		IsMeasurement:          false,
		NumberOfSamples:        0,
		Measurements:           map[string]*DeprecatedSpecMeasurement{},
		CapturedOutput:         report.CombinedOutput(),
	}
}

func (s DeprecatedSpecSummary) HasFailureState() bool {
	return s.State.Is(SpecStateFailureStates...)
}

func (s DeprecatedSpecSummary) TimedOut() bool {
	return false
}

func (s DeprecatedSpecSummary) Panicked() bool {
	return s.State == SpecStatePanicked
}

func (s DeprecatedSpecSummary) Failed() bool {
	return s.State == SpecStateFailed
}

func (s DeprecatedSpecSummary) Passed() bool {
	return s.State == SpecStatePassed
}

func (s DeprecatedSpecSummary) Skipped() bool {
	return s.State == SpecStateSkipped
}

func (s DeprecatedSpecSummary) Pending() bool {
	return s.State == SpecStatePending
}

type DeprecatedSpecFailure struct {
	Message        string
	Location       CodeLocation
	ForwardedPanic string

	ComponentIndex        int
	ComponentType         SpecComponentType
	ComponentCodeLocation CodeLocation
}

func deprecatedSpecFailureFromFailure(failure Failure) SpecFailure {
	return SpecFailure{
		Message:               failure.Message,
		Location:              failure.Location,
		ForwardedPanic:        failure.ForwardedPanic,
		ComponentIndex:        failure.NodeIndex,
		ComponentType:         failure.NodeType,
		ComponentCodeLocation: failure.Location,
	}
}

type DeprecatedSpecMeasurement struct {
	Name  string
	Info  interface{}
	Order int

	Results []float64

	Smallest     float64
	Largest      float64
	Average      float64
	StdDeviation float64

	SmallestLabel string
	LargestLabel  string
	AverageLabel  string
	Units         string
	Precision     int
}

func (s DeprecatedSpecMeasurement) PrecisionFmt() string {
	if s.Precision == 0 {
		return "%f"
	}

	str := strconv.Itoa(s.Precision)

	return "%." + str + "f"
}
