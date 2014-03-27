package internal

import (
	"fmt"
	"github.com/onsi/ginkgo/internal/codelocation"
	"github.com/onsi/ginkgo/types"
	"reflect"
	"sync"
	"time"
)

type runnableNode struct {
	isAsync          bool
	asyncFunc        func(chan<- interface{})
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
		if !(bodyType.In(0).Kind() == reflect.Chan && bodyType.In(0).Elem().Kind() == reflect.Interface) {
			panic(fmt.Sprintf("Must pass a Done channel to function at %v", codeLocation))
		}

		wrappedBody := func(done chan<- interface{}) {
			bodyValue := reflect.ValueOf(body)
			bodyValue.Call([]reflect.Value{reflect.ValueOf(done)})
		}

		return &runnableNode{
			isAsync:          true,
			asyncFunc:        wrappedBody,
			syncFunc:         nil,
			codeLocation:     codeLocation,
			timeoutThreshold: timeout,
		}
	}

	panic(fmt.Sprintf("Too many arguments to function at %v", codeLocation))
}

func (runnable *runnableNode) run() (outcome runOutcome, failure failureData) {
	done := make(chan interface{}, 1)
	lock := &sync.Mutex{}

	panicRecovery := func() {
		if e := recover(); e != nil {
			lock.Lock()
			outcome = runOutcomePanicked
			failure = failureData{
				message:        "Test Panicked",
				codeLocation:   codelocation.New(2),
				forwardedPanic: e,
			}
			lock.Unlock()
			select {
			case <-done:
				break
			default:
				close(done)
			}
		}
	}

	defer panicRecovery()

	if runnable.isAsync {
		go func() {
			defer panicRecovery()
			runnable.asyncFunc(done)
		}()
	} else {
		runnable.syncFunc()
		close(done)
	}

	select {
	case <-done:
		lock.Lock()
		if outcome != runOutcomePanicked {
			outcome = runOutcomeCompleted
		}
		lock.Unlock()
	case <-time.After(runnable.timeoutThreshold):
		lock.Lock()
		if outcome != runOutcomePanicked {
			outcome = runOutcomeTimedOut
			failure = failureData{
				message:      "Timed out",
				codeLocation: runnable.codeLocation,
			}
		}
		lock.Unlock()
	}

	return
}

//It Node

type itNode struct {
	*runnableNode

	flag FlagType
	text string
}

func newItNode(text string, body interface{}, flag FlagType, codeLocation types.CodeLocation, timeout time.Duration) *itNode {
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

func (node *itNode) getFlag() FlagType {
	return node.flag
}

func (node *itNode) getCodeLocation() types.CodeLocation {
	return node.codeLocation
}
