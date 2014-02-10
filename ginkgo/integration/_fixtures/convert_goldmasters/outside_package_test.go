package tmp_test

import (
	. "github.com/onsi/ginkgo"
)

type UselessStruct struct {
	ImportantField string
}

func init() {
	Describe("Testing with Ginkgo", func() {
		It("something important", func() {

			whatever := &UselessStruct{}
			GinkgoT().Fail(whatever.ImportantField != "SECRET_PASSWORD")
		})
	})
}
