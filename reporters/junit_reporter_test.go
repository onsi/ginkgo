package reporters_test

import (
	"encoding/xml"
	"io/ioutil"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/internal/codelocation"
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

		reporter.SpecSuiteWillBegin(config.GinkgoConfigType{}, &types.SuiteSummary{
			SuiteDescription:           "My test suite",
			NumberOfSpecsThatWillBeRun: 1,
		})
	})

	AfterEach(func() {
		os.RemoveAll(outputFile)
	})

	Describe("when configured with ReportPassed, and test has passed", func() {
		BeforeEach(func() {
			beforeSuite := &types.SetupSummary{
				State: types.SpecStatePassed,
			}
			reporter.BeforeSuiteDidRun(beforeSuite)

			afterSuite := &types.SetupSummary{
				State: types.SpecStatePassed,
			}
			reporter.AfterSuiteDidRun(afterSuite)

			// Set the ReportPassed config flag, in order to show captured output when tests have passed.
			reporter.ReporterConfig.ReportPassed = true

			spec := &types.SpecSummary{
				ComponentTexts: []string{"[Top Level]", "A", "B", "C"},
				CapturedOutput: "Test scenario...",
				State:          types.SpecStatePassed,
				RunTime:        5 * time.Second,
			}
			reporter.SpecWillRun(spec)
			reporter.SpecDidComplete(spec)

			reporter.SpecSuiteDidEnd(&types.SuiteSummary{
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
			Expect(output.TestCases[0].PassedMessage.Message).To(ContainSubstring("Test scenario"))
		})
	})

	Describe("when configured with ReportFile <file path>", func() {
		BeforeEach(func() {
			beforeSuite := &types.SetupSummary{
				State: types.SpecStatePassed,
			}
			reporter.BeforeSuiteDidRun(beforeSuite)

			afterSuite := &types.SetupSummary{
				State: types.SpecStatePassed,
			}
			reporter.AfterSuiteDidRun(afterSuite)

			reporter.SpecSuiteWillBegin(config.GinkgoConfigType{}, &types.SuiteSummary{
				SuiteDescription:           "My test suite",
				NumberOfSpecsThatWillBeRun: 1,
			})

			// Set the ReportFile config flag with a new directory and new file path to be created.
			d, err := ioutil.TempDir("", "new-junit-dir")
			Expect(err).ToNot(HaveOccurred())
			f, err := ioutil.TempFile(d, "output")
			Expect(err).ToNot(HaveOccurred())
			f.Close()
			outputFile = f.Name()
			err = os.RemoveAll(d)
			Expect(err).ToNot(HaveOccurred())
			reporter.ReporterConfig.ReportFile = outputFile

			spec := &types.SpecSummary{
				ComponentTexts: []string{"[Top Level]", "A", "B", "C"},
				CapturedOutput: "Test scenario...",
				State:          types.SpecStatePassed,
				RunTime:        5 * time.Second,
			}
			reporter.SpecWillRun(spec)
			reporter.SpecDidComplete(spec)

			reporter.SpecSuiteDidEnd(&types.SuiteSummary{
				NumberOfSpecsThatWillBeRun: 1,
				NumberOfFailedSpecs:        0,
				RunTime:                    testSuiteTime,
			})

		})

		It("should create the report (and parent directories) as specified by ReportFile path", func() {
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
		})
	})

	Describe("when the BeforeSuite fails", func() {
		var beforeSuite *types.SetupSummary

		BeforeEach(func() {
			beforeSuite = &types.SetupSummary{
				State:   types.SpecStateFailed,
				RunTime: 3 * time.Second,
				Failure: types.SpecFailure{
					Message:               "failed to setup",
					ComponentCodeLocation: codelocation.New(0),
					Location:              codelocation.New(2),
				},
			}
			reporter.BeforeSuiteDidRun(beforeSuite)

			reporter.SpecSuiteDidEnd(&types.SuiteSummary{
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
			Expect(output.TestCases[0].FailureMessage.Message).To(ContainSubstring(beforeSuite.Failure.ComponentCodeLocation.String()))
			Expect(output.TestCases[0].FailureMessage.Message).To(ContainSubstring(beforeSuite.Failure.Location.String()))
			Expect(output.TestCases[0].Skipped).To(BeNil())
		})
	})

	Describe("when the AfterSuite fails", func() {
		var afterSuite *types.SetupSummary

		BeforeEach(func() {
			afterSuite = &types.SetupSummary{
				State:   types.SpecStateFailed,
				RunTime: 3 * time.Second,
				Failure: types.SpecFailure{
					Message:               "failed to setup",
					ComponentCodeLocation: codelocation.New(0),
					Location:              codelocation.New(2),
				},
			}
			reporter.AfterSuiteDidRun(afterSuite)

			reporter.SpecSuiteDidEnd(&types.SuiteSummary{
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
			Expect(output.TestCases[0].Name).To(Equal("AfterSuite"))
			Expect(output.TestCases[0].Time).To(Equal(3.0))
			Expect(output.TestCases[0].ClassName).To(Equal("My test suite"))
			Expect(output.TestCases[0].FailureMessage.Type).To(Equal("Failure"))
			Expect(output.TestCases[0].FailureMessage.Message).To(ContainSubstring("failed to setup"))
			Expect(output.TestCases[0].FailureMessage.Message).To(ContainSubstring(afterSuite.Failure.ComponentCodeLocation.String()))
			Expect(output.TestCases[0].FailureMessage.Message).To(ContainSubstring(afterSuite.Failure.Location.String()))
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
		{types.SpecStateTimedOut, "Timeout", ""},
		{types.SpecStatePanicked, "Panic", "artifical panic"},
	}

	for _, specStateCase := range specStateCases {
		specStateCase := specStateCase
		Describe("a failing test", func() {
			var spec *types.SpecSummary
			BeforeEach(func() {
				spec = &types.SpecSummary{
					ComponentTexts: []string{"[Top Level]", "A", "B", "C"},
					State:          specStateCase.state,
					RunTime:        5 * time.Second,
					Failure: types.SpecFailure{
						ComponentCodeLocation: codelocation.New(0),
						Location:              codelocation.New(2),
						Message:               "I failed",
						ForwardedPanic:        specStateCase.forwardedPanic,
					},
				}
				reporter.SpecWillRun(spec)
				reporter.SpecDidComplete(spec)

				reporter.SpecSuiteDidEnd(&types.SuiteSummary{
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
				Expect(output.TestCases[0].FailureMessage.Message).To(ContainSubstring(spec.Failure.ComponentCodeLocation.String()))
				Expect(output.TestCases[0].FailureMessage.Message).To(ContainSubstring(spec.Failure.Location.String()))
				Expect(output.TestCases[0].Skipped).To(BeNil())
				if specStateCase.state == types.SpecStatePanicked {
					Expect(output.TestCases[0].FailureMessage.Message).To(ContainSubstring("\nPanic: " + specStateCase.forwardedPanic + "\n"))
					Expect(output.TestCases[0].FailureMessage.Message).To(ContainSubstring("\nFull stack:\n" + spec.Failure.Location.FullStackTrace))
				}
			})
		})
	}

	for _, specStateCase := range []types.SpecState{types.SpecStatePending, types.SpecStateSkipped} {
		specStateCase := specStateCase
		Describe("a skipped test", func() {
			var spec *types.SpecSummary
			BeforeEach(func() {
				spec = &types.SpecSummary{
					ComponentTexts: []string{"[Top Level]", "A", "B", "C"},
					State:          specStateCase,
					RunTime:        5 * time.Second,
					Failure: types.SpecFailure{
						Message: "skipped reason",
					},
				}
				reporter.SpecWillRun(spec)
				reporter.SpecDidComplete(spec)

				reporter.SpecSuiteDidEnd(&types.SuiteSummary{
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
