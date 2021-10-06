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
	Shutdown()
}

type NoopOutputInterceptor struct{}

func (interceptor NoopOutputInterceptor) StartInterceptingOutput()                      {}
func (interceptor NoopOutputInterceptor) StartInterceptingOutputAndForwardTo(io.Writer) {}
func (interceptor NoopOutputInterceptor) StopInterceptingAndReturnOutput() string       { return "" }
func (interceptor NoopOutputInterceptor) Shutdown()                                     {}

type pipePair struct {
	reader *os.File
	writer *os.File
}

func startPipeFactory(pipeChannel chan pipePair, shutdown chan interface{}) {
	for {
		//make the next pipe...
		pair := pipePair{}
		pair.reader, pair.writer, _ = os.Pipe()
		select {
		//...and provide it to the next consumer (they are responsible for closing the files)
		case pipeChannel <- pair:
			continue
		//...or close the files if we were told to shutdown
		case <-shutdown:
			pair.reader.Close()
			pair.writer.Close()
			return
		}
	}
}

/* This is used on windows builds but included here so it can be explicitly tested on unix systems too */

func NewOSGlobalReassigningOutputInterceptor() OutputInterceptor {
	return &OSGlobalReassigningOutputInterceptor{
		interceptedContent: make(chan string),
		pipeChannel:        make(chan pipePair),
		shutdown:           make(chan interface{}),
	}
}

type OSGlobalReassigningOutputInterceptor struct {
	intercepting bool

	stdoutClone *os.File
	stderrClone *os.File
	pipe        pipePair

	shutdown           chan interface{}
	pipeChannel        chan pipePair
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

	if interceptor.stdoutClone == nil {
		interceptor.stdoutClone = os.Stdout
		interceptor.stderrClone = os.Stderr

		interceptor.shutdown = make(chan interface{})
		go startPipeFactory(interceptor.pipeChannel, interceptor.shutdown)
	}

	interceptor.pipe = <-interceptor.pipeChannel

	go func() {
		buffer := &bytes.Buffer{}
		destination := io.MultiWriter(buffer, w)
		io.Copy(destination, interceptor.pipe.reader)
		interceptor.interceptedContent <- buffer.String()
	}()

	os.Stdout = interceptor.pipe.writer
	os.Stderr = interceptor.pipe.writer
}

func (interceptor *OSGlobalReassigningOutputInterceptor) StopInterceptingAndReturnOutput() string {
	if !interceptor.intercepting {
		return ""
	}

	os.Stdout = interceptor.stdoutClone
	os.Stderr = interceptor.stderrClone

	interceptor.pipe.writer.Close()
	content := <-interceptor.interceptedContent
	interceptor.pipe.reader.Close()

	interceptor.intercepting = false

	return content
}

func (interceptor *OSGlobalReassigningOutputInterceptor) Shutdown() {
	interceptor.StopInterceptingAndReturnOutput()
	if interceptor.stdoutClone != nil {
		close(interceptor.shutdown)
		interceptor.stdoutClone = nil
		interceptor.stderrClone = nil
	}
}
