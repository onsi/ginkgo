package tmp

import (
	. "github.com/onsi/ginkgo"
)

func init() {
	Describe("Testing with Ginkgo", func() {
		It("something less important", func() {

			strp := "hello!"
			somethingImportant(GinkgoT(), &strp)
		})
	})
}

func somethingImportant(t GinkgoTestingT, message *string) {
	t.Log("Something important happened in a test: " + *message)
}
