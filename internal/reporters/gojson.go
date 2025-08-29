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
	Action      Test2JSONAction
	Package     string   `json:",omitempty"`
	Test        string   `json:",omitempty"`
	Elapsed     *float64 `json:",omitempty"`
	Output      *string  `json:",omitempty"`
	FailedBuild string   `json:",omitempty"`
}

type Test2JSONAction string

const (
	// start  - the test binary is about to be executed
	Test2JSONStart Test2JSONAction = "start"
	// run    - the test has started running
	Test2JSONRun Test2JSONAction = "run"
	// pause  - the test has been paused
	Test2JSONPause Test2JSONAction = "pause"
	// cont   - the test has continued running
	Test2JSONCont Test2JSONAction = "cont"
	// pass   - the test passed
	Test2JSONPass Test2JSONAction = "pass"
	// bench  - the benchmark printed log output but did not fail
	Test2JSONBench Test2JSONAction = "bench"
	// fail   - the test or benchmark failed
	Test2JSONFail Test2JSONAction = "fail"
	// output - the test printed output
	Test2JSONOutput Test2JSONAction = "output"
	// skip   - the test was skipped or the package contained no tests
	Test2JSONSkip Test2JSONAction = "skip"
)

func failureToOutput(failure types.Failure) string {
	return failure.Message
}
