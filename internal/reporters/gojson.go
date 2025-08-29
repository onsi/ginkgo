package reporters

import (
	"time"

	"github.com/onsi/ginkgo/v2/types"
)

func ptr[T any](in T) *T {
	return &in
}

type encoder interface {
	Encode(v any) error
}

// test2jsonEvent matches the format from go internals
// https://github.com/golang/go/blob/master/src/cmd/internal/test2json/test2json.go#L31-L41
// https://pkg.go.dev/cmd/test2json
type test2jsonEvent struct {
	Time        *time.Time `json:",omitempty"`
	Action      GoJSONAction
	Package     string   `json:",omitempty"`
	Test        string   `json:",omitempty"`
	Elapsed     *float64 `json:",omitempty"`
	Output      *string  `json:",omitempty"`
	FailedBuild string   `json:",omitempty"`
}

type GoJSONAction string

const (
	// start  - the test binary is about to be executed
	GoJSONStart GoJSONAction = "start"
	// run    - the test has started running
	GoJSONRun GoJSONAction = "run"
	// pause  - the test has been paused
	GoJSONPause GoJSONAction = "pause"
	// cont   - the test has continued running
	GoJSONCont GoJSONAction = "cont"
	// pass   - the test passed
	GoJSONPass GoJSONAction = "pass"
	// bench  - the benchmark printed log output but did not fail
	GoJSONBench GoJSONAction = "bench"
	// fail   - the test or benchmark failed
	GoJSONFail GoJSONAction = "fail"
	// output - the test printed output
	GoJSONOutput GoJSONAction = "output"
	// skip   - the test was skipped or the package contained no tests
	GoJSONSkip GoJSONAction = "skip"
)

func failureToOutput(failure types.Failure) string {
	return failure.Message
}
