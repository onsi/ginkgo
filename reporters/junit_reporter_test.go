package reporters_test

import (
	"encoding/xml"
	"io/ioutil"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	"github.com/onsi/ginkgo/types"
	. "github.com/onsi/gomega"
)

var _ = Describe("JUnit Reporter", func() {
	var (
		outputFile string
		reporter   *reporters.JUnitReporter
	)
	testSuiteTime := 12456999 * time.Microsecond
	reportedSuiteTime := 12.456

	readOutputFile := func() reporters.JUnitTestSuite {
		bytes, err := ioutil.ReadFile(outputFile)
		Expect(err).ToNot(HaveOccurred())
		var suite reporters.JUnitTestSuite
		err = xml.Unmarshal(bytes, &suite)
		Expect(err).ToNot(HaveOccurred())
		return suite
	}

	BeforeEach(func() {
		f, err := ioutil.TempFile("", "output")
		Expect(err).ToNot(HaveOccurred())
		f.Close()
		outputFile = f.Name()

		reporter = reporters.NewJUnitReporter(outputFile)

		reporter.SuiteWillBegin(types.SuiteConfig{}, types.SuiteSummary{
			SuiteDescription:           "My test suite",
			NumberOfSpecsThatWillBeRun: 1,
		})
	})

	AfterEach(func() {
		os.RemoveAll(outputFile)
	})

	Describe("when configured with ReportPassed, and test has passed", func() {
		BeforeEach(func() {
			reporter.ReporterConfig.ReportPassed = true

			report := types.SpecReport{
				NodeTexts:                  []string{"A", "B", "C"},
				CapturedGinkgoWriterOutput: "Test scenario...",
				State:                      types.SpecStatePassed,
				RunTime:                    5 * time.Second,
			}
			reporter.WillRun(report)
			reporter.DidRun(report)

			reporter.SuiteDidEnd(types.SuiteSummary{
				NumberOfSpecsThatWillBeRun: 1,
				NumberOfFailedSpecs:        0,
				RunTime:                    testSuiteTime,
			})
		})

		It("should record the test as passing, including detailed output", func() {
			output := readOutputFile()
			Expect(output.Name).To(Equal("My test suite"))
			Expect(output.Tests).To(Equal(1))
			Expect(output.Failures).To(Equal(0))
			Expect(output.Time).To(Equal(reportedSuiteTime))
			Expect(output.Errors).To(Equal(0))
			Expect(output.TestCases).To(HaveLen(1))
			Expect(output.TestCases[0].Name).To(Equal("A B C"))
			Expect(output.TestCases[0].ClassName).To(Equal("My test suite"))
			Expect(output.TestCases[0].FailureMessage).To(BeNil())
			Expect(output.TestCases[0].Skipped).To(BeNil())
			Expect(output.TestCases[0].Time).To(Equal(5.0))
			Expect(output.TestCases[0].SystemOut).To(ContainSubstring("Test scenario"))
		})
	})

	Describe("when a BeforeSuite fails", func() {
		var beforeSuite types.SpecReport

		BeforeEach(func() {
			beforeSuite = types.SpecReport{
				LeafNodeType: types.NodeTypeBeforeSuite,
				State:        types.SpecStateFailed,
				RunTime:      3 * time.Second,
				Failure: types.Failure{
					Message:  "failed to setup",
					NodeType: types.NodeTypeJustBeforeEach,
					Location: types.NewCodeLocation(2),
				},
			}
			reporter.WillRun(beforeSuite)
			reporter.DidRun(beforeSuite)

			reporter.SuiteDidEnd(types.SuiteSummary{
				NumberOfSpecsThatWillBeRun: 1,
				NumberOfFailedSpecs:        1,
				RunTime:                    testSuiteTime,
			})
		})

		It("should record the test as having failed", func() {
			output := readOutputFile()
			Expect(output.Name).To(Equal("My test suite"))
			Expect(output.Tests).To(Equal(1))
			Expect(output.Failures).To(Equal(1))
			Expect(output.Time).To(Equal(reportedSuiteTime))
			Expect(output.Errors).To(Equal(0))
			Expect(output.TestCases[0].Name).To(Equal("BeforeSuite"))
			Expect(output.TestCases[0].Time).To(Equal(3.0))
			Expect(output.TestCases[0].ClassName).To(Equal("My test suite"))
			Expect(output.TestCases[0].FailureMessage.Type).To(Equal("Failure"))
			Expect(output.TestCases[0].FailureMessage.Message).To(ContainSubstring("failed to setup"))
			Expect(output.TestCases[0].FailureMessage.Message).To(ContainSubstring("JustBeforeEach"))
			Expect(output.TestCases[0].FailureMessage.Message).To(ContainSubstring(beforeSuite.Failure.Location.String()))
			Expect(output.TestCases[0].Skipped).To(BeNil())
		})
	})

	specStateCases := []struct {
		state   types.SpecState
		message string

		// Only for SpecStatePanicked.
		forwardedPanic string
	}{
		{types.SpecStateFailed, "Failure", ""},
		{types.SpecStatePanicked, "Panicked", "artifical panic"},
		{types.SpecStateInterrupted, "Interrupted", ""},
	}

	for _, specStateCase := range specStateCases {
		specStateCase := specStateCase
		Describe("a failing test", func() {
			var report types.SpecReport
			BeforeEach(func() {
				report = types.SpecReport{
					NodeTexts: []string{"A", "B", "C"},
					State:     specStateCase.state,
					RunTime:   5 * time.Second,
					Failure: types.Failure{
						NodeType:       types.NodeTypeJustBeforeEach,
						Location:       types.NewCodeLocation(2),
						Message:        "I failed",
						ForwardedPanic: specStateCase.forwardedPanic,
					},
				}
				reporter.WillRun(report)
				reporter.DidRun(report)

				reporter.SuiteDidEnd(types.SuiteSummary{
					NumberOfSpecsThatWillBeRun: 1,
					NumberOfFailedSpecs:        1,
					RunTime:                    testSuiteTime,
				})
			})

			It("should record test as failing", func() {
				output := readOutputFile()
				Expect(output.Name).To(Equal("My test suite"))
				Expect(output.Tests).To(Equal(1))
				Expect(output.Failures).To(Equal(1))
				Expect(output.Time).To(Equal(reportedSuiteTime))
				Expect(output.Errors).To(Equal(0))
				Expect(output.TestCases[0].Name).To(Equal("A B C"))
				Expect(output.TestCases[0].ClassName).To(Equal("My test suite"))
				Expect(output.TestCases[0].FailureMessage.Type).To(Equal(specStateCase.message))
				Expect(output.TestCases[0].FailureMessage.Message).To(ContainSubstring("I failed"))
				Expect(output.TestCases[0].FailureMessage.Message).To(ContainSubstring("JustBeforeEach"))
				Expect(output.TestCases[0].FailureMessage.Message).To(ContainSubstring(report.Failure.Location.String()))
				Expect(output.TestCases[0].Skipped).To(BeNil())
				if specStateCase.state == types.SpecStatePanicked {
					Expect(output.TestCases[0].FailureMessage.Message).To(ContainSubstring("\nPanic: " + specStateCase.forwardedPanic + "\n"))
					Expect(output.TestCases[0].FailureMessage.Message).To(ContainSubstring("\nFull stack:\n" + report.Failure.Location.FullStackTrace))
				}
			})
		})
	}

	for _, specStateCase := range []types.SpecState{types.SpecStatePending, types.SpecStateSkipped} {
		specStateCase := specStateCase
		Describe("a skipped test", func() {
			var report types.SpecReport
			BeforeEach(func() {
				report = types.SpecReport{
					NodeTexts: []string{"A", "B", "C"},
					State:     specStateCase,
					RunTime:   5 * time.Second,
					Failure: types.Failure{
						Message: "skipped reason",
					},
				}
				reporter.WillRun(report)
				reporter.DidRun(report)

				reporter.SuiteDidEnd(types.SuiteSummary{
					NumberOfSpecsThatWillBeRun: 1,
					NumberOfFailedSpecs:        0,
					RunTime:                    testSuiteTime,
				})
			})

			It("should record test as failing", func() {
				output := readOutputFile()
				Expect(output.Tests).To(Equal(1))
				Expect(output.Failures).To(Equal(0))
				Expect(output.Time).To(Equal(reportedSuiteTime))
				Expect(output.Errors).To(Equal(0))
				Expect(output.TestCases[0].Name).To(Equal("A B C"))
				Expect(output.TestCases[0].Skipped.Message).To(ContainSubstring("skipped reason"))
			})
		})
	}
})
