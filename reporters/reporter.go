package reporters

import (
	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/types"
)

type Reporter interface {
	SpecSuiteWillBegin(config config.GinkgoConfigType, summary types.SuiteSummary)
	WillRun(specSummary types.Summary)
	DidRun(specSummary types.Summary)
	SpecSuiteDidEnd(summary types.SuiteSummary)
}

// Old V1Reporter compatibility support
type V1Reporter interface {
	SpecSuiteWillBegin(config config.GinkgoConfigType, summary *types.SuiteSummary)
	BeforeSuiteDidRun(setupSummary *types.SetupSummary)
	SpecWillRun(specSummary *types.SpecSummary)
	SpecDidComplete(specSummary *types.SpecSummary)
	AfterSuiteDidRun(setupSummary *types.SetupSummary)
	SpecSuiteDidEnd(summary *types.SuiteSummary)
}

type compatiblityShim struct {
	reporter V1Reporter
}

func (cs *compatiblityShim) IsDeprecatedReporter() {}

func (cs *compatiblityShim) SpecSuiteWillBegin(config config.GinkgoConfigType, summary types.SuiteSummary) {
	s := summary
	cs.reporter.SpecSuiteWillBegin(config, &s)
}

func (cs *compatiblityShim) WillRun(summary types.Summary) {
	if summary.LeafNodeType.Is(types.NodeTypesForSuiteSetup...) {
		return
	}
	cs.reporter.SpecWillRun(types.DeprecatedSpecSummaryFromSummary(summary))
}

func (cs *compatiblityShim) DidRun(summary types.Summary) {
	if summary.LeafNodeType.Is(types.NodeTypesForSuiteSetup...) {
		if summary.LeafNodeType.Is(types.NodeTypeBeforeSuite, types.NodeTypeSynchronizedBeforeSuite) {
			cs.reporter.BeforeSuiteDidRun(types.DeprecatedSetupSummaryFromSummary(summary))
		} else {
			cs.reporter.AfterSuiteDidRun(types.DeprecatedSetupSummaryFromSummary(summary))
		}
	} else {
		cs.reporter.SpecDidComplete(types.DeprecatedSpecSummaryFromSummary(summary))
	}
}

func (cs *compatiblityShim) SpecSuiteDidEnd(summary types.SuiteSummary) {
	s := summary
	cs.reporter.SpecSuiteDidEnd(&s)
}

func ReporterFromV1Reporter(reporter V1Reporter) Reporter {
	return &compatiblityShim{
		reporter: reporter,
	}
}
