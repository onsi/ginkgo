package failer

import (
	"github.com/onsi/ginkgo/types"
	"sync"
)

type Failer struct {
	lock    *sync.Mutex
	failure types.ExampleFailure
	state   types.ExampleState
}

func New() *Failer {
	return &Failer{
		lock:  &sync.Mutex{},
		state: types.ExampleStatePassed,
	}
}

func (f *Failer) Panic(location types.CodeLocation, forwardedPanic interface{}) {
	f.lock.Lock()
	defer f.lock.Unlock()

	if f.state == types.ExampleStatePassed {
		f.state = types.ExampleStatePanicked
		f.failure = types.ExampleFailure{
			Message:        "Test Panicked",
			Location:       location,
			ForwardedPanic: forwardedPanic,
		}
	}
}

func (f *Failer) Timeout(location types.CodeLocation) {
	f.lock.Lock()
	defer f.lock.Unlock()

	if f.state == types.ExampleStatePassed {
		f.state = types.ExampleStateTimedOut
		f.failure = types.ExampleFailure{
			Message:  "Timed out",
			Location: location,
		}
	}
}

func (f *Failer) Fail(message string, location types.CodeLocation) {
	f.lock.Lock()
	defer f.lock.Unlock()

	if f.state == types.ExampleStatePassed {
		f.state = types.ExampleStateFailed
		f.failure = types.ExampleFailure{
			Message:  message,
			Location: location,
		}
	}
}

func (f *Failer) Drain(componentType types.ExampleComponentType, componentIndex int, componentCodeLocation types.CodeLocation) (types.ExampleFailure, types.ExampleState) {
	f.lock.Lock()
	defer f.lock.Unlock()

	failure := f.failure
	outcome := f.state
	if outcome != types.ExampleStatePassed {
		failure.ComponentType = componentType
		failure.ComponentIndex = componentIndex
		failure.ComponentCodeLocation = componentCodeLocation
	}

	f.state = types.ExampleStatePassed
	f.failure = types.ExampleFailure{}

	return failure, outcome
}
