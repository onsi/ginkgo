package subpackage

import (
	. "github.com/onsi/ginkgo"
)

func init() {
	Describe("Testing with Ginkgo", func() {
		It("nested sub packages", func() {
			GinkgoT().Fail(true)
		})
	})
}
