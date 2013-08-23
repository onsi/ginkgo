package godescribe

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

func generateCodeLocation(skip int) (CodeLocation, bool) {
	_, file, line, ok := runtime.Caller(skip)
	fullStackTrace := string(debug.Stack())
	return CodeLocation{FileName: file, LineNumber: line, FullStackTrace: fullStackTrace}, ok
}

func (codeLocation CodeLocation) String() string {
	return fmt.Sprintf("%s:%d", codeLocation.FileName, codeLocation.LineNumber)
}
