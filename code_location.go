package godescribe

import (
	"runtime"
)

func generateCodeLocation(skip int) (CodeLocation, bool) {
	_, file, line, ok := runtime.Caller(skip)
	return CodeLocation{FileName: file, LineNumber: line}, ok
}
