package integration_test

import (
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

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
			lines := strings.Split(string(session.Out.Contents()), "\n")
			randomSeeds := []string{}
			for _, line := range lines {
				if strings.Contains(line, "Random Seed:") {
					randomSeeds = append(randomSeeds, strings.Split(line, ": ")[1])
				}
			}
			Ω(randomSeeds[0]).ShouldNot(Equal(randomSeeds[1]))
			Ω(randomSeeds[1]).ShouldNot(Equal(randomSeeds[2]))
			Ω(randomSeeds[0]).ShouldNot(Equal(randomSeeds[2]))
		})
	})
})
