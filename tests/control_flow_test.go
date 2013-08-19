package gospec_tests

import (
	"fmt"
	. "github.com/onsi/godescribe"
)

func init() {
	Describe("A", func() {
		BeforeEach(func() {
			fmt.Println("BEFORE")
		})

		It("should work!", func() {
			fmt.Println("WORK!")
		})
	})
}
