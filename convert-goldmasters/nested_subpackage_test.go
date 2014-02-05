package subpackage

import (
	. "github.com/onsi/ginkgo"
)

func init() {
	Describe("Testing with Ginkgo", func() {
		It("TestNestedSubPackages", func() {
			GinkgoT().Fail(true)
		})
	})
}
