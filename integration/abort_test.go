package integration_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo/internal/test_helpers"
	"github.com/onsi/ginkgo/types"
)

var _ = Describe("Abort", func() {
	var session *gexec.Session
	BeforeEach(func() {
		fm.MountFixture("abort")
		session = startGinkgo(fm.PathTo("abort"), "--no-color", "--json-report=out.json", "--junit-report=out.xml")
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
		Ω(specs.Find("never runs")).Should(HaveBeenSkipped())
		Ω(specs.Find("never runs either")).Should(HaveBeenSkipped())

		junitSuites := fm.LoadJUnitReport("abort", "out.xml")
		cases := junitSuites.TestSuites[0].TestCases
		Ω(cases[0].Status).Should(Equal(types.SpecStatePassed.String()))
		Ω(cases[1].Status).Should(Equal(types.SpecStateAborted.String()))
		Ω(cases[1].Failure.Message).Should(Equal("this suite needs to end now!"))
		Ω(cases[1].Failure.Type).Should(Equal("aborted"))
		Ω(cases[2].Status).Should(Equal(types.SpecStateSkipped.String()))
		Ω(cases[3].Status).Should(Equal(types.SpecStateSkipped.String()))
	})

})
