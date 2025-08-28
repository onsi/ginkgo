package reporters

import (
	"encoding/json"
	"time"

	"github.com/onsi/ginkgo/v2/types"
)

func ptr[T any](in T) *T {
	return &in
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

type GoJSONEventWriter struct {
	enc *json.Encoder
}

func NewGoJSONEventWriter(enc *json.Encoder) *GoJSONEventWriter {
	return &GoJSONEventWriter{
		enc: enc,
	}
}

func (r *GoJSONEventWriter) writeEvent(e *test2jsonEvent) error {
	return r.enc.Encode(e)
}

func (r *GoJSONEventWriter) WriteSuiteStart(goPkg string, report types.Report) error {
	e := &test2jsonEvent{
		Time:        &report.StartTime,
		Action:      Test2JSONStart,
		Package:     goPkg,
		Output:      nil,
		FailedBuild: "",
	}
	return r.writeEvent(e)
}

func (r *GoJSONEventWriter) WriteSuiteResult(goPkg string, report types.Report) error {
	var action Test2JSONAction
	switch {
	case report.PreRunStats.SpecsThatWillRun == 0:
		action = Test2JSONSkip
	case report.SuiteSucceeded:
		action = Test2JSONPass
	default:
		action = Test2JSONFail
	}
	e := &test2jsonEvent{
		Time:        &report.EndTime,
		Action:      action,
		Package:     goPkg,
		Output:      nil,
		FailedBuild: "",
		Elapsed:     ptr(float64(report.RunTime.Seconds())),
	}
	return r.writeEvent(e)
}

func (r *GoJSONEventWriter) WriteSuiteLeafNodesOut(goPkg string, specReport types.SpecReport) error {
	events := []*test2jsonEvent{}

	combinedOutput := specReport.CombinedOutput()
	if combinedOutput != "" {
		events = append(events, &test2jsonEvent{
			Time:        &specReport.EndTime,
			Action:      Test2JSONOutput,
			Package:     goPkg,
			Output:      ptr(combinedOutput),
			FailedBuild: "",
		})
	}

	for _, ev := range events {
		err := r.writeEvent(ev)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *GoJSONEventWriter) WriteSpecStart(goPkg string, specReport types.SpecReport) error {
	testName, err := testNameFromSpecReport(specReport)
	if err != nil {
		return err
	}
	e := &test2jsonEvent{
		Time:        &specReport.StartTime,
		Action:      Test2JSONRun,
		Test:        testName,
		Package:     goPkg,
		Output:      nil,
		FailedBuild: "",
	}
	return r.writeEvent(e)
}

func (r *GoJSONEventWriter) WriteSpecOut(goPkg string, specReport types.SpecReport) error {
	events := []*test2jsonEvent{}
	testName, err := testNameFromSpecReport(specReport)
	if err != nil {
		return err
	}
	combinedOutput := specReport.CombinedOutput()
	if combinedOutput != "" {
		events = append(events, &test2jsonEvent{
			Time:        &specReport.EndTime,
			Action:      Test2JSONOutput,
			Test:        testName,
			Package:     goPkg,
			Output:      ptr(combinedOutput),
			FailedBuild: "",
		})
	}
	if specReport.Failure.Message != "" {
		events = append(events, &test2jsonEvent{
			Time:        &specReport.EndTime,
			Action:      Test2JSONOutput,
			Test:        testName,
			Package:     goPkg,
			Output:      ptr(failureToOutput(specReport.Failure)),
			FailedBuild: "",
		})
	}
	for _, ev := range events {
		err = r.writeEvent(ev)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *GoJSONEventWriter) WriteSpecResult(goPkg string, specReport types.SpecReport) error {
	testName, err := testNameFromSpecReport(specReport)
	if err != nil {
		return err
	}
	status := specStateToAction(specReport.State)
	e := &test2jsonEvent{
		Time:        &specReport.EndTime,
		Action:      status,
		Test:        testName,
		Package:     goPkg,
		Elapsed:     ptr(specReport.RunTime.Seconds()),
		Output:      nil,
		FailedBuild: "",
	}
	return r.writeEvent(e)
}
