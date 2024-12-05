package example_test

import (
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("Abnormal\"Fixture", func() {
	When("four random words: flown, authenticating, semiweekly, and overproduction", func() {
		It("has a dangling double-quote here: \"", func() {
			By("step 1")
			By("step 2")
		})
	})
})
