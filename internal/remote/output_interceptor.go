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
	redirectFile *os.File
	streamTarget *os.File
	intercepting bool
	tailer       io.Reader
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

	// Call a function in ./syscall_dup_*.go
	// If building for plan9, use Dup. If building for Windows, use SetStdHandle. If building everything
	// other than linux_arm64 or plan9 or Windows, use a "normal" syscall.Dup2(oldfd, newfd) call.
	// If building for linux_arm64 (which doesn't have syscall.Dup2), call syscall.Dup3(oldfd, newfd, 0).
	// They are nearly identical, see: http://linux.die.net/man/2/dup3
	if err := syscallDup(int(interceptor.redirectFile.Fd()), 1); err != nil {
		os.Remove(interceptor.redirectFile.Name())
		return err
	}
	if err := syscallDup(int(interceptor.redirectFile.Fd()), 2); err != nil {
		os.Remove(interceptor.redirectFile.Name())
		return err
	}

	if interceptor.streamTarget != nil {
		interceptor.tailer = io.TeeReader(interceptor.redirectFile, interceptor.streamTarget)
	}

	return nil
}

func (interceptor *outputInterceptor) StopInterceptingAndReturnOutput() (string, error) {
	if !interceptor.intercepting {
		return "", errors.New("Not intercepting output!")
	}

	interceptor.redirectFile.Close()
	output, err := ioutil.ReadFile(interceptor.redirectFile.Name())
	//os.Remove(interceptor.redirectFile.Name())

	interceptor.intercepting = false

	if interceptor.streamTarget != nil {
		// reading the redirectFile causes the io.TeeReader to write to streamTarget,
		// so we just need to sync.
		er := interceptor.streamTarget.Sync()
		if er != nil {
			err = er
		}
	}

	return string(output), err
}

func (interceptor *outputInterceptor) StreamTo(out *os.File) {
	interceptor.streamTarget = out
}
