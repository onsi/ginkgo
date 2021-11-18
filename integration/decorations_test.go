package integration_test

import (
	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Decorations", func() {
	BeforeEach(func() {
		fm.MountFixture("decorations")
	})

	It("processes the various decorations", func() {
		session := startGinkgo(fm.PathTo("decorations"), "-vv", "--no-color")
		Eventually(session).Should(gexec.Exit(types.GINKGO_FOCUS_EXIT_CODE))

		out := string(session.Out.Contents())
		Ω(out).Should(MatchRegexp(
			`P \[PENDING\]
some decorated tests
.*decorations_fixture_suite_test.go:\d+
  pending it`,
		))
		Ω(out).ShouldNot(ContainSubstring("never_see_this_file"))
	})

	It("exits with a clear error if decorations are misconfigured", func() {
		session := startGinkgo(fm.PathTo("decorations", "invalid_decorations"), "-v", "--no-color")
		Eventually(session).Should(gexec.Exit(1))
		Ω(session).Should(gbytes.Say("Invalid Combination of Decorators: Focused and Pending"))
	})
})
