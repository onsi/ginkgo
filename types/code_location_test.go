package types_test

import (
	"runtime"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/types"
	. "github.com/onsi/gomega"
)

var _ = Describe("CodeLocation", func() {
	var codeLocation types.CodeLocation
	var expectedFileName string
	var expectedLineNumber int

	caller0 := func() {
		codeLocation = types.NewCodeLocation(1)
	}

	caller1 := func() {
		_, expectedFileName, expectedLineNumber, _ = runtime.Caller(0)
		expectedLineNumber += 2
		caller0()
	}

	BeforeEach(func() {
		caller1()
	})

	It("should use the passed in skip parameter to pick out the correct file & line number", func() {
		Ω(codeLocation.FileName).Should(Equal(expectedFileName))
		Ω(codeLocation.LineNumber).Should(Equal(expectedLineNumber))
		Ω(codeLocation.FullStackTrace).Should(BeZero())
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
			prunedStack := types.PruneStack(input, 1)
			Ω(prunedStack).Should(Equal(`Oops: BlowUp()
/usr/goroot/pkg/strings/oops.go:10 (0x12341)
MyCode: Func()
/Users/whoever/gospace/src/mycode/code.go:10 (0x12341)
MyCodeTest: Func()
/Users/whoever/gospace/src/mycode/code_test.go:10 (0x12341)
TestFoo: RunSpecs(t, "Foo Suite")
/Users/whoever/gospace/src/mycode/code_suite_test.go:12 (0x37f08)`))
		})
	})
})
