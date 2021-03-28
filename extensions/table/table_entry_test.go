package table

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TableEntry", func() {
	DescribeTable("entries with invalid descriptions",
		func(description interface{}) {
			Ω(func() {
				DescribeTable("", func(_ string) {},
					Entry(description, "foobar"),
				)
			}).Should(Panic())
		},
		Entry("no description", nil),
		Entry("description function with incorrect parameters", func() string {
			return "foobaz"
		}),
		Entry("description function that does not return a string", func(_ string) {}),
	)

	It("should panic when no parameters are provided to an entry", func() {
		Ω(func() {
			DescribeTable("", func() {}, Entry())
		}).Should(Panic())
	})
})
