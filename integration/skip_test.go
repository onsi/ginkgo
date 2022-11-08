package integration_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Skipping Specs", func() {
	BeforeEach(func() {
		fm.MountFixture("skip")
	})

	It("should skip in all the possible ways", func() {
		session := startGinkgo(fm.PathTo("skip"), "--no-color", "-v")
		Eventually(session).Should(gexec.Exit(0))
		output := string(session.Out.Contents())

		Ω(output).ShouldNot(ContainSubstring("NEVER SEE THIS"))

		Ω(output).Should(ContainSubstring("a top level skip on line 9"))
		Ω(output).Should(ContainSubstring("skip_fixture_test.go:9"))

		Ω(output).Should(ContainSubstring("a sync SKIP"))

		Ω(output).Should(ContainSubstring("S [SKIPPED] ["))
		Ω(output).Should(ContainSubstring("a BeforeEach SKIP"))
		Ω(output).Should(ContainSubstring("S [SKIPPED] ["))
		Ω(output).Should(ContainSubstring("an AfterEach SKIP"))

		Ω(output).Should(ContainSubstring("0 Passed | 0 Failed | 0 Pending | 4 Skipped"))
	})
})
