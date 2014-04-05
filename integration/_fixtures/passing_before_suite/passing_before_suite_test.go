package passing_before_suite_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("PassingBeforeSuite", func() {
	It("should pass", func() {
		Ω(a).Should(Equal("ran before suite"))
	})

	It("should pass", func() {
		Ω(a).Should(Equal("ran before suite"))
	})

	It("should pass", func() {
		Ω(a).Should(Equal("ran before suite"))
	})

	It("should pass", func() {
		Ω(a).Should(Equal("ran before suite"))
	})
})
