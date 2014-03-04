package reporters_test

import (
	"bytes"
	"fmt"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/reporters"
	"github.com/onsi/ginkgo/types"
	. "github.com/onsi/gomega"
	"time"
)

var _ = Describe("TeamCity Reporter", func() {
	var (
		buffer   bytes.Buffer
		reporter Reporter
	)

	BeforeEach(func() {
		buffer.Truncate(0)
		reporter = reporters.NewTeamCityReporter(&buffer)
		reporter.SpecSuiteWillBegin(config.GinkgoConfigType{}, &types.SuiteSummary{
			SuiteDescription:              "Foo's test suite",
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
			actual := buffer.String()
			expected :=
				"##teamcity[testSuiteStarted name='Foo|'s test suite']" +
					"##teamcity[testStarted name='A B C']" +
					"##teamcity[testFinished name='A B C' duration='5000']" +
					"##teamcity[testSuiteFinished name='Foo|'s test suite']"
			Ω(actual).Should(Equal(expected))
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
				actual := buffer.String()
				expected :=
					fmt.Sprintf("##teamcity[testSuiteStarted name='Foo|'s test suite']"+
						"##teamcity[testStarted name='A B C']"+
						"##teamcity[testFailed name='A B C' message='%s' details='I failed']"+
						"##teamcity[testFinished name='A B C' duration='5000']"+
						"##teamcity[testSuiteFinished name='Foo|'s test suite']", example.Failure.ComponentCodeLocation.String())
				Ω(actual).Should(Equal(expected))
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

			It("should record test as ignored", func() {
				actual := buffer.String()
				expected :=
					"##teamcity[testSuiteStarted name='Foo|'s test suite']" +
						"##teamcity[testStarted name='A B C']" +
						"##teamcity[testIgnored name='A B C']" +
						"##teamcity[testFinished name='A B C' duration='5000']" +
						"##teamcity[testSuiteFinished name='Foo|'s test suite']"
				Ω(actual).Should(Equal(expected))
			})
		})
	}
})
