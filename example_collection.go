package ginkgo

import (
	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/types"

	"math/rand"
	"regexp"
	"sort"
	"time"
)

type exampleCollection struct {
	t                                 GinkgoTestingT
	description                       string
	examples                          []*example
	exampleCountBeforeParallelization int
	reporters                         []Reporter
	startTime                         time.Time
	suiteID                           string
	runningExample                    *example
	config                            config.GinkgoConfigType
}

func newExampleCollection(t GinkgoTestingT, description string, examples []*example, reporters []Reporter, config config.GinkgoConfigType) *exampleCollection {
	collection := &exampleCollection{
		t:           t,
		description: description,
		examples:    examples,
		reporters:   reporters,
		config:      config,
		suiteID:     types.GenerateRandomID(),
		exampleCountBeforeParallelization: len(examples),
	}

	collection.enumerateAndAssignExampleIndices()

	r := rand.New(rand.NewSource(config.RandomSeed))
	if config.RandomizeAllSpecs {
		collection.shuffle(r)
	}

	if config.FocusString == "" && config.SkipString == "" {
		collection.applyProgrammaticFocus()
	} else {
		collection.applyRegExpFocus(config.FocusString, config.SkipString)
	}

	if config.SkipMeasurements {
		collection.skipMeasurements()
	}

	if config.ParallelTotal > 1 {
		collection.trimForParallelization(config.ParallelTotal, config.ParallelNode)
	}

	return collection
}

func (collection *exampleCollection) enumerateAndAssignExampleIndices() {
	for index, example := range collection.examples {
		example.exampleIndex = index
	}
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

func (collection *exampleCollection) applyRegExpFocus(focusString string, skipString string) {
	for _, example := range collection.examples {
		matchesFocus := true
		matchesSkip := false

		stringToMatch := collection.description + " " + example.concatenatedString()

		if focusString != "" {
			focusFilter := regexp.MustCompile(focusString)
			matchesFocus = focusFilter.Match([]byte(stringToMatch))
		}

		if skipString != "" {
			skipFilter := regexp.MustCompile(skipString)
			matchesSkip = skipFilter.Match([]byte(stringToMatch))
		}

		if !matchesFocus || matchesSkip {
			example.skip()
		}
	}
}

func (collection *exampleCollection) trimForParallelization(parallelTotal int, parallelNode int) {
	startIndex, count := parallelizedIndexRange(len(collection.examples), parallelTotal, parallelNode)
	if count == 0 {
		collection.examples = make([]*example, 0)
	} else {
		collection.examples = collection.examples[startIndex : startIndex+count]
	}
}

func (collection *exampleCollection) skipMeasurements() {
	for _, example := range collection.examples {
		if example.subjectComponentType() == types.ExampleComponentTypeMeasure {
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

func (collection *exampleCollection) run() bool {
	collection.reportSuiteWillBegin()

	suiteFailed := false

	for _, example := range collection.examples {
		collection.reportExampleWillRun(example)

		if !example.skippedOrPending() {
			collection.runningExample = example
			example.run()
			if example.failed() {
				suiteFailed = true
			}
		} else if example.pending() && collection.config.FailOnPending {
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

func (collection *exampleCollection) fail(failure failureData) {
	if collection.runningExample != nil {
		collection.runningExample.fail(failure)
	}
}

func (collection *exampleCollection) currentGinkgoTestDescription() GinkgoTestDescription {
	currentExample := collection.runningExample
	if currentExample == nil {
		return GinkgoTestDescription{}
	}

	return currentExample.ginkgoTestDescription()
}

func (collection *exampleCollection) reportSuiteWillBegin() {
	collection.startTime = time.Now()
	summary := collection.summary()
	for _, reporter := range collection.reporters {
		reporter.SpecSuiteWillBegin(collection.config, summary)
	}
}

func (collection *exampleCollection) reportExampleWillRun(example *example) {
	summary := example.summary(collection.suiteID)
	for _, reporter := range collection.reporters {
		reporter.ExampleWillRun(summary)
	}
}

func (collection *exampleCollection) reportExampleDidComplete(example *example) {
	summary := example.summary(collection.suiteID)
	for _, reporter := range collection.reporters {
		reporter.ExampleDidComplete(summary)
	}
}

func (collection *exampleCollection) reportSuiteDidEnd() {
	summary := collection.summary()
	summary.RunTime = time.Since(collection.startTime)
	for _, reporter := range collection.reporters {
		reporter.SpecSuiteDidEnd(summary)
	}
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

func (collection *exampleCollection) summary() *types.SuiteSummary {
	numberOfExamplesThatWillBeRun := collection.countExamplesSatisfying(func(ex *example) bool {
		return !ex.skippedOrPending()
	})

	numberOfPendingExamples := collection.countExamplesSatisfying(func(ex *example) bool {
		return ex.state == types.ExampleStatePending
	})

	numberOfSkippedExamples := collection.countExamplesSatisfying(func(ex *example) bool {
		return ex.state == types.ExampleStateSkipped
	})

	numberOfPassedExamples := collection.countExamplesSatisfying(func(ex *example) bool {
		return ex.state == types.ExampleStatePassed
	})

	numberOfFailedExamples := collection.countExamplesSatisfying(func(ex *example) bool {
		return ex.failed()
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

		NumberOfExamplesBeforeParallelization: collection.exampleCountBeforeParallelization,
		NumberOfTotalExamples:                 len(collection.examples),
		NumberOfExamplesThatWillBeRun:         numberOfExamplesThatWillBeRun,
		NumberOfPendingExamples:               numberOfPendingExamples,
		NumberOfSkippedExamples:               numberOfSkippedExamples,
		NumberOfPassedExamples:                numberOfPassedExamples,
		NumberOfFailedExamples:                numberOfFailedExamples,
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
