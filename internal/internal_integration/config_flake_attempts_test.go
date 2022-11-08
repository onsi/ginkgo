package internal_integration_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/ginkgo/v2/internal/test_helpers"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
)

var _ = Describe("when config.FlakeAttempts is greater than 1", func() {
	var success bool
	JustBeforeEach(func() {
		var counterA, counterC int

		success, _ = RunFixture("flakey success", func() {
			It("A", rt.T("A", func() {
				counterA += 1
				if counterA < 2 {
					F(fmt.Sprintf("A - %d", counterA))
				}
			}))
			It("B", func() {})
			It("C", FlakeAttempts(1), rt.T("C", func() { //the config flag overwrites the individual test annotations
				counterC += 1
				By(fmt.Sprintf("C - attempt #%d", counterC))
				if counterC < 3 {
					F(fmt.Sprintf("C - %d", counterC))
				}
			}))
		})
	})

	Context("when a test succeeds within the correct number of attempts", func() {
		BeforeEach(func() {
			conf.FlakeAttempts = 3
		})

		It("reports that the suite passed, but with flaked specs", func() {
			Ω(success).Should(BeTrue())
			Ω(reporter.End).Should(BeASuiteSummary(NSpecs(3), NFailed(0), NPassed(3), NFlaked(2)))
		})

		It("reports that the test passed with the correct number of attempts", func() {
			Ω(reporter.Did.Find("A")).Should(HavePassed(NumAttempts(2)))
			Ω(reporter.Did.Find("B")).Should(HavePassed(NumAttempts(1)))
			Ω(reporter.Did.Find("C")).Should(HavePassed(NumAttempts(3)))
			Ω(reporter.Did.Find("C").Timeline()).Should(BeTimelineContaining(
				BeSpecEvent(types.SpecEventByStart, "C - attempt #1"),
				HaveFailed("C - 1"),
				BeSpecEvent(types.SpecEventSpecRetry, 1),
				BeSpecEvent(types.SpecEventByStart, "C - attempt #2"),
				HaveFailed("C - 2"),
				BeSpecEvent(types.SpecEventSpecRetry, 2),
				BeSpecEvent(types.SpecEventByStart, "C - attempt #3"),
			))
		})

		It("includes the intermediate failures as AdditionalFailures (this allows timeline reconstruction)", func() {
			Ω(reporter.Did.Find("C").AdditionalFailures).Should(HaveLen(2))
			Ω(reporter.Did.Find("C").AdditionalFailures[0]).Should(HaveFailed("C - 1"))
			Ω(reporter.Did.Find("C").AdditionalFailures[1]).Should(HaveFailed("C - 2"))
		})
	})

	Context("when the test fails", func() {
		BeforeEach(func() {
			conf.FlakeAttempts = 2
		})

		It("reports that the suite failed", func() {
			Ω(success).Should(BeFalse())
			Ω(reporter.End).Should(BeASuiteSummary(NSpecs(3), NFailed(1), NPassed(2), NFlaked(1)))
		})

		It("reports that the test failed with the correct number of attempts", func() {
			Ω(reporter.Did.Find("A")).Should(HavePassed(NumAttempts(2)))
			Ω(reporter.Did.Find("B")).Should(HavePassed(NumAttempts(1)))
			Ω(reporter.Did.Find("C")).Should(HaveFailed("C - 2", NumAttempts(2)))
			Ω(reporter.Did.Find("C").Timeline()).Should(BeTimelineContaining(
				BeSpecEvent(types.SpecEventByStart, "C - attempt #1"),
				HaveFailed("C - 1"),
				BeSpecEvent(types.SpecEventSpecRetry, 1),
				BeSpecEvent(types.SpecEventByStart, "C - attempt #2"),
			))
		})

		It("includes the intermediate failures as AdditionalFailure, but not the final failure (this allows timeline reconstruction)", func() {
			Ω(reporter.Did.Find("C").AdditionalFailures).Should(HaveLen(1))
			Ω(reporter.Did.Find("C").AdditionalFailures[0]).Should(HaveFailed("C - 1"))
		})

	})
})
