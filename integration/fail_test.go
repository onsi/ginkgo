package integration_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Failing Specs", func() {
	Describe("when the tests contain failures", func() {
		BeforeEach(func() {
			fm.MountFixture("fail")
		})

		It("should fail in all the possible ways", func() {
			session := startGinkgo(fm.PathTo("fail"), "--no-color")
			Eventually(session).Should(gexec.Exit(1))
			output := string(session.Out.Contents())

			Ω(output).ShouldNot(ContainSubstring("NEVER SEE THIS"))

			Ω(output).Should(ContainSubstring("a top level failure on line 12"))
			Ω(output).Should(ContainSubstring("fail_fixture_test.go:12"))

			Ω(output).Should(ContainSubstring("a sync failure"))
			Ω(output).Should(MatchRegexp(`Test Panicked`))
			Ω(output).Should(MatchRegexp(`a sync panic`))
			Ω(output).Should(ContainSubstring("a sync FAIL failure"))

			Ω(output).Should(ContainSubstring("a top level specify"))
			Ω(output).ShouldNot(ContainSubstring("ginkgo_dsl.go"))
			Ω(output).Should(ContainSubstring("fail_fixture_test.go:38"))

			Ω(output).Should(ContainSubstring("[TIMEDOUT]"))
			Ω(output).Should(MatchRegexp(`goroutine \d+ \[chan receive\]`), "from the progress report emitted by the timeout")
			Ω(output).Should(MatchRegexp(`>\s*\<\-c\.Done\(\)`), "from the progress report emitted by the timeout")

			Ω(output).Should(MatchRegexp(`a top level DescribeTable \[It\] a TableEntry constructed by Entry\n.*fail_fixture_test\.go:45`),
				"the output of a failing Entry should include its file path and line number")

			Ω(output).Should(ContainSubstring(`a helper failed`))
			Ω(output).Should(ContainSubstring(`fail_fixture_test.go:54`), "the code location reported for the helper failure - we're testing the call to GinkgoHelper() works as expected")

			Ω(output).Should(ContainSubstring("synchronous failures with GinkgoT().Fail"))
			Ω(output).Should(ContainSubstring("fail_fixture_ginkgo_t_test.go:9"))

			Ω(output).Should(ContainSubstring("GinkgoT DescribeTable"))
			Ω(output).Should(ContainSubstring("fail_fixture_ginkgo_t_test.go:15"))

			Ω(output).Should(ContainSubstring(`tracks line numbers correctly when GinkgoT().Helper() is called`))
			Ω(output).Should(ContainSubstring(`fail_fixture_ginkgo_t_test.go:21`), "the code location reported for the ginkgoT helper failure")

			Ω(output).Should(ContainSubstring(`tracks the actual line number when no helper is used`))
			Ω(output).Should(ContainSubstring(`fail_fixture_ginkgo_t_test.go:30`), "the code location reported for the ginkgoT no helper failure")

			Ω(output).Should(ContainSubstring("synchronous failures with GinkgoTB().Fail"))
			Ω(output).Should(ContainSubstring("fail_fixture_ginkgo_tb_test.go:9"))

			Ω(output).Should(ContainSubstring("GinkgoTB DescribeTable"))
			Ω(output).Should(ContainSubstring("fail_fixture_ginkgo_tb_test.go:15"))

			Ω(output).Should(ContainSubstring(`tracks line numbers correctly when GinkgoTB().Helper() is called`))
			Ω(output).Should(ContainSubstring(`fail_fixture_ginkgo_tb_test.go:21`), "the code location reported for the ginkgoTB helper failure")

			Ω(output).Should(ContainSubstring(`tracks the actual line number when no GinkgoTB helper is used`))
			Ω(output).Should(ContainSubstring(`fail_fixture_ginkgo_tb_test.go:30`), "the code location reported for the ginkgoT no helper failure")

			Ω(output).Should(ContainSubstring("0 Passed | 16 Failed"))
		})
	})

	Describe("when the tests are incorrectly structured", func() {
		BeforeEach(func() {
			fm.MountFixture("malformed")
		})

		It("exits early with a helpful error message", func() {
			session := startGinkgo(fm.PathTo("malformed"), "--no-color")
			Eventually(session).Should(gexec.Exit(1))
			output := string(session.Out.Contents())

			Ω(output).Should(ContainSubstring("Ginkgo detected an issue with your spec structure"))
			Ω(output).Should(ContainSubstring("malformed_fixture_test.go:9"))
		})

		It("emits the error message even if running in parallel", func() {
			session := startGinkgo(fm.PathTo("malformed"), "--no-color", "--procs=2")
			Eventually(session).Should(gexec.Exit(1))
			output := string(session.Out.Contents()) + string(session.Err.Contents())

			Ω(output).Should(ContainSubstring("Ginkgo detected an issue with your spec structure"))
			Ω(output).Should(ContainSubstring("malformed_fixture_test.go:9"))
		})
	})

	Describe("when By is called outside of a runnable node", func() {
		BeforeEach(func() {
			fm.MountFixture("malformed_by")
		})

		It("exits early with a helpful error message", func() {
			session := startGinkgo(fm.PathTo("malformed_by"), "--no-color", "--procs=2")
			Eventually(session).Should(gexec.Exit(1))
			output := string(session.Out.Contents()) + string(session.Err.Contents())

			Ω(output).Should(ContainSubstring("Ginkgo detected an issue with your spec structure"))
			Ω(output).Should(ContainSubstring("malformed_by_fixture_suite_test.go:16"))

		})
	})
})
