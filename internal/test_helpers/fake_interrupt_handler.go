package test_helpers

import (
	"sync"

	"github.com/onsi/ginkgo/v2/internal/interrupt_handler"
)

type FakeInterruptHandler struct {
	triggerInterrupt chan bool

	c                                  chan interface{}
	stop                               chan interface{}
	lock                               *sync.Mutex
	interrupted                        bool
	cause                              interrupt_handler.InterruptCause
	interruptPlaceholderMessage        string
	emittedInterruptPlaceholderMessage string
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
			handler.emittedInterruptPlaceholderMessage = handler.interruptPlaceholderMessage
			close(handler.c)
			handler.c = make(chan interface{})
			handler.lock.Unlock()
		}
	}()
}

func (handler *FakeInterruptHandler) Interrupt(cause interrupt_handler.InterruptCause) {
	handler.lock.Lock()
	handler.cause = cause
	handler.lock.Unlock()

	handler.triggerInterrupt <- true
}

func (handler *FakeInterruptHandler) Status() interrupt_handler.InterruptStatus {
	handler.lock.Lock()
	defer handler.lock.Unlock()

	return interrupt_handler.InterruptStatus{
		Interrupted: handler.interrupted,
		Channel:     handler.c,
		Cause:       handler.cause,
	}
}

func (handler *FakeInterruptHandler) SetInterruptPlaceholderMessage(message string) {
	handler.lock.Lock()
	defer handler.lock.Unlock()

	handler.interruptPlaceholderMessage = message
}

func (handler *FakeInterruptHandler) ClearInterruptPlaceholderMessage() {
	handler.lock.Lock()
	defer handler.lock.Unlock()

	handler.interruptPlaceholderMessage = ""
}

func (handler *FakeInterruptHandler) EmittedInterruptPlaceholderMessage() string {
	handler.lock.Lock()
	defer handler.lock.Unlock()
	return handler.emittedInterruptPlaceholderMessage
}

func (handler *FakeInterruptHandler) InterruptMessageWithStackTraces() string {
	handler.lock.Lock()
	defer handler.lock.Unlock()

	return handler.cause.String() + "\nstack trace"
}
