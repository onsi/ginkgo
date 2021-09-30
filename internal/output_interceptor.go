package internal

import (
	"bytes"
	"io"
	"os"
)

/*
The OutputInterceptor is used by to
intercept and capture all stdin and stderr output during a test run.
*/
type OutputInterceptor interface {
	StartInterceptingOutput()
	StartInterceptingOutputAndForwardTo(io.Writer)
	StopInterceptingAndReturnOutput() string
}

type NoopOutputInterceptor struct{}

func (interceptor NoopOutputInterceptor) StartInterceptingOutput()                      {}
func (interceptor NoopOutputInterceptor) StartInterceptingOutputAndForwardTo(io.Writer) {}
func (interceptor NoopOutputInterceptor) StopInterceptingAndReturnOutput() string       { return "" }

/* This is used on windows builds but included here so it can be explicitly tested on unix systems too */

func NewOSGlobalReassigningOutputInterceptor() OutputInterceptor {
	return &OSGlobalReassigningOutputInterceptor{
		interceptedContent: make(chan string),
	}
}

type OSGlobalReassigningOutputInterceptor struct {
	intercepting bool

	originalStdout *os.File
	originalStderr *os.File
	pipeWriter     *os.File
	pipeReader     *os.File

	interceptedContent chan string
}

func (interceptor *OSGlobalReassigningOutputInterceptor) StartInterceptingOutput() {
	interceptor.StartInterceptingOutputAndForwardTo(io.Discard)
}

func (interceptor *OSGlobalReassigningOutputInterceptor) StartInterceptingOutputAndForwardTo(w io.Writer) {
	if interceptor.intercepting {
		return
	}
	interceptor.intercepting = true

	interceptor.originalStdout = os.Stdout
	interceptor.originalStderr = os.Stderr

	interceptor.pipeReader, interceptor.pipeWriter, _ = os.Pipe()

	go func() {
		buffer := &bytes.Buffer{}
		destination := io.MultiWriter(buffer, w)
		io.Copy(destination, interceptor.pipeReader)
		interceptor.interceptedContent <- buffer.String()
	}()

	os.Stdout = interceptor.pipeWriter
	os.Stderr = interceptor.pipeWriter
}

func (interceptor *OSGlobalReassigningOutputInterceptor) StopInterceptingAndReturnOutput() string {
	if !interceptor.intercepting {
		return ""
	}

	os.Stdout = interceptor.originalStdout
	os.Stderr = interceptor.originalStderr

	interceptor.pipeWriter.Close()
	content := <-interceptor.interceptedContent
	interceptor.pipeReader.Close()

	interceptor.intercepting = false

	return content
}
