/*

The remote package provides the pieces to allow Ginkgo test suites to report to remote listeners.
This is used, primarily, to enable streaming parallel test output but has, in principal, broader applications (e.g. streaming test output to a browser).

*/

package parallel_support

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"os"
	"sync"

	"github.com/onsi/ginkgo/reporters"
	"github.com/onsi/ginkgo/types"
)

type beforeSuiteState struct {
	Data  []byte
	State types.SpecState
}

type ParallelIndexCounter struct {
	Index int
}

/*
Server spins up on an automatically selected port and listens for communication from the forwarding reporter.
It then forwards that communication to attached reporters.
*/
type Server struct {
	Done              chan interface{}
	OutputDestination io.Writer

	listener         net.Listener
	reporter         reporters.Reporter
	alives           []func() bool
	lock             *sync.Mutex
	beforeSuiteState beforeSuiteState
	parallelTotal    int
	counter          int
	shouldAbort      bool

	numSuiteDidBegins int
	numSuiteDidEnds   int
	aggregatedReport  types.Report
	reportHoldingArea []types.SpecReport
}

//Create a new server, automatically selecting a port
func NewServer(parallelTotal int, reporter reporters.Reporter) (*Server, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	return &Server{
		listener:          listener,
		reporter:          reporter,
		lock:              &sync.Mutex{},
		alives:            make([]func() bool, parallelTotal),
		beforeSuiteState:  beforeSuiteState{Data: nil, State: types.SpecStateInvalid},
		parallelTotal:     parallelTotal,
		OutputDestination: os.Stdout,
		Done:              make(chan interface{}),
	}, nil
}

//Start the server.  You don't need to `go s.Start()`, just `s.Start()`
func (server *Server) Start() {
	httpServer := &http.Server{}
	mux := http.NewServeMux()
	httpServer.Handler = mux

	//streaming endpoints
	mux.HandleFunc("/suite-will-begin", server.specSuiteWillBegin)
	mux.HandleFunc("/did-run", server.didRun)
	mux.HandleFunc("/suite-did-end", server.specSuiteDidEnd)
	mux.HandleFunc("/stream-output", server.streamOutput)

	//synchronization endpoints
	mux.HandleFunc("/before-suite-failed", server.handleBeforeSuiteFailed)
	mux.HandleFunc("/before-suite-succeeded", server.handleBeforeSuiteSucceeded)
	mux.HandleFunc("/before-suite-state", server.handleBeforeSuiteState)
	mux.HandleFunc("/have-nonprimary-nodes-finished", server.handleHaveNonprimaryNodesFinished)
	mux.HandleFunc("/aggregated-nonprimary-nodes-report", server.handleAggregatedNonprimaryNodesReport)
	mux.HandleFunc("/counter", server.handleCounter)
	mux.HandleFunc("/up", server.handleUp)
	mux.HandleFunc("/abort", server.handleAbort)

	go httpServer.Serve(server.listener)
}

//Stop the server
func (server *Server) Close() {
	server.listener.Close()
}

//The address the server can be reached it.  Pass this into the `ForwardingReporter`.
func (server *Server) Address() string {
	return "http://" + server.listener.Addr().String()
}

//
// Streaming Endpoints
//

//The server will forward all received messages to Ginkgo reporters registered with `RegisterReporters`
func (server *Server) decode(request *http.Request, object interface{}) error {
	defer request.Body.Close()
	return json.NewDecoder(request.Body).Decode(object)
}

func (server *Server) specSuiteWillBegin(writer http.ResponseWriter, request *http.Request) {
	server.lock.Lock()
	defer server.lock.Unlock()

	server.numSuiteDidBegins += 1

	var report types.Report
	err := server.decode(request, &report)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	// all summaries are identical, so it's fine to simply emit the last one of these
	if server.numSuiteDidBegins == server.parallelTotal {
		server.reporter.SuiteWillBegin(report)

		for _, summary := range server.reportHoldingArea {
			server.reporter.WillRun(summary)
			server.reporter.DidRun(summary)
		}

		server.reportHoldingArea = nil
	}
}

func (server *Server) didRun(writer http.ResponseWriter, request *http.Request) {
	server.lock.Lock()
	defer server.lock.Unlock()

	var report types.SpecReport
	err := server.decode(request, &report)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	if server.numSuiteDidBegins == server.parallelTotal {
		server.reporter.WillRun(report)
		server.reporter.DidRun(report)
	} else {
		server.reportHoldingArea = append(server.reportHoldingArea, report)
	}
}

func (server *Server) specSuiteDidEnd(writer http.ResponseWriter, request *http.Request) {
	server.lock.Lock()
	defer server.lock.Unlock()

	server.numSuiteDidEnds += 1

	var report types.Report
	err := server.decode(request, &report)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	if server.numSuiteDidEnds == 1 {
		server.aggregatedReport = report
	} else {
		server.aggregatedReport = server.aggregatedReport.Add(report)
	}

	if server.numSuiteDidEnds == server.parallelTotal {
		server.reporter.SuiteDidEnd(server.aggregatedReport)
		close(server.Done)
	}
}

func (server *Server) streamOutput(writer http.ResponseWriter, request *http.Request) {
	io.Copy(server.OutputDestination, request.Body)
}

//
// Synchronization Endpoints
//

func (server *Server) RegisterAlive(node int, alive func() bool) {
	server.lock.Lock()
	defer server.lock.Unlock()
	server.alives[node-1] = alive
}

func (server *Server) nodeIsAlive(node int) bool {
	server.lock.Lock()
	defer server.lock.Unlock()
	alive := server.alives[node-1]
	if alive == nil {
		return true
	}
	return alive()
}

func (server *Server) haveNonprimaryNodesFinished() bool {
	for i := 2; i <= server.parallelTotal; i++ {
		if server.nodeIsAlive(i) {
			return false
		}
	}
	return true
}

func (server *Server) handleBeforeSuiteFailed(writer http.ResponseWriter, request *http.Request) {
	server.lock.Lock()
	defer server.lock.Unlock()
	server.beforeSuiteState.State = types.SpecStateFailed
}

func (server *Server) handleBeforeSuiteSucceeded(writer http.ResponseWriter, request *http.Request) {
	server.lock.Lock()
	defer server.lock.Unlock()
	server.beforeSuiteState.State = types.SpecStatePassed
	var err error
	server.beforeSuiteState.Data, err = io.ReadAll(request.Body)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
	}
}

func (server *Server) handleBeforeSuiteState(writer http.ResponseWriter, request *http.Request) {
	node1IsAlive := server.nodeIsAlive(1)
	server.lock.Lock()
	defer server.lock.Unlock()
	beforeSuiteState := server.beforeSuiteState
	switch beforeSuiteState.State {
	case types.SpecStatePassed:
		writer.Write(beforeSuiteState.Data)
	case types.SpecStateFailed:
		writer.WriteHeader(http.StatusFailedDependency)
	case types.SpecStateInvalid:
		if node1IsAlive {
			writer.WriteHeader(http.StatusTooEarly)
		} else {
			writer.WriteHeader(http.StatusGone)
		}
	}
}

func (server *Server) handleHaveNonprimaryNodesFinished(writer http.ResponseWriter, request *http.Request) {
	if server.haveNonprimaryNodesFinished() {
		writer.WriteHeader(http.StatusOK)
	} else {
		writer.WriteHeader(http.StatusTooEarly)
	}
}

func (server *Server) handleAggregatedNonprimaryNodesReport(writer http.ResponseWriter, request *http.Request) {
	if server.haveNonprimaryNodesFinished() {
		server.lock.Lock()
		defer server.lock.Unlock()
		if server.numSuiteDidEnds == server.parallelTotal-1 {
			json.NewEncoder(writer).Encode(server.aggregatedReport)
		} else {
			writer.WriteHeader(http.StatusGone)
		}
	} else {
		writer.WriteHeader(http.StatusTooEarly)
	}
}

func (server *Server) handleCounter(writer http.ResponseWriter, request *http.Request) {
	c := ParallelIndexCounter{}
	server.lock.Lock()
	c.Index = server.counter
	server.counter++
	server.lock.Unlock()

	json.NewEncoder(writer).Encode(c)
}

func (server *Server) handleUp(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
}

func (server *Server) handleAbort(writer http.ResponseWriter, request *http.Request) {
	server.lock.Lock()
	defer server.lock.Unlock()
	if request.Method == "GET" {
		if server.shouldAbort {
			writer.WriteHeader(http.StatusGone)
		} else {
			writer.WriteHeader(http.StatusOK)
		}
	} else {
		server.shouldAbort = true
	}
}
