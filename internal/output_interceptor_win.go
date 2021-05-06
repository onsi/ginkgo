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

	interceptingWriter io.Closer
	interceptedContent chan string
}

func (interceptor *osGlobalReassigningOutputInterceptor) StartInterceptingOutput() {
	if interceptor.intercepting {
		return
	}
	interceptor.intercepting = true

	interceptor.originalStdout = os.Stdout
	interceptor.originalStderr = os.Stderr

	reader, writer, _ := os.Pipe()
	interceptor.interceptingWriter = writer

	go func() {
		buffer := &bytes.Buffer{}
		io.Copy(buffer, reader)
		interceptor.interceptedContent <- buffer.String()
	}()

	os.Stdout = writer
	os.Stderr = writer
}

func (interceptor *osGlobalReassigningOutputInterceptor) StopInterceptingAndReturnOutput() string {
	if !interceptor.intercepting {
		return ""
	}

	os.Stdout = interceptor.originalStdout
	os.Stderr = interceptor.originalStderr

	interceptor.interceptingWriter.Close()
	content := <-interceptor.interceptedContent

	interceptor.intercepting = false

	return content
}
