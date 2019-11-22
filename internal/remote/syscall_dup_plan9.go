// +build plan9

package remote

import (
	"syscall"
)

// Plan9 doesn't have syscall.Dup2 which ginkgo uses, so
// use the identical syscall.Dup instead
func syscallDup(oldfd int, newfd int) (err error) {
	_, err := syscall.Dup(oldfd, newfd, 0)
	return err
}
