// +build windows

package remote

import (
	"errors"

	"golang.org/x/sys/windows"
)

func syscallDup(oldfd int, newfd int) (err error) {
	var stdfd uint32
	switch newfd {
	case 1:
		stdfd = windows.STD_OUTPUT_HANDLE
	case 2:
		stdfd = windows.STD_ERROR_HANDLE
	default:
		return errors.New("unrecognized newfd: %d", newfd)
	}
	return windows.SetStdHandle(stdfd, windows.Handle(oldfd))
}
