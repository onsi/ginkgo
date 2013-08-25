package ginkgo

import (
	"math/rand"
	"testing"
	"time"
)

type exampleCollection struct {
	t              *testing.T
	description    string
	examples       []*example
	reporter       Reporter
	startTime      time.Time
	runningExample *example
}

func newExampleCollection(t *testing.T, description string, examples []*example, reporter Reporter) *exampleCollection {
	hasFocusedTests := false
	for _, example := range examples {
		if example.focused {
			hasFocusedTests = true
			break
		}
	}

	if hasFocusedTests {
		for _, example := range examples {
			if !example.focused {
				example.skip()
			}
		}
	}

	return &exampleCollection{
		t:           t,
		description: description,
		examples:    examples,
		reporter:    reporter,
	}
}

func (collection *exampleCollection) shuffle(r *rand.Rand) {
	permutation := r.Perm(len(collection.examples))
	shuffledExamples := make([]*example, len(collection.examples))
	for i, j := range permutation {
		shuffledExamples[i] = collection.examples[j]
	}
	collection.examples = shuffledExamples
}

func (collection *exampleCollection) run() {
	collection.reportSuiteWillBegin()

	suiteFailed := false

	for _, example := range collection.examples {
		if !example.skippedOrPending() {
			collection.runningExample = example
			example.run()
			if example.failed() {
				suiteFailed = true
			}
		}

		collection.reportExample(example)
	}

	collection.reportSuiteDidEnd()

	if suiteFailed {
		collection.t.Fail()
	}
}

func (collection *exampleCollection) fail(failure failureData) {
	if collection.runningExample != nil {
		collection.runningExample.fail(failure)
	}
}

func (collection *exampleCollection) reportSuiteWillBegin() {
	collection.startTime = time.Now()
	collection.reporter.SpecSuiteWillBegin(collection.summary())
}

func (collection *exampleCollection) reportExample(example *example) {
	collection.reporter.ExampleDidComplete(example.summary())
}

func (collection *exampleCollection) reportSuiteDidEnd() {
	summary := collection.summary()
	summary.RunTime = time.Since(collection.startTime)
	collection.reporter.SpecSuiteDidEnd(summary)
}

func (collection *exampleCollection) countExamplesSatisfying(filter func(ex *example) bool) (count int) {
	count = 0

	for _, example := range collection.examples {
		if filter(example) {
			count++
		}
	}

	return count
}

func (collection *exampleCollection) summary() *SuiteSummary {
	numberOfExamplesThatWillBeRun := collection.countExamplesSatisfying(func(ex *example) bool {
		return !ex.skippedOrPending()
	})

	numberOfPendingExamples := collection.countExamplesSatisfying(func(ex *example) bool {
		return ex.state == ExampleStatePending
	})

	numberOfSkippedExamples := collection.countExamplesSatisfying(func(ex *example) bool {
		return ex.state == ExampleStateSkipped
	})

	numberOfPassedExamples := collection.countExamplesSatisfying(func(ex *example) bool {
		return ex.state == ExampleStatePassed
	})

	numberOfFailedExamples := collection.countExamplesSatisfying(func(ex *example) bool {
		return ex.failed()
	})

	return &SuiteSummary{
		SuiteDescription: collection.description,

		NumberOfTotalExamples:         len(collection.examples),
		NumberOfExamplesThatWillBeRun: numberOfExamplesThatWillBeRun,
		NumberOfPendingExamples:       numberOfPendingExamples,
		NumberOfSkippedExamples:       numberOfSkippedExamples,
		NumberOfPassedExamples:        numberOfPassedExamples,
		NumberOfFailedExamples:        numberOfFailedExamples,
	}
}
