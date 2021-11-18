package integration_test

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

func extractRandomSeeds(content string) []string {
	lines := strings.Split(content, "\n")
	randomSeeds := []string{}
	for _, line := range lines {
		if strings.Contains(line, "Random Seed:") {
			randomSeeds = append(randomSeeds, strings.Split(line, ": ")[1])
		}
	}
	return randomSeeds
}

var _ = Describe("Repeat", func() {
	Context("when a test attempts to invoke RunSpecs twice", func() {
		BeforeEach(func() {
			fm.MountFixture("rerun_specs")
		})

		It("errors out and tells the user not to do that", func() {
			session := startGinkgo(fm.PathTo("rerun_specs"))
			Eventually(session).Should(gexec.Exit())
			Ω(session).Should(gbytes.Say("It looks like you are calling RunSpecs more than once."))
		})
	})

	Context("when told to keep going --until-it-fails", func() {
		BeforeEach(func() {
			fm.MountFixture("eventually_failing")
		})

		It("should keep rerunning the tests, until a failure occurs", func() {
			session := startGinkgo(fm.PathTo("eventually_failing"), "--until-it-fails", "--no-color")
			Eventually(session).Should(gexec.Exit(1))
			Ω(session).Should(gbytes.Say("This was attempt #1"))
			Ω(session).Should(gbytes.Say("This was attempt #2"))
			Ω(session).Should(gbytes.Say("Tests failed on attempt #3"))

			//it should change the random seed between each test
			randomSeeds := extractRandomSeeds(string(session.Out.Contents()))
			Ω(randomSeeds[0]).ShouldNot(Equal(randomSeeds[1]))
			Ω(randomSeeds[1]).ShouldNot(Equal(randomSeeds[2]))
			Ω(randomSeeds[0]).ShouldNot(Equal(randomSeeds[2]))
		})
	})

	Context("when told to --repeat", func() {
		BeforeEach(func() {
			fm.MountFixture("eventually_failing")
		})

		Context("when the test passes for N repetitions", func() {
			It("should run the tests N+1 times and report success", func() {
				session := startGinkgo(fm.PathTo("eventually_failing"), "--repeat=1", "--no-color")
				Eventually(session).Should(gexec.Exit(0))
				Ω(session).Should(gbytes.Say("This was attempt 1 of 2"))

				//it should change the random seed between each test
				randomSeeds := extractRandomSeeds(string(session.Out.Contents()))
				Ω(randomSeeds[0]).ShouldNot(Equal(randomSeeds[1]))
			})
		})

		Context("when the test eventually fails", func() {
			It("should report failure and stop running", func() {
				session := startGinkgo(fm.PathTo("eventually_failing"), "--repeat=3", "--no-color")
				Eventually(session).Should(gexec.Exit(1))
				Ω(session).Should(gbytes.Say("This was attempt 1 of 4"))
				Ω(session).Should(gbytes.Say("This was attempt 2 of 4"))
				Ω(session).Should(gbytes.Say("Tests failed on attempt #3"))

				//it should change the random seed between each test
				randomSeeds := extractRandomSeeds(string(session.Out.Contents()))
				Ω(randomSeeds[0]).ShouldNot(Equal(randomSeeds[1]))
				Ω(randomSeeds[1]).ShouldNot(Equal(randomSeeds[2]))
				Ω(randomSeeds[0]).ShouldNot(Equal(randomSeeds[2]))
			})
		})
	})

	Context("if both --repeat and --until-it-fails are set", func() {
		BeforeEach(func() {
			fm.MountFixture("eventually_failing")
		})

		It("errors out early", func() {
			session := startGinkgo(fm.PathTo("eventually_failing"), "--repeat=3", "--until-it-fails", "--no-color")
			Eventually(session).Should(gexec.Exit(1))
			Ω(session.Err).Should(gbytes.Say("--repeat and --until-it-fails are both set"))
		})
	})
})
