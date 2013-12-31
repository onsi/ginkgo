package remote

import (
	"bytes"
	"encoding/json"
	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/types"
	"io"
	"net/http"
)

//An interface to net/http's client to allow the injection of fakes under test
type Poster interface {
	Post(url string, bodyType string, body io.Reader) (resp *http.Response, err error)
}

/*
The ForwardingReporter is a Ginkgo reporter that forwards information to
a Ginkgo remote server.

When streaming parallel test output, this repoter is automatically installed by Ginkgo.

This is accomplished by passing in the GINKGO_REMOTE_REPORTING_SERVER environment variable to `go test`, the Ginkgo test runner
detects this environment variable (which should contain the host of the server) and automatically installs a ForwardingReporter
in place of Ginkgo's DefaultReporter.
*/

type ForwardingReporter struct {
	serverHost        string
	poster            Poster
	outputInterceptor OutputInterceptor
}

func NewForwardingReporter(serverHost string, poster Poster, outputInterceptor OutputInterceptor) *ForwardingReporter {
	return &ForwardingReporter{
		serverHost:        serverHost,
		poster:            poster,
		outputInterceptor: outputInterceptor,
	}
}

func (reporter *ForwardingReporter) post(path string, data interface{}) {
	encoded, _ := json.Marshal(data)
	buffer := bytes.NewBuffer(encoded)
	reporter.poster.Post("http://"+reporter.serverHost+path, "application/json", buffer)
}

func (reporter *ForwardingReporter) SpecSuiteWillBegin(conf config.GinkgoConfigType, summary *types.SuiteSummary) {
	data := struct {
		Config  config.GinkgoConfigType `json:"config"`
		Summary *types.SuiteSummary     `json:"suite-summary"`
	}{
		conf,
		summary,
	}

	reporter.post("/SpecSuiteWillBegin", data)
}

func (reporter *ForwardingReporter) ExampleWillRun(exampleSummary *types.ExampleSummary) {
	reporter.outputInterceptor.StartInterceptingOutput()
	reporter.post("/ExampleWillRun", exampleSummary)
}

func (reporter *ForwardingReporter) ExampleDidComplete(exampleSummary *types.ExampleSummary) {
	output, _ := reporter.outputInterceptor.StopInterceptingAndReturnOutput()
	exampleSummary.CapturedOutput = output
	reporter.post("/ExampleDidComplete", exampleSummary)
}

func (reporter *ForwardingReporter) SpecSuiteDidEnd(summary *types.SuiteSummary) {
	reporter.post("/SpecSuiteDidEnd", summary)
}
