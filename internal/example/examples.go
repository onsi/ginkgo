package example

import (
	"math/rand"
	"regexp"
	"sort"
)

type Examples struct {
	examples                 []*Example
	numberOfOriginalExamples int
}

func NewExamples(examples []*Example) *Examples {
	return &Examples{
		examples:                 examples,
		numberOfOriginalExamples: len(examples),
	}
}

func (e *Examples) Examples() []*Example {
	return e.examples
}

func (e *Examples) NumberOfOriginalExamples() int {
	return e.numberOfOriginalExamples
}

func (e *Examples) Shuffle(r *rand.Rand) {
	sort.Sort(e)
	permutation := r.Perm(len(e.examples))
	shuffledExamples := make([]*Example, len(e.examples))
	for i, j := range permutation {
		shuffledExamples[i] = e.examples[j]
	}
	e.examples = shuffledExamples
}

func (e *Examples) ApplyFocus(description string, focusString string, skipString string) {
	if focusString == "" && skipString == "" {
		e.applyProgrammaticFocus()
	} else {
		e.applyRegExpFocus(description, focusString, skipString)
	}
}

func (e *Examples) applyProgrammaticFocus() {
	hasFocusedTests := false
	for _, example := range e.examples {
		if example.Focused() {
			hasFocusedTests = true
			break
		}
	}

	if hasFocusedTests {
		for _, example := range e.examples {
			if !example.Focused() {
				example.Skip()
			}
		}
	}
}

func (e *Examples) applyRegExpFocus(description string, focusString string, skipString string) {
	for _, example := range e.examples {
		matchesFocus := true
		matchesSkip := false

		toMatch := []byte(description + " " + example.ConcatenatedString())

		if focusString != "" {
			focusFilter := regexp.MustCompile(focusString)
			matchesFocus = focusFilter.Match([]byte(toMatch))
		}

		if skipString != "" {
			skipFilter := regexp.MustCompile(skipString)
			matchesSkip = skipFilter.Match([]byte(toMatch))
		}

		if !matchesFocus || matchesSkip {
			example.Skip()
		}
	}
}

func (e *Examples) SkipMeasurements() {
	for _, example := range e.examples {
		if example.IsMeasurement() {
			example.Skip()
		}
	}
}

func (e *Examples) TrimForParallelization(total int, node int) {
	startIndex, count := ParallelizedIndexRange(len(e.examples), total, node)
	if count == 0 {
		e.examples = make([]*Example, 0)
	} else {
		e.examples = e.examples[startIndex : startIndex+count]
	}
}

//sort.Interface

func (e *Examples) Len() int {
	return len(e.examples)
}

func (e *Examples) Less(i, j int) bool {
	return e.examples[i].ConcatenatedString() < e.examples[j].ConcatenatedString()
}

func (e *Examples) Swap(i, j int) {
	e.examples[i], e.examples[j] = e.examples[j], e.examples[i]
}
