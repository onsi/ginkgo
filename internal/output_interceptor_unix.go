// +build freebsd openbsd netbsd dragonfly darwin linux solaris

package internal

import (
	"bytes"
	"io"
	"os"

	"golang.org/x/sys/unix"
)

func NewOutputInterceptor() OutputInterceptor {
	return &dupSyscallOutputInterceptor{
		interceptedContent: make(chan string),
		pipeChannel:        make(chan pipePair),
		shutdown:           make(chan interface{}),
	}
}

type dupSyscallOutputInterceptor struct {
	intercepting bool

	stdoutClone *os.File
	stderrClone *os.File
	pipe        pipePair

	shutdown           chan interface{}
	pipeChannel        chan pipePair
	interceptedContent chan string
}

func (interceptor *dupSyscallOutputInterceptor) StartInterceptingOutput() {
	interceptor.StartInterceptingOutputAndForwardTo(io.Discard)
}

func (interceptor *dupSyscallOutputInterceptor) StartInterceptingOutputAndForwardTo(w io.Writer) {
	if interceptor.intercepting {
		return
	}
	interceptor.intercepting = true

	if interceptor.stdoutClone == nil {
		// First, we create two clone file descriptors that point to the stdout and stderr file descriptions
		stdoutCloneFD, _ := unix.Dup(1)
		stderrCloneFD, _ := unix.Dup(2)
		// And we wrap the clone file descriptors in files so we can write to them if need be (e.g. to emit output to the console evne though we're intercepting output)
		interceptor.stdoutClone, interceptor.stderrClone = os.NewFile(uintptr(stdoutCloneFD), "stdout-clone"), os.NewFile(uintptr(stderrCloneFD), "stderr-clone")

		interceptor.shutdown = make(chan interface{})
		go startPipeFactory(interceptor.pipeChannel, interceptor.shutdown)
	}

	// Now we make a pipe, we'll use this to redirect the input to the 1 and 2 file descriptors (this is how everything else in the world is tring to log to stdout and stderr)
	// we get the pipe from our pipe factory.  it runs in the background so we can request the next pipe while the spec being intercepted is running
	interceptor.pipe = <-interceptor.pipeChannel

	//Spin up a goroutine to copy data from the pipe into a buffer, this is how we capture any output the user is emitting
	go func() {
		buffer := &bytes.Buffer{}
		destination := io.MultiWriter(buffer, w)
		io.Copy(destination, interceptor.pipe.reader)
		interceptor.interceptedContent <- buffer.String()
	}()

	// And now we call Dup2 (possibly Dup3 on some architectures) to have file descriptors 1 and 2 point to the same file description as the pipeWriter
	// This effectively shunts data written to stdout and stderr to the write end of our pipe
	unix.Dup2(int(interceptor.pipe.writer.Fd()), 1)
	unix.Dup2(int(interceptor.pipe.writer.Fd()), 2)
}

func (interceptor *dupSyscallOutputInterceptor) StopInterceptingAndReturnOutput() string {
	if !interceptor.intercepting {
		return ""
	}

	// first we have to close the write end of the pipe.  To do this we have to close all file descriptors pointing
	// to the write end.  So that would be the pipewriter itself, FD #1 and FD #2.
	interceptor.pipe.writer.Close() // the pipewriter itself
	// we also need to stop intercepting. we do that by reconnecting the stdout and stderr file descriptions back to their respective #1 and #2 file descriptors;
	// this also closes #1 and #2 before it points that their original stdout and stderr file descriptions
	unix.Dup2(int(interceptor.stdoutClone.Fd()), 1)
	unix.Dup2(int(interceptor.stderrClone.Fd()), 2)

	content := <-interceptor.interceptedContent // now wait for the goroutine to notice and return content
	interceptor.pipe.reader.Close()             //and now close the read end of the pipe so we don't leak a file descriptor

	interceptor.intercepting = false

	return content
}

func (interceptor *dupSyscallOutputInterceptor) Shutdown() {
	interceptor.StopInterceptingAndReturnOutput()

	if interceptor.stdoutClone != nil {
		close(interceptor.shutdown)
		// and now we're done with the clone file descriptors, we can close them to clean up after ourselves
		interceptor.stdoutClone.Close()
		interceptor.stderrClone.Close()
		interceptor.stdoutClone = nil
		interceptor.stderrClone = nil
	}
}
