package internal_integration_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/ginkgo/v2/internal/test_helpers"
	. "github.com/onsi/gomega"

	"github.com/onsi/ginkgo/v2/types"
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
					outputInterceptor.AppendInterceptedOutput("so flaky\n")
					writer.Println("so tasty")
					if count < 3 {
						F("fail")
					}
				}))
				It("never-passes", FlakeAttempts(2), rt.T("never-passes", func() {
					F("fail")
				}))
				It("skips", FlakeAttempts(3), rt.T("skips", func() {
					Skip("skip")
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
			"skips",
		))
	})

	Describe("Offset", func() {
		It("applies the offset when computing the codelocation", func() {
			clForOffset.LineNumber = clForOffset.LineNumber + 1
			Ω(reporter.Did.Find("is-offset").LeafNodeLocation).Should(Equal(clForOffset))
		})
	})

	Describe("FlakeAttempts", func() {
		It("reruns tests until they pass or until the number of flake attempts is exhausted, but does not rerun skipped tests", func() {
			Ω(reporter.Did.Find("flaky")).Should(HavePassed(NumAttempts(3), CapturedStdOutput("so flaky\nso flaky\nso flaky\n"), CapturedGinkgoWriterOutput("so tasty\n\nGinkgo: Attempt #1 Failed.  Retrying...\nso tasty\n\nGinkgo: Attempt #2 Failed.  Retrying...\nso tasty\n")))
			Ω(reporter.Did.Find("never-passes")).Should(HaveFailed("fail", NumAttempts(2)))
			Ω(reporter.Did.Find("skips")).Should(HaveBeenSkippedWithMessage("skip", NumAttempts(1)))
		})
	})
})
