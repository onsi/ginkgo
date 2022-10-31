package internal_integration_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/ginkgo/v2/internal/test_helpers"
	. "github.com/onsi/gomega"

	"github.com/onsi/ginkgo/v2/types"
)

var _ = Describe("Running Tests in Series - the happy path", func() {
	BeforeEach(func() {
		success, hPF := RunFixture("happy-path run suite", func() {
			BeforeSuite(rt.T("before-suite", func() {
				time.Sleep(10 * time.Millisecond)
				writer.Write([]byte("before-suite\n"))
				outputInterceptor.AppendInterceptedOutput("output-intercepted-in-before-suite")
			}))
			AfterSuite(rt.T("after-suite", func() {
				time.Sleep(20 * time.Millisecond)
				outputInterceptor.AppendInterceptedOutput("output-intercepted-in-after-suite")
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
						outputInterceptor.AppendInterceptedOutput("output-intercepted-in-C")
					}))
					It("D", rt.T("D"))
				})
				JustBeforeEach(rt.T("outer-just-before-each"))
				BeforeEach(rt.T("outer-before-each"))
				AfterEach(rt.T("outer-after-each"))
				JustAfterEach(rt.T("outer-just-after-each"))
			})
		})
		Ω(success).Should(BeTrue())
		Ω(hPF).Should(BeFalse())
	})

	It("runs all the test nodes in the expected order", func() {
		Ω(rt).Should(HaveTracked(
			"before-suite",
			"before-each", "outer-before-each", "just-before-each", "outer-just-before-each", "A", "just-after-each", "outer-just-after-each", "after-each", "after-each-2", "outer-after-each",
			"before-each", "outer-before-each", "just-before-each", "outer-just-before-each", "B", "just-after-each", "outer-just-after-each", "after-each", "after-each-2", "outer-after-each",
			"before-each", "outer-before-each", "nested-before-each", "just-before-each", "outer-just-before-each", "nested-just-before-each", "C", "nested-just-after-each", "nested-just-after-each-2", "just-after-each", "outer-just-after-each", "nested-after-each", "after-each", "after-each-2", "outer-after-each",
			"before-each", "outer-before-each", "nested-before-each", "just-before-each", "outer-just-before-each", "nested-just-before-each", "D", "nested-just-after-each", "nested-just-after-each-2", "just-after-each", "outer-just-after-each", "nested-after-each", "after-each", "after-each-2", "outer-after-each",
			"after-suite",
		))
	})

	Describe("reporting", func() {
		It("reports the suite summary correctly when starting", func() {
			Ω(reporter.Begin).Should(SatisfyAll(
				HaveField("SuitePath", "/path/to/suite"),
				HaveField("SuiteDescription", "happy-path run suite"),
				HaveField("SuiteSucceeded", BeFalse()),
				HaveField("PreRunStats.TotalSpecs", 4),
				HaveField("PreRunStats.SpecsThatWillRun", 4),
			))
		})

		It("reports the suite summary correctly when complete", func() {
			Ω(reporter.End).Should(SatisfyAll(
				HaveField("SuitePath", "/path/to/suite"),
				HaveField("SuiteDescription", "happy-path run suite"),
				HaveField("SuiteSucceeded", BeTrue()),
				HaveField("RunTime", BeNumerically(">=", time.Millisecond*(10+20+10+20))),
				HaveField("PreRunStats.TotalSpecs", 4),
				HaveField("PreRunStats.SpecsThatWillRun", 4),
			))
			Ω(reporter.End.SpecReports.WithLeafNodeType(types.NodeTypeIt).CountWithState(types.SpecStatePassed)).Should(Equal(4))
		})

		It("reports the correct suite node summaries", func() {
			Ω(reporter.Did.FindByLeafNodeType(types.NodeTypeBeforeSuite)).Should(SatisfyAll(
				HaveField("LeafNodeType", types.NodeTypeBeforeSuite),
				HaveField("State", types.SpecStatePassed),
				HaveField("RunTime", BeNumerically(">=", 10*time.Millisecond)),
				HaveField("Failure", BeZero()),
				HaveField("CapturedGinkgoWriterOutput", "before-suite\n"),
				HaveField("CapturedStdOutErr", "output-intercepted-in-before-suite"),
				HaveField("ParallelProcess", 1),
			))

			beforeSuiteReport := reporter.Did.FindByLeafNodeType(types.NodeTypeBeforeSuite)
			Ω(beforeSuiteReport.EndTime.Sub(beforeSuiteReport.StartTime)).Should(BeNumerically("~", beforeSuiteReport.RunTime))

			Ω(reporter.Did.FindByLeafNodeType(types.NodeTypeAfterSuite)).Should(SatisfyAll(
				HaveField("LeafNodeType", types.NodeTypeAfterSuite),
				HaveField("State", types.SpecStatePassed),
				HaveField("RunTime", BeNumerically(">=", 20*time.Millisecond)),
				HaveField("Failure", BeZero()),
				HaveField("CapturedGinkgoWriterOutput", BeZero()),
				HaveField("CapturedStdOutErr", "output-intercepted-in-after-suite"),
				HaveField("ParallelProcess", 1),
			))

			afterSuiteReport := reporter.Did.FindByLeafNodeType(types.NodeTypeAfterSuite)
			Ω(afterSuiteReport.EndTime.Sub(afterSuiteReport.StartTime)).Should(BeNumerically("~", afterSuiteReport.RunTime))
		})

		It("reports about each just before it runs", func() {
			Ω(reporter.Will.Names()).Should(Equal([]string{"A", "B", "C", "D"}))
		})

		It("reports about each test after it completes", func() {
			Ω(reporter.Did.Names()).Should(Equal([]string{"A", "B", "C", "D"}))
			Ω(reporter.Did.WithState(types.SpecStatePassed).Names()).Should(Equal([]string{"A", "B", "C", "D"}))

			//spot-check
			Ω(reporter.Did.Find("C")).Should(SatisfyAll(
				HaveField("LeafNodeType", types.NodeTypeIt),
				HaveField("LeafNodeText", "C"),
				HaveField("ContainerHierarchyTexts", []string{"top-level-container", "nested-container"}),
				HaveField("State", types.SpecStatePassed),
				HaveField("Failure", BeZero()),
				HaveField("NumAttempts", 1),
				HaveField("CapturedGinkgoWriterOutput", "before-each\nC\n"),
				HaveField("CapturedStdOutErr", "output-intercepted-in-C"),
				HaveField("ParallelProcess", 1),
				HaveField("RunningInParallel", false),
			))

		})

		It("computes start times, end times, and run times", func() {
			Ω(reporter.Did.Find("A").RunTime).Should(BeNumerically(">=", 10*time.Millisecond))
			Ω(reporter.Did.Find("B").RunTime).Should(BeNumerically(">=", 20*time.Millisecond))

			reportA := reporter.Did.Find("A")
			Ω(reportA.EndTime.Sub(reportA.StartTime)).Should(BeNumerically("~", reportA.RunTime))
		})
	})
})
