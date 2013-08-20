package godescribe

import (
	"fmt"
	"reflect"
	"time"
)

type runOutcome uint

const (
	runOutcomeInvalid runOutcome = iota
	runOutcomePassed
	runOutcomeFailed
	runOutcomePanicked
	runOutcomeTimedOut
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

func (runnable *runnableNode) run() (outcome runOutcome, failure failureData) {
	done := make(chan interface{}, 1)

	defer func() {
		if e := recover(); e != nil {
			outcome = runOutcomePanicked
			failure = failureData{
				message:        "Test Panicked",
				codeLocation:   runnable.codeLocation, //todo: can we get the code location that threw the panic from within the defer
				forwardedPanic: e,
			}
		}
	}()

	if runnable.isAsync {
		go runnable.asyncFunc(done)
	} else {
		runnable.syncFunc()
		done <- true
	}

	select {
	case <-done:
		outcome = runOutcomePassed
	case <-time.After(runnable.timeoutThreshold):
		outcome = runOutcomeTimedOut
		failure = failureData{
			message:      "Timed out",
			codeLocation: runnable.codeLocation,
		}
	}

	return
}

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
