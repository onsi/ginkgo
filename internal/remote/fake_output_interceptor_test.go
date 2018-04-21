package remote_test

type fakeOutputInterceptor struct {
	DidStartInterceptingOutput bool
	DidStopInterceptingOutput  bool
	InterceptedOutput          string
	ReturnedOutput             bool
}

func (interceptor *fakeOutputInterceptor) StartInterceptingOutput() error {
	interceptor.DidStartInterceptingOutput = true
	return nil
}

func (interceptor *fakeOutputInterceptor) StopInterceptingAndReturnOutput() (string, error) {
	interceptor.DidStopInterceptingOutput = true
	return interceptor.InterceptedOutput, nil
}

func (interceptor *fakeOutputInterceptor) Output() (string, error) {
	interceptor.ReturnedOutput = true
	return interceptor.InterceptedOutput, nil
}
