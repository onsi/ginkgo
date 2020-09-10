package parallel_support

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"

	"github.com/onsi/ginkgo/internal"
	"github.com/onsi/ginkgo/reporters"

	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/types"
)

/*
The OutputInterceptor is used by the ForwardingReporter to
intercept and capture all stdin and stderr output during a test run.
*/
type OutputInterceptor interface {
	StartInterceptingOutput() error
	StopInterceptingAndReturnOutput() (string, error)
}

type ConfigAndSummary struct {
	Config  config.GinkgoConfigType `json:"config"`
	Summary types.SuiteSummary      `json:"suite-summary"`
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
	outputInterceptor OutputInterceptor
	debugFile         *os.File
	nestedReporter    *reporters.DefaultReporter
}

func NewForwardingReporter(config config.DefaultReporterConfigType, serverHost string, outputInterceptor OutputInterceptor, ginkgoWriter *internal.Writer, debugFile string) *ForwardingReporter {
	reporter := &ForwardingReporter{
		serverHost:        serverHost,
		outputInterceptor: outputInterceptor,
		debugFile:         nil,
	}

	if debugFile != "" {
		var err error
		reporter.debugFile, err = os.Create(debugFile)
		if err != nil {
			panic(err.Error())
		}

		config.Succinct = false
		config.NoColor = true
		config.Verbose = true
		config.FullTrace = true
		reporter.nestedReporter = reporters.NewDefaultReporter(config, reporter.debugFile)
	}

	return reporter
}

func (reporter *ForwardingReporter) post(path string, data interface{}) {
	encoded, _ := json.Marshal(data)
	buffer := bytes.NewBuffer(encoded)
	http.Post(reporter.serverHost+path, "application/json", buffer)
}

func (reporter *ForwardingReporter) SpecSuiteWillBegin(conf config.GinkgoConfigType, summary types.SuiteSummary) {
	data := ConfigAndSummary{Config: conf, Summary: summary}

	reporter.outputInterceptor.StartInterceptingOutput()
	if reporter.debugFile != nil {
		reporter.nestedReporter.SpecSuiteWillBegin(conf, summary)
		reporter.debugFile.Sync()
	}
	reporter.post("/SpecSuiteWillBegin", data)
}

func (reporter *ForwardingReporter) WillRun(summary types.Summary) {
	if reporter.debugFile != nil {
		reporter.nestedReporter.WillRun(summary)
		reporter.debugFile.Sync()
	}
}

func (reporter *ForwardingReporter) DidRun(summary types.Summary) {
	output, _ := reporter.outputInterceptor.StopInterceptingAndReturnOutput()
	reporter.outputInterceptor.StartInterceptingOutput()
	summary.CapturedStdOutErr = output
	if reporter.debugFile != nil {
		reporter.nestedReporter.DidRun(summary)
		reporter.debugFile.Sync()
	}
	reporter.post("/DidRun", summary)
}

func (reporter *ForwardingReporter) SpecSuiteDidEnd(summary types.SuiteSummary) {
	reporter.outputInterceptor.StopInterceptingAndReturnOutput()
	if reporter.debugFile != nil {
		reporter.nestedReporter.SpecSuiteDidEnd(summary)
		reporter.debugFile.Sync()
	}
	reporter.post("/SpecSuiteDidEnd", summary)
}
