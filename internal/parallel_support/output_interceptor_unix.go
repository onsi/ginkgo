// +build freebsd openbsd netbsd dragonfly darwin linux solaris

package parallel_support

import (
	"errors"
	"io/ioutil"
	"os"

	"golang.org/x/sys/unix"
)

func NewOutputInterceptor() OutputInterceptor {
	return &outputInterceptor{}
}

type outputInterceptor struct {
	redirectFile *os.File
	intercepting bool
	doneTailing  chan bool

	stdoutClone int
	stderrClone int
}

func (interceptor *outputInterceptor) StartInterceptingOutput() error {
	if interceptor.intercepting {
		return errors.New("Already intercepting output!")
	}
	interceptor.intercepting = true

	var err error

	interceptor.redirectFile, err = ioutil.TempFile("", "ginkgo-output")
	if err != nil {
		return err
	}

	interceptor.stdoutClone, _ = unix.Dup(1)
	interceptor.stderrClone, _ = unix.Dup(2)

	// This might call Dup3 if the dup2 syscall is not available, e.g. on
	// linux/arm64 or linux/riscv64
	unix.Dup2(int(interceptor.redirectFile.Fd()), 1)
	unix.Dup2(int(interceptor.redirectFile.Fd()), 2)

	return nil
}

func (interceptor *outputInterceptor) StopInterceptingAndReturnOutput() (string, error) {
	if !interceptor.intercepting {
		return "", errors.New("Not intercepting output!")
	}

	interceptor.redirectFile.Close()
	output, err := ioutil.ReadFile(interceptor.redirectFile.Name())
	os.Remove(interceptor.redirectFile.Name())

	unix.Dup2(interceptor.stdoutClone, 1)
	unix.Dup2(interceptor.stderrClone, 2)

	interceptor.intercepting = false

	return string(output), err
}
