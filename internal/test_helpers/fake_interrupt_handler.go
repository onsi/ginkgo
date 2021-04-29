package test_helpers

import (
	"sync"

	"github.com/onsi/ginkgo/internal"
)

type FakeInterruptHandler struct {
	triggerInterrupt chan bool

	c                       chan interface{}
	stop                    chan interface{}
	lock                    *sync.Mutex
	interrupted             bool
	cause                   string
	interruptMessage        string
	emittedInterruptMessage string
}

func NewFakeInterruptHandler() *FakeInterruptHandler {
	handler := &FakeInterruptHandler{
		triggerInterrupt: make(chan bool),
		c:                make(chan interface{}),
		lock:             &sync.Mutex{},
		interrupted:      false,
		stop:             make(chan interface{}),
	}
	handler.registerForInterrupts()
	return handler
}

func (handler *FakeInterruptHandler) Stop() {
	close(handler.stop)
}

func (handler *FakeInterruptHandler) registerForInterrupts() {
	go func() {
		for {
			select {
			case <-handler.triggerInterrupt:
			case <-handler.stop:
				return
			}
			handler.lock.Lock()
			handler.interrupted = true
			handler.emittedInterruptMessage = handler.interruptMessage
			close(handler.c)
			handler.c = make(chan interface{})
			handler.lock.Unlock()
		}
	}()
}

func (handler *FakeInterruptHandler) Interrupt(cause string) {
	handler.lock.Lock()
	handler.cause = cause
	handler.lock.Unlock()

	handler.triggerInterrupt <- true
}

func (handler *FakeInterruptHandler) Status() internal.InterruptStatus {
	handler.lock.Lock()
	defer handler.lock.Unlock()

	return internal.InterruptStatus{
		Interrupted: handler.interrupted,
		Channel:     handler.c,
		Cause:       handler.cause,
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

func (handler *FakeInterruptHandler) InterruptMessageWithStackTraces() string {
	handler.lock.Lock()
	defer handler.lock.Unlock()

	return handler.cause + "\nstack trace"
}
