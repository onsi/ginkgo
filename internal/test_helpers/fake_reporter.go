package test_helpers

import (
	"reflect"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	"github.com/onsi/ginkgo/types"
)

/*

A FakeReporter and collection of matchers to match against reported suite and spec summaries

*/

type Reports []types.SpecReport

func (s Reports) FindByLeafNodeType(nodeType ...types.NodeType) types.SpecReport {
	for _, summary := range s {
		if summary.LeafNodeType.Is(nodeType...) {
			return summary
		}
	}

	return types.SpecReport{}
}

func (s Reports) Find(name string) types.SpecReport {
	for _, summary := range s {
		if len(summary.NodeTexts) > 0 && summary.NodeTexts[len(summary.NodeTexts)-1] == name {
			return summary
		}
	}

	return types.SpecReport{}
}

func (s Reports) Names() []string {
	out := []string{}
	for _, summary := range s {
		if len(summary.NodeTexts) > 0 {
			out = append(out, summary.NodeTexts[len(summary.NodeTexts)-1])
		}
	}
	return out
}

func (s Reports) WithState(state types.SpecState) Reports {
	out := Reports{}
	for _, summary := range s {
		if summary.State == state {
			out = append(out, summary)
		}
	}
	return out
}

type FakeReporter struct {
	Begin types.Report
	Will  Reports
	Did   Reports
	End   types.Report
}

func (r *FakeReporter) SpecSuiteWillBegin(report types.Report) {
	r.Begin = report
}

func (r *FakeReporter) WillRun(report types.SpecReport) {
	r.Will = append(r.Will, report)
}

func (r *FakeReporter) DidRun(report types.SpecReport) {
	r.Did = append(r.Did, report)
}

func (r *FakeReporter) SpecSuiteDidEnd(report types.Report) {
	r.End = report
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
			Failed:       specs.CountWithState(types.SpecStateFailureStates...),
			Pending:      specs.CountWithState(types.SpecStatePending),
			Flaked:       specs.CountOfFlakedSpecs(),
		}
	}, MatchFields(IgnoreExtras, fields))
}

type CapturedOutput string
type NumAttempts int

func HavePassed(options ...interface{}) OmegaMatcher {
	fields := Fields{
		"State":   Equal(types.SpecStatePassed),
		"Failure": BeZero(),
	}
	for _, option := range options {
		t := reflect.TypeOf(option)
		if t == reflect.TypeOf(CapturedOutput("")) {
			fields["CapturedGinkgoWriterOutput"] = Equal(string(option.(CapturedOutput)))
		} else if t == reflect.TypeOf(NumAttempts(0)) {
			fields["NumAttempts"] = Equal(int(option.(NumAttempts)))
		}
	}
	return MatchFields(IgnoreExtras, fields)
}

func BePending() OmegaMatcher {
	return MatchFields(IgnoreExtras, Fields{
		"State":   Equal(types.SpecStatePending),
		"Failure": BeZero(),
	})
}

func HaveBeenSkipped() OmegaMatcher {
	return MatchFields(IgnoreExtras, Fields{
		"State":   Equal(types.SpecStateSkipped),
		"Failure": BeZero(),
	})
}

func HaveFailed(options ...interface{}) OmegaMatcher {
	fields := Fields{
		"State": Equal(types.SpecStateFailed),
	}
	failureFields := Fields{}
	for _, option := range options {
		t := reflect.TypeOf(option)
		if t == reflect.TypeOf(CapturedOutput("")) {
			fields["CapturedGinkgoWriterOutput"] = Equal(string(option.(CapturedOutput)))
		} else if t.Kind() == reflect.String {
			failureFields["Message"] = Equal(option.(string))
		} else if t == reflect.TypeOf(types.CodeLocation{}) {
			failureFields["Location"] = Equal(option.(types.CodeLocation))
		} else if t == reflect.TypeOf(NumAttempts(0)) {
			fields["NumAttempts"] = Equal(int(option.(NumAttempts)))
		}
	}
	if len(failureFields) > 0 {
		fields["Failure"] = MatchFields(IgnoreExtras, failureFields)
	}
	return MatchFields(IgnoreExtras, fields)
}

func HavePanicked(options ...interface{}) OmegaMatcher {
	fields := Fields{
		"State": Equal(types.SpecStatePanicked),
	}
	failureFields := Fields{}
	for _, option := range options {
		t := reflect.TypeOf(option)
		if t == reflect.TypeOf(CapturedOutput("")) {
			fields["CapturedGinkgoWriterOutput"] = Equal(string(option.(CapturedOutput)))
		} else if t.Kind() == reflect.String {
			failureFields["ForwardedPanic"] = Equal(option.(string))
		} else if t == reflect.TypeOf(NumAttempts(0)) {
			fields["NumAttempts"] = Equal(int(option.(NumAttempts)))
		}
	}
	if len(failureFields) > 0 {
		fields["Failure"] = MatchFields(IgnoreExtras, failureFields)
	}
	return MatchFields(IgnoreExtras, fields)
}
