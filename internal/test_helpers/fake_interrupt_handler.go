package test_helpers

import (
	"sync"

	"github.com/onsi/ginkgo/internal"
)

type FakeInterruptHandler struct {
	triggerInterrupt chan bool

	c                       chan interface{}
	lock                    *sync.Mutex
	interrupted             bool
	interruptMessage        string
	emittedInterruptMessage string
}

func NewFakeInterruptHandler() *FakeInterruptHandler {
	handler := &FakeInterruptHandler{
		triggerInterrupt: make(chan bool),
		c:                make(chan interface{}),
		lock:             &sync.Mutex{},
		interrupted:      false,
	}
	handler.registerForInterrupts()
	return handler
}

func (handler *FakeInterruptHandler) registerForInterrupts() {
	go func() {
		for {
			<-handler.triggerInterrupt
			handler.lock.Lock()
			handler.interrupted = true
			handler.emittedInterruptMessage = handler.interruptMessage
			close(handler.c)
			handler.c = make(chan interface{})
			handler.lock.Unlock()
		}
	}()
}

func (handler *FakeInterruptHandler) Interrupt() {
	handler.triggerInterrupt <- true
}

func (handler *FakeInterruptHandler) Status() internal.InterruptStatus {
	handler.lock.Lock()
	defer handler.lock.Unlock()

	return internal.InterruptStatus{
		Interrupted: handler.interrupted,
		Channel:     handler.c,
	}
}

func (handler *FakeInterruptHandler) SetInterruptMessage(message string) {
	handler.lock.Lock()
	defer handler.lock.Unlock()

	handler.interruptMessage = message
}

func (handler *FakeInterruptHandler) ClearInterruptMessage() {
	handler.lock.Lock()
	defer handler.lock.Unlock()

	handler.interruptMessage = ""
}

func (handler *FakeInterruptHandler) EmittedInterruptMessage() string {
	handler.lock.Lock()
	defer handler.lock.Unlock()
	return handler.emittedInterruptMessage
}
