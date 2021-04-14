package reporters

import (
	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/types"
)

type Reporter interface {
	SpecSuiteWillBegin(report types.Report)
	WillRun(report types.SpecReport)
	DidRun(report types.SpecReport)
	SpecSuiteDidEnd(report types.Report)
}

// TODO: FIX
// Deprecated Custom Reporters in V2

// Deprecated: DeprecatedReporter was how Ginkgo V1 provided support for CustomReporters
// this has been removed in V2.
// Please read the documentation at:
// https://github.com/onsi/ginkgo/blob/v2/docs/MIGRATING_TO_V2.md#removed-custom-reporters
// for Ginkgo's new behavior and for a migration path.
type DeprecatedReporter interface {
	SpecSuiteWillBegin(config config.GinkgoConfigType, summary *types.SuiteSummary)
	BeforeSuiteDidRun(setupSummary *types.SetupSummary)
	SpecWillRun(specSummary *types.SpecSummary)
	SpecDidComplete(specSummary *types.SpecSummary)
	AfterSuiteDidRun(setupSummary *types.SetupSummary)
	SpecSuiteDidEnd(summary *types.SuiteSummary)
}

// ReportViaDeprecatedReporter takes a V1 custom reporter and a V2 report and
// calls the custom reporter's methods with appropriately transformed data from the V2 report.
//
// ReportViaDeprecatedReporter should be called in a `ReportAfterSuite()`
//
// Deprecated: ReportViaDeprecatedReporter method exists to help developer bridge between deprecated V1 functionality and the new
// reporting support in V2.  It will be removed in a future minor version of Ginkgo.
func ReportViaDeprecatedReporter(reporter DeprecatedReporter, report types.SuiteSummary) {
	//TODO flesh this out

	// type compatiblityShim struct {
	// 	reporter V1Reporter
	// }

	// func (cs *compatiblityShim) IsDeprecatedReporter() {}

	// func (cs *compatiblityShim) SpecSuiteWillBegin(config types.SuiteConfig, summary types.SuiteSummary) {
	// 	s := summary
	// 	cs.reporter.SpecSuiteWillBegin(config, &s)
	// }

	// func (cs *compatiblityShim) WillRun(summary types.Summary) {
	// 	if summary.LeafNodeType.Is(types.NodeTypesForSuiteSetup...) {
	// 		return
	// 	}
	// 	cs.reporter.SpecWillRun(types.DeprecatedSpecSummaryFromSummary(summary))
	// }

	// func (cs *compatiblityShim) DidRun(summary types.Summary) {
	// 	if summary.LeafNodeType.Is(types.NodeTypesForSuiteSetup...) {
	// 		if summary.LeafNodeType.Is(types.NodeTypeBeforeSuite, types.NodeTypeSynchronizedBeforeSuite) {
	// 			cs.reporter.BeforeSuiteDidRun(types.DeprecatedSetupSummaryFromSummary(summary))
	// 		} else {
	// 			cs.reporter.AfterSuiteDidRun(types.DeprecatedSetupSummaryFromSummary(summary))
	// 		}
	// 	} else {
	// 		cs.reporter.SpecDidComplete(types.DeprecatedSpecSummaryFromSummary(summary))
	// 	}
	// }

	// func (cs *compatiblityShim) SpecSuiteDidEnd(summary types.SuiteSummary) {
	// 	s := summary
	// 	cs.reporter.SpecSuiteDidEnd(&s)
	// }
}
