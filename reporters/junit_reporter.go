/*

JUnit XML Reporter for Ginkgo

For usage instructions: http://onsi.github.io/ginkgo/#generating_junit_xml_output

*/

package reporters

import (
	"encoding/xml"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/types"
)

type JUnitTestSuite struct {
	XMLName   xml.Name        `xml:"testsuite"`
	TestCases []JUnitTestCase `xml:"testcase"`
	Name      string          `xml:"name,attr"`
	Tests     int             `xml:"tests,attr"`
	Failures  int             `xml:"failures,attr"`
	Errors    int             `xml:"errors,attr"`
	Time      float64         `xml:"time,attr"`
}

type JUnitTestCase struct {
	Name           string               `xml:"name,attr"`
	ClassName      string               `xml:"classname,attr"`
	FailureMessage *JUnitFailureMessage `xml:"failure,omitempty"`
	Skipped        *JUnitSkipped        `xml:"skipped,omitempty"`
	Time           float64              `xml:"time,attr"`
	SystemOut      string               `xml:"system-out,omitempty"`
}

type JUnitFailureMessage struct {
	Type    string `xml:"type,attr"`
	Message string `xml:",chardata"`
}

type JUnitSkipped struct {
	Message string `xml:",chardata"`
}

type JUnitReporter struct {
	suite          JUnitTestSuite
	filename       string
	testSuiteName  string
	ReporterConfig config.DefaultReporterConfigType
}

//NewJUnitReporter creates a new JUnit XML reporter.  The XML will be stored in the passed in filename.
func NewJUnitReporter(filename string) *JUnitReporter {
	return &JUnitReporter{
		filename: filename,
	}
}

func (reporter *JUnitReporter) SpecSuiteWillBegin(ginkgoConfig config.GinkgoConfigType, summary types.SuiteSummary) {
	reporter.suite = JUnitTestSuite{
		Name:      summary.SuiteDescription,
		TestCases: []JUnitTestCase{},
	}
	reporter.testSuiteName = summary.SuiteDescription
	reporter.ReporterConfig = config.DefaultReporterConfig
}

func (reporter *JUnitReporter) WillRun(_ types.Summary) {
}

func (reporter *JUnitReporter) DidRun(summary types.Summary) {
	testCase := JUnitTestCase{
		ClassName: reporter.testSuiteName,
	}
	if summary.LeafNodeType.Is(types.NodeTypesForSuiteSetup...) {
		if summary.State.Is(types.SpecStatePassed) {
			return
		}
		testCase.Name = summary.LeafNodeType.String()
	} else {
		testCase.Name = strings.Join(summary.NodeTexts, " ")
	}
	if reporter.ReporterConfig.ReportPassed && summary.State == types.SpecStatePassed {
		testCase.SystemOut = summary.CombinedOutput()
	}
	if summary.State.Is(types.SpecStateFailureStates...) {
		testCase.FailureMessage = &JUnitFailureMessage{
			Type:    reporter.failureTypeForState(summary.State),
			Message: reporter.failureMessage(summary.Failure),
		}
		if summary.State.Is(types.SpecStatePanicked) {
			testCase.FailureMessage.Message += fmt.Sprintf("\n\nPanic: %s\n\nFull stack:\n%s",
				summary.Failure.ForwardedPanic,
				summary.Failure.Location.FullStackTrace)
		}
		testCase.SystemOut = summary.CombinedOutput()
	}
	if summary.State == types.SpecStateSkipped || summary.State == types.SpecStatePending {
		testCase.Skipped = &JUnitSkipped{}
		if summary.Failure.Message != "" {
			testCase.Skipped.Message = reporter.failureMessage(summary.Failure)
		}
	}
	testCase.Time = summary.RunTime.Seconds()
	reporter.suite.TestCases = append(reporter.suite.TestCases, testCase)
}

func (reporter *JUnitReporter) SpecSuiteDidEnd(summary types.SuiteSummary) {
	reporter.suite.Tests = summary.NumberOfSpecsThatWillBeRun
	reporter.suite.Time = math.Trunc(summary.RunTime.Seconds()*1000) / 1000
	reporter.suite.Failures = summary.NumberOfFailedSpecs
	reporter.suite.Errors = 0
	filePath, _ := filepath.Abs(reporter.filename)
	dirPath := filepath.Dir(filePath)
	err := os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		return
	}
	file, err := os.Create(filePath)
	if err != nil {
		return
	}
	defer file.Close()
	file.WriteString(xml.Header)
	encoder := xml.NewEncoder(file)
	encoder.Indent("  ", "    ")
	encoder.Encode(reporter.suite)
}

func (reporter *JUnitReporter) failureMessage(failure types.Failure) string {
	return fmt.Sprintf("%s\n%s\n%s", failure.NodeType.String(), failure.Message, failure.Location.String())
}

func (reporter *JUnitReporter) failureTypeForState(state types.SpecState) string {
	switch state {
	case types.SpecStateFailed:
		return "Failure"
	case types.SpecStatePanicked:
		return "Panicked"
	case types.SpecStateInterrupted:
		return "Interrupted"
	default:
		return ""
	}
}
