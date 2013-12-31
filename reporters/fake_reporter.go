package reporters

import (
	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/types"
)

//FakeReporter is useful for testing purposes
type FakeReporter struct {
	Config config.GinkgoConfigType

	BeginSummary            *types.SuiteSummary
	ExampleWillRunSummaries []*types.ExampleSummary
	ExampleSummaries        []*types.ExampleSummary
	EndSummary              *types.SuiteSummary
}

func NewFakeReporter() *FakeReporter {
	return &FakeReporter{
		ExampleWillRunSummaries: make([]*types.ExampleSummary, 0),
		ExampleSummaries:        make([]*types.ExampleSummary, 0),
	}
}

func (fakeR *FakeReporter) SpecSuiteWillBegin(config config.GinkgoConfigType, summary *types.SuiteSummary) {
	fakeR.Config = config
	fakeR.BeginSummary = summary
}

func (fakeR *FakeReporter) ExampleWillRun(exampleSummary *types.ExampleSummary) {
	fakeR.ExampleWillRunSummaries = append(fakeR.ExampleWillRunSummaries, exampleSummary)
}

func (fakeR *FakeReporter) ExampleDidComplete(exampleSummary *types.ExampleSummary) {
	fakeR.ExampleSummaries = append(fakeR.ExampleSummaries, exampleSummary)
}

func (fakeR *FakeReporter) SpecSuiteDidEnd(summary *types.SuiteSummary) {
	fakeR.EndSummary = summary
}
