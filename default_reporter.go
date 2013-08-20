package godescribe

import (
	"fmt"
)

type defaultReporter struct {
	noColor           bool
	slowSpecThreshold float64
}

func newDefaultReporter(noColor bool, slowSpecThreshold float64) *defaultReporter {
	return &defaultReporter{
		noColor:           noColor,
		slowSpecThreshold: slowSpecThreshold,
	}
}

func (reporter *defaultReporter) RandomizationStrategy(randomSeed int64, randomizeAllExamples bool) {
	fmt.Println(randomSeed, randomizeAllExamples)
}

func (reporter *defaultReporter) SpecSuiteWillBegin(summary *SuiteSummary) {
	fmt.Printf("%#v\n", summary)
}

func (reporter *defaultReporter) ExampleDidComplete(exampleSummary *ExampleSummary) {
	fmt.Printf("%#v\n", exampleSummary)
}

func (reporter *defaultReporter) SpecSuiteDidEnd(summary *SuiteSummary) {
	fmt.Printf("%#v\n", summary)
}
