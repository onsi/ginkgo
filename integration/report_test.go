package integration_test

import (
	"io/ioutil"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Report", func() {
	BeforeEach(func() {
		fm.MountFixture("reporting")
	})

	It("reports on each test via ReportAfterEach", func() {
		session := startGinkgo(fm.PathTo("reporting"), "--no-color")
		Eventually(session).Should(gexec.Exit(1))

		report, err := ioutil.ReadFile(fm.PathTo("reporting", "report-after-each.out"))
		Ω(err).ShouldNot(HaveOccurred())
		lines := strings.Split(string(report), "\n")
		Ω(lines).Should(ConsistOf(
			"passes - passed",
			"fails - failed",
			"panics - panicked",
			"is pending - pending",
			"is skipped - skipped",
			"",
		))
	})
})
