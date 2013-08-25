package ginkgo

import (
	"fmt"
	"runtime"
	"runtime/debug"
)

type CodeLocation struct {
	FileName       string
	LineNumber     int
	FullStackTrace string
}

func generateCodeLocation(skip int) CodeLocation {
	_, file, line, _ := runtime.Caller(skip + 1)
	fullStackTrace := string(debug.Stack())
	return CodeLocation{FileName: file, LineNumber: line, FullStackTrace: fullStackTrace}
}

func (codeLocation CodeLocation) String() string {
	return fmt.Sprintf("%s:%d", codeLocation.FileName, codeLocation.LineNumber)
}
