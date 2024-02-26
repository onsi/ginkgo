package internal_integration_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/internal/interrupt_handler"
	. "github.com/onsi/ginkgo/v2/internal/test_helpers"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
)

var _ = Describe("Sending reports to ReportBeforeEach and ReportAfterEach nodes", func() {
	var reports map[string]Reports
	BeforeEach(func() {
		conf.SkipStrings = []string{"flag-skipped"}
		reports = map[string]Reports{}
		success, hPF := RunFixture("suite with reporting nodes", func() {
			BeforeSuite(rt.T("before-suite"))
			AfterSuite(rt.T("after-suite"))
			ReportAfterEach(func(report types.SpecReport) {
				rt.Run("outer-RAE")
				reports["outer-RAE"] = append(reports["outer-RAE"], report)
			})
			Describe("top-level container", func() {
				ReportBeforeEach(func(report types.SpecReport) {
					rt.Run("inner-RBE")
					reports["inner-RBE"] = append(reports["inner-RBE"], report)
				})
				ReportAfterEach(func(report types.SpecReport) {
					rt.Run("inner-RAE")
					reports["inner-RAE"] = append(reports["inner-RAE"], report)
				})
				It("passes", rt.T("passes"))
				It("fails", rt.T("fails", func() {
					F("fail")
				}))
				It("panics", rt.T("panics", func() {
					panic("boom")
				}))
				PIt("is pending", rt.T("pending"))
				It("is Skip()ed", func() {
					rt.Run("skipped")
					FixtureSkip("nah...")
				})
				It("is flag-skipped", rt.T("flag-skipped"))
				Context("when the ReportAfterEach node fails", func() {
					It("also passes", rt.T("also-passes"))
					ReportAfterEach(func(report types.SpecReport) {
						rt.Run("failing-RAE")
						reports["failing-RAE"] = append(reports["failing-RAE"], report)
						F("fail")
					})
				})
				Context("when the ReportAfterEach node fails in a skipped test", func() {
					It("is also flag-skipped", rt.T("also-flag-skipped"))
					ReportAfterEach(func(report types.SpecReport) {
						rt.Run("failing-in-skip-RAE")
						reports["failing-skipped-RAE"] = append(reports["failing-skipped-RAE"], report)
						F("fail")
					})
				})
				Context("when stuff is emitted to writers and stdout/stderr", func() {
					It("writes stuff", rt.T("writer", func() {
						writer.Println("GinkgoWriter from It")
						outputInterceptor.AppendInterceptedOutput("Output from It\n")
					}))
					ReportAfterEach(func(report types.SpecReport) {
						rt.Run("writing-reporter")
						reports["writing"] = append(reports["writing"], report)
						writer.Println("GinkgoWriter from ReportAfterEach")
						outputInterceptor.AppendInterceptedOutput("Output from ReportAfterEach\n")
					})
				})
				Context("when a ReportBeforeEach fails", func() {
					ReportBeforeEach(func(report types.SpecReport) {
						rt.Run("failing-RBE")
						reports["failing-RBE"] = append(reports["failing-RBE"], report)
						F("fail")
					})
					ReportBeforeEach(func(report types.SpecReport) {
						rt.Run("not-failing-RBE")
						reports["not-failing-RBE"] = append(reports["not-failing-RBE"], report)
					})

					It("does not run", rt.T("does-not-run"))
				})
				Context("when a reporter is interrupted", func() {
					It("passes yet again", rt.T("passes-yet-again"))
					It("skipped by interrupt", rt.T("skipped-by-interrupt"))
					ReportAfterEach(func(report types.SpecReport) {
						if interruptHandler.Status().Level == interrupt_handler.InterruptLevelUninterrupted {
							interruptHandler.Interrupt(interrupt_handler.InterruptCauseSignal)
							time.Sleep(time.Hour)
						}
						rt.Run("interrupt-reporter")
						reports["interrupt"] = append(reports["interrupt"], report)
					})
				})
				Context("when a after each reporter times out", func() {
					It("passes", rt.T("passes"))
					ReportAfterEach(func(ctx SpecContext, report types.SpecReport) {
						select {
						case <-ctx.Done():
							rt.Run("timeout-reporter")
							reports["timeout"] = append(reports["timeout"], report)
						case <-time.After(100 * time.Millisecond):
						}
					}, NodeTimeout(10*time.Millisecond))
				})
			})
			ReportBeforeEach(func(report types.SpecReport) {
				rt.Run("outer-RBE")
				reports["outer-RBE"] = append(reports["outer-RBE"], report)
			})
		})
		Ω(success).Should(BeFalse())
		Ω(hPF).Should(BeFalse())
	})

	It("runs ReportAfterEach blocks in the correct order", func() {
		Ω(rt).Should(HaveTracked(
			"before-suite",
			"outer-RBE", "inner-RBE", "passes", "inner-RAE", "outer-RAE",
			"outer-RBE", "inner-RBE", "fails", "inner-RAE", "outer-RAE",
			"outer-RBE", "inner-RBE", "panics", "inner-RAE", "outer-RAE",
			"outer-RBE", "inner-RBE", "inner-RAE", "outer-RAE", // pending test
			"outer-RBE", "inner-RBE", "skipped", "inner-RAE", "outer-RAE",
			"outer-RBE", "inner-RBE", "inner-RAE", "outer-RAE", // flag-skipped test
			"outer-RBE", "inner-RBE", "also-passes", "failing-RAE", "inner-RAE", "outer-RAE",
			"outer-RBE", "inner-RBE", "failing-in-skip-RAE", "inner-RAE", "outer-RAE", // is also flag-skipped
			"outer-RBE", "inner-RBE", "writer", "writing-reporter", "inner-RAE", "outer-RAE",
			"outer-RBE", "inner-RBE", "failing-RBE", "not-failing-RBE", "inner-RAE", "outer-RAE",
			"outer-RBE", "inner-RBE", "passes-yet-again", "inner-RAE", "outer-RAE",
			"outer-RBE", "inner-RBE", "interrupt-reporter", "inner-RAE", "outer-RAE", // skipped by interrupt
			"outer-RBE", "inner-RBE", "timeout-reporter", "inner-RAE", "outer-RAE", // skipped by timeout
			"after-suite",
		))
	})

	It("does not include the before-suite or after-suite reports", func() {
		Ω(reports["outer-RAE"].FindByLeafNodeType(types.NodeTypeBeforeSuite)).Should(BeZero())
		Ω(reports["outer-RAE"].FindByLeafNodeType(types.NodeTypeAfterSuite)).Should(BeZero())
	})

	It("submits the correct reports to the reporters", func() {
		format.MaxLength = 10000
		for _, name := range []string{"passes", "fails", "panics", "is Skip()ed", "is flag-skipped", "is also flag-skipped"} {

			expected := reporter.Did.Find(name)
			expected.SpecEvents = nil

			actual := reports["outer-RAE"].Find(name)
			actual.SpecEvents = nil
			Ω(actual).Should(Equal(expected))

			actual = reports["inner-RAE"].Find(name)
			actual.SpecEvents = nil
			Ω(actual).Should(Equal(expected))
		}

		Ω(reports["outer-RBE"].Find("passes")).ShouldNot(BeZero())
		Ω(reports["outer-RBE"].Find("fails")).ShouldNot(BeZero())
		Ω(reports["outer-RBE"].Find("panics")).ShouldNot(BeZero())
		Ω(reports["outer-RBE"].Find("is pending")).Should(BePending())
		Ω(reports["outer-RAE"].Find("is flag-skipped")).Should(HaveBeenSkipped())

		Ω(reports["outer-RAE"].Find("passes")).Should(HavePassed())
		Ω(reports["outer-RAE"].Find("fails")).Should(HaveFailed("fail"))
		Ω(reports["outer-RAE"].Find("panics")).Should(HavePanicked("boom"))
		Ω(reports["outer-RAE"].Find("is pending")).Should(BePending())
		Ω(reports["outer-RAE"].Find("is Skip()ed").State).Should(Equal(types.SpecStateSkipped))
		Ω(reports["outer-RAE"].Find("is Skip()ed").Failure.Message).Should(Equal("nah..."))
		Ω(reports["outer-RAE"].Find("is flag-skipped")).Should(HaveBeenSkipped())
	})

	It("handles reporters that fail", func() {
		Ω(reports["failing-RAE"].Find("also passes")).Should(HavePassed())
		Ω(reports["outer-RAE"].Find("also passes")).Should(HaveFailed("fail"))
		Ω(reports["inner-RAE"].Find("also passes")).Should(HaveFailed("fail"))
		Ω(reporter.Did.Find("also passes")).Should(HaveFailed("fail"), FailureNodeType(types.NodeTypeReportAfterEach))

		Ω(reports["failing-RBE"].Find("does not run")).ShouldNot(BeZero())
		Ω(reports["not-failing-RBE"].Find("does not run")).Should(HaveFailed("fail"))
		Ω(reports["outer-RAE"].Find("does not run")).Should(HaveFailed("fail"))
		Ω(reports["inner-RAE"].Find("does not run")).Should(HaveFailed("fail"))
		Ω(reporter.Did.Find("does not run")).Should(HaveFailed("fail", FailureNodeType(types.NodeTypeReportBeforeEach)))
	})

	It("handles reporters that fail, even in skipped specs", func() {
		Ω(reports["failing-skipped-RAE"].Find("is also flag-skipped")).Should(HaveBeenSkipped())
		Ω(reports["outer-RAE"].Find("is also flag-skipped")).Should(HaveFailed("fail"))
		Ω(reports["inner-RAE"].Find("is also flag-skipped")).Should(HaveFailed("fail"))
		Ω(reporter.Did.Find("is also flag-skipped")).Should(HaveFailed("fail"))
	})

	It("captures output from reporter nodes, but only sends them to the DefaultReporter, not the subsequent nodes", func() {
		Ω(reports["writing"].Find("writes stuff").CapturedGinkgoWriterOutput).Should((Equal("GinkgoWriter from It\n")))
		Ω(reports["writing"].Find("writes stuff").CapturedStdOutErr).Should((Equal("Output from It\n")))
		Ω(reports["outer-RAE"].Find("writes stuff").CapturedGinkgoWriterOutput).Should(Equal("GinkgoWriter from It\nGinkgoWriter from ReportAfterEach\n"))
		Ω(reports["outer-RAE"].Find("writes stuff").CapturedStdOutErr).Should(Equal("Output from It\nOutput from ReportAfterEach\n"))
		Ω(reports["inner-RAE"].Find("writes stuff").CapturedGinkgoWriterOutput).Should(Equal("GinkgoWriter from It\nGinkgoWriter from ReportAfterEach\n"))
		Ω(reports["inner-RAE"].Find("writes stuff").CapturedStdOutErr).Should(Equal("Output from It\nOutput from ReportAfterEach\n"))

		// but a report containing the additional output will be send to Ginkgo's reporter...
		Ω(reporter.Did.Find("writes stuff").CapturedGinkgoWriterOutput).Should((Equal("GinkgoWriter from It\nGinkgoWriter from ReportAfterEach\n")))
		Ω(reporter.Did.Find("writes stuff").CapturedStdOutErr).Should((Equal("Output from It\nOutput from ReportAfterEach\n")))
	})
})
