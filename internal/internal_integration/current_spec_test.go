package internal_integration_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/types"
	. "github.com/onsi/gomega"
)

var _ = Describe("CurrentSpec", func() {
	var specs map[string]types.Summary
	BeforeEach(func() {
		specs = map[string]types.Summary{}
		logCurrentSpec := func(key string, andRun ...func()) func() {
			return func() {
				specs[key] = CurrentSpec()
				if len(andRun) > 0 {
					andRun[0]()
				}
			}
		}

		RunFixture("current test description", func() {
			BeforeSuite(logCurrentSpec("before-suite"))
			Context("a passing test", func() {
				BeforeEach(logCurrentSpec("bef-A", func() {
					writer.Println("hello bef-A")
				}))
				It("A", logCurrentSpec("it-A", func() {
					writer.Println("hello it-A")
					time.Sleep(20 * time.Millisecond)
				}))
				AfterEach(logCurrentSpec("aft-A"))
			})
			Context("a failing test", func() {
				BeforeEach(logCurrentSpec("bef-B"))
				It("B", logCurrentSpec("it-B", func() {
					writer.Println("hello it-B")
					F("failed")
				}))
				AfterEach(logCurrentSpec("aft-B"))
			})
			AfterSuite(logCurrentSpec("after-suite"))
		})
	})

	It("returns an empty GinkgoTestDescription in the before suite and after suite", func() {
		Ω(specs["before-suite"]).Should(BeZero())
		Ω(specs["after-suite"]).Should(BeZero())
	})

	It("reports as passed while the test is passing", func() {
		Ω(specs["bef-A"].Failed()).Should(BeFalse())
		Ω(specs["it-A"].Failed()).Should(BeFalse())
		Ω(specs["aft-A"].Failed()).Should(BeFalse())
	})

	It("reports as failed when the test fails", func() {
		Ω(specs["bef-B"].Failed()).Should(BeFalse())
		Ω(specs["it-B"].Failed()).Should(BeFalse())
		Ω(specs["aft-B"].Failed()).Should(BeTrue())
	})

	It("captures GinkgoWriter output", func() {
		Ω(specs["bef-A"].CapturedGinkgoWriterOutput).Should(BeZero())
		Ω(specs["it-A"].CapturedGinkgoWriterOutput).Should(Equal("hello bef-A\n"))
		Ω(specs["aft-A"].CapturedGinkgoWriterOutput).Should(Equal("hello bef-A\nhello it-A\n"))

		Ω(specs["bef-B"].CapturedGinkgoWriterOutput).Should(BeZero())
		Ω(specs["it-B"].CapturedGinkgoWriterOutput).Should(BeZero())
		Ω(specs["aft-B"].CapturedGinkgoWriterOutput).Should(Equal("hello it-B\n"))
	})

	It("captures test details correctly", func() {
		spec := specs["aft-A"]
		Ω(spec.NodeTexts).Should(Equal([]string{"a passing test", "A"}))
		Ω(spec.FullText()).Should(Equal("a passing test A"))
		Ω(spec.SpecText()).Should(Equal("A"))
		locations := reporter.Did.Find("A").NodeLocations
		location := locations[len(locations)-1]
		Ω(spec.FileName()).Should(Equal(location.FileName))
		Ω(spec.LineNumber()).Should(Equal(location.LineNumber))
		Ω(spec.RunTime).Should(BeNumerically(">=", time.Millisecond*20))
	})
})
