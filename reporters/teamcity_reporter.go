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

	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/types"
)

const (
	teamcityMessageId = "##teamcity"
)

type TeamCityReporter struct {
	writer         io.Writer
	testSuiteName  string
	ReporterConfig config.DefaultReporterConfigType
}

func NewTeamCityReporter(writer io.Writer) *TeamCityReporter {
	return &TeamCityReporter{
		writer: writer,
	}
}

func (reporter *TeamCityReporter) SpecSuiteWillBegin(conf config.GinkgoConfigType, summary types.SuiteSummary) {
	reporter.testSuiteName = reporter.escape(summary.SuiteDescription)
	reporter.ReporterConfig = config.DefaultReporterConfig

	fmt.Fprintf(reporter.writer, "%s[testSuiteStarted name='%s']\n", teamcityMessageId, reporter.testSuiteName)
}

func (reporter *TeamCityReporter) testNameFor(summary types.Summary) string {
	if summary.LeafNodeType.Is(types.NodeTypesForSuiteSetup...) {
		return reporter.escape(summary.LeafNodeType.String())
	} else {
		return reporter.escape(strings.Join(summary.NodeTexts, " "))
	}
}

func (reporter *TeamCityReporter) WillRun(summary types.Summary) {
	fmt.Fprintf(reporter.writer, "%s[testStarted name='%s']\n", teamcityMessageId, reporter.testNameFor(summary))
}

func (reporter *TeamCityReporter) DidRun(summary types.Summary) {
	testName := reporter.testNameFor(summary)

	if reporter.ReporterConfig.ReportPassed && summary.State == types.SpecStatePassed {
		details := reporter.escape(summary.CombinedOutput())
		fmt.Fprintf(reporter.writer, "%s[testPassed name='%s' details='%s']\n", teamcityMessageId, testName, details)
	}
	if summary.State.Is(types.SpecStateFailureStates...) {
		message := reporter.failureMessage(summary.Failure)
		details := reporter.failureDetails(summary.Failure)
		fmt.Fprintf(reporter.writer, "%s[testFailed name='%s' message='%s' details='%s']\n", teamcityMessageId, testName, message, details)
	}
	if summary.State == types.SpecStateSkipped || summary.State == types.SpecStatePending {
		fmt.Fprintf(reporter.writer, "%s[testIgnored name='%s']\n", teamcityMessageId, testName)
	}

	durationInMilliseconds := summary.RunTime.Seconds() * 1000
	fmt.Fprintf(reporter.writer, "%s[testFinished name='%s' duration='%v']\n", teamcityMessageId, testName, durationInMilliseconds)
}

func (reporter *TeamCityReporter) SpecSuiteDidEnd(summary types.SuiteSummary) {
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
