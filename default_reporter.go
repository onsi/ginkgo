package godescribe

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

func (*defaultReporter) SpecSuiteWillBegin(summary *SuiteSummary) {

}

func (*defaultReporter) ExampleDidComplete(exampleSummary *ExampleSummary, summary *SuiteSummary) {

}

func (*defaultReporter) SpecSuiteDidEnd(summary *SuiteSummary) {

}
