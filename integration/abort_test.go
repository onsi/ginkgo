package integration_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	"github.com/onsi/ginkgo/v2/internal/interrupt_handler"
	. "github.com/onsi/ginkgo/v2/internal/test_helpers"
	"github.com/onsi/ginkgo/v2/reporters"
	"github.com/onsi/ginkgo/v2/types"
)

var _ = Describe("Abort", SpecPriority(10), func() {
	var session *gexec.Session
	BeforeEach(func() {
		fm.MountFixture("abort")
		session = startGinkgo(fm.PathTo("abort"), "--no-color", "--json-report=out.json", "--junit-report=out.xml", "--procs=2")
		Eventually(session).Should(gexec.Exit(1))
	})

	It("aborts the suite and does not run any subsequent tests", func() {
		Ω(session).Should(gbytes.Say("this suite needs to end now!"))
		Ω(string(session.Out.Contents())).ShouldNot(ContainSubstring("SHOULD NOT SEE THIS"))
	})

	It("reports on the test run correctly", func() {
		report := fm.LoadJSONReports("abort", "out.json")[0]
		specs := Reports(report.SpecReports)
		Ω(specs.Find("runs and passes")).Should(HavePassed())
		Ω(specs.Find("aborts")).Should(HaveAborted("this suite needs to end now!"))
		Ω(specs.Find("never runs")).Should(HaveBeenInterrupted(interrupt_handler.InterruptCauseAbortByOtherProcess))
		Ω(specs.Find("never runs either")).Should(HaveBeenSkipped())

		junitSuites := fm.LoadJUnitReport("abort", "out.xml")
		cases := junitSuites.TestSuites[0].TestCases
		var abortCase reporters.JUnitTestCase
		for _, testCase := range cases {
			if testCase.Status == types.SpecStateAborted.String() {
				abortCase = testCase
			}
		}
		Ω(abortCase.Failure.Message).Should(Equal("this suite needs to end now!"))
		Ω(abortCase.Failure.Type).Should(Equal("aborted"))
	})
})
