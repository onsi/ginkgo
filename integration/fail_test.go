package integration_test

import (
	. "github.com/onsi/ginkgo"
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

			Ω(output).Should(ContainSubstring("a top level failure on line 10"))
			Ω(output).Should(ContainSubstring("fail_fixture_test.go:10"))

			Ω(output).Should(ContainSubstring("a sync failure"))
			Ω(output).Should(MatchRegexp(`Test Panicked`))
			Ω(output).Should(MatchRegexp(`a sync panic`))
			Ω(output).Should(ContainSubstring("a sync FAIL failure"))

			Ω(output).Should(ContainSubstring("a top level specify"))
			Ω(output).ShouldNot(ContainSubstring("ginkgo_dsl.go"))
			Ω(output).Should(ContainSubstring("fail_fixture_test.go:31"))

			Ω(output).ShouldNot(ContainSubstring("table.go"))
			Ω(output).Should(MatchRegexp(`a top level DescribeTable\n.*fail_fixture_test\.go:35`),
				"the output of a failing DescribeTable should include its file path and line number")
			Ω(output).ShouldNot(ContainSubstring("table_entry.go"))
			Ω(output).Should(MatchRegexp(`\[It\] a TableEntry constructed by Entry\n.*fail_fixture_test\.go:39`),
				"the output of a failing Entry should include its file path and line number")

			Ω(output).Should(ContainSubstring("0 Passed | 7 Failed"))
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

			Ω(output).Should(ContainSubstring("Ginkgo detected an issue with your test structure"))
			Ω(output).Should(ContainSubstring("malformed_fixture_test.go:9"))
		})
	})
})
