package internal_integration_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/internal/test_helpers"
	. "github.com/onsi/gomega"

	"github.com/onsi/ginkgo/types"
)

var _ = Describe("Decorations test", func() {
	var clForOffset types.CodeLocation
	BeforeEach(func() {
		customIt := func() {
			It("is-offset", rt.T("is-offset"), Offset(1))
		}
		var count = 0
		success, _ := RunFixture("happy-path decoration test", func() {
			Describe("top-level-container", func() {
				clForOffset = types.NewCodeLocation(0)
				customIt()
				It("flaky", FlakeAttempts(4), rt.T("flaky", func() {
					count += 1
					if count < 3 {
						F("fail")
					}
				}))
				It("never-passes", FlakeAttempts(2), rt.T("never-passes", func() {
					F("fail")
				}))
			})
		})
		Ω(success).Should(BeFalse())
	})

	It("runs all the test nodes in the expected order", func() {
		Ω(rt).Should(HaveTracked(
			"is-offset",
			"flaky", "flaky", "flaky",
			"never-passes", "never-passes",
		))
	})

	Describe("Offset", func() {
		It("applies the offset when computing the codelocation", func() {
			clForOffset.LineNumber = clForOffset.LineNumber + 1
			Ω(reporter.Did.Find("is-offset").LeafNodeLocation).Should(Equal(clForOffset))
		})
	})

	Describe("FlakeAttempts", func() {
		It("reruns tests until they pass or until the number of flake attempts is exhausted", func() {
			Ω(reporter.Did.Find("flaky")).Should(HavePassed(NumAttempts(3)))
			Ω(reporter.Did.Find("never-passes")).Should(HaveFailed("fail", NumAttempts(2)))
		})
	})
})
