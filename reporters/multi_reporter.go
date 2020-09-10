package reporters

import (
	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/types"
)

type MultiReporter struct {
	reporters []Reporter
}

func NewMultiReporter(reporters ...Reporter) *MultiReporter {
	return &MultiReporter{
		reporters: reporters,
	}
}

func (mr *MultiReporter) SpecSuiteWillBegin(config config.GinkgoConfigType, summary types.SuiteSummary) {
	for _, reporter := range mr.reporters {
		reporter.SpecSuiteWillBegin(config, summary)
	}
}

func (mr *MultiReporter) WillRun(summary types.Summary) {
	for _, reporter := range mr.reporters {
		reporter.WillRun(summary)
	}
}

func (mr *MultiReporter) DidRun(summary types.Summary) {
	for i := len(mr.reporters) - 1; i >= 0; i-- {
		mr.reporters[i].DidRun(summary)
	}
}

func (mr *MultiReporter) SpecSuiteDidEnd(summary types.SuiteSummary) {
	for _, reporter := range mr.reporters {
		reporter.SpecSuiteDidEnd(summary)
	}
}
