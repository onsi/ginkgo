package ginkgo

import (
	. "github.com/onsi/gomega"
	"runtime"
)

func init() {
	Describe("CodeLocation", func() {
		var (
			codeLocation       CodeLocation
			expectedFileName   string
			expectedLineNumber int
		)

		caller0 := func() {
			codeLocation = generateCodeLocation(1)
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
		})

		It("should also have the full stack trace", func() {
			Ω(codeLocation.FullStackTrace).Should(ContainSubstring(expectedFileName))
			Ω(codeLocation.FullStackTrace).Should(ContainSubstring("caller0()"))
			Ω(codeLocation.FullStackTrace).Should(ContainSubstring(codeLocation.String()), "The code location string actually appears in the stack trace.")
		})

		Describe("stringer behavior", func() {
			It("should stringify nicely", func() {
				Ω(codeLocation.String()).Should(ContainSubstring("code_location_test.go:%d", expectedLineNumber))
			})
		})
	})
}
