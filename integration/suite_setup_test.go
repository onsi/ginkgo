package integration_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("SuiteSetup", func() {
	Context("With passing synchronized before and after suites", func() {
		BeforeEach(func() {
			fm.MountFixture("synchronized_setup_tests")
		})

		Context("when run with one node", func() {
			It("should do all the work on that one node", func() {
				session := startGinkgo(fm.PathTo("synchronized_setup_tests"), "--no-color")
				Eventually(session).Should(gexec.Exit(0))
				output := string(session.Out.Contents())

				Ω(output).Should(ContainSubstring("BEFORE_A_1\nBEFORE_B_1: DATA"))
				Ω(output).Should(ContainSubstring("AFTER_A_1\nAFTER_B_1"))
			})
		})

		Context("when run across multiple nodes", func() {
			It("should run the first BeforeSuite function (BEFORE_A) on node 1, the second (BEFORE_B) on all the nodes, the first AfterSuite (AFTER_A) on all the nodes, and then the second (AFTER_B) on Node 1 *after* everything else is finished", func() {
				session := startGinkgo(fm.PathTo("synchronized_setup_tests"), "--no-color", "--nodes=3")
				Eventually(session).Should(gexec.Exit(0))
				output := string(session.Out.Contents())

				Ω(output).Should(ContainSubstring("BEFORE_A_1"))
				Ω(output).Should(ContainSubstring("BEFORE_B_1: DATA"))
				Ω(output).Should(ContainSubstring("BEFORE_B_2: DATA"))
				Ω(output).Should(ContainSubstring("BEFORE_B_3: DATA"))

				Ω(output).ShouldNot(ContainSubstring("BEFORE_A_2"))
				Ω(output).ShouldNot(ContainSubstring("BEFORE_A_3"))

				Ω(output).Should(ContainSubstring("AFTER_A_1"))
				Ω(output).Should(ContainSubstring("AFTER_A_2"))
				Ω(output).Should(ContainSubstring("AFTER_A_3"))
				Ω(output).Should(ContainSubstring("AFTER_B_1"))

				Ω(output).ShouldNot(ContainSubstring("AFTER_B_2"))
				Ω(output).ShouldNot(ContainSubstring("AFTER_B_3"))
			})
		})
	})

	Context("With a failing synchronized before suite", func() {
		BeforeEach(func() {
			fm.MountFixture("exiting_synchronized_setup")
		})

		It("should fail and let the user know that node 1 disappeared prematurely", func() {
			session := startGinkgo(fm.PathTo("exiting_synchronized_setup"), "--no-color", "--nodes=3")
			Eventually(session).Should(gexec.Exit(1))
			output := string(session.Out.Contents()) + string(session.Err.Contents())

			Ω(output).Should(ContainSubstring("SynchronizedBeforeSuite on Node 1 disappeared before it could report back"))
			Ω(output).Should(ContainSubstring("Ginkgo timed out waiting for all parallel nodes to report back"))
		})
	})
})
