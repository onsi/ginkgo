package types

import (
	"fmt"
	"regexp"
	"runtime"
	"runtime/debug"
	"strings"
)

type CodeLocation struct {
	FileName       string
	LineNumber     int
	FullStackTrace string
}

func GenerateCodeLocation(skip int) CodeLocation {
	_, file, line, _ := runtime.Caller(skip + 1)
	stackTrace := PruneStack(string(debug.Stack()), skip)
	return CodeLocation{FileName: file, LineNumber: line, FullStackTrace: stackTrace}
}

func (codeLocation CodeLocation) String() string {
	return fmt.Sprintf("%s:%d", codeLocation.FileName, codeLocation.LineNumber)
}

func PruneStack(fullStackTrace string, skip int) string {
	stack := strings.Split(fullStackTrace, "\n")
	if len(stack) > 2*(skip+1) {
		stack = stack[2*(skip+1):]
	}
	prunedStack := []string{}
	re := regexp.MustCompile(`\/ginkgo\/|\/pkg\/testing\/|\/pkg\/runtime\/`)
	for i := 0; i < len(stack)/2; i++ {
		if !re.Match([]byte(stack[i*2])) {
			prunedStack = append(prunedStack, stack[i*2])
			prunedStack = append(prunedStack, stack[i*2+1])
		}
	}
	return strings.Join(prunedStack, "\n")
}
