package types_test

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
)

var _ = Describe("FileFilters", func() {

	Describe("Parsing Filters", func() {
		DescribeTable("Failure cases",
			func(filter string) {
				ffs, err := types.ParseFileFilters([]string{filter})
				Ω(ffs).Should(BeZero())
				Ω(err).Should(Equal(types.GinkgoErrors.InvalidFileFilter(filter)))
			},
			Entry(nil, ""),
			Entry(nil, "floop:woop:wibble"),
			Entry(nil, "floop:1asd"),
			Entry(nil, "floop:1asd-3"),
			Entry(nil, "floop:1-3asd"),
			Entry(nil, "floop:1-2-3"),
		)

		DescribeTable("Successful cases",
			func(matches bool, filters []string, clsArgs ...interface{}) {
				ffs, err := types.ParseFileFilters(filters)
				Ω(err).ShouldNot(HaveOccurred())

				cls := []types.CodeLocation{}
				for i := 0; i < len(clsArgs); {
					cls = append(cls, types.CodeLocation{
						FileName:   clsArgs[i].(string),
						LineNumber: clsArgs[i+1].(int),
					})
					i += 2
				}

				if matches {
					Ω(ffs.Matches(cls)).Should(BeTrue())
				} else {
					Ω(ffs.Matches(cls)).Should(BeFalse())
				}
			},
			func(matches bool, filters []string, clsArgs ...interface{}) string {
				return "When the filters are " + strings.Join(filters, " | ")
			},
			//without line numbers
			Entry(nil, true, []string{"foo"}, "foo_test.go", 10),
			Entry(nil, true, []string{"foo"}, "foo/bar_test.go", 10),
			Entry(nil, false, []string{"foo"}, "bar_test.go", 10),
			Entry(nil, true, []string{"foo"}, "foo_test.go", 10, "bar_test.go", 11),
			Entry(nil, true, []string{"foo", "bar"}, "bar_test.go", 10),
			//with line numbers
			Entry(nil, true, []string{"foo:10"}, "foo_test.go", 9, "foo_test.go", 10),
			Entry(nil, false, []string{"foo:11"}, "foo_test.go", 10),
			Entry(nil, false, []string{"foo:10"}, "bar_test.go", 10),
			Entry(nil, true, []string{"foo:10", "foo:11"}, "foo_test.go", 10),
			//with multiple line numbers
			Entry(nil, true, []string{"foo:10,11"}, "foo_test.go", 10),
			Entry(nil, true, []string{"foo:10,11"}, "foo_test.go", 11),
			Entry(nil, false, []string{"foo:10,11"}, "foo_test.go", 12),
			Entry(nil, false, []string{"foo:10,11"}, "foo_test.go", 12),
			//with line ranges
			Entry(nil, false, []string{"foo:10-12"}, "foo_test.go", 9),
			Entry(nil, true, []string{"foo:10-12"}, "foo_test.go", 10),
			Entry(nil, true, []string{"foo:10-12"}, "foo_test.go", 11),
			Entry(nil, false, []string{"foo:10-12"}, "foo_test.go", 12),
			//with all the things
			Entry(nil, false, []string{"foo:7,10-12,15"}, "foo_test.go", 9),
			Entry(nil, true, []string{"foo:7,10-12,15"}, "foo_test.go", 7),
			Entry(nil, false, []string{"foo:7,10-12,15"}, "bar_test.go", 7),
			Entry(nil, true, []string{"foo:7,10-12,15"}, "foo/bar_test.go", 7),
			Entry(nil, true, []string{"foo:7,10-12,15"}, "foo/bar_test.go", 10),
			Entry(nil, true, []string{"foo:7,10-12,15"}, "foo/bar_test.go", 11),
			Entry(nil, false, []string{"foo:7,10-12,15"}, "foo/bar_test.go", 12),
			Entry(nil, false, []string{"foo:7,10-12,15"}, "foo/bar_test.go", 13),
			Entry(nil, true, []string{"foo:7,10-12,15"}, "foo/bar_test.go", 15),
			Entry(nil, true, []string{"foo:7,10-12", "bar:15"}, "foo/bar_test.go", 15),
		)
	})
})
