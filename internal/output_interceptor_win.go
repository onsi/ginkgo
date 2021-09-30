// +build windows

package internal

import (
	"bytes"
	"io"
	"os"
)

func NewOutputInterceptor() OutputInterceptor {
	return NewOSGlobalReassigningOutputInterceptor()
}

