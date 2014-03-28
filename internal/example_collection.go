package internal

import (
	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/internal/example"
	"github.com/onsi/ginkgo/internal/randomid"
	"github.com/onsi/ginkgo/internal/types"
	"github.com/onsi/ginkgo/reporters"
	"github.com/onsi/ginkgo/types"

	"math/rand"
	"regexp"
	"sort"
	"time"
)

type exampleCollection struct {
	t                                 internaltypes.GinkgoTestingT
	description                       string
	examples                          []*example.Example
	exampleCountBeforeParallelization int
	reporters                         []reporters.Reporter
	startTime                         time.Time
	suiteID                           string
	runningExample                    *example.Example
	writer                            ginkgoWriterInterface
	config                            config.GinkgoConfigType
}

func newExampleCollection(t internaltypes.GinkgoTestingT, description string, examples []*example.Example, reporters []reporters.Reporter, writer ginkgoWriterInterface, config config.GinkgoConfigType) *exampleCollection {
	collection := &exampleCollection{
		t:           t,
		description: description,
		examples:    examples,
		reporters:   reporters,
		writer:      writer,
		config:      config,
		suiteID:     randomid.New(),
		exampleCountBeforeParallelization: len(examples),
	}

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

func (collection *exampleCollection) applyProgrammaticFocus() {
	hasFocusedTests := false
	for _, example := range collection.examples {
		if example.Focused() {
			hasFocusedTests = true
			break
		}
	}

	if hasFocusedTests {
		for _, example := range collection.examples {
			if !example.Focused() {
				example.Skip()
			}
		}
	}
}

func (collection *exampleCollection) applyRegExpFocus(focusString string, skipString string) {
	for _, example := range collection.examples {
		matchesFocus := true
		matchesSkip := false

		stringToMatch := collection.description + " " + example.ConcatenatedString()

		if focusString != "" {
			focusFilter := regexp.MustCompile(focusString)
			matchesFocus = focusFilter.Match([]byte(stringToMatch))
		}

		if skipString != "" {
			skipFilter := regexp.MustCompile(skipString)
			matchesSkip = skipFilter.Match([]byte(stringToMatch))
		}

		if !matchesFocus || matchesSkip {
			example.Skip()
		}
	}
}

func (collection *exampleCollection) trimForParallelization(parallelTotal int, parallelNode int) {
	startIndex, count := parallelizedIndexRange(len(collection.examples), parallelTotal, parallelNode)
	if count == 0 {
		collection.examples = make([]*example.Example, 0)
	} else {
		collection.examples = collection.examples[startIndex : startIndex+count]
	}
}

func (collection *exampleCollection) skipMeasurements() {
	for _, example := range collection.examples {
		if example.IsMeasurement() {
			example.Skip()
		}
	}
}

func (collection *exampleCollection) shuffle(r *rand.Rand) {
	sort.Sort(collection)
	permutation := r.Perm(len(collection.examples))
	shuffledExamples := make([]*example.Example, len(collection.examples))
	for i, j := range permutation {
		shuffledExamples[i] = collection.examples[j]
	}
	collection.examples = shuffledExamples
}

func (collection *exampleCollection) run() bool {
	collection.reportSuiteWillBegin()
	suiteFailed := false

	for _, example := range collection.examples {
		collection.writer.Truncate()

		collection.reportExampleWillRun(example)

		if !example.Skipped() && !example.Pending() {
			collection.runningExample = example
			example.Run()
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

func (collection *exampleCollection) currentExampleSummary() (*types.ExampleSummary, bool) {
	if collection.runningExample == nil {
		return nil, false
	}

	return collection.runningExample.Summary(collection.suiteID), true
}

func (collection *exampleCollection) reportSuiteWillBegin() {
	collection.startTime = time.Now()
	summary := collection.summary()
	for _, reporter := range collection.reporters {
		reporter.SpecSuiteWillBegin(collection.config, summary)
	}
}

func (collection *exampleCollection) reportExampleWillRun(example *example.Example) {
	summary := example.Summary(collection.suiteID)
	for _, reporter := range collection.reporters {
		reporter.ExampleWillRun(summary)
	}
}

func (collection *exampleCollection) reportExampleDidComplete(example *example.Example) {
	summary := example.Summary(collection.suiteID)
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

func (collection *exampleCollection) countExamplesSatisfying(filter func(ex *example.Example) bool) (count int) {
	count = 0

	for _, example := range collection.examples {
		if filter(example) {
			count++
		}
	}

	return count
}

func (collection *exampleCollection) summary() *types.SuiteSummary {
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
	return collection.examples[i].ConcatenatedString() < collection.examples[j].ConcatenatedString()
}

func (collection *exampleCollection) Swap(i, j int) {
	collection.examples[i], collection.examples[j] = collection.examples[j], collection.examples[i]
}
