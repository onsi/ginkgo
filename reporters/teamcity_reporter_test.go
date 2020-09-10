package reporters_test

import (
	"bytes"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/reporters"
	"github.com/onsi/ginkgo/types"
	. "github.com/onsi/gomega"
)

var _ = Describe("TeamCity Reporter", func() {
	var (
		buffer   bytes.Buffer
		reporter *reporters.TeamCityReporter
	)

	BeforeEach(func() {
		buffer.Truncate(0)
		reporter = reporters.NewTeamCityReporter(&buffer)
		reporter.SpecSuiteWillBegin(config.GinkgoConfigType{}, types.SuiteSummary{
			SuiteDescription:           "Foo's test suite",
			NumberOfSpecsThatWillBeRun: 1,
		})
	})

	Describe("a passing test", func() {
		BeforeEach(func() {
			// Set the ReportPassed config flag, in order to show captured output when tests have passed.
			reporter.ReporterConfig.ReportPassed = true

			spec := types.Summary{
				NodeTexts:                  []string{"A", "B", "C"},
				CapturedGinkgoWriterOutput: "Test scenario...",
				State:                      types.SpecStatePassed,
				RunTime:                    5 * time.Second,
			}
			reporter.WillRun(spec)
			reporter.DidRun(spec)

			reporter.SpecSuiteDidEnd(types.SuiteSummary{
				NumberOfSpecsThatWillBeRun: 1,
				NumberOfFailedSpecs:        0,
				RunTime:                    10 * time.Second,
			})
		})

		It("should record the test as passing", func() {
			actual := buffer.String()
			expected :=
				"##teamcity[testSuiteStarted name='Foo|'s test suite']\n" +
					"##teamcity[testStarted name='A B C']\n" +
					"##teamcity[testPassed name='A B C' details='Test scenario...']\n" +
					"##teamcity[testFinished name='A B C' duration='5000']\n" +
					"##teamcity[testSuiteFinished name='Foo|'s test suite']\n"
			Ω(actual).Should(Equal(expected))
		})
	})

	Describe("when the BeforeSuite fails", func() {
		var beforeSuite types.Summary

		BeforeEach(func() {
			beforeSuite = types.Summary{
				LeafNodeType: types.NodeTypeBeforeSuite,
				State:        types.SpecStateFailed,
				RunTime:      3 * time.Second,
				Failure: types.Failure{
					Message:  "failed to setup\n",
					NodeType: types.NodeTypeBeforeSuite,
					Location: types.NewCodeLocation(2),
				},
			}
			reporter.WillRun(beforeSuite)
			reporter.DidRun(beforeSuite)

			reporter.SpecSuiteDidEnd(types.SuiteSummary{
				NumberOfSpecsThatWillBeRun: 1,
				NumberOfFailedSpecs:        1,
				RunTime:                    10 * time.Second,
			})
		})

		It("should record the test as having failed", func() {
			actual := buffer.String()
			expected := fmt.Sprintf(
				"##teamcity[testSuiteStarted name='Foo|'s test suite']\n"+
					"##teamcity[testStarted name='BeforeSuite']\n"+
					"##teamcity[testFailed name='BeforeSuite' message='BeforeSuite' details='failed to setup|n|n%s']\n"+
					"##teamcity[testFinished name='BeforeSuite' duration='3000']\n"+
					"##teamcity[testSuiteFinished name='Foo|'s test suite']\n",
				beforeSuite.Failure.Location.String(),
			)
			Ω(actual).Should(Equal(expected))
		})
	})

	specStateCases := []struct {
		state   types.SpecState
		message string
	}{
		{types.SpecStateFailed, "Failure"},
		{types.SpecStatePanicked, "Panicked"},
		{types.SpecStateInterrupted, "interrupted"},
	}

	for _, specStateCase := range specStateCases {
		specStateCase := specStateCase
		Describe("a failing test", func() {
			var spec types.Summary
			BeforeEach(func() {
				spec = types.Summary{
					NodeTexts: []string{"A", "B", "C"},
					State:     specStateCase.state,
					RunTime:   5 * time.Second,
					Failure: types.Failure{
						NodeType: types.NodeTypeJustBeforeEach,
						Location: types.NewCodeLocation(2),
						Message:  "I failed",
					},
				}
				reporter.WillRun(spec)
				reporter.DidRun(spec)

				reporter.SpecSuiteDidEnd(types.SuiteSummary{
					NumberOfSpecsThatWillBeRun: 1,
					NumberOfFailedSpecs:        1,
					RunTime:                    10 * time.Second,
				})
			})

			It("should record test as failing", func() {
				actual := buffer.String()
				expected :=
					fmt.Sprintf("##teamcity[testSuiteStarted name='Foo|'s test suite']\n"+
						"##teamcity[testStarted name='A B C']\n"+
						"##teamcity[testFailed name='A B C' message='JustBeforeEach' details='I failed|n%s']\n"+
						"##teamcity[testFinished name='A B C' duration='5000']\n"+
						"##teamcity[testSuiteFinished name='Foo|'s test suite']\n",
						spec.Failure.Location.String(),
					)
				Ω(actual).Should(Equal(expected))
			})
		})
	}

	for _, specStateCase := range []types.SpecState{types.SpecStatePending, types.SpecStateSkipped} {
		specStateCase := specStateCase
		Describe("a skipped test", func() {
			var spec types.Summary
			BeforeEach(func() {
				spec = types.Summary{
					NodeTexts: []string{"A", "B", "C"},
					State:     specStateCase,
					RunTime:   5 * time.Second,
				}
				reporter.WillRun(spec)
				reporter.DidRun(spec)

				reporter.SpecSuiteDidEnd(types.SuiteSummary{
					NumberOfSpecsThatWillBeRun: 1,
					NumberOfFailedSpecs:        0,
					RunTime:                    10 * time.Second,
				})
			})

			It("should record test as ignored", func() {
				actual := buffer.String()
				expected :=
					"##teamcity[testSuiteStarted name='Foo|'s test suite']\n" +
						"##teamcity[testStarted name='A B C']\n" +
						"##teamcity[testIgnored name='A B C']\n" +
						"##teamcity[testFinished name='A B C' duration='5000']\n" +
						"##teamcity[testSuiteFinished name='Foo|'s test suite']\n"
				Ω(actual).Should(Equal(expected))
			})
		})
	}
})
