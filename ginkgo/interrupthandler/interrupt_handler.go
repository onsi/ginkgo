package interrupthandler

import (
	"os"
	"os/signal"
	"syscall"
)

type InterruptHandler struct {
	c chan interface{}
}

func NewInterruptHandler() *InterruptHandler {
	handler := &InterruptHandler{
		c: make(chan interface{}),
	}

	handler.registerForInterrupts()
	SwallowSigQuit()

	return handler
}

func (handler *InterruptHandler) registerForInterrupts() {
	didRegister := make(chan struct{})
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		close(didRegister)

		<-c

		close(handler.c)
	}()
	<-didRegister
}

func (handler *InterruptHandler) WasInterrupted() bool {
	select {
	case <-handler.c:
		return true
	default:
		return false
	}
}

func (handler *InterruptHandler) InterruptChannel() chan interface{} {
	return handler.c
}
