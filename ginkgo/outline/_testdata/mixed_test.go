package example_test

import (
	. "github.com/onsi/ginkgo"
)

var _ = FDescribe("unfocused", func() {
	FContext("unfocused", func() {
		It("unfocused", func() {

		})
		FIt("focused", func() {

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

	PContext("unfocused", func() {
		FIt("unfocused", func() {
			By("unfocused")
			By("unfocused")
		})
		It("unfocused", func() {
			By("unfocused")
			By("unfocused")
		})
	})
})
