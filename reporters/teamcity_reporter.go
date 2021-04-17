/*

TeamCity Reporter for Ginkgo

Makes use of TeamCity's support for Service Messages
http://confluence.jetbrains.com/display/TCD7/Build+Script+Interaction+with+TeamCity#BuildScriptInteractionwithTeamCity-ReportingTests
*/

package reporters

import (
	"fmt"
	"io"
	"strings"

	"github.com/onsi/ginkgo/types"
)

const (
	teamcityMessageId = "##teamcity"
)

type TeamCityReporter struct {
	writer         io.Writer
	testSuiteName  string
	ReporterConfig types.ReporterConfig
}

func NewTeamCityReporter(writer io.Writer) *TeamCityReporter {
	return &TeamCityReporter{
		writer: writer,
	}
}

func (reporter *TeamCityReporter) SuiteWillBegin(conf types.SuiteConfig, summary types.SuiteSummary) {
	reporter.testSuiteName = reporter.escape(summary.SuiteDescription)
	// reporter.ReporterConfig = config.DefaultReporterConfig //TODO: NEED TO REPLICATE THIS LESS TERRIBLY

	fmt.Fprintf(reporter.writer, "%s[testSuiteStarted name='%s']\n", teamcityMessageId, reporter.testSuiteName)
}

func (reporter *TeamCityReporter) testNameFor(report types.SpecReport) string {
	if report.LeafNodeType.Is(types.NodeTypesForSuiteLevelNodes...) {
		return reporter.escape(report.LeafNodeType.String())
	} else {
		return reporter.escape(strings.Join(report.NodeTexts, " "))
	}
}

func (reporter *TeamCityReporter) WillRun(report types.SpecReport) {
	fmt.Fprintf(reporter.writer, "%s[testStarted name='%s']\n", teamcityMessageId, reporter.testNameFor(report))
}

func (reporter *TeamCityReporter) DidRun(report types.SpecReport) {
	testName := reporter.testNameFor(report)

	if reporter.ReporterConfig.ReportPassed && report.State == types.SpecStatePassed {
		details := reporter.escape(report.CombinedOutput())
		fmt.Fprintf(reporter.writer, "%s[testPassed name='%s' details='%s']\n", teamcityMessageId, testName, details)
	}
	if report.State.Is(types.SpecStateFailureStates...) {
		message := reporter.failureMessage(report.Failure)
		details := reporter.failureDetails(report.Failure)
		fmt.Fprintf(reporter.writer, "%s[testFailed name='%s' message='%s' details='%s']\n", teamcityMessageId, testName, message, details)
	}
	if report.State == types.SpecStateSkipped || report.State == types.SpecStatePending {
		fmt.Fprintf(reporter.writer, "%s[testIgnored name='%s']\n", teamcityMessageId, testName)
	}

	durationInMilliseconds := report.RunTime.Seconds() * 1000
	fmt.Fprintf(reporter.writer, "%s[testFinished name='%s' duration='%v']\n", teamcityMessageId, testName, durationInMilliseconds)
}

func (reporter *TeamCityReporter) SuiteDidEnd(summary types.SuiteSummary) {
	fmt.Fprintf(reporter.writer, "%s[testSuiteFinished name='%s']\n", teamcityMessageId, reporter.testSuiteName)
}

func (reporter *TeamCityReporter) failureMessage(failure types.Failure) string {
	return reporter.escape(failure.NodeType.String())
}

func (reporter *TeamCityReporter) failureDetails(failure types.Failure) string {
	return reporter.escape(fmt.Sprintf("%s\n%s", failure.Message, failure.Location.String()))
}

func (reporter *TeamCityReporter) escape(output string) string {
	output = strings.Replace(output, "|", "||", -1)
	output = strings.Replace(output, "'", "|'", -1)
	output = strings.Replace(output, "\n", "|n", -1)
	output = strings.Replace(output, "\r", "|r", -1)
	output = strings.Replace(output, "[", "|[", -1)
	output = strings.Replace(output, "]", "|]", -1)
	return output
}
