// +build !linux !arm64
// +build !windows
// +build !solaris
// +build !plan9

package remote

import "syscall"

func syscallDup(oldfd int, newfd int) (err error) {
	return syscall.Dup2(oldfd, newfd)
}
