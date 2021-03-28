package table_test

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"strings"
)

var _ = Describe("ItTable", func() {
	Describe("a simple table", func() {
		ItTable(func(x, y int, expected bool) string {
			return fmt.Sprintf("should assert that '%d > %d' is %t", x, y, expected)
		},
			func(x, y int, expected bool) {
				Ω(x > y).Should(Equal(expected))
			},
			Entry(1, 0, true),
			Entry(0, 0, false),
			Entry(0, 1, false),
		)
	})

	Describe("a more complicated table", func() {
		ItTable(func(c ComplicatedThings) string {
			r := fmt.Sprintf("with %d matching substructure", c.Count)
			if c.Count != 1 {
				r += "s"
			}
			return r
		},
			func(c ComplicatedThings) {
				Ω(strings.Count(c.Superstructure, c.Substructure)).Should(BeNumerically("==", c.Count))
			},
			Entry(ComplicatedThings{
				Superstructure: "the sixth sheikh's sixth sheep's sick",
				Substructure:   "emir",
				Count:          0,
			}),
			Entry(ComplicatedThings{
				Superstructure: "the sixth sheikh's sixth sheep's sick",
				Substructure:   "sheep",
				Count:          1,
			}),
			Entry(ComplicatedThings{
				Superstructure: "the sixth sheikh's sixth sheep's sick",
				Substructure:   "si",
				Count:          3,
			}),
		)
	})

	ItTable("should panic when the table description is invalid",
		func(desc interface{}) {
			Ω(func() {
				ItTable(desc, func(_ string) {},
					Entry("foobar"),
				)
			}).Should(Panic())
		},
		Entry(nil),
		Entry(func() string { return "" }),
		Entry(func(_ string) {}),
	)
})
