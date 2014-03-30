package specrunner

import (
	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/internal/example"
	"github.com/onsi/ginkgo/internal/randomid"
	"github.com/onsi/ginkgo/internal/types"
	Writer "github.com/onsi/ginkgo/internal/writer"
	"github.com/onsi/ginkgo/reporters"
	"github.com/onsi/ginkgo/types"

	"time"
)

type SpecRunner struct {
	t              internaltypes.GinkgoTestingT
	description    string
	examples       *example.Examples
	reporters      []reporters.Reporter
	startTime      time.Time
	suiteID        string
	runningExample *example.Example
	writer         Writer.WriterInterface
	config         config.GinkgoConfigType
}

func New(t internaltypes.GinkgoTestingT, description string, examples *example.Examples, reporters []reporters.Reporter, writer Writer.WriterInterface, config config.GinkgoConfigType) *SpecRunner {
	return &SpecRunner{
		t:           t,
		description: description,
		examples:    examples,
		reporters:   reporters,
		writer:      writer,
		config:      config,
		suiteID:     randomid.New(),
	}
}

func (collection *SpecRunner) Run() bool {
	collection.reportSuiteWillBegin()
	suiteFailed := false

	for _, example := range collection.examples.Examples() {
		collection.writer.Truncate()

		collection.reportExampleWillRun(example)

		if !example.Skipped() && !example.Pending() {
			collection.runningExample = example
			example.Run()
			collection.runningExample = nil
			if example.Failed() {
				suiteFailed = true
				collection.writer.DumpOut()
			}
		} else if example.Pending() && collection.config.FailOnPending {
			suiteFailed = true
		}

		collection.reportExampleDidComplete(example)
	}

	collection.reportSuiteDidEnd()

	if suiteFailed {
		collection.t.Fail()
	}

	return !suiteFailed
}

func (collection *SpecRunner) CurrentExampleSummary() (*types.ExampleSummary, bool) {
	if collection.runningExample == nil {
		return nil, false
	}

	return collection.runningExample.Summary(collection.suiteID), true
}

func (collection *SpecRunner) reportSuiteWillBegin() {
	collection.startTime = time.Now()
	summary := collection.summary()
	for _, reporter := range collection.reporters {
		reporter.SpecSuiteWillBegin(collection.config, summary)
	}
}

func (collection *SpecRunner) reportExampleWillRun(example *example.Example) {
	summary := example.Summary(collection.suiteID)
	for _, reporter := range collection.reporters {
		reporter.ExampleWillRun(summary)
	}
}

func (collection *SpecRunner) reportExampleDidComplete(example *example.Example) {
	summary := example.Summary(collection.suiteID)
	for _, reporter := range collection.reporters {
		reporter.ExampleDidComplete(summary)
	}
}

func (collection *SpecRunner) reportSuiteDidEnd() {
	summary := collection.summary()
	summary.RunTime = time.Since(collection.startTime)
	for _, reporter := range collection.reporters {
		reporter.SpecSuiteDidEnd(summary)
	}
}

func (collection *SpecRunner) countExamplesSatisfying(filter func(ex *example.Example) bool) (count int) {
	count = 0

	for _, example := range collection.examples.Examples() {
		if filter(example) {
			count++
		}
	}

	return count
}

func (collection *SpecRunner) summary() *types.SuiteSummary {
	numberOfExamplesThatWillBeRun := collection.countExamplesSatisfying(func(ex *example.Example) bool {
		return !ex.Skipped() && !ex.Pending()
	})

	numberOfPendingExamples := collection.countExamplesSatisfying(func(ex *example.Example) bool {
		return ex.Pending()
	})

	numberOfSkippedExamples := collection.countExamplesSatisfying(func(ex *example.Example) bool {
		return ex.Skipped()
	})

	numberOfPassedExamples := collection.countExamplesSatisfying(func(ex *example.Example) bool {
		return ex.Passed()
	})

	numberOfFailedExamples := collection.countExamplesSatisfying(func(ex *example.Example) bool {
		return ex.Failed()
	})

	success := true

	if numberOfFailedExamples > 0 {
		success = false
	} else if numberOfPendingExamples > 0 && collection.config.FailOnPending {
		success = false
	}

	return &types.SuiteSummary{
		SuiteDescription: collection.description,
		SuiteSucceeded:   success,
		SuiteID:          collection.suiteID,

		NumberOfExamplesBeforeParallelization: collection.examples.NumberOfOriginalExamples(),
		NumberOfTotalExamples:                 len(collection.examples.Examples()),
		NumberOfExamplesThatWillBeRun:         numberOfExamplesThatWillBeRun,
		NumberOfPendingExamples:               numberOfPendingExamples,
		NumberOfSkippedExamples:               numberOfSkippedExamples,
		NumberOfPassedExamples:                numberOfPassedExamples,
		NumberOfFailedExamples:                numberOfFailedExamples,
	}
}
