package reporters_test

import (
	"encoding/json"
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

var _ = Describe("Json Reporter", func() {
	var (
		outputFile string
		reporter   Reporter
	)

	readOutputFile := func() reporters.JsonTestSuite {
		bytes, err := ioutil.ReadFile(outputFile)
		Ω(err).ShouldNot(HaveOccurred())
		var suite reporters.JsonTestSuite
		err = json.Unmarshal(bytes, &suite)
		Ω(err).ShouldNot(HaveOccurred())
		return suite
	}

	BeforeEach(func() {
		f, err := ioutil.TempFile("", "output")
		Ω(err).ShouldNot(HaveOccurred())
		f.Close()
		outputFile = f.Name()

		reporter = reporters.NewJsonReporter(outputFile)

		reporter.SpecSuiteWillBegin(config.GinkgoConfigType{}, &types.SuiteSummary{
			SuiteDescription:           "My test suite",
			NumberOfSpecsThatWillBeRun: 1,
		})
	})

	AfterEach(func() {
		os.RemoveAll(outputFile)
	})

	Describe("a passing test", func() {
		BeforeEach(func() {
			beforeSuite := &types.SetupSummary{
				State: types.SpecStatePassed,
			}
			reporter.BeforeSuiteDidRun(beforeSuite)

			afterSuite := &types.SetupSummary{
				State: types.SpecStatePassed,
			}
			reporter.AfterSuiteDidRun(afterSuite)

			spec := &types.SpecSummary{
				ComponentTexts: []string{"[Top Level]", "A", "B", "C"},
				State:          types.SpecStatePassed,
				RunTime:        5 * time.Second,
			}
			reporter.SpecWillRun(spec)
			reporter.SpecDidComplete(spec)

			reporter.SpecSuiteDidEnd(&types.SuiteSummary{
				NumberOfSpecsThatWillBeRun: 1,
				NumberOfTotalSpecs:         1,
				NumberOfFailedSpecs:        0,
				RunTime:                    10 * time.Second,
			})
		})

		It("should record the test as passing", func() {
			output := readOutputFile()
			Ω(output.Tests).Should(Equal(1))
			Ω(output.Failures.Number).Should(Equal(0))
			Ω(output.Failures.Names).Should(BeEmpty())
			Ω(output.Time).Should(Equal(10.0))
			Ω(output.TestCases).Should(HaveLen(1))
			Ω(output.TestCases[0].TestName).Should(Equal("A B C"))
			Ω(output.TestCases[0].SuiteName).Should(Equal("My test suite"))
			Ω(output.TestCases[0].Passed).Should(BeTrue())
			Ω(output.TestCases[0].Failed).Should(BeNil())
			Ω(output.TestCases[0].Skipped).Should(BeFalse())
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
				},
			}
			reporter.BeforeSuiteDidRun(beforeSuite)

			reporter.SpecSuiteDidEnd(&types.SuiteSummary{
				NumberOfSpecsThatWillBeRun: 1,
				NumberOfTotalSpecs:         1,
				NumberOfFailedSpecs:        1,
				RunTime:                    10 * time.Second,
			})
		})

		It("should record the test as having failed", func() {
			output := readOutputFile()
			Ω(output.Tests).Should(Equal(1))
			Ω(output.Failures.Number).Should(Equal(1))
			Ω(output.Failures.Names[0]).Should(Equal("BeforeSuite"))
			Ω(output.Time).Should(Equal(10.0))
			Ω(output.TestCases[0].TestName).Should(Equal("BeforeSuite"))
			Ω(output.TestCases[0].Time).Should(Equal(3.0))
			Ω(output.TestCases[0].SuiteName).Should(Equal("My test suite"))
			Ω(output.TestCases[0].Failed.Type).Should(Equal("Failure"))
			Ω(output.TestCases[0].Failed.Message).Should(ContainSubstring("failed to setup"))
			Ω(output.TestCases[0].Failed.Where).Should(ContainSubstring(beforeSuite.Failure.ComponentCodeLocation.String()))
			Ω(output.TestCases[0].Skipped).Should(BeFalse())
			Ω(output.TestCases[0].Passed).Should(BeFalse())
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
				},
			}
			reporter.AfterSuiteDidRun(afterSuite)

			reporter.SpecSuiteDidEnd(&types.SuiteSummary{
				NumberOfSpecsThatWillBeRun: 1,
				NumberOfTotalSpecs:         1,
				NumberOfFailedSpecs:        1,
				RunTime:                    10 * time.Second,
			})
		})

		It("should record the test as having failed", func() {
			output := readOutputFile()
			Ω(output.Tests).Should(Equal(1))
			Ω(output.Failures.Number).Should(Equal(1))
			Ω(output.Failures.Names[0]).Should(Equal("AfterSuite"))
			Ω(output.Time).Should(Equal(10.0))
			Ω(output.TestCases[0].TestName).Should(Equal("AfterSuite"))
			Ω(output.TestCases[0].Time).Should(Equal(3.0))
			Ω(output.TestCases[0].SuiteName).Should(Equal("My test suite"))
			Ω(output.TestCases[0].Failed.Type).Should(Equal("Failure"))
			Ω(output.TestCases[0].Failed.Message).Should(ContainSubstring("failed to setup"))
			Ω(output.TestCases[0].Failed.Where).Should(ContainSubstring(afterSuite.Failure.ComponentCodeLocation.String()))
			Ω(output.TestCases[0].Skipped).Should(BeFalse())
			Ω(output.TestCases[0].Passed).Should(BeFalse())
		})
	})

	specStateCases := []struct {
		state   types.SpecState
		message string
	}{
		{types.SpecStateFailed, "Failure"},
		{types.SpecStateTimedOut, "Timeout"},
		{types.SpecStatePanicked, "Panic"},
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
						Message:               "I failed",
					},
				}
				reporter.SpecWillRun(spec)
				reporter.SpecDidComplete(spec)

				reporter.SpecSuiteDidEnd(&types.SuiteSummary{
					NumberOfSpecsThatWillBeRun: 1,
					NumberOfTotalSpecs:         1,
					NumberOfFailedSpecs:        1,
					RunTime:                    10 * time.Second,
				})
			})

			It("should record test as failing", func() {
				output := readOutputFile()
				Ω(output.Tests).Should(Equal(1))
				Ω(output.Failures.Number).Should(Equal(1))
				Ω(output.Failures.Names[0]).Should(Equal("A B C"))
				Ω(output.Time).Should(Equal(10.0))
				Ω(output.TestCases[0].TestName).Should(Equal("A B C"))
				Ω(output.TestCases[0].SuiteName).Should(Equal("My test suite"))
				Ω(output.TestCases[0].Failed.Type).Should(Equal(specStateCase.message))
				Ω(output.TestCases[0].Failed.Message).Should(ContainSubstring("I failed"))
				Ω(output.TestCases[0].Failed.Where).Should(ContainSubstring(spec.Failure.ComponentCodeLocation.String()))
				Ω(output.TestCases[0].Skipped).Should(BeFalse())
				Ω(output.TestCases[0].Passed).Should(BeFalse())
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
					NumberOfTotalSpecs:         1,
					NumberOfFailedSpecs:        0,
					RunTime:                    10 * time.Second,
				})
			})

			It("should record test as failing", func() {
				output := readOutputFile()
				Ω(output.Tests).Should(Equal(1))
				Ω(output.Failures.Number).Should(Equal(0))
				Ω(output.Failures.Names).Should(BeEmpty())
				Ω(output.Time).Should(Equal(10.0))
				Ω(output.TestCases[0].TestName).Should(Equal("A B C"))
				Ω(output.TestCases[0].Skipped).Should(BeTrue())
				Ω(output.TestCases[0].Passed).Should(BeFalse())
			})
		})
	}
})
