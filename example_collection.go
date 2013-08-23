package ginkgo

import (
	"math/rand"
	"testing"
	"time"
)

type exampleCollection struct {
	t               *testing.T
	description     string
	examples        []*example
	reporter        Reporter
	hasFocusedTests bool
	startTime       time.Time
	runningExample  *example
}

func newExampleCollection(t *testing.T, description string, examples []*example, reporter Reporter) *exampleCollection {
	hasFocusedTests := false
	for _, example := range examples {
		if example.hasFocusFlag {
			hasFocusedTests = true
			break
		}
	}

	return &exampleCollection{
		t:               t,
		description:     description,
		examples:        examples,
		reporter:        reporter,
		hasFocusedTests: hasFocusedTests,
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
	collection.reportBeginning()

	suiteFailed := false

	for _, example := range collection.examples {
		if collection.hasFocusedTests {
			if example.hasFocusFlag {
				if !example.hasPendingFlag {
					collection.runningExample = example
					collection.runningExample.run()
				}
			} else {
				example.skip()
			}
		} else {
			if !example.hasPendingFlag {
				collection.runningExample = example
				collection.runningExample.run()
			}
		}
		if example.outcome != runOutcomePassed {
			suiteFailed = true
		}

		collection.reportExample(example)
	}

	collection.reportEnding()

	if suiteFailed {
		collection.t.Fail()
	}
}

func (collection *exampleCollection) fail(failure failureData) {
	collection.runningExample.fail(failure)
}

func (collection *exampleCollection) numberOfPendingExamples() (count int) {
	for _, example := range collection.examples {
		if collection.hasFocusedTests {
			if example.hasPendingFlag && example.hasFocusFlag {
				count++
			}
		} else if example.hasPendingFlag {
			count++
		}
	}

	return
}

func (collection *exampleCollection) numberOfSkippedExamples() (count int) {
	if !collection.hasFocusedTests {
		return 0
	}

	for _, example := range collection.examples {
		if !example.hasFocusFlag {
			count++
		}
	}

	return
}

func (collection *exampleCollection) numberOfExamplesThatWillBeRun() (count int) {
	for _, example := range collection.examples {
		if collection.hasFocusedTests {
			if example.hasFocusFlag && !example.hasPendingFlag {
				count++
			}
		} else if !example.hasPendingFlag {
			count++
		}
	}

	return
}

func (collection *exampleCollection) numberOfPassedExamples() (count int) {
	for _, example := range collection.examples {
		if example.outcome == runOutcomePassed {
			count++
		}
	}

	return count
}

func (collection *exampleCollection) numberOfFailedExamples() (count int) {
	for _, example := range collection.examples {
		if example.outcome == runOutcomeFailed || example.outcome == runOutcomeTimedOut || example.outcome == runOutcomePanicked {
			count++
		}
	}

	return count
}

func (collection *exampleCollection) reportBeginning() {
	collection.startTime = time.Now()

	summary := &SuiteSummary{
		SuiteDescription: collection.description,

		NumberOfTotalExamples:         len(collection.examples),
		NumberOfExamplesThatWillBeRun: collection.numberOfExamplesThatWillBeRun(),
		NumberOfPendingExamples:       collection.numberOfPendingExamples(),
		NumberOfSkippedExamples:       collection.numberOfSkippedExamples(),
		NumberOfPassedExamples:        0,
		NumberOfFailedExamples:        0,
		RunTime:                       0,
	}

	collection.reporter.SpecSuiteWillBegin(summary)
}

func (collection *exampleCollection) reportEnding() {
	runTime := time.Since(collection.startTime)

	summary := &SuiteSummary{
		SuiteDescription: collection.description,

		NumberOfTotalExamples:         len(collection.examples),
		NumberOfExamplesThatWillBeRun: collection.numberOfExamplesThatWillBeRun(),
		NumberOfPendingExamples:       collection.numberOfPendingExamples(),
		NumberOfSkippedExamples:       collection.numberOfSkippedExamples(),
		NumberOfPassedExamples:        collection.numberOfPassedExamples(),
		NumberOfFailedExamples:        collection.numberOfFailedExamples(),
		RunTime:                       runTime,
	}

	collection.reporter.SpecSuiteDidEnd(summary)
}

func (collection *exampleCollection) reportExample(example *example) {
	collection.reporter.ExampleDidComplete(example.summary())
}
