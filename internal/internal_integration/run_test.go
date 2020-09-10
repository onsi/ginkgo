package internal_integration_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/internal/test_helpers"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	"github.com/onsi/ginkgo/types"
)

var _ = Describe("Running Tests in Series - the happy path", func() {
	BeforeEach(func() {
		success, hPF := RunFixture("happy-path run suite", func() {
			BeforeSuite(rt.T("before-suite", func() {
				time.Sleep(10 * time.Millisecond)
				writer.Write([]byte("before-suite\n"))
			}))
			AfterSuite(rt.T("after-suite", func() {
				time.Sleep(20 * time.Millisecond)
			}))
			Describe("top-level-container", func() {
				JustBeforeEach(rt.T("just-before-each"))
				BeforeEach(rt.T("before-each", func() {
					writer.Write([]byte("before-each\n"))
				}))
				AfterEach(rt.T("after-each"))
				AfterEach(rt.T("after-each-2"))
				JustAfterEach(rt.T("just-after-each"))
				It("A", rt.T("A", func() {
					time.Sleep(10 * time.Millisecond)
				}))
				It("B", rt.T("B", func() {
					time.Sleep(20 * time.Millisecond)
				}))
				Describe("nested-container", func() {
					JustBeforeEach(rt.T("nested-just-before-each"))
					BeforeEach(rt.T("nested-before-each"))
					AfterEach(rt.T("nested-after-each"))
					JustAfterEach(rt.T("nested-just-after-each"))
					JustAfterEach(rt.T("nested-just-after-each-2"))
					It("C", rt.T("C", func() {
						writer.Write([]byte("C\n"))
					}))
					It("D", rt.T("D"))
				})
			})
		})
		Ω(success).Should(BeTrue())
		Ω(hPF).Should(BeFalse())
	})

	It("runs all the test nodes in the expected order", func() {
		Ω(rt).Should(HaveTracked(
			"before-suite",
			"before-each", "just-before-each", "A", "just-after-each", "after-each", "after-each-2",
			"before-each", "just-before-each", "B", "just-after-each", "after-each", "after-each-2",
			"before-each", "nested-before-each", "just-before-each", "nested-just-before-each", "C", "nested-just-after-each", "nested-just-after-each-2", "just-after-each", "nested-after-each", "after-each", "after-each-2",
			"before-each", "nested-before-each", "just-before-each", "nested-just-before-each", "D", "nested-just-after-each", "nested-just-after-each-2", "just-after-each", "nested-after-each", "after-each", "after-each-2",
			"after-suite",
		))
	})

	Describe("reporting", func() {
		It("reports the suite summary correctly when starting", func() {
			Ω(reporter.Begin).Should(MatchFields(IgnoreExtras, Fields{
				"SuiteDescription":           Equal("happy-path run suite"),
				"SuiteSucceeded":             BeFalse(),
				"NumberOfTotalSpecs":         Equal(4),
				"NumberOfSpecsThatWillBeRun": Equal(4),
			}))
		})

		It("reports the suite summary correctly when complete", func() {
			Ω(reporter.End).Should(MatchFields(IgnoreExtras, Fields{
				"SuiteDescription":           Equal("happy-path run suite"),
				"SuiteSucceeded":             BeTrue(),
				"NumberOfTotalSpecs":         Equal(4),
				"NumberOfSpecsThatWillBeRun": Equal(4),
				"NumberOfSkippedSpecs":       Equal(0),
				"NumberOfPassedSpecs":        Equal(4),
				"NumberOfFailedSpecs":        Equal(0),
				"NumberOfPendingSpecs":       Equal(0),
				"NumberOfFlakedSpecs":        Equal(0),
				"RunTime":                    BeNumerically(">=", time.Millisecond*(10+20+10+20)),
			}))
		})

		It("reports the correct suite node summaries", func() {
			Ω(reporter.Did.FindByLeafNodeType(types.NodeTypeBeforeSuite)).Should(MatchFields(IgnoreExtras, Fields{
				"LeafNodeType":               Equal(types.NodeTypeBeforeSuite),
				"State":                      Equal(types.SpecStatePassed),
				"RunTime":                    BeNumerically(">=", 10*time.Millisecond),
				"Failure":                    BeZero(),
				"CapturedGinkgoWriterOutput": Equal("before-suite\n"),
			}))

			Ω(reporter.Did.FindByLeafNodeType(types.NodeTypeAfterSuite)).Should(MatchFields(IgnoreExtras, Fields{
				"LeafNodeType":               Equal(types.NodeTypeAfterSuite),
				"State":                      Equal(types.SpecStatePassed),
				"RunTime":                    BeNumerically(">=", 20*time.Millisecond),
				"Failure":                    BeZero(),
				"CapturedGinkgoWriterOutput": BeZero(),
			}))
		})

		It("reports about each just before it runs", func() {
			Ω(reporter.Will.Names()).Should(Equal([]string{"A", "B", "C", "D"}))
		})

		It("reports about each test after it completes", func() {
			Ω(reporter.Did.Names()).Should(Equal([]string{"A", "B", "C", "D"}))
			Ω(reporter.Did.WithState(types.SpecStatePassed).Names()).Should(Equal([]string{"A", "B", "C", "D"}))

			//spot-check
			Ω(reporter.Did.Find("C")).Should(MatchFields(IgnoreExtras, Fields{
				"LeafNodeType":               Equal(types.NodeTypeIt),
				"NodeTexts":                  Equal([]string{"top-level-container", "nested-container", "C"}),
				"State":                      Equal(types.SpecStatePassed),
				"Failure":                    BeZero(),
				"NumAttempts":                Equal(1),
				"CapturedGinkgoWriterOutput": Equal("before-each\nC\n"),
			}))
		})

		It("computes run times", func() {
			Ω(reporter.Did.Find("A").RunTime).Should(BeNumerically(">=", 10*time.Millisecond))
			Ω(reporter.Did.Find("B").RunTime).Should(BeNumerically(">=", 20*time.Millisecond))
		})
	})
})
