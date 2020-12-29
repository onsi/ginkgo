package example_test

import (
	. "github.com/onsi/ginkgo"
)

var _ = FDescribe("unfocused", func() {
	FContext("unfocused", func() {
		It("unfocused", func() {
			By("unfocused")
			By("unfocused")
		})
		FIt("focused", func() {
			By("focused")
			By("focused")
		})
	})

	Context("unfocused", func() {
		FIt("focused", func() {

		})
		It("unfocused", func() {

		})
	})

	FContext("focused", func() {
		It("focused", func() {

		})
		It("focused", func() {

		})
	})
})
