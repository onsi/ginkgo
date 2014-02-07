package tmp

import (
	. "github.com/onsi/ginkgo"
)

type UselessStruct struct {
	ImportantField string
	T              GinkgoTestingT
}

var testFunc = func(t GinkgoTestingT, arg *string) {}

func init() {
	Describe("Testing with Ginkgo", func() {
		It("TestSomethingImportant", func() {

			whatever := &UselessStruct{
				T:              GinkgoT(),
				ImportantField: "twisty maze of passages",
			}
			app := "string value"
			something := &UselessStruct{ImportantField: app}
			GinkgoT().Fail(whatever.ImportantField != "SECRET_PASSWORD")
			assert.Equal(GinkgoT(), whatever.ImportantField, "SECRET_PASSWORD")
			var foo = func(t GinkgoTestingT) {}
			foo()
			testFunc(GinkgoT(), "something")
		})
	})
}
