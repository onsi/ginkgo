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

	DescribeTable("ParseSemVerFilter", func(version string, constraints []string, expectedErrMsg string, shouldPass bool) {
		filterFn, err := types.ParseSemVerFilter(version)
		if expectedErrMsg != "" {
			Expect(err.Error()).Should(ContainSubstring(expectedErrMsg))
			return
		}
		Expect(err).ShouldNot(HaveOccurred())
		Expect(filterFn(constraints)).To(Equal(shouldPass))
	},
		Entry("no semantic version filter", "", []string{"> 1.0.0"}, "", true),
		Entry("no semantic version constraints", "2.0.0", []string{}, "", true),
		Entry("invalid semantic version filter", "a1.0.0", []string{"> 1.0.0"}, "invalid filter version", false),
		Entry("matched semantic version with single constraint", "2.0.0", []string{"> 1.0.0"}, "", true),
		Entry("matched semantic version with multiple constraints", "2.0.0", []string{"> 1.0.0", "< 3.0.0"}, "", true),
		Entry("matched semantic version with complex constraint", "2.0.0", []string{"> 1.0.0, < 3.0.0"}, "", true),
		Entry("unmatched semantic version with constraint", "0.1.0", []string{"> 1.0.0, < 3.0.0"}, "", false),
	)

	Describe("MustParseSemVerFilter", func() {
		It("panics if passed an invalid filter", func() {
			Ω(types.MustParseSemVerFilter("1.2.3")([]string{"> 1.0.0"})).Should(BeTrue())
			Ω(types.MustParseSemVerFilter("3.2.0")([]string{"> 1.0.0", "< 2.0.0"})).Should(BeFalse())
			Ω(func() {
				types.MustParseSemVerFilter("a1.2.3")
			}).Should(Panic())
		})
	})
})
