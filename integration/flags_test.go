package integration_test

import (
	"strings"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/types"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Flags Specs", func() {
	BeforeEach(func() {
		fm.MountFixture("flags")
	})

	getRandomOrders := func(output string) []int {
		return []int{strings.Index(output, "RANDOM_A"), strings.Index(output, "RANDOM_B"), strings.Index(output, "RANDOM_C")}
	}

	It("normally passes, prints out noisy pendings, does not randomize tests, and honors the programmatic focus", func() {
		session := startGinkgo(fm.PathTo("flags"), "--no-color")
		Eventually(session).Should(gexec.Exit(types.GINKGO_FOCUS_EXIT_CODE))
		output := string(session.Out.Contents())

		Ω(output).Should(ContainSubstring("10 Passed"))
		Ω(output).Should(ContainSubstring("0 Failed"))
		Ω(output).Should(ContainSubstring("1 Pending"))
		Ω(output).Should(ContainSubstring("3 Skipped"))
		Ω(output).Should(ContainSubstring("[PENDING]"))
		Ω(output).Should(ContainSubstring("marshmallow"))
		Ω(output).Should(ContainSubstring("chocolate"))
		Ω(output).Should(ContainSubstring("CUSTOM_FLAG: default"))
		Ω(output).Should(ContainSubstring("Detected Programmatic Focus - setting exit status to %d", types.GINKGO_FOCUS_EXIT_CODE))
		Ω(output).ShouldNot(ContainSubstring("smores"))
		Ω(output).ShouldNot(ContainSubstring("SLOW TEST"))
		Ω(output).ShouldNot(ContainSubstring("should honor -slow-spec-threshold"))

		orders := getRandomOrders(output)
		Ω(orders[0]).Should(BeNumerically("<", orders[1]))
		Ω(orders[1]).Should(BeNumerically("<", orders[2]))
	})

	It("should fail when there are pending tests and it is passed --fail-on-pending", func() {
		session := startGinkgo(fm.PathTo("flags"), "--no-color", "--fail-on-pending")
		Eventually(session).Should(gexec.Exit(1))
		output := string(session.Out.Contents())
		Ω(output).Should(ContainSubstring("Detected pending specs and --fail-on-pending is set"))
	})

	PIt("should fail if the test suite takes longer than the timeout", func() {
		session := startGinkgo(fm.PathTo("flags"), "--no-color", "--timeout=1ms")
		Eventually(session).Should(gexec.Exit(1))
	})

	It("should override the programmatic focus when told to focus", func() {
		session := startGinkgo(fm.PathTo("flags"), "--no-color", "--focus=smores")
		Eventually(session).Should(gexec.Exit(0))
		output := string(session.Out.Contents())

		Ω(output).Should(ContainSubstring("marshmallow"))
		Ω(output).Should(ContainSubstring("chocolate"))
		Ω(output).Should(ContainSubstring("smores"))
		Ω(output).Should(ContainSubstring("3 Passed"))
		Ω(output).Should(ContainSubstring("0 Failed"))
		Ω(output).Should(ContainSubstring("1 Pending"))
		Ω(output).Should(ContainSubstring("10 Skipped"))
	})

	It("should override the programmatic focus when told to skip", func() {
		session := startGinkgo(fm.PathTo("flags"), "--no-color", "--skip=marshmallow|failing|flaky")
		Eventually(session).Should(gexec.Exit(0))
		output := string(session.Out.Contents())

		Ω(output).ShouldNot(ContainSubstring("marshmallow"))
		Ω(output).Should(ContainSubstring("chocolate"))
		Ω(output).Should(ContainSubstring("smores"))
		Ω(output).Should(ContainSubstring("10 Passed"))
		Ω(output).Should(ContainSubstring("0 Failed"))
		Ω(output).Should(ContainSubstring("1 Pending"))
		Ω(output).Should(ContainSubstring("3 Skipped"))
	})

	It("should override the programmatic focus when told to skip (multiple options)", func() {
		session := startGinkgo(fm.PathTo("flags"), "--no-color", "--skip=marshmallow", "--skip=failing", "--skip=flaky")
		Eventually(session).Should(gexec.Exit(0))
		output := string(session.Out.Contents())

		Ω(output).ShouldNot(ContainSubstring("marshmallow"))
		Ω(output).Should(ContainSubstring("chocolate"))
		Ω(output).Should(ContainSubstring("smores"))
		Ω(output).Should(ContainSubstring("10 Passed"))
		Ω(output).Should(ContainSubstring("0 Failed"))
		Ω(output).Should(ContainSubstring("1 Pending"))
		Ω(output).Should(ContainSubstring("3 Skipped"))
	})

	It("should ignore empty skip and focus variables", func() {
		session := startGinkgo(fm.PathTo("flags"), "--noColor", "--skip=", "--focus=")
		Eventually(session).Should(gexec.Exit(types.GINKGO_FOCUS_EXIT_CODE))
		output := string(session.Out.Contents())
		Ω(output).Should(ContainSubstring("marshmallow"))
		Ω(output).Should(ContainSubstring("chocolate"))
	})

	It("should run the race detector when told to", func() {
		if !raceDetectorSupported() {
			Skip("race detection is not supported")
		}
		session := startGinkgo(fm.PathTo("flags"), "--no-color", "--race")
		Eventually(session).Should(gexec.Exit(types.GINKGO_FOCUS_EXIT_CODE))
		output := string(session.Out.Contents())

		Ω(output).Should(ContainSubstring("WARNING: DATA RACE"))
	})

	It("should randomize tests when told to", func() {
		session := startGinkgo(fm.PathTo("flags"), "--no-color", "--randomize-all", "--seed=1")
		Eventually(session).Should(gexec.Exit(types.GINKGO_FOCUS_EXIT_CODE))
		output := string(session.Out.Contents())

		orders := getRandomOrders(output)
		Ω(orders[0]).ShouldNot(BeNumerically("<", orders[1]))
	})

	It("should watch for slow specs", func() {
		session := startGinkgo(fm.PathTo("flags"), "--slow-spec-threshold=0.05")
		Eventually(session).Should(gexec.Exit(types.GINKGO_FOCUS_EXIT_CODE))
		output := string(session.Out.Contents())

		Ω(output).Should(ContainSubstring("SLOW TEST"))
		Ω(output).Should(ContainSubstring("should honor -slow-spec-threshold"))
	})

	It("should pass additional arguments in", func() {
		session := startGinkgo(fm.PathTo("flags"), "--", "--customFlag=madagascar")
		Eventually(session).Should(gexec.Exit(types.GINKGO_FOCUS_EXIT_CODE))
		output := string(session.Out.Contents())

		Ω(output).Should(ContainSubstring("CUSTOM_FLAG: madagascar"))
	})

	It("should print out full stack traces for failures when told to", func() {
		session := startGinkgo(fm.PathTo("flags"), "--focus=a failing test", "--trace")
		Eventually(session).Should(gexec.Exit(1))
		output := string(session.Out.Contents())

		Ω(output).Should(ContainSubstring("Full Stack Trace"))
	})

	It("should fail fast when told to", func() {
		fm.MountFixture("fail")
		session := startGinkgo(fm.PathTo("fail"), "--fail-fast")
		Eventually(session).Should(gexec.Exit(1))
		output := string(session.Out.Contents())

		Ω(output).Should(ContainSubstring("1 Failed"))
		Ω(output).Should(ContainSubstring("6 Skipped"))
	})

	Context("with a flaky test", func() {
		It("should normally fail", func() {
			session := startGinkgo(fm.PathTo("flags"), "--focus=flaky")
			Eventually(session).Should(gexec.Exit(1))
		})

		It("should pass if retries are requested", func() {
			session := startGinkgo(fm.PathTo("flags"), "--focus=flaky --flake-attempts=2")
			Eventually(session).Should(gexec.Exit(0))
		})
	})

	It("should perform a dry run when told to", func() {
		fm.MountFixture("fail")
		session := startGinkgo(fm.PathTo("fail"), "--dry-run", "-v")
		Eventually(session).Should(gexec.Exit(0))
		output := string(session.Out.Contents())

		Ω(output).Should(ContainSubstring("synchronous failures"))
		Ω(output).Should(ContainSubstring("7 Specs"))
		Ω(output).Should(ContainSubstring("7 Passed"))
		Ω(output).Should(ContainSubstring("0 Failed"))
	})

	regextest := func(regexOption string, skipOrFocus string) string {
		fm.MountFixture("passing_ginkgo_tests")
		session := startGinkgo(fm.PathTo("passing_ginkgo_tests"), regexOption, "--dry-run", "-v", skipOrFocus)
		Eventually(session).Should(gexec.Exit(0))
		return string(session.Out.Contents())
	}

	It("regexScansFilePath (enabled) should skip and focus on file names", func() {
		output := regextest("-regexScansFilePath=true", "-skip=/passing") // everything gets skipped (nothing runs)
		Ω(output).Should(ContainSubstring("0 of 4 Specs"))
		output = regextest("-regexScansFilePath=true", "-focus=/passing") // everything gets focused (everything runs)
		Ω(output).Should(ContainSubstring("4 of 4 Specs"))
	})

	It("regexScansFilePath (disabled) should not effect normal filtering", func() {
		output := regextest("-regexScansFilePath=false", "-skip=/passing") // nothing gets skipped (everything runs)
		Ω(output).Should(ContainSubstring("4 of 4 Specs"))
		output = regextest("-regexScansFilePath=false", "-focus=/passing") // nothing gets focused (nothing runs)
		Ω(output).Should(ContainSubstring("0 of 4 Specs"))
	})

	It("should honor compiler flags", func() {
		session := startGinkgo(fm.PathTo("flags"), "-gcflags=-importmap 'math=math/cmplx'")
		Eventually(session).Should(gexec.Exit(types.GINKGO_FOCUS_EXIT_CODE))
		output := string(session.Out.Contents())
		Ω(output).Should(ContainSubstring("NaN returns complex128"))
	})
})
