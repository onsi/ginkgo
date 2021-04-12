package test_helpers

type FakeOutputInterceptor struct {
	intercepting      bool
	InterceptedOutput string
}

func (interceptor *FakeOutputInterceptor) StartInterceptingOutput() {
	interceptor.intercepting = true
}

func (interceptor *FakeOutputInterceptor) StopInterceptingAndReturnOutput() string {
	if interceptor.intercepting {
		interceptor.intercepting = false
		return interceptor.InterceptedOutput
	}

	return ""
}
