/*

The remote package provides the pieces to allow Ginkgo test suites to report to remote listeners.
This is used, primarily, to enable streaming parallel test output but has, in principal, broader applications (e.g. streaming test output to a browser).

*/

package parallel_support

import (
	"encoding/json"
	"net"
	"net/http"
	"sync"

	"github.com/onsi/ginkgo/internal"

	"github.com/onsi/ginkgo/reporters"
	"github.com/onsi/ginkgo/types"
)

/*
Server spins up on an automatically selected port and listens for communication from the forwarding reporter.
It then forwards that communication to attached reporters.
*/
type Server struct {
	Done chan interface{}

	listener        net.Listener
	reporter        reporters.Reporter
	alives          []func() bool
	lock            *sync.Mutex
	beforeSuiteData types.RemoteBeforeSuiteData
	parallelTotal   int
	counter         int

	numSuiteDidBegins         int
	numSuiteDidEnds           int
	aggregatedSuiteEndSummary types.SuiteSummary
	reportHoldingArea         []types.SpecReport
}

//Create a new server, automatically selecting a port
func NewServer(parallelTotal int, reporter reporters.Reporter) (*Server, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	return &Server{
		listener:        listener,
		reporter:        reporter,
		lock:            &sync.Mutex{},
		alives:          make([]func() bool, parallelTotal),
		beforeSuiteData: types.RemoteBeforeSuiteData{Data: nil, State: types.RemoteBeforeSuiteStatePending},
		parallelTotal:   parallelTotal,
		Done:            make(chan interface{}),
	}, nil
}

//Start the server.  You don't need to `go s.Start()`, just `s.Start()`
func (server *Server) Start() {
	httpServer := &http.Server{}
	mux := http.NewServeMux()
	httpServer.Handler = mux

	//streaming endpoints
	mux.HandleFunc("/SpecSuiteWillBegin", server.specSuiteWillBegin)
	mux.HandleFunc("/DidRun", server.didRun)
	mux.HandleFunc("/SpecSuiteDidEnd", server.specSuiteDidEnd)

	//synchronization endpoints
	mux.HandleFunc("/BeforeSuiteState", server.handleBeforeSuiteState)
	mux.HandleFunc("/AfterSuiteState", server.handleRemoteAfterSuiteData)
	mux.HandleFunc("/counter", server.handleCounter)
	mux.HandleFunc("/up", server.handleUp)

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

	var data SuiteConfigAndSummary
	err := server.decode(request, &data)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	// all summaries are identical, so it's fine to simply emit the last one of these
	if server.numSuiteDidBegins == server.parallelTotal {
		server.reporter.SpecSuiteWillBegin(data.SuiteConfig, data.Summary)

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

	var summary types.SuiteSummary
	err := server.decode(request, &summary)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	if server.numSuiteDidEnds == 1 {
		server.aggregatedSuiteEndSummary = summary
	} else {
		server.aggregatedSuiteEndSummary = server.aggregatedSuiteEndSummary.Add(summary)
	}

	if server.numSuiteDidEnds == server.parallelTotal {
		server.reporter.SpecSuiteDidEnd(server.aggregatedSuiteEndSummary)
		close(server.Done)
	}
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

func (server *Server) handleBeforeSuiteState(writer http.ResponseWriter, request *http.Request) {
	if request.Method == "POST" {
		server.lock.Lock()
		dec := json.NewDecoder(request.Body)
		dec.Decode(&(server.beforeSuiteData))
		server.lock.Unlock()
	} else {
		server.lock.Lock()
		beforeSuiteData := server.beforeSuiteData
		server.lock.Unlock()
		if beforeSuiteData.State == types.RemoteBeforeSuiteStatePending && !server.nodeIsAlive(1) {
			beforeSuiteData.State = types.RemoteBeforeSuiteStateDisappeared
		}
		enc := json.NewEncoder(writer)
		enc.Encode(beforeSuiteData)
	}
}

func (server *Server) handleRemoteAfterSuiteData(writer http.ResponseWriter, request *http.Request) {
	afterSuiteData := types.RemoteAfterSuiteData{
		CanRun: true,
	}
	for i := 2; i <= server.parallelTotal; i++ {
		afterSuiteData.CanRun = afterSuiteData.CanRun && !server.nodeIsAlive(i)
	}

	enc := json.NewEncoder(writer)
	enc.Encode(afterSuiteData)
}

func (server *Server) handleCounter(writer http.ResponseWriter, request *http.Request) {
	c := internal.Counter{}
	server.lock.Lock()
	c.Index = server.counter
	server.counter++
	server.lock.Unlock()

	json.NewEncoder(writer).Encode(c)
}

func (server *Server) handleUp(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
}
