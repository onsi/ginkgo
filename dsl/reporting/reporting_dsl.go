/*
Ginkgo is usually dot-imported via:

	import . "github.com/onsi/ginkgo/v2"

however some parts of the DSL may conflict with existing symbols in the user's code.

To mitigate this without losing the brevity of dot-importing Ginkgo the various packages in the
dsl directory provide pieces of the Ginkgo DSL that can be dot-imported separately.

This "reporting" package pulls in the reporting-related pieces of the Ginkgo DSL.
*/
package reporting

import (
	"github.com/onsi/ginkgo/v2"
)

type Report = ginkgo.Report
type SpecReport = ginkgo.SpecReport
type ReportEntryVisibility = ginkgo.ReportEntryVisibility

const ReportEntryVisibilityAlways, ReportEntryVisibilityFailureOrVerbose, ReportEntryVisibilityNever = ginkgo.ReportEntryVisibilityAlways, ginkgo.ReportEntryVisibilityFailureOrVerbose, ginkgo.ReportEntryVisibilityNever

var CurrentSpecReport = ginkgo.CurrentSpecReport
var AddReportEntry = ginkgo.AddReportEntry

var ReportBeforeEach = ginkgo.ReportBeforeEach
var ReportAfterEach = ginkgo.ReportAfterEach
var ReportBeforeSuite = ginkgo.ReportBeforeSuite
var ReportAfterSuite = ginkgo.ReportAfterSuite
