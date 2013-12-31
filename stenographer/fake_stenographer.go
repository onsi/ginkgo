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

func (stenographer *FakeStenographer) AnnounceSuite(description string, randomSeed int64, randomizingAll bool) {
	stenographer.registerCall("AnnounceSuite", description, randomSeed, randomizingAll)
}

func (stenographer *FakeStenographer) AnnounceAggregatedParallelRun(nodes int) {
	stenographer.registerCall("AnnounceAggregatedParallelRun", nodes)
}

func (stenographer *FakeStenographer) AnnounceParallelRun(node int, nodes int, specsToRun int, totalSpecs int) {
	stenographer.registerCall("AnnounceParallelRun", node, nodes, specsToRun, totalSpecs)
}

func (stenographer *FakeStenographer) AnnounceNumberOfSpecs(specsToRun int, total int) {
	stenographer.registerCall("AnnounceNumberOfSpecs", specsToRun, total)
}

func (stenographer *FakeStenographer) AnnounceSpecRunCompletion(summary *types.SuiteSummary) {
	stenographer.registerCall("AnnounceSpecRunCompletion", summary)
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

func (stenographer *FakeStenographer) AnnouncePendingExample(example *types.ExampleSummary, noisy bool, succinct bool) {
	stenographer.registerCall("AnnouncePendingExample", example, noisy, succinct)
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
