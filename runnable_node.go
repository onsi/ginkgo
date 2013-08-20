package godescribe

import (
	"fmt"
	"reflect"
	"time"
)

type runState uint

const (
	runStateInvalid runState = iota
	runStatePassed
	runStateFailed
	runStatePanicked
	runStateTimedOut
)

type runnableNode struct {
	isAsync          bool
	asyncFunc        func(Done)
	syncFunc         func()
	codeLocation     CodeLocation
	timeoutThreshold time.Duration
}

func newRunnableNode(body interface{}, codeLocation CodeLocation) *runnableNode {
	bodyType := reflect.TypeOf(body)
	if bodyType.Kind() != reflect.Func {
		panic(fmt.Sprintf("Expected a function but got something else at %v", codeLocation))
	}

	switch bodyType.NumIn() {
	case 0:
		return &runnableNode{
			isAsync:          false,
			asyncFunc:        nil,
			syncFunc:         body.(func()),
			codeLocation:     codeLocation,
			timeoutThreshold: 5.0 * time.Second,
		}
	case 1:
		if bodyType.In(0) != reflect.TypeOf((*Done)(nil)).Elem() {
			panic(fmt.Sprintf("Must pass a Done channel to function at %v", codeLocation))
		}

		return &runnableNode{
			isAsync:          true,
			asyncFunc:        body.(func(Done)),
			syncFunc:         nil,
			codeLocation:     codeLocation,
			timeoutThreshold: 5.0 * time.Second,
		}
	}

	panic(fmt.Sprintf("Too many arguments to function at %v", codeLocation))
}

func (runnable *runnableNode) run() (state runState, runTime time.Duration, failure failureData) {
	done := make(chan interface{}, 1)
	startTime := time.Now()

	if runnable.isAsync {
		go runnable.asyncFunc(done)
	} else {
		runnable.syncFunc()
		done <- true
	}

	defer func() {
		runTime = time.Since(startTime)
		if e := recover(); e != nil {
			if reflect.TypeOf(e) == reflect.TypeOf((*failureData)(nil)).Elem() {
				state = runStateFailed
				failure = e.(failureData)
			} else {
				state = runStatePanicked
				failure = failureData{
					message:        "Panic",
					codeLocation:   runnable.codeLocation, //todo: can we get the code location that threw the panic from within the defer
					forwardedPanic: e,
				}
			}
		}
	}()

	select {
	case <-done:
		state = runStatePassed
		runTime = time.Since(startTime)
		failure = failureData{}
	case <-time.After(runnable.timeoutThreshold):
		state = runStateTimedOut
		runTime = time.Since(startTime)
		failure = failureData{
			message:      "Timed out",
			codeLocation: runnable.codeLocation,
		}
	}

	return
}

// beforeEach

type beforeEachNode struct {
	*runnableNode
}

func newBeforeEachNode(body interface{}, codeLocation CodeLocation) *beforeEachNode {
	return &beforeEachNode{
		runnableNode: newRunnableNode(body, codeLocation),
	}
}

// justBeforeEach

type justBeforeEachNode struct {
	*runnableNode
}

func newJustBeforeEachNode(body interface{}, codeLocation CodeLocation) *justBeforeEachNode {
	return &justBeforeEachNode{
		runnableNode: newRunnableNode(body, codeLocation),
	}
}

// afterEach

type afterEachNode struct {
	*runnableNode
}

func newAfterEachNode(body interface{}, codeLocation CodeLocation) *afterEachNode {
	return &afterEachNode{
		runnableNode: newRunnableNode(body, codeLocation),
	}
}

// it

type itNode struct {
	*runnableNode

	flag flagType
	text string
}

func newItNode(text string, body interface{}, flag flagType, codeLocation CodeLocation) *itNode {
	return &itNode{
		runnableNode: newRunnableNode(body, codeLocation),
		flag:         flag,
		text:         text,
	}
}

func (node *itNode) isContainerNode() bool {
	return false
}

func (node *itNode) isItNode() bool {
	return true
}
