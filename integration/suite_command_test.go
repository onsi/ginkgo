package integration_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("After Run Hook Specs", func() {
	BeforeEach(func() {
		fm.MountFixture("after_run_hook")
	})

	It("Runs command after suite echoing out suite data, properly reporting suite name and passing status in successful command output", func() {
		command := "-after-run-hook=echo THIS IS A (ginkgo-suite-passed) TEST OF THE (ginkgo-suite-name) SYSTEM, THIS IS ONLY A TEST"
		expected := "THIS IS A [PASS] TEST OF THE after_run_hook SYSTEM, THIS IS ONLY A TEST"
		session := startGinkgo(fm.PathTo("after_run_hook"), command)
		Eventually(session).Should(gexec.Exit(0))
		output := string(session.Out.Contents())

		Ω(output).Should(ContainSubstring("1 Passed"))
		Ω(output).Should(ContainSubstring("0 Failed"))
		Ω(output).Should(ContainSubstring("1 Pending"))
		Ω(output).Should(ContainSubstring("0 Skipped"))
		Ω(output).Should(ContainSubstring("Test Suite Passed"))
		Ω(output).Should(ContainSubstring("After-run-hook succeeded:"))
		Ω(output).Should(ContainSubstring(expected))
	})

	It("Runs command after suite reporting that command failed", func() {
		command := "-after-run-hook=exit 1"
		session := startGinkgo(fm.PathTo("after_run_hook"), command)
		Eventually(session).Should(gexec.Exit(0))
		output := string(session.Out.Contents())

		Ω(output).Should(ContainSubstring("1 Passed"))
		Ω(output).Should(ContainSubstring("0 Failed"))
		Ω(output).Should(ContainSubstring("1 Pending"))
		Ω(output).Should(ContainSubstring("0 Skipped"))
		Ω(output).Should(ContainSubstring("Test Suite Passed"))
		Ω(output).Should(ContainSubstring("After-run-hook failed:"))
	})

	It("Runs command after suite echoing out suite data, properly reporting suite name and failing status in successful command output", func() {
		command := "-after-run-hook=echo THIS IS A (ginkgo-suite-passed) TEST OF THE (ginkgo-suite-name) SYSTEM, THIS IS ONLY A TEST"
		expected := "THIS IS A [FAIL] TEST OF THE after_run_hook SYSTEM, THIS IS ONLY A TEST"
		session := startGinkgo(fm.PathTo("after_run_hook"), "-fail-on-pending=true", command)
		Eventually(session).Should(gexec.Exit(1))
		output := string(session.Out.Contents())

		Ω(output).Should(ContainSubstring("1 Passed"))
		Ω(output).Should(ContainSubstring("0 Failed"))
		Ω(output).Should(ContainSubstring("1 Pending"))
		Ω(output).Should(ContainSubstring("0 Skipped"))
		Ω(output).Should(ContainSubstring("Test Suite Failed"))
		Ω(output).Should(ContainSubstring("After-run-hook succeeded:"))
		Ω(output).Should(ContainSubstring(expected))
	})

})
