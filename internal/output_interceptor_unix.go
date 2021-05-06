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
	}
}

type dupSyscallOutputInterceptor struct {
	intercepting bool

	stdoutClone int
	stderrClone int

	interceptingWriter io.Closer
	interceptedContent chan string
}

func (interceptor *dupSyscallOutputInterceptor) StartInterceptingOutput() {
	if interceptor.intercepting {
		return
	}
	interceptor.intercepting = true

	interceptor.stdoutClone, _ = unix.Dup(1)
	interceptor.stderrClone, _ = unix.Dup(2)

	reader, writer, _ := os.Pipe()
	interceptor.interceptingWriter = writer

	go func() {
		buffer := &bytes.Buffer{}
		io.Copy(buffer, reader)
		interceptor.interceptedContent <- buffer.String()
	}()

	// This might call Dup3 if the dup2 syscall is not available, e.g. on
	// linux/arm64 or linux/riscv64
	unix.Dup2(int(writer.Fd()), 1)
	unix.Dup2(int(writer.Fd()), 2)
}

func (interceptor *dupSyscallOutputInterceptor) StopInterceptingAndReturnOutput() string {
	if !interceptor.intercepting {
		return ""
	}

	unix.Dup2(interceptor.stdoutClone, 1)
	unix.Dup2(interceptor.stderrClone, 2)
	unix.Close(interceptor.stdoutClone)
	unix.Close(interceptor.stderrClone)

	interceptor.interceptingWriter.Close()
	content := <-interceptor.interceptedContent

	interceptor.intercepting = false

	return content
}
