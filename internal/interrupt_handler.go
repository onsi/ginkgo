package internal

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type InterruptStatus struct {
	Interrupted bool
	Channel     chan interface{}
}

type InterruptHandlerInterface interface {
	Status() InterruptStatus
}

type InterruptHandler struct {
	c           chan interface{}
	lock        *sync.Mutex
	interrupted bool
}

func NewInterruptHandler() *InterruptHandler {
	handler := &InterruptHandler{
		c:           make(chan interface{}),
		lock:        &sync.Mutex{},
		interrupted: false,
	}
	handler.registerForInterrupts()
	return handler
}

func (handler *InterruptHandler) registerForInterrupts() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		for {
			<-c
			handler.lock.Lock()
			handler.interrupted = true
			close(handler.c)
			handler.c = make(chan interface{})
			handler.lock.Unlock()
		}
	}()
}

func (handler *InterruptHandler) Status() InterruptStatus {
	handler.lock.Lock()
	defer handler.lock.Unlock()

	return InterruptStatus{
		Interrupted: handler.interrupted,
		Channel:     handler.c,
	}
}
