package test_helpers

import (
	"reflect"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	"github.com/onsi/ginkgo/config"
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
	Config config.GinkgoConfigType
	Begin  types.SuiteSummary
	Will   Reports
	Did    Reports
	End    types.SuiteSummary
}

func (r *FakeReporter) SpecSuiteWillBegin(conf config.GinkgoConfigType, summary types.SuiteSummary) {
	r.Begin = summary
	r.Config = conf
}

func (r *FakeReporter) WillRun(report types.SpecReport) {
	r.Will = append(r.Will, report)
}

func (r *FakeReporter) DidRun(report types.SpecReport) {
	r.Did = append(r.Did, report)
}

func (r *FakeReporter) SpecSuiteDidEnd(summary types.SuiteSummary) {
	r.End = summary
}

type NSpecs int
type NWillRun int
type NPassed int
type NSkipped int
type NFailed int
type NPending int
type NFlaked int

func BeASuiteSummary(options ...interface{}) OmegaMatcher {
	fields := Fields{}
	fields["NumberOfTotalSpecs"] = Equal(0)
	fields["NumberOfPassedSpecs"] = Equal(0)
	fields["NumberOfSkippedSpecs"] = Equal(0)
	fields["NumberOfFailedSpecs"] = Equal(0)
	fields["NumberOfPendingSpecs"] = Equal(0)
	fields["NumberOfFlakedSpecs"] = Equal(0)

	for _, option := range options {
		t := reflect.TypeOf(option)
		if t.Kind() == reflect.Bool {
			if option.(bool) {
				fields["SuiteSucceeded"] = BeTrue()
			} else {
				fields["SuiteSucceeded"] = BeFalse()
			}
		} else if t == reflect.TypeOf(NSpecs(0)) {
			fields["NumberOfTotalSpecs"] = Equal(int(option.(NSpecs)))
		} else if t == reflect.TypeOf(NWillRun(0)) {
			fields["NumberOfSpecsThatWillBeRun"] = Equal(int(option.(NWillRun)))
		} else if t == reflect.TypeOf(NPassed(0)) {
			fields["NumberOfPassedSpecs"] = Equal(int(option.(NPassed)))
		} else if t == reflect.TypeOf(NSkipped(0)) {
			fields["NumberOfSkippedSpecs"] = Equal(int(option.(NSkipped)))
		} else if t == reflect.TypeOf(NFailed(0)) {
			fields["NumberOfFailedSpecs"] = Equal(int(option.(NFailed)))
		} else if t == reflect.TypeOf(NPending(0)) {
			fields["NumberOfPendingSpecs"] = Equal(int(option.(NPending)))
		} else if t == reflect.TypeOf(NFlaked(0)) {
			fields["NumberOfFlakedSpecs"] = Equal(int(option.(NFlaked)))
		}
	}

	return MatchFields(IgnoreExtras, fields)
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
