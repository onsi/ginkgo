package types_test

import (
	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
)

var _ = Describe("SemVerFilter", func() {
	DescribeTable("ValidateAndCleanupSemVerConstraint", func(semVerConstraint, errMsg string) {
		cl := types.NewCodeLocation(0)

		constraint, err := types.ValidateAndCleanupSemVerConstraint(semVerConstraint, cl)
		if len(errMsg) != 0 {
			Expect(err.Error()).Should(ContainSubstring(errMsg))
			return
		}
		Expect(constraint).To(Equal(semVerConstraint))
	},
		Entry("no constraints", "", "SemVerConstraint cannot be empty"),
		Entry("invalid constraint", "a1.2.3", "is an invalid SemVerConstraint"),
		Entry("valid constraint", "1.0.0", ""),
		Entry("valid constraint", "2.x.0", ""),
	)

	type input struct {
		version        string
		component      string
		constraints    []string
		expectedErrMsg string
		shouldPass     bool
	}
	DescribeTable("ParseSemVerFilter", func(i input) {
		filterFn, err := types.ParseSemVerFilter(i.version)
		if i.expectedErrMsg != "" {
			Expect(err.Error()).Should(ContainSubstring(i.expectedErrMsg))
			return
		}
		Expect(err).ShouldNot(HaveOccurred())
		Expect(filterFn(i.component, i.constraints)).To(Equal(i.shouldPass))
	},
		Entry("no semantic version filter for constraints", input{
			version:        "",
			component:      "",
			constraints:    []string{"> 1.0.0"},
			expectedErrMsg: "",
			shouldPass:     true,
		}),
		Entry("no semantic version filter for component constraints", input{
			version:        "",
			component:      "componentA",
			constraints:    []string{"> 1.0.0"},
			expectedErrMsg: "",
			shouldPass:     true,
		}),
		Entry("no semantic version constraints", input{
			version:        "2.0.0",
			component:      "",
			constraints:    []string{},
			expectedErrMsg: "",
			shouldPass:     true,
		}),
		Entry("no semantic version constraints for a component", input{
			version:        "2.0.0",
			component:      "componentA",
			constraints:    []string{},
			expectedErrMsg: "",
			shouldPass:     true,
		}),
		Entry("invalid semantic version filter", input{
			version:        "a1.0.0",
			component:      "",
			constraints:    []string{"> 1.0.0"},
			expectedErrMsg: "invalid filter version",
			shouldPass:     false,
		}),
		Entry("matched semantic version with single constraint", input{
			version:        "2.0.0",
			component:      "",
			constraints:    []string{"> 1.0.0"},
			expectedErrMsg: "",
			shouldPass:     true,
		}),
		Entry("matched semantic version with component single constraint", input{
			version:        "compA=2.0.0",
			component:      "compA",
			constraints:    []string{"> 1.0.0"},
			expectedErrMsg: "",
			shouldPass:     true,
		}),
		Entry("matched semantic version with multiple constraints", input{
			version:        "2.0.0",
			component:      "",
			constraints:    []string{"> 1.0.0", "< 3.0.0"},
			expectedErrMsg: "",
			shouldPass:     true,
		}),
		Entry("matched semantic version with component multiple constraints", input{
			version:        "compA=2.0.0",
			component:      "compA",
			constraints:    []string{"> 1.0.0", "< 3.0.0"},
			expectedErrMsg: "",
			shouldPass:     true,
		}),
		Entry("ignore non-component semantic version with component constraint", input{
			version:        "2.0.0",
			component:      "compA",
			constraints:    []string{"== 1.0.0"},
			expectedErrMsg: "",
			shouldPass:     true,
		}),
		Entry("matched semantic version with complex constraint", input{
			version:        "2.0.0",
			component:      "",
			constraints:    []string{"> 1.0.0, < 3.0.0"},
			expectedErrMsg: "",
			shouldPass:     true,
		}),
		Entry("matched semantic version with component complex constraint", input{
			version:        "compA=2.0.0",
			component:      "compA",
			constraints:    []string{"> 1.0.0, < 3.0.0"},
			expectedErrMsg: "",
			shouldPass:     true,
		}),
		Entry("matched mixed semantic version with constraints", input{
			version:        "1.2.0, compA=2.0.0",
			component:      "",
			constraints:    []string{"> 1.0.0, < 3.0.0"},
			expectedErrMsg: "",
			shouldPass:     true,
		}),
		Entry("matched mixed semantic version with component constraints", input{
			version:        "1.0.0, compA=2.0.0",
			component:      "compA",
			constraints:    []string{"> 1.0.0, < 3.0.0"},
			expectedErrMsg: "",
			shouldPass:     true,
		}),
		Entry("unmatched semantic version with constraint", input{
			version:        "0.1.0",
			component:      "",
			constraints:    []string{"> 1.0.0, < 3.0.0"},
			expectedErrMsg: "",
			shouldPass:     false,
		}),
		Entry("unmatched component semantic version with component constraint", input{
			version:        "compA=0.1.0",
			component:      "compA",
			constraints:    []string{"> 1.0.0, < 3.0.0"},
			expectedErrMsg: "",
			shouldPass:     false,
		}),
	)

	Describe("MustParseSemVerFilter", func() {
		It("panics if passed an invalid filter", func() {
			Ω(types.MustParseSemVerFilter("1.2.3")("", []string{"> 1.0.0"})).Should(BeTrue())
			Ω(types.MustParseSemVerFilter("3.2.0")("", []string{"> 1.0.0", "< 2.0.0"})).Should(BeFalse())
			Ω(func() {
				types.MustParseSemVerFilter("a1.2.3")
			}).Should(Panic())
			Ω(types.MustParseSemVerFilter("compA=1.2.3")("compA", []string{"> 1.0.0"})).Should(BeTrue())
			Ω(types.MustParseSemVerFilter("compA=3.2.0")("compA", []string{"> 1.0.0", "< 2.0.0"})).Should(BeFalse())
			Ω(types.MustParseSemVerFilter("1.0.0")("compA", []string{"> 2.0.0"})).Should(BeTrue())
		})
	})
})
