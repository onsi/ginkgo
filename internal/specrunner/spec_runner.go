package specrunner

import (
	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/internal/leafnodes"
	"github.com/onsi/ginkgo/internal/spec"
	Writer "github.com/onsi/ginkgo/internal/writer"
	"github.com/onsi/ginkgo/reporters"
	"github.com/onsi/ginkgo/types"

	"time"
)

type SpecRunner struct {
	description     string
	beforeSuiteNode *leafnodes.SuiteNode
	specs           *spec.Specs
	reporters       []reporters.Reporter
	startTime       time.Time
	suiteID         string
	runningSpec     *spec.Spec
	writer          Writer.WriterInterface
	config          config.GinkgoConfigType
}

func New(description string, beforeSuiteNode *leafnodes.SuiteNode, specs *spec.Specs, reporters []reporters.Reporter, writer Writer.WriterInterface, config config.GinkgoConfigType) *SpecRunner {
	return &SpecRunner{
		description:     description,
		beforeSuiteNode: beforeSuiteNode,
		specs:           specs,
		reporters:       reporters,
		writer:          writer,
		config:          config,
		suiteID:         randomID(),
	}
}

func (runner *SpecRunner) Run() bool {
	runner.reportSuiteWillBegin()

	suitePassed := runner.runBeforeSuite()
	if suitePassed {
		suitePassed = runner.runSpecs()
	}

	runner.reportSuiteDidEnd()

	return suitePassed
}

func (runner *SpecRunner) runBeforeSuite() bool {
	if runner.beforeSuiteNode == nil {
		return true
	}

	passed := runner.beforeSuiteNode.Run()
	runner.reportBeforeSuite(runner.beforeSuiteNode.Summary())
	return passed
}

func (runner *SpecRunner) runSpecs() bool {
	suiteFailed := false
	for _, spec := range runner.specs.Specs() {
		runner.writer.Truncate()

		runner.reportSpecWillRun(spec)

		if !spec.Skipped() && !spec.Pending() {
			runner.runningSpec = spec
			spec.Run()
			runner.runningSpec = nil
			if spec.Failed() {
				suiteFailed = true
				runner.writer.DumpOut()
			}
		} else if spec.Pending() && runner.config.FailOnPending {
			suiteFailed = true
		}

		runner.reportSpecDidComplete(spec)
	}

	return !suiteFailed
}

func (runner *SpecRunner) CurrentSpecSummary() (*types.SpecSummary, bool) {
	if runner.runningSpec == nil {
		return nil, false
	}

	return runner.runningSpec.Summary(runner.suiteID), true
}

func (runner *SpecRunner) reportSuiteWillBegin() {
	runner.startTime = time.Now()
	summary := runner.summary()
	for _, reporter := range runner.reporters {
		reporter.SpecSuiteWillBegin(runner.config, summary)
	}
}

func (runner *SpecRunner) reportBeforeSuite(summary *types.SetupSummary) {
	for _, reporter := range runner.reporters {
		reporter.BeforeSuiteDidRun(summary)
	}
}

func (runner *SpecRunner) reportSpecWillRun(spec *spec.Spec) {
	summary := spec.Summary(runner.suiteID)
	for _, reporter := range runner.reporters {
		reporter.SpecWillRun(summary)
	}
}

func (runner *SpecRunner) reportSpecDidComplete(spec *spec.Spec) {
	summary := spec.Summary(runner.suiteID)
	for _, reporter := range runner.reporters {
		reporter.SpecDidComplete(summary)
	}
}

func (runner *SpecRunner) reportSuiteDidEnd() {
	summary := runner.summary()
	summary.RunTime = time.Since(runner.startTime)
	for _, reporter := range runner.reporters {
		reporter.SpecSuiteDidEnd(summary)
	}
}

func (runner *SpecRunner) countSpecsSatisfying(filter func(ex *spec.Spec) bool) (count int) {
	count = 0

	for _, spec := range runner.specs.Specs() {
		if filter(spec) {
			count++
		}
	}

	return count
}

func (runner *SpecRunner) summary() *types.SuiteSummary {
	numberOfSpecsThatWillBeRun := runner.countSpecsSatisfying(func(ex *spec.Spec) bool {
		return !ex.Skipped() && !ex.Pending()
	})

	numberOfPendingSpecs := runner.countSpecsSatisfying(func(ex *spec.Spec) bool {
		return ex.Pending()
	})

	numberOfSkippedSpecs := runner.countSpecsSatisfying(func(ex *spec.Spec) bool {
		return ex.Skipped()
	})

	numberOfPassedSpecs := runner.countSpecsSatisfying(func(ex *spec.Spec) bool {
		return ex.Passed()
	})

	numberOfFailedSpecs := runner.countSpecsSatisfying(func(ex *spec.Spec) bool {
		return ex.Failed()
	})

	success := true

	if numberOfFailedSpecs > 0 {
		success = false
	} else if numberOfPendingSpecs > 0 && runner.config.FailOnPending {
		success = false
	} else if runner.beforeSuiteNode != nil && !runner.beforeSuiteNode.Passed() {
		success = false
		numberOfFailedSpecs = numberOfSpecsThatWillBeRun
	}

	return &types.SuiteSummary{
		SuiteDescription: runner.description,
		SuiteSucceeded:   success,
		SuiteID:          runner.suiteID,

		NumberOfSpecsBeforeParallelization: runner.specs.NumberOfOriginalSpecs(),
		NumberOfTotalSpecs:                 len(runner.specs.Specs()),
		NumberOfSpecsThatWillBeRun:         numberOfSpecsThatWillBeRun,
		NumberOfPendingSpecs:               numberOfPendingSpecs,
		NumberOfSkippedSpecs:               numberOfSkippedSpecs,
		NumberOfPassedSpecs:                numberOfPassedSpecs,
		NumberOfFailedSpecs:                numberOfFailedSpecs,
	}
}
