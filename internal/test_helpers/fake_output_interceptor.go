package test_helpers

import (
	"io"
	"sync"
)

type FakeOutputInterceptor struct {
	intercepting      bool
	forwardingWriter  io.Writer
	interceptedOutput string
	lock              *sync.Mutex
}

func NewFakeOutputInterceptor() *FakeOutputInterceptor {
	return &FakeOutputInterceptor{
		lock:             &sync.Mutex{},
		forwardingWriter: io.Discard,
	}
}

func (interceptor *FakeOutputInterceptor) AppendInterceptedOutput(s string) {
	interceptor.lock.Lock()
	defer interceptor.lock.Unlock()
	interceptor.interceptedOutput += s
	interceptor.forwardingWriter.Write([]byte(s))
}

func (interceptor *FakeOutputInterceptor) StartInterceptingOutput() {
	interceptor.StartInterceptingOutputAndForwardTo(io.Discard)
}

func (interceptor *FakeOutputInterceptor) StartInterceptingOutputAndForwardTo(w io.Writer) {
	interceptor.lock.Lock()
	defer interceptor.lock.Unlock()
	interceptor.forwardingWriter = w
	interceptor.intercepting = true
	interceptor.interceptedOutput = ""
}

func (interceptor *FakeOutputInterceptor) PauseIntercepting() {
	interceptor.lock.Lock()
	defer interceptor.lock.Unlock()
	interceptor.intercepting = false
}

func (interceptor *FakeOutputInterceptor) ResumeIntercepting() {
	interceptor.lock.Lock()
	defer interceptor.lock.Unlock()
	interceptor.intercepting = true
}

func (interceptor *FakeOutputInterceptor) StopInterceptingAndReturnOutput() string {
	interceptor.lock.Lock()
	defer interceptor.lock.Unlock()
	interceptor.intercepting = false
	return interceptor.interceptedOutput
}

func (interceptor *FakeOutputInterceptor) Shutdown() {
}
