package tmp

import (
	. "github.com/onsi/ginkgo"
)

func somethingImportant(t GinkgoTestingT, message *string) {
	t.Log("Something important happened in a test: " + *message)
}
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestSomethingLessImportant", func() {
			somethingImportant(GinkgoT(), &"hello!")
		})
	})
}
