package table_test

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/extensions/table"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Table", func() {
	DescribeTable("a simple table",
		func(x int, y int, expected bool) {
			Ω(x > y).Should(Equal(expected))
		},
		Entry("x > y", 1, 0, true),
		Entry("x == y", 0, 0, false),
		Entry("x < y", 0, 1, false),
	)

	type ComplicatedThings struct {
		Superstructure string
		Substructure   string
		Count          int
	}

	DescribeTable("a more complicated table",
		func(c ComplicatedThings) {
			Ω(strings.Count(c.Superstructure, c.Substructure)).Should(BeNumerically("==", c.Count))
		},
		Entry("with no matching substructures", ComplicatedThings{
			Superstructure: "the sixth sheikh's sixth sheep's sick",
			Substructure:   "emir",
			Count:          0,
		}),
		Entry("with one matching substructure", ComplicatedThings{
			Superstructure: "the sixth sheikh's sixth sheep's sick",
			Substructure:   "sheep",
			Count:          1,
		}),
		Entry("with many matching substructures", ComplicatedThings{
			Superstructure: "the sixth sheikh's sixth sheep's sick",
			Substructure:   "si",
			Count:          3,
		}),
	)

	PDescribeTable("a failure",
		func(value bool) {
			Ω(value).Should(BeFalse())
		},
		Entry("when true", true),
		Entry("when false", false),
		Entry("when malformed", 2),
	)

	DescribeTable("an untyped nil as an entry",
		func(x interface{}) {
			Expect(x).To(BeNil())
		},
		Entry("nil", nil),
	)
})

var _ = Describe("TableWithParametricDescription", func() {
	describe := func(desc string) func(int, int, bool) string {
		return func(x, y int, expected bool) string {
			return fmt.Sprintf("%s x=%d y=%d expected:%t", desc, x, y, expected)
		}
	}

	DescribeTable("a simple table",
		func(x int, y int, expected bool) {
			Ω(x > y).Should(Equal(expected))
		},
		Entry(describe("x > y"), 1, 0, true),
		Entry(describe("x == y"), 0, 0, false),
		Entry(describe("x < y"), 0, 1, false),
	)

	type ComplicatedThings struct {
		Superstructure string
		Substructure   string
		Count          int
	}

	describeComplicated := func(desc string) func(ComplicatedThings) string {
		return func(things ComplicatedThings) string {
			return fmt.Sprintf("%s things=%v", desc, things)
		}
	}

	DescribeTable("a more complicated table",
		func(c ComplicatedThings) {
			Ω(strings.Count(c.Superstructure, c.Substructure)).Should(BeNumerically("==", c.Count))
		},
		Entry(describeComplicated("with no matching substructures"), ComplicatedThings{
			Superstructure: "the sixth sheikh's sixth sheep's sick",
			Substructure:   "emir",
			Count:          0,
		}),
		Entry(describeComplicated("with one matching substructure"), ComplicatedThings{
			Superstructure: "the sixth sheikh's sixth sheep's sick",
			Substructure:   "sheep",
			Count:          1,
		}),
		Entry(describeComplicated("with many matching substructures"), ComplicatedThings{
			Superstructure: "the sixth sheikh's sixth sheep's sick",
			Substructure:   "si",
			Count:          3,
		}),
	)
})
