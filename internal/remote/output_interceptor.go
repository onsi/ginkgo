package remote

import (
	"bytes"
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
	return &outputInterceptor{buffer: &bytes.Buffer{}}
}

type outputInterceptor struct {
	origStdout     *os.File
	origStderr     *os.File
	readSources    [2]io.ReadCloser // stores the reader pipes form os.Pipe() for closure.
	streamTarget   *os.File
	combinedReader io.Reader
	intercepting   bool
	tailer         io.Reader
	buffer         *bytes.Buffer
}

func (interceptor *outputInterceptor) StartInterceptingOutput() error {
	if interceptor.intercepting {
		return errors.New("Already intercepting output!")
	}
	interceptor.origStdout = os.Stdout
	interceptor.origStderr = os.Stderr
	stdoutRead, stdoutWrite, err := os.Pipe()
	if err != nil {
		return err
	}
	stderrRead, stderrWrite, err := os.Pipe()
	if err != nil {
		return err
	}

	os.Stdout = stdoutWrite
	os.Stderr = stderrWrite

	interceptor.intercepting = true

	interceptor.readSources = [2]io.ReadCloser{stderrRead, stderrRead}
	interceptor.combinedReader = io.TeeReader(io.MultiReader(stderrRead, stdoutRead), interceptor.buffer)

	// if interceptor.streamTarget != nil {
	// 	interceptor.tailer = io.TeeReader(interceptor.combinedReader, interceptor.streamTarget)
	// }

	return nil
}

func (interceptor *outputInterceptor) StopInterceptingAndReturnOutput() (string, error) {
	if !interceptor.intercepting {
		return "", errors.New("Not intercepting output!")
	}

	output, err := ioutil.ReadAll(interceptor.buffer)

	currStdout := os.Stdout
	currStdout.Close()
	currStderr := os.Stderr
	currStderr.Close()

	os.Stdout = interceptor.origStdout
	os.Stderr = interceptor.origStderr

	for _, r := range interceptor.readSources {
		if closerErr := r.Close(); closerErr != nil {
			err = closerErr
		}
	}

	interceptor.intercepting = false

	// if interceptor.streamTarget != nil {
	// 	interceptor.streamTarget.Sync()
	// }

	return string(output), err
}

func (interceptor *outputInterceptor) StreamTo(out *os.File) {
	interceptor.streamTarget = out
}
