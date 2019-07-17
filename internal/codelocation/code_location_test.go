package codelocation_test

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/internal/codelocation"
	"github.com/onsi/ginkgo/types"
	. "github.com/onsi/gomega"
)

var (
	codeLocation       types.CodeLocation
	expectedFileName   string
	expectedLineNumber int
	fullStackTrace     string
)

func caller0() {
	fullStackTrace = string(debug.Stack())
	codeLocation = codelocation.New(1)
}

func caller1() {
	_, expectedFileName, expectedLineNumber, _ = runtime.Caller(0)
	expectedLineNumber += 2
	caller0()
}

var _ = Describe("CodeLocation", func() {
	BeforeEach(func() {
		caller1()
	})

	It("should use the passed in skip parameter to pick out the correct file & line number", func() {
		Ω(codeLocation.FileName).Should(Equal(expectedFileName))
		Ω(codeLocation.LineNumber).Should(Equal(expectedLineNumber))
	})

	Describe("stringer behavior", func() {
		It("should stringify nicely", func() {
			Ω(codeLocation.String()).Should(ContainSubstring("code_location_test.go:%d", expectedLineNumber))
		})
	})

	Describe("PruneStack", func() {
		It("should remove any references to ginkgo and pkg/testing and pkg/runtime", func() {
			// Hard-coded string, loosely based on what debug.Stack() produces.
			input := `Skip: skip()
/Skip/me
Skip: skip()
/Skip/me
Something: Func()
/Users/whoever/gospace/src/github.com/onsi/ginkgo/whatever.go:10 (0x12314)
SomethingInternalToGinkgo: Func()
/Users/whoever/gospace/src/github.com/onsi/ginkgo/whatever_else.go:10 (0x12314)
Oops: BlowUp()
/usr/goroot/pkg/strings/oops.go:10 (0x12341)
MyCode: Func()
/Users/whoever/gospace/src/mycode/code.go:10 (0x12341)
MyCodeTest: Func()
/Users/whoever/gospace/src/mycode/code_test.go:10 (0x12341)
TestFoo: RunSpecs(t, "Foo Suite")
/Users/whoever/gospace/src/mycode/code_suite_test.go:12 (0x37f08)
TestingT: Blah()
/usr/goroot/pkg/testing/testing.go:12 (0x37f08)
Something: Func()
/usr/goroot/pkg/runtime/runtime.go:12 (0x37f08)
`
			prunedStack := codelocation.PruneStack(input, 1)
			Ω(prunedStack).Should(Equal(`Oops: BlowUp()
/usr/goroot/pkg/strings/oops.go:10 (0x12341)
MyCode: Func()
/Users/whoever/gospace/src/mycode/code.go:10 (0x12341)
MyCodeTest: Func()
/Users/whoever/gospace/src/mycode/code_test.go:10 (0x12341)
TestFoo: RunSpecs(t, "Foo Suite")
/Users/whoever/gospace/src/mycode/code_suite_test.go:12 (0x37f08)`))
		})

		It("should skip correctly for a Go runtime stack", func() {
			// Actual string from debug.Stack(), something like:
			// "goroutine 5 [running]:",
			// "runtime/debug.Stack(0x0, 0x0, 0x0)",
			// "\t/nvme/gopath/go/src/runtime/debug/stack.go:24 +0xa1",
			// "github.com/onsi/ginkgo/internal/codelocation_test.caller0()",
			// "\t/work/gopath.ginkgo/src/github.com/onsi/XXXXXX/internal/codeloc...+36 more",
			// "github.com/onsi/ginkgo/internal/codelocation_test.caller1()",
			// "\t/work/gopath.ginkgo/src/github.com/onsi/XXXXXX/internal/codeloc...+36 more",
			// "github.com/onsi/ginkgo/internal/codelocation_test.glob..func1.1(...+1 more",
			// "\t/work/gopath.ginkgo/src/github.com/onsi/XXXXXX/internal/codeloc...+36 more",
			//
			// To avoid pruning of our test functions
			// above, we change the expected filename (= this file) into
			// something that isn't special for PruneStack().
			fakeFileName := "github.com/onsi/XXXXXX/internal/codelocation/code_location_test.go"
			mangledStackTrace := strings.Replace(fullStackTrace,
				expectedFileName,
				fakeFileName,
				-1)
			stack := strings.Split(codelocation.PruneStack(mangledStackTrace, 1), "\n")
			Ω(len(stack)).To(BeNumerically(">=", 2), "not enough entries in stack: %s", stack)
			Ω(stack[0]).To(Equal("github.com/onsi/ginkgo/internal/codelocation_test.caller1()"))
			Ω(strings.TrimLeft(stack[1], " \t")).To(HavePrefix(fmt.Sprintf("%s:%d ", fakeFileName, expectedLineNumber)))
		})
	})
})
