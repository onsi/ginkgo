package ginkgo

import (
	"fmt"
	"github.com/onsi/ginkgo/types"
	"reflect"
	"time"
)

type runnableNode struct {
	isAsync          bool
	asyncFunc        func(Done)
	syncFunc         func()
	codeLocation     types.CodeLocation
	timeoutThreshold time.Duration
}

func newRunnableNode(body interface{}, codeLocation types.CodeLocation, timeout time.Duration) *runnableNode {
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
			timeoutThreshold: timeout,
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
			timeoutThreshold: timeout,
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
				codeLocation:   types.GenerateCodeLocation(2),
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
		outcome = runOutcomeCompleted
	case <-time.After(runnable.timeoutThreshold):
		outcome = runOutcomeTimedOut
		failure = failureData{
			message:      "Timed out",
			codeLocation: runnable.codeLocation,
		}
	}

	return
}

//It Node

type itNode struct {
	*runnableNode

	flag flagType
	text string
}

func newItNode(text string, body interface{}, flag flagType, codeLocation types.CodeLocation, timeout time.Duration) *itNode {
	return &itNode{
		runnableNode: newRunnableNode(body, codeLocation, timeout),
		flag:         flag,
		text:         text,
	}
}

func (node *itNode) nodeType() nodeType {
	return nodeTypeIt
}

func (node *itNode) getText() string {
	return node.text
}

func (node *itNode) getFlag() flagType {
	return node.flag
}

func (node *itNode) getCodeLocation() types.CodeLocation {
	return node.codeLocation
}
