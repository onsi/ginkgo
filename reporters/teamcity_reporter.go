/*

TeamCity Reporter for Ginkgo

Makes use of TeamCity's support for Service Messages
http://confluence.jetbrains.com/display/TCD7/Build+Script+Interaction+with+TeamCity#BuildScriptInteractionwithTeamCity-ReportingTests
*/

package reporters

import (
	"fmt"
	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/types"
	"io"
	"strings"
)

const (
	messageId = "##teamcity"
)

type TeamCityReporter struct {
	writer        io.Writer
	testSuiteName string
}

func NewTeamCityReporter(writer io.Writer) *TeamCityReporter {
	return &TeamCityReporter{
		writer: writer,
	}
}

func (reporter *TeamCityReporter) SpecSuiteWillBegin(config config.GinkgoConfigType, summary *types.SuiteSummary) {
	reporter.testSuiteName = escape(summary.SuiteDescription)
	fmt.Fprintf(reporter.writer, "%s[testSuiteStarted name='%s']", messageId, reporter.testSuiteName)
}

func (reporter *TeamCityReporter) ExampleWillRun(exampleSummary *types.ExampleSummary) {
	testName := escape(strings.Join(exampleSummary.ComponentTexts[1:], " "))
	fmt.Fprintf(reporter.writer, "%s[testStarted name='%s']", messageId, testName)
}

func (reporter *TeamCityReporter) ExampleDidComplete(exampleSummary *types.ExampleSummary) {
	testName := escape(strings.Join(exampleSummary.ComponentTexts[1:], " "))

	if exampleSummary.State == types.ExampleStateFailed || exampleSummary.State == types.ExampleStateTimedOut || exampleSummary.State == types.ExampleStatePanicked {
		message := escape(exampleSummary.Failure.ComponentCodeLocation.String())
		details := escape(exampleSummary.Failure.Message)
		fmt.Fprintf(reporter.writer, "%s[testFailed name='%s' message='%s' details='%s']", messageId, testName, message, details)
	}
	if exampleSummary.State == types.ExampleStateSkipped || exampleSummary.State == types.ExampleStatePending {
		fmt.Fprintf(reporter.writer, "%s[testIgnored name='%s']", messageId, testName)
	}

	durationInMilliseconds := exampleSummary.RunTime.Seconds() * 1000
	fmt.Fprintf(reporter.writer, "%s[testFinished name='%s' duration='%v']", messageId, testName, durationInMilliseconds)
}

func (reporter *TeamCityReporter) SpecSuiteDidEnd(summary *types.SuiteSummary) {
	fmt.Fprintf(reporter.writer, "%s[testSuiteFinished name='%s']", messageId, reporter.testSuiteName)
}

func escape(output string) string {
	output = strings.Replace(output, "|", "||", -1)
	output = strings.Replace(output, "'", "|'", -1)
	output = strings.Replace(output, "\n", "|n", -1)
	output = strings.Replace(output, "\r", "|r", -1)
	output = strings.Replace(output, "[", "|[", -1)
	output = strings.Replace(output, "]", "|]", -1)
	return output
}
