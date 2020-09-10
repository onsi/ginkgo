package internal_integration_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Current Test Descriptions", func() {
	var descriptions map[string]GinkgoTestDescription
	BeforeEach(func() {
		descriptions = map[string]GinkgoTestDescription{}
		logDescription := func(key string, andRun ...func()) func() {
			return func() {
				descriptions[key] = CurrentGinkgoTestDescription()
				if len(andRun) > 0 {
					andRun[0]()
				}
			}
		}

		RunFixture("current test description", func() {
			BeforeSuite(logDescription("before-suite"))
			Context("a passing test", func() {
				BeforeEach(logDescription("bef-A"))
				It("A", logDescription("it-A", func() {
					time.Sleep(20 * time.Millisecond)
				}))
				AfterEach(logDescription("aft-A"))
			})
			Context("a failing test", func() {
				BeforeEach(logDescription("bef-B"))
				It("B", logDescription("it-B", func() {
					F("failed")
				}))
				AfterEach(logDescription("aft-B"))
			})
			AfterSuite(logDescription("after-suite"))
		})
	})

	It("returns an empty GinkgoTestDescription in the before suite and after suite", func() {
		Ω(descriptions["before-suite"]).Should(BeZero())
		Ω(descriptions["after-suite"]).Should(BeZero())
	})

	It("reports as passed while the test is passing", func() {
		Ω(descriptions["bef-A"].Failed).Should(BeFalse())
		Ω(descriptions["it-A"].Failed).Should(BeFalse())
		Ω(descriptions["aft-A"].Failed).Should(BeFalse())
	})

	It("reports as failed when the test fails", func() {
		Ω(descriptions["bef-B"].Failed).Should(BeFalse())
		Ω(descriptions["it-B"].Failed).Should(BeFalse())
		Ω(descriptions["aft-B"].Failed).Should(BeTrue())
	})

	It("captures test details correctly", func() {
		description := descriptions["aft-A"]
		Ω(description.ComponentTexts).Should(Equal([]string{"a passing test", "A"}))
		Ω(description.FullTestText).Should(Equal("a passing test A"))
		Ω(description.TestText).Should(Equal("A"))
		locations := reporter.Did.Find("A").NodeLocations
		location := locations[len(locations)-1]
		Ω(description.FileName).Should(Equal(location.FileName))
		Ω(description.LineNumber).Should(Equal(location.LineNumber))
		Ω(description.Duration).Should(BeNumerically(">=", time.Millisecond*20))
	})
})
