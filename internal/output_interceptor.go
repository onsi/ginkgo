package internal

/*
The OutputInterceptor is used by to
intercept and capture all stdin and stderr output during a test run.
*/
type OutputInterceptor interface {
	StartInterceptingOutput()
	StopInterceptingAndReturnOutput() string
}

func NewNoopOutputInterceptor() OutputInterceptor {
	return noopOutputInterceptor{}
}

type noopOutputInterceptor struct{}

func (interceptor noopOutputInterceptor) StartInterceptingOutput()                {}
func (interceptor noopOutputInterceptor) StopInterceptingAndReturnOutput() string { return "" }
