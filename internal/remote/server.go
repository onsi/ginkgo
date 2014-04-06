/*

The remote package provides the pieces to allow Ginkgo test suites to report to remote listeners.
This is used, primarily, to enable streaming parallel test output but has, in principal, broader applications (e.g. streaming test output to a browser).

*/

package remote

import (
	"encoding/json"
	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/reporters"
	"github.com/onsi/ginkgo/types"
	"io/ioutil"
	"net"
	"net/http"
)

/*
Server spins up on an automatically selected port and listens for communication from the forwarding reporter.
It then forwards that communication to attached reporters.
*/
type Server struct {
	listener  net.Listener
	reporters []reporters.Reporter
}

//Create a new server, automatically selecting a port
func NewServer() (*Server, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	return &Server{
		listener: listener,
	}, nil
}

//Start the server.  You don't need to `go s.Start()`, just `s.Start()`
func (server *Server) Start() {
	httpServer := &http.Server{}
	mux := http.NewServeMux()
	httpServer.Handler = mux

	mux.HandleFunc("/SpecSuiteWillBegin", func(writer http.ResponseWriter, request *http.Request) {
		defer request.Body.Close()
		body, _ := ioutil.ReadAll(request.Body)
		server.specSuiteWillBegin(body)
		writer.WriteHeader(200)
	})

	mux.HandleFunc("/BeforeSuiteDidRun", func(writer http.ResponseWriter, request *http.Request) {
		defer request.Body.Close()
		body, _ := ioutil.ReadAll(request.Body)
		server.beforeSuiteDidRun(body)
		writer.WriteHeader(200)
	})

	mux.HandleFunc("/AfterSuiteDidRun", func(writer http.ResponseWriter, request *http.Request) {
		defer request.Body.Close()
		body, _ := ioutil.ReadAll(request.Body)
		server.afterSuiteDidRun(body)
		writer.WriteHeader(200)
	})

	mux.HandleFunc("/SpecWillRun", func(writer http.ResponseWriter, request *http.Request) {
		defer request.Body.Close()
		body, _ := ioutil.ReadAll(request.Body)
		server.specWillRun(body)
		writer.WriteHeader(200)
	})

	mux.HandleFunc("/SpecDidComplete", func(writer http.ResponseWriter, request *http.Request) {
		defer request.Body.Close()
		body, _ := ioutil.ReadAll(request.Body)
		server.specDidComplete(body)
		writer.WriteHeader(200)
	})

	mux.HandleFunc("/SpecSuiteDidEnd", func(writer http.ResponseWriter, request *http.Request) {
		defer request.Body.Close()
		body, _ := ioutil.ReadAll(request.Body)
		server.specSuiteDidEnd(body)
		writer.WriteHeader(200)
	})

	go httpServer.Serve(server.listener)
}

//Stop the server
func (server *Server) Stop() {
	server.listener.Close()
}

//The address the server can be reached it.  Pass this into the `ForwardingReporter`.
func (server *Server) Address() string {
	return server.listener.Addr().String()
}

//The server will forward all received messages to Ginkgo reporters registered with `RegisterReporters`
func (server *Server) RegisterReporters(reporters ...reporters.Reporter) {
	server.reporters = reporters
}

func (server *Server) specSuiteWillBegin(body []byte) {
	var data struct {
		Config  config.GinkgoConfigType `json:"config"`
		Summary *types.SuiteSummary     `json:"suite-summary"`
	}

	json.Unmarshal(body, &data)

	for _, reporter := range server.reporters {
		reporter.SpecSuiteWillBegin(data.Config, data.Summary)
	}
}

func (server *Server) beforeSuiteDidRun(body []byte) {
	var setupSummary *types.SetupSummary
	json.Unmarshal(body, &setupSummary)

	for _, reporter := range server.reporters {
		reporter.BeforeSuiteDidRun(setupSummary)
	}
}

func (server *Server) afterSuiteDidRun(body []byte) {
	var setupSummary *types.SetupSummary
	json.Unmarshal(body, &setupSummary)

	for _, reporter := range server.reporters {
		reporter.AfterSuiteDidRun(setupSummary)
	}
}

func (server *Server) specWillRun(body []byte) {
	var specSummary *types.SpecSummary
	json.Unmarshal(body, &specSummary)

	for _, reporter := range server.reporters {
		reporter.SpecWillRun(specSummary)
	}
}

func (server *Server) specDidComplete(body []byte) {
	var specSummary *types.SpecSummary
	json.Unmarshal(body, &specSummary)

	for _, reporter := range server.reporters {
		reporter.SpecDidComplete(specSummary)
	}
}

func (server *Server) specSuiteDidEnd(body []byte) {
	var suiteSummary *types.SuiteSummary
	json.Unmarshal(body, &suiteSummary)

	for _, reporter := range server.reporters {
		reporter.SpecSuiteDidEnd(suiteSummary)
	}
}
