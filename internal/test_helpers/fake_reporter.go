package test_helpers

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gcustom"
	. "github.com/onsi/gomega/gstruct"

	"github.com/onsi/ginkgo/v2/internal/interrupt_handler"
	"github.com/onsi/ginkgo/v2/types"
)

type OmegaMatcherWithDescription struct {
	OmegaMatcher
	Description string
}

func (o OmegaMatcherWithDescription) GomegaString() string {
	return o.Description
}

/*

A FakeReporter and collection of matchers to match against reported suite and spec summaries

*/

type Reports []types.SpecReport

func (s Reports) FindByLeafNodeType(nodeTypes types.NodeType) types.SpecReport {
	for _, report := range s {
		if report.LeafNodeType.Is(nodeTypes) {
			return report
		}
	}

	return types.SpecReport{}
}

func (s Reports) Find(name string) types.SpecReport {
	for _, report := range s {
		if report.LeafNodeText == name {
			return report
		}
	}

	return types.SpecReport{}
}

func (s Reports) FindByFullText(text string) types.SpecReport {
	for _, report := range s {
		if report.FullText() == text {
			return report
		}
	}

	return types.SpecReport{}
}

func (s Reports) Names() []string {
	out := []string{}
	for _, report := range s {
		if report.LeafNodeText != "" {
			out = append(out, report.LeafNodeText)
		}
	}
	return out
}

func (s Reports) WithState(state types.SpecState) Reports {
	out := Reports{}
	for _, report := range s {
		if report.State == state {
			out = append(out, report)
		}
	}
	return out
}

func (s Reports) WithLeafNodeType(nodeTypes types.NodeType) Reports {
	out := Reports{}
	for _, report := range s {
		if report.LeafNodeType.Is(nodeTypes) {
			out = append(out, report)
		}
	}
	return out
}

type FakeReporter struct {
	Begin           types.Report
	Will            Reports
	Did             Reports
	End             types.Report
	ProgressReports []types.ProgressReport
	ReportEntries   []types.ReportEntry
	SpecEvents      []types.SpecEvent
	Failures        []types.AdditionalFailure
	lock            *sync.Mutex
}

func NewFakeReporter() *FakeReporter {
	return &FakeReporter{
		lock: &sync.Mutex{},
	}
}

func (r *FakeReporter) SuiteWillBegin(report types.Report) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.Begin = report
}

func (r *FakeReporter) WillRun(report types.SpecReport) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.Will = append(r.Will, report)
}

func (r *FakeReporter) DidRun(report types.SpecReport) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.Did = append(r.Did, report)
}

func (r *FakeReporter) SuiteDidEnd(report types.Report) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.End = report
}
func (r *FakeReporter) EmitProgressReport(progressReport types.ProgressReport) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.ProgressReports = append(r.ProgressReports, progressReport)
}
func (r *FakeReporter) EmitFailure(state types.SpecState, failure types.Failure) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.Failures = append(r.Failures, types.AdditionalFailure{Failure: failure, State: state})
}
func (r *FakeReporter) EmitReportEntry(reportEntry types.ReportEntry) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.ReportEntries = append(r.ReportEntries, reportEntry)
}
func (r *FakeReporter) EmitSpecEvent(specEvent types.SpecEvent) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.SpecEvents = append(r.SpecEvents, specEvent)
}

type NSpecs int
type NWillRun int
type NPassed int
type NSkipped int
type NFailed int
type NPending int
type NFlaked int

func BeASuiteSummary(options ...interface{}) OmegaMatcher {
	type ReportStats struct {
		Succeeded    bool
		TotalSpecs   int
		WillRunSpecs int
		Passed       int
		Skipped      int
		Failed       int
		Pending      int
		Flaked       int
	}
	fields := Fields{
		"Passed":     Equal(0),
		"Skipped":    Equal(0),
		"Failed":     Equal(0),
		"Pending":    Equal(0),
		"Flaked":     Equal(0),
		"TotalSpecs": Equal(0),
	}
	for _, option := range options {
		t := reflect.TypeOf(option)
		if t.Kind() == reflect.Bool {
			if option.(bool) {
				fields["Succeeded"] = BeTrue()
			} else {
				fields["Succeeded"] = BeFalse()
			}
		} else if t == reflect.TypeOf(NSpecs(0)) {
			fields["TotalSpecs"] = Equal(int(option.(NSpecs)))
		} else if t == reflect.TypeOf(NWillRun(0)) {
			fields["WillRunSpecs"] = Equal(int(option.(NWillRun)))
		} else if t == reflect.TypeOf(NPassed(0)) {
			fields["Passed"] = Equal(int(option.(NPassed)))
		} else if t == reflect.TypeOf(NSkipped(0)) {
			fields["Skipped"] = Equal(int(option.(NSkipped)))
		} else if t == reflect.TypeOf(NFailed(0)) {
			fields["Failed"] = Equal(int(option.(NFailed)))
		} else if t == reflect.TypeOf(NPending(0)) {
			fields["Pending"] = Equal(int(option.(NPending)))
		} else if t == reflect.TypeOf(NFlaked(0)) {
			fields["Flaked"] = Equal(int(option.(NFlaked)))
		}
	}
	return WithTransform(func(report types.Report) ReportStats {
		specs := report.SpecReports.WithLeafNodeType(types.NodeTypeIt)
		return ReportStats{
			Succeeded:    report.SuiteSucceeded,
			TotalSpecs:   report.PreRunStats.TotalSpecs,
			WillRunSpecs: report.PreRunStats.SpecsThatWillRun,
			Passed:       specs.CountWithState(types.SpecStatePassed),
			Skipped:      specs.CountWithState(types.SpecStateSkipped),
			Failed:       specs.CountWithState(types.SpecStateFailureStates),
			Pending:      specs.CountWithState(types.SpecStatePending),
			Flaked:       specs.CountOfFlakedSpecs(),
		}
	}, MatchFields(IgnoreExtras, fields))
}

type CapturedGinkgoWriterOutput string
type CapturedStdOutput string
type NumAttempts int

func HavePassed(options ...interface{}) OmegaMatcher {
	matchers := []OmegaMatcher{
		HaveField("State", types.SpecStatePassed),
		HaveField("Failure", BeZero()),
	}
	for _, option := range options {
		var matcher OmegaMatcher
		switch v := option.(type) {
		case CapturedGinkgoWriterOutput:
			matcher = HaveField("CapturedGinkgoWriterOutput", string(v))
		case CapturedStdOutput:
			matcher = HaveField("CapturedStdOutErr", string(v))
		case types.NodeType:
			matcher = HaveField("LeafNodeType", v)
		case NumAttempts:
			matcher = HaveField("NumAttempts", int(v))
		}
		if matcher != nil {
			matchers = append(matchers, matcher)
		}
	}

	return And(matchers...)
}

func BePending() OmegaMatcher {
	return And(
		HaveField("State", types.SpecStatePending),
		HaveField("Failure", BeZero()),
	)
}

func HaveBeenSkipped() OmegaMatcher {
	return And(
		HaveField("State", types.SpecStateSkipped),
		HaveField("Failure", BeZero()),
	)
}

func HaveBeenSkippedWithMessage(message string, options ...interface{}) OmegaMatcher {
	matchers := []OmegaMatcher{
		HaveField("State", types.SpecStateSkipped),
		HaveField("Failure.Message", Equal(message)),
	}

	for _, option := range options {
		switch v := option.(type) {
		case NumAttempts:
			matchers = append(matchers, HaveField("NumAttempts", int(v)))
		}
	}
	return And(matchers...)
}

func HaveBeenInterrupted(cause interrupt_handler.InterruptCause) OmegaMatcher {
	return And(
		HaveField("State", types.SpecStateInterrupted),
		HaveField("Failure.Message", HavePrefix(cause.String())),
	)
}

type FailureNodeType types.NodeType

func failureMatcherForState(state types.SpecState, messageField string, options ...interface{}) OmegaMatcher {
	matchers := []OmegaMatcher{
		HaveField("State", state),
	}
	for _, option := range options {
		var matcher OmegaMatcher
		switch v := option.(type) {
		case CapturedGinkgoWriterOutput:
			matcher = HaveField("CapturedGinkgoWriterOutput", string(v))
		case CapturedStdOutput:
			matcher = HaveField("CapturedStdOutErr", string(v))
		case types.NodeType:
			matcher = HaveField("LeafNodeType", v)
		case types.FailureNodeContext:
			matcher = HaveField("Failure.FailureNodeContext", v)
		case string:
			matcher = HaveField(messageField, ContainSubstring(v))
		case OmegaMatcher:
			matcher = HaveField(messageField, v)
		case types.CodeLocation:
			matcher = HaveField("Failure.Location", v)
		case FailureNodeType:
			matcher = HaveField("Failure.FailureNodeType", types.NodeType(v))
		case NumAttempts:
			matcher = HaveField("NumAttempts", int(v))
		case types.TimelineLocation:
			matcher = HaveField("Failure.TimelineLocation.Offset", v.Offset)
		}
		if matcher != nil {
			matchers = append(matchers, matcher)
		}
	}

	return And(matchers...)
}

func HaveFailed(options ...interface{}) OmegaMatcher {
	return failureMatcherForState(types.SpecStateFailed, "Failure.Message", options...)
}

func HaveTimedOut(options ...interface{}) OmegaMatcher {
	return failureMatcherForState(types.SpecStateTimedout, "Failure.Message", options...)
}

func HaveAborted(options ...interface{}) OmegaMatcher {
	return failureMatcherForState(types.SpecStateAborted, "Failure.Message", options...)
}

func HavePanicked(options ...interface{}) OmegaMatcher {
	return failureMatcherForState(types.SpecStatePanicked, "Failure.ForwardedPanic", options...)
}

func TLWithOffset[O int | string](o O) types.TimelineLocation {
	t := types.TimelineLocation{}
	switch x := any(o).(type) {
	case int:
		t.Offset = x
	case string:
		t.Offset = len(x)
	}
	return t
}

func BeSpecEvent(options ...interface{}) OmegaMatcher {
	description := []string{"BeSpecEvent"}
	matchers := []OmegaMatcher{}
	for _, option := range options {
		var matcher OmegaMatcher
		switch x := option.(type) {
		case types.SpecEventType:
			matcher = HaveField("SpecEventType", x)
			description = append(description, "["+x.String()+" SpecEvent]")
		case types.CodeLocation:
			matcher = HaveField("CodeLocation", x)
			description = append(description, "CL="+x.String())
		case types.TimelineLocation:
			matcher = HaveField("TimelineLocation.Offset", x.Offset)
			description = append(description, fmt.Sprintf("TL.Offset=%d", x.Offset))
		case string:
			matcher = HaveField("Message", ContainSubstring(x))
			description = append(description, `Message="`+x+`"`)
		case int:
			matcher = HaveField("Attempt", x)
			description = append(description, fmt.Sprintf("Attempt=%d", x))
		case time.Duration:
			matcher = HaveField("Duration", BeNumerically("~", x, time.Duration(float64(x)*0.2)))
			description = append(description, "Duration="+x.String())
		case types.NodeType:
			matcher = HaveField("NodeType", x)
			description = append(description, "NodeType="+x.String())
		}
		if matcher != nil {
			matchers = append(matchers, matcher)
		}
	}
	return OmegaMatcherWithDescription{OmegaMatcher: And(matchers...), Description: strings.Join(description, " ")}
}

func BeProgressReport(options ...interface{}) OmegaMatcher {
	description := []string{"BeProgressReport"}
	matchers := []OmegaMatcher{}
	for _, option := range options {
		var matcher OmegaMatcher
		switch x := option.(type) {
		case string:
			matcher = HaveField("Message", ContainSubstring(x))
			description = append(description, `Message="`+x+`"`)
		case types.TimelineLocation:
			matcher = HaveField("TimelineLocation.Offset", x.Offset)
			description = append(description, fmt.Sprintf("TL.Offset=%d", x.Offset))
		case types.CodeLocation:
			matcher = HaveField("CurrentNodeLocation", x)
			description = append(description, "CurrentNodeLocation="+x.String())
		case types.NodeType:
			matcher = HaveField("CurrentNodeType", x)
			description = append(description, "CurrentNodeType="+x.String())
		}
		if matcher != nil {
			matchers = append(matchers, matcher)
		}
	}
	return OmegaMatcherWithDescription{OmegaMatcher: And(matchers...), Description: strings.Join(description, " ")}
}

func BeReportEntry(options ...interface{}) OmegaMatcher {
	description := []string{"BeReportEntry"}
	matchers := []OmegaMatcher{}
	for _, option := range options {
		var matcher OmegaMatcher
		switch x := option.(type) {
		case string:
			matcher = HaveField("Name", ContainSubstring(x))
			description = append(description, `Name="`+x+`"`)
		case types.TimelineLocation:
			matcher = HaveField("TimelineLocation.Offset", x.Offset)
			description = append(description, fmt.Sprintf("TL.Offset=%d", x.Offset))
		case types.ReportEntryVisibility:
			matcher = HaveField("Visibility", x)
			description = append(description, "Visibility="+x.String())
		}
		if matcher != nil {
			matchers = append(matchers, matcher)
		}
	}
	return OmegaMatcherWithDescription{OmegaMatcher: And(matchers...), Description: strings.Join(description, " ")}
}

func BeTimelineContaining(matchers ...OmegaMatcher) OmegaMatcher {
	return gcustom.MakeMatcher(func(timeline types.Timeline) (bool, error) {
		timelineIdx := 0
		for _, matcher := range matchers {
			for {
				if timelineIdx >= len(timeline) {
					return false, nil
				}
				event := timeline[timelineIdx]
				timelineIdx += 1
				success, err := matcher.Match(event)
				if success && err == nil {
					break
				}
			}
		}
		return true, nil
	}).WithTemplate("Expected:\n{{.FormattedActual}}\n{{.To}} contain events matching (in order):\n{{format .Data 1}}", matchers)
}

func BeTimelineExactlyMatching(matchers ...OmegaMatcher) OmegaMatcher {
	data := map[string]any{}
	data["Matchers"] = matchers
	return gcustom.MakeMatcher(func(timeline types.Timeline) (bool, error) {
		for idx, matcher := range matchers {
			if idx == len(timeline) {
				data["LengthMismatch"] = "Not enough timeline entries"
				return false, nil
			}
			event := timeline[idx]
			success, err := matcher.Match(event)
			if !(success && err == nil) {
				data["Failure"] = matcher.FailureMessage(event)
				data["FailedIndex"] = idx
				return false, nil
			}
		}

		if len(matchers) < len(timeline) {
			data["LengthMismatch"] = "Not enough matcher entries"
			return false, nil
		}

		return true, nil
	}).WithTemplate(`Timeline failed to match:
{{if .Data.LengthMismatch}}
  {{.Data.LengthMismatch}}
  Timeline has {{len .Actual}} events:
{{ range .Actual }}{{ printf "    %T\n" . }}{{ end }}
  Matchers has {{len .Data.Matchers}} entries.
{{else}}
  Failed at index {{.Data.FailedIndex}}:
{{format (index .Actual .Data.FailedIndex) 2}}
{{.Data.Failure}}
{{end}}`, data)
}
