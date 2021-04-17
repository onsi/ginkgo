package internal_integration_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/internal/test_helpers"
	"github.com/onsi/ginkgo/types"
	. "github.com/onsi/gomega"
)

var _ = Describe("Sending reports to ReportAfterEach nodes", func() {
	var reports map[string]Reports
	BeforeEach(func() {
		conf.SkipStrings = []string{"flag-skipped"}
		reports = map[string]Reports{}
		success, hPF := RunFixture("suite with reporting nodes", func() {
			BeforeSuite(rt.T("before-suite"))
			AfterSuite(rt.T("after-suite"))
			ReportAfterEach(func(report types.SpecReport) {
				rt.Run("outer-reporter")
				reports["outer"] = append(reports["outer"], report)
			})
			Describe("top-level container", func() {
				ReportAfterEach(func(report types.SpecReport) {
					rt.Run("inner-reporter")
					reports["inner"] = append(reports["inner"], report)
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
				Context("when the report node fails", func() {
					It("also passes", rt.T("also-passes"))
					ReportAfterEach(func(report types.SpecReport) {
						rt.Run("failing-reporter")
						reports["failing"] = append(reports["failing"], report)
						F("fail")
					})
				})
				Context("when stuff is emitted to writers and stdout/stderr", func() {
					It("writes stuff", rt.T("writer", func() {
						writer.Println("GinkgoWriter from It")
						outputInterceptor.InterceptedOutput = "Output from It\n"
					}))
					ReportAfterEach(func(report types.SpecReport) {
						rt.Run("writing-reporter")
						reports["writing"] = append(reports["writing"], report)
						writer.Println("GinkgoWriter from ReportAfterEach")
						outputInterceptor.InterceptedOutput = "Output from ReportAfterEach\n"
					})
				})
				Context("when a reporter is interrupted", func() {
					It("passes yet again", rt.T("passes-yet-again"))
					It("skipped by interrupt", rt.T("skipped-by-interrupt"))
					ReportAfterEach(func(report types.SpecReport) {
						interruptHandler.Interrupt()
						time.Sleep(100 * time.Millisecond)
						rt.RunWithData("interrupt-reporter", "interrupt-message", interruptHandler.EmittedInterruptMessage())
						reports["interrupt"] = append(reports["interrupt"], report)
					})
				})
			})
		})
		Ω(success).Should(BeFalse())
		Ω(hPF).Should(BeFalse())
	})

	It("runs ReportAfterEach blocks in the correct order", func() {
		Ω(rt).Should(HaveTracked(
			"before-suite",
			"passes", "inner-reporter", "outer-reporter",
			"fails", "inner-reporter", "outer-reporter",
			"panics", "inner-reporter", "outer-reporter",
			"inner-reporter", "outer-reporter", //pending test
			"skipped", "inner-reporter", "outer-reporter",
			"inner-reporter", "outer-reporter", //flag-skipped test
			"also-passes", "failing-reporter", "inner-reporter", "outer-reporter",
			"writer", "writing-reporter", "inner-reporter", "outer-reporter",
			"passes-yet-again", "interrupt-reporter", "inner-reporter", "outer-reporter",
			"interrupt-reporter", "inner-reporter", "outer-reporter",
			"after-suite",
		))
	})

	It("does not include the before-suite or after-suite reports", func() {
		Ω(reports["outer"].FindByLeafNodeType(types.NodeTypeBeforeSuite)).Should(BeZero())
		Ω(reports["outer"].FindByLeafNodeType(types.NodeTypeAfterSuite)).Should(BeZero())
	})

	It("submits the correct reports to the reporters", func() {
		for _, name := range []string{"passes", "fails", "panics", "is Skip()ed", "is flag-skipped"} {
			Ω(reports["outer"].Find(name)).Should(Equal(reporter.Did.Find(name)))
			Ω(reports["inner"].Find(name)).Should(Equal(reporter.Did.Find(name)))
		}

		Ω(reports["outer"].Find("passes")).Should(HavePassed())
		Ω(reports["outer"].Find("fails")).Should(HaveFailed("fail"))
		Ω(reports["outer"].Find("panics")).Should(HavePanicked("boom"))
		Ω(reports["outer"].Find("is pending")).Should(BePending())
		Ω(reports["outer"].Find("is Skip()ed").State).Should(Equal(types.SpecStateSkipped))
		Ω(reports["outer"].Find("is Skip()ed").Failure.Message).Should(Equal("nah..."))
		Ω(reports["outer"].Find("is flag-skipped")).Should(HaveBeenSkipped())
	})

	It("handles reporters that fail", func() {
		//The "failing" reporter actually fails
		//the identical passing report will be send to all reporters in the chain...
		Ω(reports["failing"].Find("also passes")).Should(HavePassed())
		Ω(reports["outer"].Find("also passes")).Should(HavePassed())
		Ω(reports["inner"].Find("also passes")).Should(HavePassed())
		Ω(reports["failing"].Find("also passes")).Should(Equal(reports["outer"].Find("also passes")))
		Ω(reports["failing"].Find("also passes")).Should(Equal(reports["inner"].Find("also passes")))

		//but the failing report will be sent to Ginkgo's reporter...
		Ω(reporter.Did.Find("also passes")).Should(HaveFailed("fail"))
	})

	It("captures output from reporter nodes, but only sends them to the DefaultReporter, not the subsequent nodes", func() {
		//The "writing" reporter follows a test that wrote to the GinkgoWriter and stdout/err
		//But the reporter, itself, emits things to GinkgoWriter and stdout/err
		//the identical report containing only the test content will be send to all reporters in the chain...
		Ω(reports["writing"].Find("writes stuff").CapturedGinkgoWriterOutput).Should((Equal("GinkgoWriter from It\n")))
		Ω(reports["writing"].Find("writes stuff").CapturedStdOutErr).Should((Equal("Output from It\n")))
		Ω(reports["writing"].Find("writes stuff")).Should(Equal(reports["outer"].Find("writes stuff")))
		Ω(reports["writing"].Find("writes stuff")).Should(Equal(reports["inner"].Find("writes stuff")))

		//but a report containing the additional output will be send to Ginkgo's reporter...
		Ω(reporter.Did.Find("writes stuff").CapturedGinkgoWriterOutput).Should((Equal("GinkgoWriter from It\nGinkgoWriter from ReportAfterEach\n")))
		Ω(reporter.Did.Find("writes stuff").CapturedStdOutErr).Should((Equal("Output from It\nOutput from ReportAfterEach\n")))
	})

	It("ignores interrupts and soldiers on", func() {
		//The "interrupt" reporter is interrupted by the user - but keeps running (instead, the user sees a message emitted that they are attempting to interrupt a reporter and will just need to wait)
		//The interrupt is, however, honored and subsequent tests are skipped.  These skipped tests, however, are still reported to the reporter node
		Ω(reports["interrupt"].Find("passes yet again")).ShouldNot(BeZero())
		Ω(reports["interrupt"].Find("passes yet again")).Should(HavePassed())
		Ω(reports["interrupt"].Find("skipped by interrupt")).Should(HaveBeenSkipped())
		Ω(reports["interrupt"].Find("passes yet again")).Should(Equal(reports["inner"].Find("passes yet again")))
		Ω(reports["interrupt"].Find("passes yet again")).Should(Equal(reports["outer"].Find("passes yet again")))
		Ω(reports["interrupt"].Find("skipped by interrupt")).Should(Equal(reports["inner"].Find("skipped by interrupt")))
		Ω(reports["interrupt"].Find("skipped by interrupt")).Should(Equal(reports["outer"].Find("skipped by interrupt")))

		cl := types.NewCodeLocation(0)
		Ω(rt.DataFor("interrupt-reporter")["interrupt-message"]).Should(ContainSubstring("The running ReportAfterEach node is at:\n%s", cl.FileName))
	})
})
