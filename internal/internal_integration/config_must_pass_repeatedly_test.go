package internal_integration_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/ginkgo/v2/internal/test_helpers"
	. "github.com/onsi/gomega"
)

var _ = Describe("when config.MustPassRepeatedly is greater than 1", func() {
	var success bool
	JustBeforeEach(func() {
		var counterB int
		success, _ = RunFixture("flakey success", func() {
			It("A", func() {})
			It("B", func() {
				counterB += 1
				if counterB == 8 {
					F(fmt.Sprintf("C - %d", counterB))
				}
			})
		})
	})

	Context("when all tests pass", func() {
		BeforeEach(func() {
			conf.MustPassRepeatedly = 5
		})

		It("reports that the suite passed", func() {
			Ω(success).Should(BeTrue())
			Ω(reporter.End).Should(BeASuiteSummary(NSpecs(2), NFailed(0), NPassed(2)))
		})

		It("reports that the tests passed with the correct number of attempts", func() {
			Ω(reporter.Did.Find("A")).Should(HavePassed(NumAttempts(5)))
			Ω(reporter.Did.Find("B")).Should(HavePassed(NumAttempts(5)))
		})
	})

	Context("when a test fails", func() {
		BeforeEach(func() {
			conf.MustPassRepeatedly = 10
		})

		It("reports that the suite failed", func() {
			Ω(success).Should(BeFalse())
			Ω(reporter.End).Should(BeASuiteSummary(NSpecs(2), NFailed(1), NPassed(1)))
		})

		It("reports that the tests failed with the correct number of attempts", func() {
			Ω(reporter.Did.Find("A")).Should(HavePassed(NumAttempts(10)))
			Ω(reporter.Did.Find("B")).Should(HaveFailed(NumAttempts(8)))
		})
	})
})
