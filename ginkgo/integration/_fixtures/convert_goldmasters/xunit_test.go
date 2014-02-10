package tmp

import (
	. "github.com/onsi/ginkgo"
)

func init() {
	Describe("Testing with Ginkgo", func() {
		It("something important", func() {

			whatever := &UselessStruct{
				T:              GinkgoT(),
				ImportantField: "SECRET_PASSWORD",
			}
			something := &UselessStruct{ImportantField: "string value"}
			assertEqual(GinkgoT(), whatever.ImportantField, "SECRET_PASSWORD")
			assertEqual(GinkgoT(), something.ImportantField, "string value")

			var foo = func(t GinkgoTestingT) {}
			foo(GinkgoT())

			strp := "something"
			testFunc(GinkgoT(), &strp)
			GinkgoT().Fail()
		})
	})
}

type UselessStruct struct {
	ImportantField string
	T              GinkgoTestingT
}

var testFunc = func(t GinkgoTestingT, arg *string) {}

func assertEqual(t GinkgoTestingT, arg1, arg2 interface{}) {
	if arg1 != arg2 {
		t.Fail()
	}
}
