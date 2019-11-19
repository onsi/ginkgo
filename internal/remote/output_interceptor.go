package remote

import (
	"errors"
	"io"
	"io/ioutil"
	"os"
)

/*
The OutputInterceptor is used by the ForwardingReporter to
intercept and capture all stdin and stderr output during a test run.
*/
type OutputInterceptor interface {
	StartInterceptingOutput() error
	StopInterceptingAndReturnOutput() (string, error)
	StreamTo(*os.File)
}

func NewOutputInterceptor() OutputInterceptor {
	return &outputInterceptor{}
}

type outputInterceptor struct {
	readPipe     *os.File
	writePipe    *os.File
	streamTarget *os.File
	intercepting bool
	tailer       io.Reader
}

func (interceptor *outputInterceptor) StartInterceptingOutput() error {
	if interceptor.intercepting {
		return errors.New("Already intercepting output!")
	}
	interceptor.intercepting = true

	if interceptor.streamTarget != nil {
		r, w, err := os.Pipe()
		if err != nil {
			return err
		}
		interceptor.readPipe = r
		interceptor.writePipe = w
		interceptor.tailer = io.TeeReader(r, w)
	}

	return nil
}

func (interceptor *outputInterceptor) StopInterceptingAndReturnOutput() (string, error) {
	if !interceptor.intercepting {
		return "", errors.New("Not intercepting output!")
	}

	output, err := ioutil.ReadAll(interceptor.tailer)
	interceptor.writePipe.Close()
	interceptor.readPipe.Close()

	interceptor.intercepting = false

	if interceptor.streamTarget != nil {
		interceptor.streamTarget.Sync()
	}

	return string(output), err
}

func (interceptor *outputInterceptor) StreamTo(out *os.File) {
	interceptor.streamTarget = out
}
