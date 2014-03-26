package stenographer

import (
	"github.com/onsi/ginkgo/types"
)

func NewFakeStenographerCall(method string, args ...interface{}) FakeStenographerCall {
	return FakeStenographerCall{
		Method: method,
		Args:   args,
	}
}

type FakeStenographer struct {
	Calls []FakeStenographerCall
}

type FakeStenographerCall struct {
	Method string
	Args   []interface{}
}

func NewFakeStenographer() *FakeStenographer {
	stenographer := &FakeStenographer{}
	stenographer.Reset()
	return stenographer
}

func (stenographer *FakeStenographer) Reset() {
	stenographer.Calls = make([]FakeStenographerCall, 0)
}

func (stenographer *FakeStenographer) CallsTo(method string) []FakeStenographerCall {
	results := make([]FakeStenographerCall, 0)
	for _, call := range stenographer.Calls {
		if call.Method == method {
			results = append(results, call)
		}
	}

	return results
}

func (stenographer *FakeStenographer) registerCall(method string, args ...interface{}) {
	stenographer.Calls = append(stenographer.Calls, NewFakeStenographerCall(method, args...))
}

func (stenographer *FakeStenographer) AnnounceSuite(description string, randomSeed int64, randomizingAll bool, succinct bool) {
	stenographer.registerCall("AnnounceSuite", description, randomSeed, randomizingAll, succinct)
}

func (stenographer *FakeStenographer) AnnounceAggregatedParallelRun(nodes int, succinct bool) {
	stenographer.registerCall("AnnounceAggregatedParallelRun", nodes, succinct)
}

func (stenographer *FakeStenographer) AnnounceParallelRun(node int, nodes int, specsToRun int, totalSpecs int, succinct bool) {
	stenographer.registerCall("AnnounceParallelRun", node, nodes, specsToRun, totalSpecs, succinct)
}

func (stenographer *FakeStenographer) AnnounceNumberOfSpecs(specsToRun int, total int, succinct bool) {
	stenographer.registerCall("AnnounceNumberOfSpecs", specsToRun, total, succinct)
}

func (stenographer *FakeStenographer) AnnounceSpecRunCompletion(summary *types.SuiteSummary, succinct bool) {
	stenographer.registerCall("AnnounceSpecRunCompletion", summary, succinct)
}

func (stenographer *FakeStenographer) AnnounceExampleWillRun(example *types.ExampleSummary) {
	stenographer.registerCall("AnnounceExampleWillRun", example)
}

func (stenographer *FakeStenographer) AnnounceCapturedOutput(example *types.ExampleSummary) {
	stenographer.registerCall("AnnounceCapturedOutput", example)
}

func (stenographer *FakeStenographer) AnnounceSuccesfulExample(example *types.ExampleSummary) {
	stenographer.registerCall("AnnounceSuccesfulExample", example)
}

func (stenographer *FakeStenographer) AnnounceSuccesfulSlowExample(example *types.ExampleSummary, succinct bool) {
	stenographer.registerCall("AnnounceSuccesfulSlowExample", example, succinct)
}

func (stenographer *FakeStenographer) AnnounceSuccesfulMeasurement(example *types.ExampleSummary, succinct bool) {
	stenographer.registerCall("AnnounceSuccesfulMeasurement", example, succinct)
}

func (stenographer *FakeStenographer) AnnouncePendingExample(example *types.ExampleSummary, noisy bool) {
	stenographer.registerCall("AnnouncePendingExample", example, noisy)
}

func (stenographer *FakeStenographer) AnnounceSkippedExample(example *types.ExampleSummary) {
	stenographer.registerCall("AnnounceSkippedExample", example)
}

func (stenographer *FakeStenographer) AnnounceExampleTimedOut(example *types.ExampleSummary, succinct bool) {
	stenographer.registerCall("AnnounceExampleTimedOut", example, succinct)
}

func (stenographer *FakeStenographer) AnnounceExamplePanicked(example *types.ExampleSummary, succinct bool) {
	stenographer.registerCall("AnnounceExamplePanicked", example, succinct)
}

func (stenographer *FakeStenographer) AnnounceExampleFailed(example *types.ExampleSummary, succinct bool) {
	stenographer.registerCall("AnnounceExampleFailed", example, succinct)
}
