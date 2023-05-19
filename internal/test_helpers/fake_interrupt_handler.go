package test_helpers

import (
	"sync"

	"github.com/onsi/ginkgo/v2/internal/interrupt_handler"
)

type FakeInterruptHandler struct {
	c                                  chan any
	lock                               *sync.Mutex
	level                              interrupt_handler.InterruptLevel
	cause                              interrupt_handler.InterruptCause
	interruptPlaceholderMessage        string
	emittedInterruptPlaceholderMessage string
}

func NewFakeInterruptHandler() *FakeInterruptHandler {
	handler := &FakeInterruptHandler{
		c:     make(chan any),
		lock:  &sync.Mutex{},
		level: interrupt_handler.InterruptLevelUninterrupted,
	}
	return handler
}

func (handler *FakeInterruptHandler) Interrupt(cause interrupt_handler.InterruptCause) {
	handler.lock.Lock()
	handler.cause = cause
	handler.level += 1
	if handler.level > interrupt_handler.InterruptLevelBailOut {
		handler.level = interrupt_handler.InterruptLevelBailOut
	} else {
		close(handler.c)
		handler.c = make(chan any)
	}
	handler.lock.Unlock()
}

func (handler *FakeInterruptHandler) Status() interrupt_handler.InterruptStatus {
	handler.lock.Lock()
	defer handler.lock.Unlock()

	return interrupt_handler.InterruptStatus{
		Channel: handler.c,
		Level:   handler.level,
		Cause:   handler.cause,
	}
}
