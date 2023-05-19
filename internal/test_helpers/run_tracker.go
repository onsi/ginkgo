package test_helpers

import (
	"fmt"
	"strings"
	"sync"

	"github.com/onsi/ginkgo/v2/formatter"
	"github.com/onsi/ginkgo/v2/internal"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

/*

RunTracker tracks invocations of functions - useful to assert orders in which nodes run

*/

type RunTracker struct {
	lock        *sync.Mutex
	trackedRuns []string
	trackedData map[string]map[string]any
}

func NewRunTracker() *RunTracker {
	return &RunTracker{
		lock:        &sync.Mutex{},
		trackedData: map[string]map[string]any{},
	}
}

func (rt *RunTracker) Reset() {
	rt.lock.Lock()
	defer rt.lock.Unlock()
	rt.trackedRuns = []string{}
}

func (rt *RunTracker) Run(text string) {
	rt.lock.Lock()
	defer rt.lock.Unlock()
	rt.trackedRuns = append(rt.trackedRuns, text)
}

func (rt *RunTracker) RunWithData(text string, kv ...any) {
	rt.lock.Lock()
	defer rt.lock.Unlock()
	rt.trackedRuns = append(rt.trackedRuns, text)
	data := map[string]any{}
	for i := 0; i < len(kv); i += 2 {
		key := kv[i].(string)
		value := kv[i+1]
		data[key] = value
	}
	rt.trackedData[text] = data
}

func (rt *RunTracker) TrackedRuns() []string {
	rt.lock.Lock()
	defer rt.lock.Unlock()
	trackedRuns := make([]string, len(rt.trackedRuns))
	copy(trackedRuns, rt.trackedRuns)
	return trackedRuns
}

func (rt *RunTracker) DataFor(text string) map[string]any {
	rt.lock.Lock()
	defer rt.lock.Unlock()
	return rt.trackedData[text]
}

func (rt *RunTracker) T(text string, callback ...func()) func() {
	return func() {
		rt.Run(text)
		if len(callback) > 0 {
			callback[0]()
		}
	}
}

func (rt *RunTracker) TSC(text string, callback ...func(internal.SpecContext)) func(internal.SpecContext) {
	return func(c internal.SpecContext) {
		rt.Run(text)
		if len(callback) > 0 {
			callback[0](c)
		}
	}
}

func (rt *RunTracker) C(text string, callback ...func()) func(args []string, additionalArgs []string) {
	return func(args []string, additionalArgs []string) {
		rt.RunWithData(text, "Args", args, "AdditionalArgs", additionalArgs)
		if len(callback) > 0 {
			callback[0]()
		}
	}
}

func HaveRun(run string) OmegaMatcher {
	return WithTransform(func(rt *RunTracker) []string {
		return rt.TrackedRuns()
	}, ContainElement(run))
}

func HaveRunWithData(run string, kv ...any) OmegaMatcher {
	matchers := []types.GomegaMatcher{}
	for i := 0; i < len(kv); i += 2 {
		matchers = append(matchers, HaveKeyWithValue(kv[i], kv[i+1]))
	}
	return And(
		HaveRun(run),
		WithTransform(func(rt *RunTracker) map[string]any {
			return rt.DataFor(run)
		}, And(matchers...)),
	)
}

func HaveTrackedNothing() OmegaMatcher {
	return WithTransform(func(rt *RunTracker) []string {
		return rt.TrackedRuns()
	}, BeEmpty())
}

type HaveTrackedMatcher struct {
	expectedRuns []string
	message      string
}

func (m *HaveTrackedMatcher) Match(actual any) (bool, error) {
	rt, ok := actual.(*RunTracker)
	if !ok {
		return false, fmt.Errorf("HaveTracked() must be passed a RunTracker - got %T instead", actual)
	}
	actualRuns := rt.TrackedRuns()
	n := len(actualRuns)
	if n < len(m.expectedRuns) {
		n = len(m.expectedRuns)
	}
	failureMessage, success := &strings.Builder{}, true
	fmt.Fprintf(failureMessage, "{{/}}%10s == %-10s{{/}}\n", "Actual", "Expected")
	fmt.Fprintf(failureMessage, "{{/}}========================\n{{/}}")
	for i := 0; i < n; i++ {
		var expected, actual string
		if i < len(actualRuns) {
			actual = actualRuns[i]
		}
		if i < len(m.expectedRuns) {
			expected = m.expectedRuns[i]
		}
		if actual != expected {
			success = false
			fmt.Fprintf(failureMessage, "{{red}}%10s != %-10s{{/}}\n", actual, expected)
		} else {
			fmt.Fprintf(failureMessage, "{{green}}%10s == %-10s{{/}}\n", actual, expected)
		}

	}
	m.message = failureMessage.String()
	return success, nil

}
func (m *HaveTrackedMatcher) FailureMessage(actual any) string {
	return "Expected runs did not match tracked runs:\n" + formatter.F(m.message)

}
func (m *HaveTrackedMatcher) NegatedFailureMessage(actual any) string {
	return "Expected runs matched tracked runs:\n" + formatter.F(m.message)
}

func HaveTracked(runs ...string) OmegaMatcher {
	return &HaveTrackedMatcher{expectedRuns: runs}
}
