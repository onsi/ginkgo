package ginkgo

import (
	"math/rand"
	"regexp"
	"sort"
	"time"
)

type testingT interface {
	Fail()
}

type exampleCollection struct {
	t              testingT
	description    string
	examples       []*example
	reporter       Reporter
	startTime      time.Time
	runningExample *example
	config         GinkgoConfigType
}

func newExampleCollection(t testingT, description string, examples []*example, focusFilter *regexp.Regexp, reporter Reporter, config GinkgoConfigType) *exampleCollection {
	collection := &exampleCollection{
		t:           t,
		description: description,
		examples:    examples,
		reporter:    reporter,
		config:      config,
	}

	if focusFilter == nil {
		collection.applyProgrammaticFocus()
	} else {
		collection.applyRegExpFocus(focusFilter)
	}

	return collection
}

func (collection *exampleCollection) applyProgrammaticFocus() {
	hasFocusedTests := false
	for _, example := range collection.examples {
		if example.focused {
			hasFocusedTests = true
			break
		}
	}

	if hasFocusedTests {
		for _, example := range collection.examples {
			if !example.focused {
				example.skip()
			}
		}
	}
}

func (collection *exampleCollection) applyRegExpFocus(focusFilter *regexp.Regexp) {
	for _, example := range collection.examples {
		if !focusFilter.Match([]byte(example.concatenatedString())) {
			example.skip()
		}
	}
}

func (collection *exampleCollection) shuffle(r *rand.Rand) {
	sort.Sort(collection)
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
	collection.reporter.SpecSuiteWillBegin(collection.config, collection.summary())
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

//sort.Interface

func (collection *exampleCollection) Len() int {
	return len(collection.examples)
}

func (collection *exampleCollection) Less(i, j int) bool {
	return collection.examples[i].concatenatedString() < collection.examples[j].concatenatedString()
}

func (collection *exampleCollection) Swap(i, j int) {
	collection.examples[i], collection.examples[j] = collection.examples[j], collection.examples[i]
}
