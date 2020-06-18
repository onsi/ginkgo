package integration_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Failing Specs", func() {
	var pathToTest string

	BeforeEach(func() {
		pathToTest = tmpPath("failing")
		copyIn(fixturePath("fail_fixture"), pathToTest, false)
	})

	It("should fail in all the possible ways", func() {
		session := startGinkgo(pathToTest, "--noColor")
		Eventually(session).Should(gexec.Exit(1))
		output := string(session.Out.Contents())

		Ω(output).ShouldNot(ContainSubstring("NEVER SEE THIS"))

		Ω(output).Should(ContainSubstring("a top level failure on line 10"))
		Ω(output).Should(ContainSubstring("fail_fixture_test.go:10"))
		Ω(output).Should(ContainSubstring("an async top level failure on line 15"))
		Ω(output).Should(ContainSubstring("fail_fixture_test.go:15"))
		Ω(output).Should(ContainSubstring("a top level goroutine failure on line 22"))
		Ω(output).Should(ContainSubstring("fail_fixture_test.go:22"))

		Ω(output).Should(ContainSubstring("a sync failure"))
		Ω(output).Should(MatchRegexp(`Test Panicked\n\s+a sync panic`))
		Ω(output).Should(ContainSubstring("a sync FAIL failure"))
		Ω(output).Should(ContainSubstring("async timeout [It]"))
		Ω(output).Should(ContainSubstring("Timed out"))
		Ω(output).Should(ContainSubstring("an async failure"))
		Ω(output).Should(MatchRegexp(`Test Panicked\n\s+an async panic`))
		Ω(output).Should(ContainSubstring("an async FAIL failure"))
		Ω(output).Should(ContainSubstring("a goroutine FAIL failure"))
		Ω(output).Should(ContainSubstring("a goroutine failure"))
		Ω(output).Should(MatchRegexp(`Test Panicked\n\s+a goroutine panic`))
		Ω(output).Should(ContainSubstring("a measure failure"))
		Ω(output).Should(ContainSubstring("a measure FAIL failure"))
		Ω(output).Should(MatchRegexp(`Test Panicked\n\s+a measure panic`))

		Ω(output).Should(ContainSubstring("a top level specify"))
		Ω(output).ShouldNot(ContainSubstring("ginkgo_dsl.go"))
		// depending on the go version this could be the first line of the Specify
		// block (>= go1.9) or the last line of the Specify block (< go1.9)
		Ω(output).Should(Or(ContainSubstring("fail_fixture_test.go:102"), ContainSubstring("fail_fixture_test.go:104")))
		Ω(output).Should(ContainSubstring("fail_fixture_test.go:103"))

		Ω(output).ShouldNot(ContainSubstring("table.go"))
		Ω(output).Should(MatchRegexp(`a top level DescribeTable\n.*fail_fixture_test\.go:106`),
			"the output of a failing DescribeTable should include its file path and line number")
		Ω(output).ShouldNot(ContainSubstring("table_entry.go"))
		Ω(output).Should(MatchRegexp(`a TableEntry constructed by Entry \[It\]\n.*fail_fixture_test\.go:110`),
			"the output of a failing Entry should include its file path and line number")

		Ω(output).Should(ContainSubstring("0 Passed | 19 Failed"))
	})
})
