package spechierarchy_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSpecHierarchy(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SpecHierarchy Suite", SemVerConstraint("> 2.0.0"))
}

var _ = Describe("Spec Hierarchy Semantic Version Filtering", func() {
	It("should inherit spec constraint", func() {})

	It("should narrow down spec constraint", SemVerConstraint(">= 3.0.0, < 4.0.0"), func() {})
})
