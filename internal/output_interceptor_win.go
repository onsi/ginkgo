// +build windows

package internal

import (
	"bytes"
	"io"
	"os"
)

func NewOutputInterceptor() OutputInterceptor {
	return &osGlobalReassigningOutputInterceptor{
		interceptedContent: make(chan string),
	}
}

type osGlobalReassigningOutputInterceptor struct {
	intercepting bool

	originalStdout *os.File
	originalStderr *os.File
	pipeWriter     *os.File
	pipeReader     *os.File

	interceptedContent chan string
}

func (interceptor *osGlobalReassigningOutputInterceptor) StartInterceptingOutput() {
	if interceptor.intercepting {
		return
	}
	interceptor.intercepting = true

	interceptor.originalStdout = os.Stdout
	interceptor.originalStderr = os.Stderr

	interceptor.pipeReader, interceptor.pipeWriter, _ = os.Pipe()

	go func() {
		buffer := &bytes.Buffer{}
		io.Copy(buffer, interceptor.pipeReader)
		interceptor.interceptedContent <- buffer.String()
	}()

	os.Stdout = interceptor.pipeWriter
	os.Stderr = interceptor.pipeWriter
}

func (interceptor *osGlobalReassigningOutputInterceptor) StopInterceptingAndReturnOutput() string {
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
