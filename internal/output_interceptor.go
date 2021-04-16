package internal

/*
The OutputInterceptor is used by to
intercept and capture all stdin and stderr output during a test run.
*/
type OutputInterceptor interface {
	StartInterceptingOutput()
	StopInterceptingAndReturnOutput() string
}

type NoopOutputInterceptor struct{}

func (interceptor NoopOutputInterceptor) StartInterceptingOutput()                {}
func (interceptor NoopOutputInterceptor) StopInterceptingAndReturnOutput() string { return "" }
