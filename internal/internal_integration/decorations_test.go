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
		otherCustomIt := func() {
			GinkgoHelper()
			It("is-also-offset", rt.T("is-also-offset"))
		}
		var countFlaky = 0
		var countRepeat = 0
		success, _ := RunFixture("happy-path decoration test", func() {
			Describe("top-level-container", func() {
				clForOffset = types.NewCodeLocation(0)
				customIt()
				otherCustomIt()
				It("flaky", FlakeAttempts(4), rt.T("flaky", func() {
					countFlaky += 1
					outputInterceptor.AppendInterceptedOutput("so flaky\n")
					writer.Println("so tasty")
					if countFlaky < 3 {
						F("fail")
					}
				}))
				It("flaky-never-passes", FlakeAttempts(2), rt.T("flaky-never-passes", func() {
					F("fail")
				}))
				It("flaky-skips", FlakeAttempts(3), rt.T("flaky-skips", func() {
					Skip("skip")
				}))
				It("repeat", MustPassRepeatedly(4), rt.T("repeat", func() {
					countRepeat += 1
					outputInterceptor.AppendInterceptedOutput("repeats a bit\n")
					writer.Println("here we go")
					if countRepeat >= 3 {
						F("fail")
					}
				}))
				It("repeat-never-fails", MustPassRepeatedly(2), rt.T("repeat-never-passes", func() {
					// F("fail")
				}))
				It("repeat-skips", MustPassRepeatedly(3), rt.T("repeat-skips", func() {
					Skip("skip")
				}))
			})
		})
		Ω(success).Should(BeFalse())
	})

	It("runs all the test nodes in the expected order", func() {
		Ω(rt).Should(HaveTracked(
			"is-offset",
			"is-also-offset",
			"flaky", "flaky", "flaky",
			"flaky-never-passes", "flaky-never-passes",
			"flaky-skips",
			"repeat", "repeat", "repeat",
			"repeat-never-passes", "repeat-never-passes",
			"repeat-skips",
		))
	})

	Describe("Offset", func() {
		It("applies the offset when computing the codelocation", func() {
			clForOffset.LineNumber = clForOffset.LineNumber + 1
			Ω(reporter.Did.Find("is-offset").LeafNodeLocation).Should(Equal(clForOffset))
		})
	})

	Describe("GinkgoHelper", func() {
		It("correctly skips through the stack trace when computing the codelocation", func() {
			clForOffset.LineNumber = clForOffset.LineNumber + 2
			Ω(reporter.Did.Find("is-also-offset").LeafNodeLocation).Should(Equal(clForOffset))
		})
	})

	Describe("FlakeAttempts", func() {
		It("reruns specs until they pass or until the number of flake attempts is exhausted, but does not rerun skipped specs", func() {
			Ω(reporter.Did.Find("flaky")).Should(HavePassed(NumAttempts(3), CapturedStdOutput("so flaky\nso flaky\nso flaky\n"), CapturedGinkgoWriterOutput("so tasty\nso tasty\nso tasty\n")))
			Ω(reporter.Did.Find("flaky").Timeline()).Should(BeTimelineContaining(
				BeSpecEvent(types.SpecEventSpecRetry, 1),
				BeSpecEvent(types.SpecEventSpecRetry, 2),
			))
			Ω(reporter.Did.Find("flaky-never-passes")).Should(HaveFailed("fail", NumAttempts(2)))
			Ω(reporter.Did.Find("flaky-never-passes").Timeline()).Should(BeTimelineContaining(
				BeSpecEvent(types.SpecEventSpecRetry, 1),
			))
			Ω(reporter.Did.Find("flaky-skips")).Should(HaveBeenSkippedWithMessage("skip", NumAttempts(1)))
			Ω(reporter.Did.Find("flaky-skips").Timeline()).ShouldNot(BeTimelineContaining(
				BeSpecEvent(types.SpecEventSpecRetry, 1),
			))
		})
	})

	Describe("MustPassRepeatedly", func() {
		It("reruns specs until they fail or until the number of MustPassRepeatedly attempts is exhausted, but does not rerun skipped specs", func() {
			Ω(reporter.Did.Find("repeat")).Should(HaveFailed(NumAttempts(3), CapturedStdOutput("repeats a bit\nrepeats a bit\nrepeats a bit\n"), CapturedGinkgoWriterOutput("here we go\nhere we go\nhere we go\n")))
			Ω(reporter.Did.Find("repeat").Timeline()).Should(BeTimelineContaining(
				BeSpecEvent(types.SpecEventSpecRepeat, 1, TLWithOffset("here we go\n")),
				BeSpecEvent(types.SpecEventSpecRepeat, 2, TLWithOffset("here we go\nhere we go\n")),
			))

			Ω(reporter.Did.Find("repeat-never-fails")).Should(HavePassed("passed", NumAttempts(2)))
			Ω(reporter.Did.Find("repeat-never-fails").Timeline()).Should(BeTimelineContaining(
				BeSpecEvent(types.SpecEventSpecRepeat, 1),
			))

			Ω(reporter.Did.Find("repeat-skips")).Should(HaveBeenSkippedWithMessage("skip", NumAttempts(1)))
		})
	})
})
