package specrunner

import (
	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/internal/example"
	Writer "github.com/onsi/ginkgo/internal/writer"
	"github.com/onsi/ginkgo/reporters"
	"github.com/onsi/ginkgo/types"

	"time"
)

type SpecRunner struct {
	description    string
	examples       *example.Examples
	reporters      []reporters.Reporter
	startTime      time.Time
	suiteID        string
	runningExample *example.Example
	writer         Writer.WriterInterface
	config         config.GinkgoConfigType
}

func New(description string, examples *example.Examples, reporters []reporters.Reporter, writer Writer.WriterInterface, config config.GinkgoConfigType) *SpecRunner {
	return &SpecRunner{
		description: description,
		examples:    examples,
		reporters:   reporters,
		writer:      writer,
		config:      config,
		suiteID:     randomID(),
	}
}

func (runner *SpecRunner) Run() bool {
	runner.reportSuiteWillBegin()
	suiteFailed := false

	for _, example := range runner.examples.Examples() {
		runner.writer.Truncate()

		runner.reportExampleWillRun(example)

		if !example.Skipped() && !example.Pending() {
			runner.runningExample = example
			example.Run()
			runner.runningExample = nil
			if example.Failed() {
				suiteFailed = true
				runner.writer.DumpOut()
			}
		} else if example.Pending() && runner.config.FailOnPending {
			suiteFailed = true
		}

		runner.reportExampleDidComplete(example)
	}

	runner.reportSuiteDidEnd()

	return !suiteFailed
}

func (runner *SpecRunner) CurrentExampleSummary() (*types.ExampleSummary, bool) {
	if runner.runningExample == nil {
		return nil, false
	}

	return runner.runningExample.Summary(runner.suiteID), true
}

func (runner *SpecRunner) reportSuiteWillBegin() {
	runner.startTime = time.Now()
	summary := runner.summary()
	for _, reporter := range runner.reporters {
		reporter.SpecSuiteWillBegin(runner.config, summary)
	}
}

func (runner *SpecRunner) reportExampleWillRun(example *example.Example) {
	summary := example.Summary(runner.suiteID)
	for _, reporter := range runner.reporters {
		reporter.ExampleWillRun(summary)
	}
}

func (runner *SpecRunner) reportExampleDidComplete(example *example.Example) {
	summary := example.Summary(runner.suiteID)
	for _, reporter := range runner.reporters {
		reporter.ExampleDidComplete(summary)
	}
}

func (runner *SpecRunner) reportSuiteDidEnd() {
	summary := runner.summary()
	summary.RunTime = time.Since(runner.startTime)
	for _, reporter := range runner.reporters {
		reporter.SpecSuiteDidEnd(summary)
	}
}

func (runner *SpecRunner) countExamplesSatisfying(filter func(ex *example.Example) bool) (count int) {
	count = 0

	for _, example := range runner.examples.Examples() {
		if filter(example) {
			count++
		}
	}

	return count
}

func (runner *SpecRunner) summary() *types.SuiteSummary {
	numberOfExamplesThatWillBeRun := runner.countExamplesSatisfying(func(ex *example.Example) bool {
		return !ex.Skipped() && !ex.Pending()
	})

	numberOfPendingExamples := runner.countExamplesSatisfying(func(ex *example.Example) bool {
		return ex.Pending()
	})

	numberOfSkippedExamples := runner.countExamplesSatisfying(func(ex *example.Example) bool {
		return ex.Skipped()
	})

	numberOfPassedExamples := runner.countExamplesSatisfying(func(ex *example.Example) bool {
		return ex.Passed()
	})

	numberOfFailedExamples := runner.countExamplesSatisfying(func(ex *example.Example) bool {
		return ex.Failed()
	})

	success := true

	if numberOfFailedExamples > 0 {
		success = false
	} else if numberOfPendingExamples > 0 && runner.config.FailOnPending {
		success = false
	}

	return &types.SuiteSummary{
		SuiteDescription: runner.description,
		SuiteSucceeded:   success,
		SuiteID:          runner.suiteID,

		NumberOfExamplesBeforeParallelization: runner.examples.NumberOfOriginalExamples(),
		NumberOfTotalExamples:                 len(runner.examples.Examples()),
		NumberOfExamplesThatWillBeRun:         numberOfExamplesThatWillBeRun,
		NumberOfPendingExamples:               numberOfPendingExamples,
		NumberOfSkippedExamples:               numberOfSkippedExamples,
		NumberOfPassedExamples:                numberOfPassedExamples,
		NumberOfFailedExamples:                numberOfFailedExamples,
	}
}
