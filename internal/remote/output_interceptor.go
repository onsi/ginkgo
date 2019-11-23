package remote

import (
	"bufio"
	"errors"
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

type tailing struct {
	src         *os.File
	dest        *os.File
	doneTailing chan bool
}

type outputInterceptor struct {
	redirectFile *os.File
	intercepting bool
	tail         *tailing
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

	defer func() {
		if err != nil {
			// in all of our scenarios, if we're exiting with an error
			// the redirectFile shouldn't stay.
			os.Remove(interceptor.redirectFile.Name())
		}
	}()
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

	if interceptor.tail != nil {
		interceptor.tail.src, err = os.Open(interceptor.redirectFile.Name())
		if err != nil {
			return err
		}
		scanner := bufio.NewScanner(interceptor.tail.src)
		interceptor.tail.doneTailing = make(chan bool)
		go func() {
			for {
				select {
				case <-interceptor.tail.doneTailing:
					// drain the scanner into the streamed-to file
					for scanner.Scan() {
						interceptor.tail.dest.WriteString(scanner.Text() + "\n")
					}
					interceptor.tail.src.Close()
					return
				default:
					if scanner.Scan() {
						interceptor.tail.dest.WriteString(scanner.Text() + "\n")
					}
				}
			}
		}()
	}

	return nil
}

func (interceptor *outputInterceptor) StopInterceptingAndReturnOutput() (string, error) {
	if !interceptor.intercepting {
		return "", errors.New("Not intercepting output!")
	}

	interceptor.redirectFile.Close()
	output, err := ioutil.ReadFile(interceptor.redirectFile.Name())
	os.Remove(interceptor.redirectFile.Name())

	interceptor.intercepting = false

	if interceptor.tail != nil {
		// reading the redirectFile causes the io.TeeReader to write to streamTarget,
		// so we just need to sync.
		close(interceptor.tail.doneTailing)
		er := interceptor.tail.dest.Sync()
		if er != nil {
			err = er
		}
	}

	return string(output), err
}

func (interceptor *outputInterceptor) StreamTo(out *os.File) {
	interceptor.tail = &tailing{
		dest: out,
	}
}
