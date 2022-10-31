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

	It("processes the Offset, Focus and Pending decorations", func() {
		session := startGinkgo(fm.PathTo("decorations", "offset_focus_pending"), "-vv", "--no-color")
		Eventually(session).Should(gexec.Exit(types.GINKGO_FOCUS_EXIT_CODE))

		out := string(session.Out.Contents())
		Ω(out).Should(MatchRegexp(
			`P \[PENDING\]
some decorated specs
.*offset_focus_pending_fixture_suite_test.go:\d+
  pending it`,
		))

		Ω(out).ShouldNot(ContainSubstring("never_see_this_file"))
	})

	It("processes the FlakeAttempts and the MustPassRepeatedly decorations", func() {
		session := startGinkgo(fm.PathTo("decorations", "flaky_repeated"), "-vv", "--no-color")
		Eventually(session).Should(gexec.Exit(1))

		Ω(session).Should(gbytes.Say("Attempt #1 Failed.  Retrying"))
		Ω(session).Should(gbytes.Say("Attempt #2 Failed.  Retrying"))

		Ω(session).Should(gbytes.Say("Attempt #1 Passed.  Repeating"))
		Ω(session).Should(gbytes.Say("Attempt #2 Passed.  Repeating"))
		Ω(session).Should(gbytes.Say("failed on attempt #3"))
	})

	It("exits with a clear error if decorations are misconfigured - focus and pending error", func() {
		session := startGinkgo(fm.PathTo("decorations", "invalid_decorations_focused_pending"), "-v", "--no-color")
		Eventually(session).Should(gexec.Exit(1))
		Ω(session).Should(gbytes.Say("Invalid Combination of Decorators: Focused and Pending"))
	})

	It("exits with a clear error if decorations are misconfigured - flakeattempts and mustpassrepeatedly error", func() {
		session := startGinkgo(fm.PathTo("decorations", "invalid_decorations_flakeattempts_mustpassrepeatedly"), "-v", "--no-color")
		Eventually(session).Should(gexec.Exit(1))
		Ω(session).Should(gbytes.Say("Invalid Combination of Decorators: FlakeAttempts and MustPassRepeatedly"))
	})
})
