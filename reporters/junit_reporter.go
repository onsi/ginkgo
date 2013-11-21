/*

JUnit XML Reporter for Ginkgo

For usage instructions: http://onsi.github.io/ginkgo/#generating_junit_xml_output

*/

package reporters

import (
	"encoding/xml"
	"fmt"
	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/types"
	"os"
	"strings"
)

type JUnitTestSuite struct {
	XMLName   xml.Name        `xml:"testsuite"`
	TestCases []JUnitTestCase `xml:"testcase"`
	Tests     int             `xml:"tests,attr"`
	Failures  int             `xml:"failures,attr"`
	Time      float64         `xml:"time,attr"`
}

type JUnitTestCase struct {
	Name           string               `xml:"name,attr"`
	ClassName      string               `xml:"classname,attr"`
	FailureMessage *JUnitFailureMessage `xml:"failure,omitempty"`
	Skipped        *JUnitSkipped        `xml:"skipped,omitempty"`
	Time           float64              `xml:"time,attr"`
}

type JUnitFailureMessage struct {
	Type    string `xml:"type,attr"`
	Message string `xml:",chardata"`
}

type JUnitSkipped struct {
	XMLName xml.Name `xml:"skipped"`
}

type JUnitReporter struct {
	suite         JUnitTestSuite
	filename      string
	testSuiteName string
}

//NewJUnitReporter creates a new JUnit XML reporter.  The XML will be stored in the passed in filename.
func NewJUnitReporter(filename string) *JUnitReporter {
	return &JUnitReporter{
		filename: filename,
	}
}

func (reporter *JUnitReporter) SpecSuiteWillBegin(config config.GinkgoConfigType, summary *types.SuiteSummary) {
	reporter.suite = JUnitTestSuite{
		Tests:     summary.NumberOfExamplesThatWillBeRun,
		TestCases: []JUnitTestCase{},
	}
	reporter.testSuiteName = summary.SuiteDescription
}

func (reporter *JUnitReporter) ExampleWillRun(exampleSummary *types.ExampleSummary) {
}

func (reporter *JUnitReporter) ExampleDidComplete(exampleSummary *types.ExampleSummary) {
	testCase := JUnitTestCase{
		Name:      strings.Join(exampleSummary.ComponentTexts[1:], " "),
		ClassName: reporter.testSuiteName,
	}
	if exampleSummary.State == types.ExampleStateFailed || exampleSummary.State == types.ExampleStateTimedOut || exampleSummary.State == types.ExampleStatePanicked {
		failureType := ""
		switch exampleSummary.State {
		case types.ExampleStateFailed:
			failureType = "Failure"
		case types.ExampleStateTimedOut:
			failureType = "Timeout"
		case types.ExampleStatePanicked:
			failureType = "Panic"
		}
		testCase.FailureMessage = &JUnitFailureMessage{
			Type:    failureType,
			Message: fmt.Sprintf("%s\n%s", exampleSummary.Failure.ComponentCodeLocation.String(), exampleSummary.Failure.Message),
		}
	}
	if exampleSummary.State == types.ExampleStateSkipped || exampleSummary.State == types.ExampleStatePending {
		testCase.Skipped = &JUnitSkipped{}
	}
	testCase.Time = exampleSummary.RunTime.Seconds()
	reporter.suite.TestCases = append(reporter.suite.TestCases, testCase)
}

func (reporter *JUnitReporter) SpecSuiteDidEnd(summary *types.SuiteSummary) {
	reporter.suite.Time = summary.RunTime.Seconds()
	reporter.suite.Failures = summary.NumberOfFailedExamples
	file, err := os.Create(reporter.filename)
	if err != nil {
		fmt.Printf("Failed to create JUnit report file: %s\n\t%s", reporter.filename, err.Error())
	}
	defer file.Close()
	file.WriteString(xml.Header)
	encoder := xml.NewEncoder(file)
	encoder.Indent("  ", "    ")
	err = encoder.Encode(reporter.suite)
	if err != nil {
		fmt.Printf("Failed to generate JUnit report\n\t%s", err.Error())
	}
}
