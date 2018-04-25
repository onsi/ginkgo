/*

The remote package provides the pieces to allow Ginkgo test suites to report to remote listeners.
This is used, primarily, to enable streaming parallel test output but has, in principal, broader applications (e.g. streaming test output to a browser).

*/

package remote

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/onsi/ginkgo/internal/spec_iterator"

	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/types"
)

/*
Server spins up on an automatically selected port and listens for communication from the forwarding reporter.
It then forwards that communication to attached reporters.
*/
type Server struct {
	listener        net.Listener
	aggregator      *Aggregator
	alives          []func() bool
	signalDebug     []func()
	lock            *sync.Mutex
	beforeSuiteData types.RemoteBeforeSuiteData
	parallelTotal   int
	counter         int
}

//Create a new server, automatically selecting a port
func NewServer(parallelTotal int, aggregator *Aggregator) (*Server, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	return &Server{
		aggregator:      aggregator,
		listener:        listener,
		lock:            &sync.Mutex{},
		alives:          make([]func() bool, parallelTotal),
		signalDebug:     make([]func(), parallelTotal),
		beforeSuiteData: types.RemoteBeforeSuiteData{Data: nil, State: types.RemoteBeforeSuiteStatePending},
		parallelTotal:   parallelTotal,
	}, nil
}

//Start the server.  You don't need to `go s.Start()`, just `s.Start()`
func (server *Server) Start() {
	httpServer := &http.Server{}
	mux := http.NewServeMux()
	httpServer.Handler = mux

	//streaming endpoints
	mux.HandleFunc("/debug", server.debug)
	mux.HandleFunc("/SpecSuiteWillBegin", server.specSuiteWillBegin)
	mux.HandleFunc("/BeforeSuiteDidRun", server.beforeSuiteDidRun)
	mux.HandleFunc("/AfterSuiteDidRun", server.afterSuiteDidRun)
	mux.HandleFunc("/SpecWillRun", server.specWillRun)
	mux.HandleFunc("/UpdateSpecDebugOutput", server.updateSpecDebugOutput)
	mux.HandleFunc("/SpecDidComplete", server.specDidComplete)
	mux.HandleFunc("/SpecSuiteDidEnd", server.specSuiteDidEnd)

	//synchronization endpoints
	mux.HandleFunc("/BeforeSuiteState", server.handleBeforeSuiteState)
	mux.HandleFunc("/RemoteAfterSuiteData", server.handleRemoteAfterSuiteData)
	mux.HandleFunc("/counter", server.handleCounter)
	mux.HandleFunc("/has-counter", server.handleHasCounter) //for backward compatibility

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
func (server *Server) readAll(request *http.Request) []byte {
	defer request.Body.Close()
	body, _ := ioutil.ReadAll(request.Body)
	return body
}

func (server *Server) RegisterSignalDebug(node int, debug func()) {
	server.lock.Lock()
	defer server.lock.Unlock()
	server.signalDebug[node-1] = debug
}

func (server *Server) debug(writer http.ResponseWriter, request *http.Request) {
	server.lock.Lock()
	for cpu := 1; cpu <= server.parallelTotal; cpu++ {
		server.signalDebug[cpu-1]()
	}
	server.lock.Unlock()

	//Give nodes a chance to report back
	time.Sleep(time.Second)

	writer.Write(server.aggregator.DebugReport())
}

func (server *Server) specSuiteWillBegin(writer http.ResponseWriter, request *http.Request) {
	body := server.readAll(request)

	var data struct {
		Config  config.GinkgoConfigType `json:"config"`
		Summary *types.SuiteSummary     `json:"suite-summary"`
	}

	json.Unmarshal(body, &data)

	if server.aggregator != nil {
		server.aggregator.SpecSuiteWillBegin(data.Config, data.Summary)
	}
}

func (server *Server) beforeSuiteDidRun(writer http.ResponseWriter, request *http.Request) {
	body := server.readAll(request)
	var setupSummary *types.SetupSummary
	json.Unmarshal(body, &setupSummary)

	if server.aggregator != nil {
		server.aggregator.BeforeSuiteDidRun(setupSummary)
	}
}

func (server *Server) afterSuiteDidRun(writer http.ResponseWriter, request *http.Request) {
	body := server.readAll(request)
	var setupSummary *types.SetupSummary
	json.Unmarshal(body, &setupSummary)

	if server.aggregator != nil {
		server.aggregator.AfterSuiteDidRun(setupSummary)
	}
}

func (server *Server) specWillRun(writer http.ResponseWriter, request *http.Request) {
	body := server.readAll(request)
	var specSummary *types.SpecSummary
	json.Unmarshal(body, &specSummary)

	if server.aggregator != nil {
		server.aggregator.SpecWillRun(specSummary)
	}
}

func (server *Server) updateSpecDebugOutput(writer http.ResponseWriter, request *http.Request) {
	body := server.readAll(request)
	var debugOutput *types.ParallelSpecDebugOutput
	json.Unmarshal(body, &debugOutput)

	if server.aggregator != nil {
		server.aggregator.UpdateSpecDebugOutput(debugOutput)
	}
}

func (server *Server) specDidComplete(writer http.ResponseWriter, request *http.Request) {
	body := server.readAll(request)
	var specSummary *types.SpecSummary
	json.Unmarshal(body, &specSummary)

	if server.aggregator != nil {
		server.aggregator.SpecDidComplete(specSummary)
	}
}

func (server *Server) specSuiteDidEnd(writer http.ResponseWriter, request *http.Request) {
	body := server.readAll(request)
	var suiteSummary *types.SuiteSummary
	json.Unmarshal(body, &suiteSummary)

	if server.aggregator != nil {
		server.aggregator.SpecSuiteDidEnd(suiteSummary)
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
		dec := json.NewDecoder(request.Body)
		dec.Decode(&(server.beforeSuiteData))
	} else {
		beforeSuiteData := server.beforeSuiteData
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
	c := spec_iterator.Counter{}
	server.lock.Lock()
	c.Index = server.counter
	server.counter = server.counter + 1
	server.lock.Unlock()

	json.NewEncoder(writer).Encode(c)
}

func (server *Server) handleHasCounter(writer http.ResponseWriter, request *http.Request) {
	writer.Write([]byte(""))
}
