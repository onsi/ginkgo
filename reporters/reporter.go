package reporters

import (
	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/types"
)

type Reporter interface {
	SpecSuiteWillBegin(config config.GinkgoConfigType, summary *types.SuiteSummary)
	ExampleWillRun(exampleSummary *types.ExampleSummary)
	ExampleDidComplete(exampleSummary *types.ExampleSummary)
	SpecSuiteDidEnd(summary *types.SuiteSummary)
}
