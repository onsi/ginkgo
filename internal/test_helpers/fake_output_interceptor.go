package test_helpers

import "sync"

type FakeOutputInterceptor struct {
	intercepting      bool
	InterceptedOutput string
	lock              *sync.Mutex
}

func NewFakeOutputInterceptor() *FakeOutputInterceptor {
	return &FakeOutputInterceptor{
		lock: &sync.Mutex{},
	}
}

func (interceptor *FakeOutputInterceptor) StartInterceptingOutput() {
	interceptor.lock.Lock()
	defer interceptor.lock.Unlock()
	interceptor.intercepting = true
	interceptor.InterceptedOutput = ""
}

func (interceptor *FakeOutputInterceptor) StopInterceptingAndReturnOutput() string {
	interceptor.lock.Lock()
	defer interceptor.lock.Unlock()
	interceptor.intercepting = false
	return interceptor.InterceptedOutput
}
