package godescribe

import (
	"fmt"
	"runtime"
)

type CodeLocation struct {
	FileName   string
	LineNumber int
}

func generateCodeLocation(skip int) (CodeLocation, bool) {
	_, file, line, ok := runtime.Caller(skip)
	return CodeLocation{FileName: file, LineNumber: line}, ok
}

func (codeLocation CodeLocation) String() string {
	return fmt.Sprintf("%s:%d", codeLocation.FileName, codeLocation.LineNumber)
}
