package leafnodes

import (
	"encoding/json"
	"net"
	"net/http"
	"sync"
)

type RemoteStateState int

const (
	RemoteStateStateInvalid RemoteStateState = iota

	RemoteStateStatePending
	RemoteStateStatePassed
	RemoteStateStateFailed
	RemoteStateStateDisappeared
)

type RemoteState struct {
	Data  []byte
	State RemoteStateState
}

func (r RemoteState) ToJSON() []byte {
	data, _ := json.Marshal(r)
	return data
}

type AfterSuiteCanRun struct {
	CanRun bool
}

type CompoundServer struct {
	listener         net.Listener
	alives           []func() bool
	lock             *sync.Mutex
	beforeSuiteState RemoteState
	numNodes         int
}

func NewCompoundServer(numNodes int) (*CompoundServer, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	return &CompoundServer{
		listener:         listener,
		lock:             &sync.Mutex{},
		alives:           make([]func() bool, numNodes),
		beforeSuiteState: RemoteState{nil, RemoteStateStatePending},
		numNodes:         numNodes,
	}, nil
}

func (server *CompoundServer) Start() {
	httpServer := &http.Server{}
	mux := http.NewServeMux()
	httpServer.Handler = mux

	mux.HandleFunc("/BeforeSuiteState", server.handleBeforeSuiteState)
	mux.HandleFunc("/AfterSuiteCanRun", server.handleAfterSuiteCanRun)

	go httpServer.Serve(server.listener)
}

func (server *CompoundServer) Close() {
	server.listener.Close()
}

func (server *CompoundServer) Address() string {
	return server.listener.Addr().String()
}

func (server *CompoundServer) RegisterAlive(node int, alive func() bool) {
	server.lock.Lock()
	defer server.lock.Unlock()
	server.alives[node-1] = alive
}

func (server *CompoundServer) nodeIsAlive(node int) bool {
	server.lock.Lock()
	defer server.lock.Unlock()
	alive := server.alives[node-1]
	if alive == nil {
		return true
	}
	return alive()
}

func (server *CompoundServer) handleBeforeSuiteState(writer http.ResponseWriter, request *http.Request) {
	if request.Method == "POST" {
		dec := json.NewDecoder(request.Body)
		dec.Decode(&(server.beforeSuiteState))
	} else {
		state := server.beforeSuiteState
		if state.State == RemoteStateStatePending && !server.nodeIsAlive(1) {
			state.State = RemoteStateStateDisappeared
		}
		enc := json.NewEncoder(writer)
		enc.Encode(state)
	}
}

func (server *CompoundServer) handleAfterSuiteCanRun(writer http.ResponseWriter, request *http.Request) {
	canRun := AfterSuiteCanRun{
		CanRun: true,
	}
	for i := 2; i <= server.numNodes; i++ {
		canRun.CanRun = canRun.CanRun && !server.nodeIsAlive(i)
	}

	enc := json.NewEncoder(writer)
	enc.Encode(canRun)
}
