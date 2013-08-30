package ginkgo

import (
	. "github.com/onsi/gomega"
	"testing"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Ginkgo")
}

//Fakes

type fakeTestingT struct {
	didFail bool
}

func (fakeT *fakeTestingT) Fail() {
	fakeT.didFail = true
}

type fakeReporter struct {
	randomSeed           int64
	randomizeAllExamples bool

	beginSummary     *SuiteSummary
	exampleSummaries []*ExampleSummary
	endSummary       *SuiteSummary
}

func (fakeR *fakeReporter) RandomizationStrategy(randomSeed int64, randomizeAllExamples bool) {
	fakeR.randomSeed = randomSeed
	fakeR.randomizeAllExamples = randomizeAllExamples
}

func (fakeR *fakeReporter) SpecSuiteWillBegin(summary *SuiteSummary) {
	fakeR.beginSummary = summary
}

func (fakeR *fakeReporter) ExampleDidComplete(exampleSummary *ExampleSummary) {
	fakeR.exampleSummaries = append(fakeR.exampleSummaries, exampleSummary)
}

func (fakeR *fakeReporter) SpecSuiteDidEnd(summary *SuiteSummary) {
	fakeR.endSummary = summary
}
