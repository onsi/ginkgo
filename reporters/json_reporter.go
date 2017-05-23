package reporters

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/types"
)

type JsonReporter struct {
	suite         JsonTestSuite
	filename      string
	testSuiteName string
}

type JsonTestSuite struct {
	SuiteName string         `json:"suite_name"`
	TestCases []JsonTestCase `json:"test_cases"`
	Tests     int            `json:"tests"`
	Passed    int            `json:"passed"`
	Skipped   int            `json:"skipped"`
	Failures  *JsonFailure   `json:"failures"`
	Time      float64        `json:"total_time"`
}

type JsonFailure struct {
	Number int      `json:"number"`
	Names  []string `json:"names"`
}

type JsonTestCase struct {
	TestName  string          `json:"test_name"`
	SuiteName string          `json:"suite_name"`
	Passed    bool            `json:"passed"`
	Skipped   bool            `json:"skipped"`
	Failed    *JsonFailedTest `json:"failed"`
	Time      float64         `json:"time"`
}

type JsonFailedTest struct {
	Type    string `json:"type"`
	Message string `json:"error_message"`
	Where   string `json:"error_line"`
}

//NewJsonReporter creates a new Json reporter.  The Json will be stored in the passed in filename.
func NewJsonReporter(filename string) *JsonReporter {
	return &JsonReporter{
		filename: filename,
	}
}

func (reporter *JsonReporter) SpecSuiteWillBegin(config config.GinkgoConfigType, summary *types.SuiteSummary) {
	reporter.suite = JsonTestSuite{
		SuiteName: summary.SuiteDescription,
		TestCases: []JsonTestCase{},
		Failures:  &JsonFailure{},
	}
	reporter.testSuiteName = summary.SuiteDescription
}

func (reporter *JsonReporter) BeforeSuiteDidRun(setupSummary *types.SetupSummary) {
	reporter.handleSetupSummary("BeforeSuite", setupSummary)
}

func (reporter *JsonReporter) AfterSuiteDidRun(setupSummary *types.SetupSummary) {
	reporter.handleSetupSummary("AfterSuite", setupSummary)
}

func (reporter *JsonReporter) handleSetupSummary(name string, setupSummary *types.SetupSummary) {
	if setupSummary.State != types.SpecStatePassed {
		testCase := JsonTestCase{
			TestName:  name,
			SuiteName: reporter.testSuiteName,
		}

		testCase.Failed = &JsonFailedTest{
			Type:    reporter.failureTypeForState(setupSummary.State),
			Message: setupSummary.Failure.Message,
			Where:   setupSummary.Failure.ComponentCodeLocation.String(),
		}
		testCase.Time = setupSummary.RunTime.Seconds()
		reporter.suite.TestCases = append(reporter.suite.TestCases, testCase)

		reporter.suite.Failures.Names = append(reporter.suite.Failures.Names, name)
	}
}

func (reporter *JsonReporter) SpecWillRun(specSummary *types.SpecSummary) {
}

func (reporter *JsonReporter) SpecDidComplete(specSummary *types.SpecSummary) {
	testCase := JsonTestCase{
		TestName:  strings.Join(specSummary.ComponentTexts[1:], " "),
		SuiteName: reporter.testSuiteName,
		Passed:    true,
	}
	if specSummary.State == types.SpecStateFailed || specSummary.State == types.SpecStateTimedOut || specSummary.State == types.SpecStatePanicked {
		testCase.Failed = &JsonFailedTest{
			Type:    reporter.failureTypeForState(specSummary.State),
			Message: specSummary.Failure.Message,
			Where:   specSummary.Failure.ComponentCodeLocation.String(),
		}
		testCase.Passed = false
		reporter.suite.Failures.Names = append(reporter.suite.Failures.Names, testCase.TestName)
	}
	if specSummary.State == types.SpecStateSkipped || specSummary.State == types.SpecStatePending {
		testCase.Skipped = true
		testCase.Passed = false
	}
	testCase.Time = specSummary.RunTime.Seconds()
	reporter.suite.TestCases = append(reporter.suite.TestCases, testCase)
}

func (reporter *JsonReporter) SpecSuiteDidEnd(summary *types.SuiteSummary) {
	reporter.suite.Tests = summary.NumberOfTotalSpecs
	reporter.suite.Time = summary.RunTime.Seconds()
	reporter.suite.Passed = summary.NumberOfPassedSpecs
	reporter.suite.Failures.Number = summary.NumberOfFailedSpecs
	reporter.suite.Skipped = summary.NumberOfSkippedSpecs
	file, err := os.Create(reporter.filename)
	if err != nil {
		fmt.Printf("Failed to create Json report file: %s\n\t%s", reporter.filename, err.Error())
	}
	defer file.Close()
	jsonBytes, err := json.MarshalIndent(reporter.suite, "", "  ")
	if err != nil {
		fmt.Printf("Failed to generate Json report\n\t%s", err.Error())
	}

	_, err1 := file.Write(jsonBytes)
	if err1 != nil {
		fmt.Printf("Failed to generate Json report\n\t%s", err.Error())
	}

}

func (reporter *JsonReporter) failureTypeForState(state types.SpecState) string {
	switch state {
	case types.SpecStateFailed:
		return "Failure"
	case types.SpecStateTimedOut:
		return "Timeout"
	case types.SpecStatePanicked:
		return "Panic"
	default:
		return ""
	}
}
