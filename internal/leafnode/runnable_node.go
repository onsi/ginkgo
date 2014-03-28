package leafnode

import (
	"fmt"
	"github.com/onsi/ginkgo/internal/codelocation"
	"github.com/onsi/ginkgo/types"
	"reflect"
	"sync"
	"time"
)

type runner struct {
	isAsync          bool
	asyncFunc        func(chan<- interface{})
	syncFunc         func()
	codeLocation     types.CodeLocation
	timeoutThreshold time.Duration
}

func newRunner(body interface{}, codeLocation types.CodeLocation, timeout time.Duration) *runner {
	bodyType := reflect.TypeOf(body)
	if bodyType.Kind() != reflect.Func {
		panic(fmt.Sprintf("Expected a function but got something else at %v", codeLocation))
	}

	switch bodyType.NumIn() {
	case 0:
		return &runner{
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

		return &runner{
			isAsync:          true,
			asyncFunc:        wrappedBody,
			syncFunc:         nil,
			codeLocation:     codeLocation,
			timeoutThreshold: timeout,
		}
	}

	panic(fmt.Sprintf("Too many arguments to function at %v", codeLocation))
}

func (r *runner) run() (outcome types.ExampleState, failure types.ExampleFailure) {
	done := make(chan interface{}, 1)
	lock := &sync.Mutex{}

	panicRecovery := func() {
		if e := recover(); e != nil {
			lock.Lock()
			outcome = types.ExampleStatePanicked
			failure = types.ExampleFailure{
				Message:        "Test Panicked",
				Location:       codelocation.New(2),
				ForwardedPanic: e,
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

	if r.isAsync {
		go func() {
			defer panicRecovery()
			r.asyncFunc(done)
		}()
	} else {
		r.syncFunc()
		close(done)
	}

	select {
	case <-done:
		lock.Lock()
		if outcome != types.ExampleStatePanicked {
			outcome = types.ExampleStatePassed
		}
		lock.Unlock()
	case <-time.After(r.timeoutThreshold):
		lock.Lock()
		if outcome != types.ExampleStatePanicked {
			outcome = types.ExampleStateTimedOut
			failure = types.ExampleFailure{
				Message:  "Timed out",
				Location: r.codeLocation,
			}
		}
		lock.Unlock()
	}

	return
}
