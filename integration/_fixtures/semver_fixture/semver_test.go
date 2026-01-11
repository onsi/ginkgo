package semver_fixture_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Semantic Version Filtering", func() {
	It("should run without constraints", func() {})

	It("should run with version in range [2.0.0, ~)", SemVerConstraint(">= 2.0.0"), func() {})

	It("should run with version in range [2.0.0, 3.0.0)", SemVerConstraint(">= 2.0.0, < 3.0.0"), func() {})

	It("should run with version in range [2.0.0, 4.0.0)", SemVerConstraint(">= 2.0.0", "< 4.0.0"), func() {})

	It("should run with version in range [2.0.0, 5.0.0)", SemVerConstraint(">= 2.0.0"), SemVerConstraint("< 5.0.0"), func() {})

	It("should run with ComponentA in range [2.0.0, ~)", ComponentSemVerConstraint("ComponentA", ">= 2.0.0"), func() {})

	It("should run with ComponentA in range [2.0.0, 3.0.0)", ComponentSemVerConstraint("ComponentA", ">= 2.0.0, < 3.0.0"), func() {})

	It("should run with ComponentA in range [2.0.0, 4.0.0)", ComponentSemVerConstraint("ComponentA", ">= 2.0.0", "< 4.0.0"), func() {})

	It("should run with ComponentA in range [2.0.0, 5.0.0)", ComponentSemVerConstraint("ComponentA", ">= 2.0.0"), ComponentSemVerConstraint("ComponentA", "< 5.0.0"), func() {})

	It("should run with a mixed of component-specific and non-component constraints", SemVerConstraint(">= 2.0.0"), ComponentSemVerConstraint("ComponentA", ">= 2.0.0"), func() {})

	It("shouldn't run with version in a conflict range", SemVerConstraint("2.0.0 - 6.0.0"), SemVerConstraint("<= 1.0.0"), func() {})

	It("shouldn't run with ComponentA in a conflict range", ComponentSemVerConstraint("ComponentA", "2.0.0 - 6.0.0"), ComponentSemVerConstraint("ComponentA", "<= 1.0.0"), func() {})

	It("shouldn't run with a mixed of component-specific and non-component conflict constraints", SemVerConstraint(">= 2.0.0"), ComponentSemVerConstraint("ComponentA", "<= 1.0.0"), func() {})
})

var _ = Describe("Hierarchy Semantic Version Filtering", func() {
	Context("with container constraints", SemVerConstraint(">= 2.0.0", "< 3.0.0"), func() {
		It("should inherit container constraint", func() {})

		It("should narrow down the constraint", SemVerConstraint(">= 2.1.0, < 2.8.0"), func() {})

		It("shouldn't expand the constraint", SemVerConstraint("< 4.0.0"), func() {
			// If you pass '--sem-ver-filter=3.5.0', then the whole Context would be skipped since it doesn't match the top level SemVerConstraints.
			// But if you pass '--sem-ver-filter=2.5.0', this test case would keep running since it matches the combined constraint '>= 2.0.0, < 3.0.0, < 4.0.0'
		})

		It("shouldn't combine with a conflict constraint", SemVerConstraint("< 1.0.0"), func() {
			// The new combined constraint is '>= 2.0.0, < 3.0.0, <1.0.0', there's no such a version can match this constraint.
			// So, this test case would be skipped.
		})
	})

	Context("with container component-specific constraints", ComponentSemVerConstraint("ComponentA", ">= 2.0.0", "< 3.0.0"), func() {
		It("should inherit container component-specific constraint", func() {})

		It("should narrow down the component-specific constraint", ComponentSemVerConstraint("ComponentA", ">= 2.1.0, < 2.8.0"), func() {})

		It("shouldn't expand the component-specific constraint", ComponentSemVerConstraint("ComponentA", "< 4.0.0"), func() {
			// If you pass '--sem-ver-filter=3.5.0', then the whole Context would be skipped since it doesn't match the top level ComponentSemVerConstraints.
			// But if you pass '--sem-ver-filter=2.5.0', this test case would keep running since it matches the combined constraint '>= 2.0.0, < 3.0.0, < 4.0.0'
		})

		It("shouldn't combine with a component-specific conflict constraint", ComponentSemVerConstraint("ComponentA", "< 1.0.0"), func() {
			// The new combined constraint is '>= 2.0.0, < 3.0.0, <1.0.0', there's no such a version can match this constraint.
			// So, this test case would be skipped.
		})
	})

	Context("with mixed container constraints", SemVerConstraint(">= 2.0.0", "< 3.0.0"),
		ComponentSemVerConstraint("ComponentA", ">= 2.0.0", "< 3.0.0"), func() {
			It("should inherit container constraints and a new component-specific constraint", ComponentSemVerConstraint("ComponentB", ">= 0.1.0"), func() {})
		})
})

var _ = DescribeTable("Semantic Version Filtering in table-driven spec", func() {
	Expect(true).To(BeTrue())
},
	Entry("should run without constraints by table driven"),
	Entry("should run with version in range [2.0.0, ~) by table driven", SemVerConstraint(">= 2.0.0")),
	Entry("shouldn't run with version in a conflict range by table driven", SemVerConstraint(">= 2.0.0"), SemVerConstraint("~1.2.3")),
	Entry("should run with ComponentA in range [2.0.0, ~) by table driven", ComponentSemVerConstraint("ComponentA", ">= 2.0.0")),
	Entry("shouldn't run with ComponentA in a conflict range by table driven", ComponentSemVerConstraint("ComponentA", ">= 2.0.0"), ComponentSemVerConstraint("ComponentA", "~1.2.3")),
	Entry("should run with mixed constraints by table driven", SemVerConstraint(">= 2.0.0"), ComponentSemVerConstraint("ComponentA", ">= 2.0.0")),
)
