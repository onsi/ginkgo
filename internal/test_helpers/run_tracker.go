package test_helpers

import (
	"sync"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

/*

RunTracker tracks invocations of functions - useful to assert orders in which nodes run

*/

type RunTracker struct {
	lock        *sync.Mutex
	trackedRuns []string
	trackedData map[string]map[string]interface{}
}

func NewRunTracker() *RunTracker {
	return &RunTracker{
		lock:        &sync.Mutex{},
		trackedData: map[string]map[string]interface{}{},
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

func (rt *RunTracker) RunWithData(text string, kv ...interface{}) {
	rt.lock.Lock()
	defer rt.lock.Unlock()
	rt.trackedRuns = append(rt.trackedRuns, text)
	data := map[string]interface{}{}
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

func (rt *RunTracker) DataFor(text string) map[string]interface{} {
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

func HaveRunWithData(run string, kv ...interface{}) OmegaMatcher {
	matchers := []types.GomegaMatcher{}
	for i := 0; i < len(kv); i += 2 {
		matchers = append(matchers, HaveKeyWithValue(kv[i], kv[i+1]))
	}
	return And(
		HaveRun(run),
		WithTransform(func(rt *RunTracker) map[string]interface{} {
			return rt.DataFor(run)
		}, And(matchers...)),
	)
}

func HaveTracked(runs ...string) OmegaMatcher {
	return WithTransform(func(rt *RunTracker) []string {
		return rt.TrackedRuns()
	}, Equal(runs))
}

func HaveTrackedNothing() OmegaMatcher {
	return WithTransform(func(rt *RunTracker) []string {
		return rt.TrackedRuns()
	}, BeEmpty())
}
