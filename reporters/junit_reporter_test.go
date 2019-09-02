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
		Ω(err).ShouldNot(HaveOccurred())
		var suite reporters.JUnitTestSuite
		err = xml.Unmarshal(bytes, &suite)
		Ω(err).ShouldNot(HaveOccurred())
		return suite
	}

	BeforeEach(func() {
		f, err := ioutil.TempFile("", "output")
		Ω(err).ShouldNot(HaveOccurred())
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
			Ω(output.Name).Should(Equal("My test suite"))
			Ω(output.Tests).Should(Equal(1))
			Ω(output.Failures).Should(Equal(0))
			Ω(output.Time).Should(Equal(reportedSuiteTime))
			Ω(output.Errors).Should(Equal(0))
			Ω(output.TestCases).Should(HaveLen(1))
			Ω(output.TestCases[0].Name).Should(Equal("A B C"))
			Ω(output.TestCases[0].ClassName).Should(Equal("My test suite"))
			Ω(output.TestCases[0].FailureMessage).Should(BeNil())
			Ω(output.TestCases[0].Skipped).Should(BeNil())
			Ω(output.TestCases[0].Time).Should(Equal(5.0))
			Ω(output.TestCases[0].PassedMessage.Message).Should(ContainSubstring("Test scenario"))
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
			Ω(err).ShouldNot(HaveOccurred())
			f, err := ioutil.TempFile(d, "output")
			Ω(err).ShouldNot(HaveOccurred())
			f.Close()
			outputFile = f.Name()
			err = os.RemoveAll(d)
			Ω(err).ShouldNot(HaveOccurred())
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
			Ω(output.Name).Should(Equal("My test suite"))
			Ω(output.Tests).Should(Equal(1))
			Ω(output.Failures).Should(Equal(0))
			Ω(output.Time).Should(Equal(reportedSuiteTime))
			Ω(output.Errors).Should(Equal(0))
			Ω(output.TestCases).Should(HaveLen(1))
			Ω(output.TestCases[0].Name).Should(Equal("A B C"))
			Ω(output.TestCases[0].ClassName).Should(Equal("My test suite"))
			Ω(output.TestCases[0].FailureMessage).Should(BeNil())
			Ω(output.TestCases[0].Skipped).Should(BeNil())
			Ω(output.TestCases[0].Time).Should(Equal(5.0))
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
			Ω(output.Name).Should(Equal("My test suite"))
			Ω(output.Tests).Should(Equal(1))
			Ω(output.Failures).Should(Equal(1))
			Ω(output.Time).Should(Equal(reportedSuiteTime))
			Ω(output.Errors).Should(Equal(0))
			Ω(output.TestCases[0].Name).Should(Equal("BeforeSuite"))
			Ω(output.TestCases[0].Time).Should(Equal(3.0))
			Ω(output.TestCases[0].ClassName).Should(Equal("My test suite"))
			Ω(output.TestCases[0].FailureMessage.Type).Should(Equal("Failure"))
			Ω(output.TestCases[0].FailureMessage.Message).Should(ContainSubstring("failed to setup"))
			Ω(output.TestCases[0].FailureMessage.Message).Should(ContainSubstring(beforeSuite.Failure.ComponentCodeLocation.String()))
			Ω(output.TestCases[0].FailureMessage.Message).Should(ContainSubstring(beforeSuite.Failure.Location.String()))
			Ω(output.TestCases[0].Skipped).Should(BeNil())
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
			Ω(output.Name).Should(Equal("My test suite"))
			Ω(output.Tests).Should(Equal(1))
			Ω(output.Failures).Should(Equal(1))
			Ω(output.Time).Should(Equal(reportedSuiteTime))
			Ω(output.Errors).Should(Equal(0))
			Ω(output.TestCases[0].Name).Should(Equal("AfterSuite"))
			Ω(output.TestCases[0].Time).Should(Equal(3.0))
			Ω(output.TestCases[0].ClassName).Should(Equal("My test suite"))
			Ω(output.TestCases[0].FailureMessage.Type).Should(Equal("Failure"))
			Ω(output.TestCases[0].FailureMessage.Message).Should(ContainSubstring("failed to setup"))
			Ω(output.TestCases[0].FailureMessage.Message).Should(ContainSubstring(afterSuite.Failure.ComponentCodeLocation.String()))
			Ω(output.TestCases[0].FailureMessage.Message).Should(ContainSubstring(afterSuite.Failure.Location.String()))
			Ω(output.TestCases[0].Skipped).Should(BeNil())
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
				Ω(output.Name).Should(Equal("My test suite"))
				Ω(output.Tests).Should(Equal(1))
				Ω(output.Failures).Should(Equal(1))
				Ω(output.Time).Should(Equal(reportedSuiteTime))
				Ω(output.Errors).Should(Equal(0))
				Ω(output.TestCases[0].Name).Should(Equal("A B C"))
				Ω(output.TestCases[0].ClassName).Should(Equal("My test suite"))
				Ω(output.TestCases[0].FailureMessage.Type).Should(Equal(specStateCase.message))
				Ω(output.TestCases[0].FailureMessage.Message).Should(ContainSubstring("I failed"))
				Ω(output.TestCases[0].FailureMessage.Message).Should(ContainSubstring(spec.Failure.ComponentCodeLocation.String()))
				Ω(output.TestCases[0].FailureMessage.Message).Should(ContainSubstring(spec.Failure.Location.String()))
				Ω(output.TestCases[0].Skipped).Should(BeNil())
				if specStateCase.state == types.SpecStatePanicked {
					Ω(output.TestCases[0].FailureMessage.Message).Should(ContainSubstring("\nPanic: " + specStateCase.forwardedPanic + "\n"))
					Ω(output.TestCases[0].FailureMessage.Message).Should(ContainSubstring("\nFull stack:\n" + spec.Failure.Location.FullStackTrace))
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
				Ω(output.Tests).Should(Equal(1))
				Ω(output.Failures).Should(Equal(0))
				Ω(output.Time).Should(Equal(reportedSuiteTime))
				Ω(output.Errors).Should(Equal(0))
				Ω(output.TestCases[0].Name).Should(Equal("A B C"))
				Ω(output.TestCases[0].Skipped).ShouldNot(BeNil())
			})
		})
	}
})
