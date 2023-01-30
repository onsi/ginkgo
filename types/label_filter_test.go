package types_test

import (
	"fmt"
	"reflect"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/onsi/ginkgo/v2/types"
)

var _ = Describe("LabelFilter", func() {
	BeforeEach(func() {
		types.DEBUG_LABEL_FILTER_PARSING = false
	})

	DescribeTable("Catching and communicating syntax errors",
		func(filter string, location int, message string) {
			_, err := types.ParseLabelFilter(filter)
			Ω(err).Should(MatchError(types.GinkgoErrors.SyntaxErrorParsingLabelFilter(filter, location, message)))
		},
		func(filter string, location int, message string) string {
			return fmt.Sprintf("%s => %s", filter, message)
		},
		Entry(nil, "(A && B) || ((C && D) && E", 12, "Mismatched '(' - could not find matching ')'."),
		Entry(nil, "A || B) && C", 6, "Mismatched ')' - could not find matching '('."),
		Entry(nil, "A && (   )", 9, "Found empty '()' group."),
		Entry(nil, "A && (((   )))", 11, "Found empty '()' group."),
		Entry(nil, "A && /[a/", 5, "RegExp compilation error: error parsing regexp: missing closing ]: `[a`"),
		Entry(nil, "A &&", -1, "Unexpected EOF."),
		Entry(nil, "A & B", 2, "Invalid token '&'.  Did you mean '&&'?"),
		Entry(nil, "A | B", 2, "Invalid token '|'.  Did you mean '||'?"),
		Entry(nil, "(A) B", 4, "Found two adjacent labels.  You need an operator between them."),
		Entry(nil, "A (B)", 2, "Invalid token '('."),
		Entry(nil, "A !B", 2, "Invalid token '!'."),
		Entry(nil, "A !B", 2, "Invalid token '!'."),
		Entry(nil, " && B", 1, "Operator '&&' missing left hand operand."),
		Entry(nil, " || B", 1, "Operator '||' missing left hand operand."),
		Entry(nil, "&&", 0, "Operator '&&' missing left hand operand."),
		Entry(nil, "&& || B", 0, "Operator '&&' missing left hand operand."),
	)

	type matchingLabels []string
	type nonMatchingLabels []string

	M := func(l ...string) matchingLabels {
		return matchingLabels(l)
	}

	NM := func(l ...string) nonMatchingLabels {
		return nonMatchingLabels(l)
	}

	DescribeTable("Generating correct LabelFilter",
		func(filter string, samples ...interface{}) {
			lf, err := types.ParseLabelFilter(filter)
			Ω(err).ShouldNot(HaveOccurred())
			for _, sample := range samples {
				switch reflect.TypeOf(sample) {
				case reflect.TypeOf(matchingLabels{}):
					labels := []string(sample.(matchingLabels))
					Ω(lf(labels)).Should(BeTrue(), strings.Join(labels, ","))
				case reflect.TypeOf(nonMatchingLabels{}):
					labels := []string(sample.(nonMatchingLabels))
					Ω(lf(labels)).Should(BeFalse(), strings.Join(labels, ","))
				}
			}
		},
		Entry("An empty label", "",
			M("cat"), M("cat", "dog"), M("dog", "cat"),
			M(), M("cow"),
		),
		Entry("A single label", "cat",
			M("cat"), M("cat", "dog"), M("dog", "cat"),
			NM(), NM("cow"),
		),
		Entry("Trimming whitespace", "  cat  ",
			M("cat"), M("cat", "dog"), M("dog", "cat"),
			NM(), NM("cow"),
		),
		Entry("Ignoring case", "cat",
			M("CAT"),
		),
		Entry("A simple ||", "cat || dog ",
			M("cat"), M("cat", "cow", "dog"), M("dog", "cow", "cat"), M("dog"),
			NM(), NM("cow"), NM("cat dog"),
		),
		Entry("A simple ||", "cat||dog ",
			M("cat"), M("cat", "cow", "dog"), M("dog", "cow", "cat"), M("dog"),
			NM(), NM("cow"),
		),
		Entry("A simple ,", "cat, dog ",
			M("cat"), M("cat", "cow", "dog"), M("dog", "cow", "cat"), M("dog"),
			NM(), NM("cow"),
		),
		Entry("Multiple ORs ,", "cat,dog||cow,fruit ",
			M("cat"), M("cat", "cow", "dog"), M("dog"), M("fruit"), M("cow", "aardvark"),
			NM(), NM("aardvark"),
		),
		Entry("A simple NOT", "!cat",
			M("dog"), M(),
			NM("cat"), NM("cat", "dog"),
		),
		Entry("A double negative", "!!cat",
			M("cat"), M("cat", "dog"),
			NM(), NM("dog"),
		),
		Entry("A simple AND", "cat && dog",
			M("cat", "dog"), M("cat", "dog", "cow"),
			NM(), NM("cat"), NM("dog"), NM("cow"), NM("cat dog"),
		),
		Entry("Multiple ANDs", "cat && dog && cow fruit",
			M("cat", "dog", "cow fruit"), M("cat", "dog", "cow fruit", "aardvark"),
			NM(), NM("cat", "dog", "cow", "fruit"),
		),
		Entry("&& has > precedence than ||", "cat || dog && cow",
			M("cat"), M("dog", "cow"),
			NM(), NM("dog"),
		),
		Entry("&& has > precedence than || but () overrides", "(cat || dog) && cow",
			M("cat", "cow"), M("dog", "cow"),
			NM(), NM("dog"), NM("cat"), NM("cow"), NM("cat", "dog"),
		),
		Entry("&& has > precedence than ||", "cat && dog || cow",
			M("cat", "dog"), M("cow"),
			NM(), NM("cat"), NM("dog"),
		),
		Entry("&& has > precedence than || but () overrides", "cat && (dog || cow)",
			M("cat", "dog"), M("cat", "cow"),
			NM(), NM("cat"), NM("dog"), NM("cow"),
		),
		Entry("! has > precedence than &&", "!cat && dog",
			M("dog"), M("dog", "cow"),
			NM(), NM("cat", "dog"), NM("cat"), NM("cow"),
		),
		Entry("! has > precedence than && but () overrides", "!(cat && dog)",
			M(), M("cow"), M("cat"), M("dog"), M("dog", "cow"),
			NM("cat", "dog"), NM("cat", "dog", "cow"),
		),
		Entry("! has > precedence than ||", "!cat || dog",
			M(), M("dog"), M("cow"),
			NM("cat"), NM("cat", "cow"),
		),
		Entry("! has > precedence than || but () overrides", "!(cat || dog)",
			M(), M("cow"),
			NM("cat"), NM("dog"), NM("cat", "dog"), NM("cat", "dog", "cow"),
		),
		Entry("it can handle multiple groups", "(!(cat || dog) && fruit) || (cow && !aardvark)",
			M("cow"), M("fruit"), M("fruit", "cow", "aardvark"), M("cow", "dog", "fruit"),
			NM(), NM("cow", "aardvark"), NM("cat", "fruit"), NM("dog", "fruit"), NM("dog", "cat", "fruit"), NM("cat", "fruit", "cow", "aardvark"),
		),
		Entry("Coalescing groups", "(((cat)))",
			M("cat"), M("cat", "dog"), M("dog", "cat"),
			NM(), NM("cow"),
		),
		Entry("Comping whitespace around a simple group", "  (cat)  ",
			M("cat"), M("cat", "dog"), M("dog", "cat"),
			NM(), NM("cow"),
		),
		Entry("Supporting regular expressions", "/c[ao]/ && dog",
			M("dog", "cat"), M("dog", "cow"), M("cat", "cow", "dog"), M("dog", "orca"),
			NM("dog"), NM("cow"), NM("cat"), NM("dog", "fruit"), NM("dog", "cup"),
		),
	)

	cl := types.NewCodeLocation(0)
	DescribeTable("Validating Labels",
		func(label string, expected string, expectedError error) {
			out, err := types.ValidateAndCleanupLabel(label, cl)
			Ω(out).Should(Equal(expected))
			if expectedError == nil {
				Ω(err).Should(BeNil())
			} else {
				Ω(err).Should(Equal(expectedError))
			}
		},
		func(label string, expected string, expectedError error) string {
			return label
		},
		Entry(nil, "cow", "cow", nil),
		Entry(nil, "  cow dog  ", "cow dog", nil),
		Entry(nil, "", "", types.GinkgoErrors.InvalidEmptyLabel(cl)),
		Entry(nil, "   ", "", types.GinkgoErrors.InvalidEmptyLabel(cl)),
		Entry(nil, "cow&", "", types.GinkgoErrors.InvalidLabel("cow&", cl)),
		Entry(nil, "cow|", "", types.GinkgoErrors.InvalidLabel("cow|", cl)),
		Entry(nil, "cow,", "", types.GinkgoErrors.InvalidLabel("cow,", cl)),
		Entry(nil, "cow(", "", types.GinkgoErrors.InvalidLabel("cow(", cl)),
		Entry(nil, "cow()", "", types.GinkgoErrors.InvalidLabel("cow()", cl)),
		Entry(nil, "cow)", "", types.GinkgoErrors.InvalidLabel("cow)", cl)),
		Entry(nil, "cow/", "", types.GinkgoErrors.InvalidLabel("cow/", cl)),
	)

	Describe("MustParseLabelFilter", func() {
		It("panics if passed an invalid filter", func() {
			Ω(types.MustParseLabelFilter("dog")([]string{"dog"})).Should(BeTrue())
			Ω(types.MustParseLabelFilter("dog")([]string{"cat"})).Should(BeFalse())
			Ω(func() {
				types.MustParseLabelFilter("!")
			}).Should(Panic())
		})
	})
})
