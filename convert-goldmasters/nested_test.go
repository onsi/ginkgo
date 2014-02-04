package nested

import (
	. "github.com/onsi/ginkgo"
)

func init() {
	Describe("Testing with ginkgo", func() {
		It("TestSomethingLessImportant", func() {

			whatever := &UselessStruct{}
			GinkgoT().Fail(whatever.ImportantField != "SECRET_PASSWORD")
		})
	})
}
