package integration_test

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"
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

		Ω(output).Should(ContainSubstring("9 Passed"))
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

	It("should run the race detector when told to", Label("slow"), func() {
		if !raceDetectorSupported() {
			Skip("race detection is not supported")
		}
		session := startGinkgo(fm.PathTo("flags"), "--no-color", "--race")
		Eventually(session).Should(gexec.Exit(types.GINKGO_FOCUS_EXIT_CODE))
		output := string(session.Out.Contents())

		Ω(output).Should(ContainSubstring("WARNING: DATA RACE"))
	})

	It("should randomize tests when told to", func() {
		session := startGinkgo(fm.PathTo("flags"), "--no-color", "--randomize-all", "--seed=40")
		Eventually(session).Should(gexec.Exit(types.GINKGO_FOCUS_EXIT_CODE))
		output := string(session.Out.Contents())

		orders := getRandomOrders(output)
		Ω(orders[0]).ShouldNot(BeNumerically("<", orders[1]))
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

	Describe("--fail-fast", func() {
		BeforeEach(func() {
			fm.MountFixture("fail_then_hang")
		})

		Context("when running in series", func() {
			It("should fail fast when told to", func() {
				session := startGinkgo(fm.PathTo("fail_then_hang"), "--fail-fast")
				Eventually(session).Should(gexec.Exit(1))
				output := string(session.Out.Contents())

				Ω(output).Should(ContainSubstring("1 Failed"))
				Ω(output).Should(ContainSubstring("2 Skipped"))
			})
		})

		Context("when running in parallel", func() {
			It("should fail fast when told to", func() {
				session := startGinkgo(fm.PathTo("fail_then_hang"), "--fail-fast", "--procs=2")
				Eventually(session).Should(gexec.Exit(1))
				output := string(session.Out.Contents())

				Ω(output).Should(ContainSubstring("2 Failed")) //one fails, the other is interrupted
				Ω(output).Should(ContainSubstring("1 Skipped"))
			})
		})
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
		Ω(output).Should(ContainSubstring("8 Specs"))
		Ω(output).Should(ContainSubstring("8 Passed"))
		Ω(output).Should(ContainSubstring("0 Failed"))
	})

	It("should allow configuration overrides", func() {
		fm.MountFixture("config_override")
		session := startGinkgo(fm.PathTo("config_override"), "--label-filter=NORUN", "--no-color")
		Eventually(session).Should(gexec.Exit(0), "Succeeds because --label-filter is overridden by the test suite itself.")
		output := string(session.Out.Contents())
		Ω(output).Should(ContainSubstring("2 Specs"))
		Ω(output).Should(ContainSubstring("1 Skipped"))
		Ω(output).Should(ContainSubstring("1 Passed"))
	})

	It("should emit node start/end events when running with --show-node-events", func() {
		session := startGinkgo(fm.PathTo("flags"), "--no-color", "-v", "--show-node-events")
		Eventually(session).Should(gexec.Exit(types.GINKGO_FOCUS_EXIT_CODE))
		output := string(session.Out.Contents())

		Eventually(output).Should(ContainSubstring("> Enter [It] should honor -cover"))
		Eventually(output).Should(ContainSubstring("< Exit [It] should honor -cover"))

		fm.MountFixture("fail")
		session = startGinkgo(fm.PathTo("fail"), "--no-color", "--show-node-events")
		Eventually(session).Should(gexec.Exit(1))
		output = string(session.Out.Contents())
		Ω(output).Should(ContainSubstring("> Enter [It] a top level specify"))
		Ω(output).Should(ContainSubstring("< Exit [It] a top level specify"))

		session = startGinkgo(fm.PathTo("fail"), "--no-color")
		Eventually(session).Should(gexec.Exit(1))
		output = string(session.Out.Contents())
		Ω(output).ShouldNot(ContainSubstring("> Enter [It] a top level specify"))
		Ω(output).ShouldNot(ContainSubstring("< Exit [It] a top level specify"))
	})
})
