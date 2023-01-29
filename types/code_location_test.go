package types_test

import (
	"fmt"
	"runtime"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
)

var _ = Describe("CodeLocation", func() {
	clWithSkip := func(skip int) types.CodeLocation {
		return types.NewCodeLocation(skip)
	}

	helperFunction := func() types.CodeLocation {
		GinkgoHelper()
		return types.NewCodeLocation(0)
	}

	Describe("Creating CodeLocations", func() {
		Context("when skip is 0", func() {
			It("returns the location at which types.NewCodeLocation was called", func() {
				_, fname, lnumber, _ := runtime.Caller(0)
				cl := types.NewCodeLocation(0)
				Ω(cl).Should(Equal(types.CodeLocation{
					FileName:   fname,
					LineNumber: lnumber + 1,
				}))
			})
		})

		Context("when skip is > 0", func() {
			It("returns the appropriate location from the stack", func() {
				_, fname, lnumber, _ := runtime.Caller(0)
				cl := clWithSkip(1)
				Ω(cl).Should(Equal(types.CodeLocation{
					FileName:   fname,
					LineNumber: lnumber + 1,
				}))

				_, fname, lnumber, _ = runtime.Caller(0)
				cl = func() types.CodeLocation {
					return clWithSkip(2)
				}() // this is the line that's expected
				Ω(cl).Should(Equal(types.CodeLocation{
					FileName:   fname,
					LineNumber: lnumber + 3,
				}))
			})
		})

		Describe("when a function has been marked as a helper", func() {
			It("does not include that function when generating a code location", func() {
				_, fname, lnumber, _ := runtime.Caller(0)
				cl := helperFunction()
				Ω(cl).Should(Equal(types.CodeLocation{
					FileName:   fname,
					LineNumber: lnumber + 1,
				}))

				_, fname, lnumber, _ = runtime.Caller(0)
				cl = func() types.CodeLocation {
					GinkgoHelper()
					return func() types.CodeLocation {
						types.MarkAsHelper()
						return helperFunction()
					}()
				}() // this is the line that's expected
				Ω(cl).Should(Equal(types.CodeLocation{
					FileName:   fname,
					LineNumber: lnumber + 7,
				}))
			})
		})
	})

	Describe("stringer behavior", func() {
		It("should stringify nicely", func() {
			_, fname, lnumber, _ := runtime.Caller(0)
			cl := types.NewCodeLocation(0)
			Ω(cl.String()).Should(Equal(fmt.Sprintf("%s:%d", fname, lnumber+1)))
		})
	})

	Describe("with a custom message", func() {
		It("emits the custom message", func() {
			cl := types.NewCustomCodeLocation("I'm right here.")
			Ω(cl.String()).Should(Equal("I'm right here."))
		})
	})

	Describe("Fetching the line from the file in question", func() {
		It("works", func() {
			cl := types.NewCodeLocation(0)
			cl.LineNumber = cl.LineNumber - 2
			Ω(cl.ContentsOfLine()).Should(Equal("\tDescribe(\"Fetching the line from the file in question\", func() {"))
		})

		It("returns empty string if the line is not found or is out of bounds", func() {
			cl := types.CodeLocation{
				FileName:   "foo.go",
				LineNumber: 0,
			}
			Ω(cl.ContentsOfLine()).Should(BeZero())

			cl = types.NewCodeLocation(0)
			cl.LineNumber = cl.LineNumber + 1000000
			Ω(cl.ContentsOfLine()).Should(BeZero())
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
