package reporters_test

import (
	"encoding/xml"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/reporters"
	"github.com/onsi/ginkgo/types"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"time"
)

var _ = Describe("JUnit Reporter", func() {
	var (
		outputFile string
		reporter   Reporter
	)

	readOutputFile := func() reporters.JUnitTestSuite {
		bytes, err := ioutil.ReadFile(outputFile)
		Ω(err).ShouldNot(HaveOccurred())
		var suite reporters.JUnitTestSuite
		err = xml.Unmarshal(bytes, &suite)
		Ω(err).ShouldNot(HaveOccurred())
		return suite
	}

	BeforeEach(func() {
		outputFile = "/tmp/test.xml"
		reporter = reporters.NewJUnitReporter(outputFile)

		reporter.SpecSuiteWillBegin(config.GinkgoConfigType{}, &types.SuiteSummary{
			SuiteDescription:              "My test suite",
			NumberOfExamplesThatWillBeRun: 1,
		})
	})

	Describe("a passing test", func() {
		BeforeEach(func() {
			example := &types.ExampleSummary{
				ComponentTexts: []string{"[Top Level]", "A", "B", "C"},
				State:          types.ExampleStatePassed,
				RunTime:        5 * time.Second,
			}
			reporter.ExampleWillRun(example)
			reporter.ExampleDidComplete(example)

			reporter.SpecSuiteDidEnd(&types.SuiteSummary{
				NumberOfExamplesThatWillBeRun: 1,
				NumberOfFailedExamples:        0,
				RunTime:                       10 * time.Second,
			})
		})

		It("should record the test as passing", func() {
			output := readOutputFile()
			Ω(output.Tests).Should(Equal(1))
			Ω(output.Failures).Should(Equal(0))
			Ω(output.Time).Should(Equal(10.0))
			Ω(output.TestCases).Should(HaveLen(1))
			Ω(output.TestCases[0].Name).Should(Equal("A B C"))
			Ω(output.TestCases[0].ClassName).Should(Equal("My test suite"))
			Ω(output.TestCases[0].FailureMessage).Should(BeNil())
			Ω(output.TestCases[0].Skipped).Should(BeNil())
			Ω(output.TestCases[0].Time).Should(Equal(5.0))
		})
	})

	exampleStateCases := []struct {
		state   types.ExampleState
		message string
	}{
		{types.ExampleStateFailed, "Failure"},
		{types.ExampleStateTimedOut, "Timeout"},
		{types.ExampleStatePanicked, "Panic"},
	}

	for _, exampleStateCase := range exampleStateCases {
		exampleStateCase := exampleStateCase
		Describe("a failing test", func() {
			var example *types.ExampleSummary
			BeforeEach(func() {
				example = &types.ExampleSummary{
					ComponentTexts: []string{"[Top Level]", "A", "B", "C"},
					State:          exampleStateCase.state,
					RunTime:        5 * time.Second,
					Failure: types.ExampleFailure{
						ComponentCodeLocation: types.GenerateCodeLocation(0),
						Message:               "I failed",
					},
				}
				reporter.ExampleWillRun(example)
				reporter.ExampleDidComplete(example)

				reporter.SpecSuiteDidEnd(&types.SuiteSummary{
					NumberOfExamplesThatWillBeRun: 1,
					NumberOfFailedExamples:        1,
					RunTime:                       10 * time.Second,
				})
			})

			It("should record test as failing", func() {
				output := readOutputFile()
				Ω(output.Tests).Should(Equal(1))
				Ω(output.Failures).Should(Equal(1))
				Ω(output.Time).Should(Equal(10.0))
				Ω(output.TestCases[0].Name).Should(Equal("A B C"))
				Ω(output.TestCases[0].ClassName).Should(Equal("My test suite"))
				Ω(output.TestCases[0].FailureMessage.Type).Should(Equal(exampleStateCase.message))
				Ω(output.TestCases[0].FailureMessage.Message).Should(ContainSubstring("I failed"))
				Ω(output.TestCases[0].FailureMessage.Message).Should(ContainSubstring(example.Failure.ComponentCodeLocation.String()))
				Ω(output.TestCases[0].Skipped).Should(BeNil())
			})
		})
	}

	for _, exampleStateCase := range []types.ExampleState{types.ExampleStatePending, types.ExampleStateSkipped} {
		exampleStateCase := exampleStateCase
		Describe("a skipped test", func() {
			var example *types.ExampleSummary
			BeforeEach(func() {
				example = &types.ExampleSummary{
					ComponentTexts: []string{"[Top Level]", "A", "B", "C"},
					State:          exampleStateCase,
					RunTime:        5 * time.Second,
				}
				reporter.ExampleWillRun(example)
				reporter.ExampleDidComplete(example)

				reporter.SpecSuiteDidEnd(&types.SuiteSummary{
					NumberOfExamplesThatWillBeRun: 1,
					NumberOfFailedExamples:        0,
					RunTime:                       10 * time.Second,
				})
			})

			It("should record test as failing", func() {
				output := readOutputFile()
				Ω(output.Tests).Should(Equal(1))
				Ω(output.Failures).Should(Equal(0))
				Ω(output.Time).Should(Equal(10.0))
				Ω(output.TestCases[0].Name).Should(Equal("A B C"))
				Ω(output.TestCases[0].Skipped).ShouldNot(BeNil())
			})
		})
	}
})
