package internal

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/onsi/ginkgo/formatter"
)

const TIMEOUT_REPEAT_INTERRUPT_MAXIMUM_DURATION = 30 * time.Second
const TIMEOUT_REPEAT_INTERRUPT_FRACTION_OF_TIMEOUT = 10

type InterruptStatus struct {
	Interrupted bool
	Channel     chan interface{}
	Cause       string
}

type InterruptHandlerInterface interface {
	Status() InterruptStatus
	SetInterruptMessage(string)
	ClearInterruptMessage()
	InterruptMessageWithStackTraces() string
}

type InterruptHandler struct {
	c                chan interface{}
	lock             *sync.Mutex
	interrupted      bool
	interruptMessage string
	interruptCause   string
	stop             chan interface{}
}

func NewInterruptHandler(timeout time.Duration) *InterruptHandler {
	handler := &InterruptHandler{
		c:           make(chan interface{}),
		lock:        &sync.Mutex{},
		interrupted: false,
		stop:        make(chan interface{}),
	}
	handler.registerForInterrupts(timeout)
	return handler
}

func (handler *InterruptHandler) Stop() {
	close(handler.stop)
}

func (handler *InterruptHandler) registerForInterrupts(timeout time.Duration) {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	var timeoutChannel <-chan time.Time
	var timer *time.Timer
	if timeout > 0 {
		timer = time.NewTimer(timeout)
		timeoutChannel = timer.C
	}
	go func() {
		for {
			var interruptCause string
			select {
			case <-signalChannel:
				interruptCause = "Interrupted by User"
			case <-timeoutChannel:
				interruptCause = "Interrupted by Timeout"
				repeatInterruptTimeout := timeout / time.Duration(TIMEOUT_REPEAT_INTERRUPT_FRACTION_OF_TIMEOUT)
				if repeatInterruptTimeout > TIMEOUT_REPEAT_INTERRUPT_MAXIMUM_DURATION {
					repeatInterruptTimeout = TIMEOUT_REPEAT_INTERRUPT_MAXIMUM_DURATION
				}
				timer = time.NewTimer(repeatInterruptTimeout)
				timeoutChannel = timer.C
			case <-handler.stop:
				if timer != nil {
					timer.Stop()
				}
				signal.Stop(signalChannel)
				return
			}
			handler.lock.Lock()
			handler.interruptCause = interruptCause
			if handler.interruptMessage != "" {
				fmt.Println(handler.interruptMessage)
			}
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
		Cause:       handler.interruptCause,
	}
}

func (handler *InterruptHandler) SetInterruptMessage(message string) {
	handler.lock.Lock()
	defer handler.lock.Unlock()

	handler.interruptMessage = message
}

func (handler *InterruptHandler) ClearInterruptMessage() {
	handler.lock.Lock()
	defer handler.lock.Unlock()

	handler.interruptMessage = ""
}

func (handler *InterruptHandler) InterruptMessageWithStackTraces() string {
	handler.lock.Lock()
	out := fmt.Sprintf("%s\n\n", handler.interruptCause)
	defer handler.lock.Unlock()
	out += "Here's a stack trace of all running goroutines:\n"
	buf := make([]byte, 8192)
	for {
		n := runtime.Stack(buf, true)
		if n < len(buf) {
			buf = buf[:n]
			break
		}
		buf = make([]byte, 2*len(buf))
	}
	out += formatter.Fi(1, "%s", string(buf))
	return out
}
