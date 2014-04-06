package integration_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"strings"
)

var _ = Describe("SuiteSetup", func() {
	var pathToTest string

	Context("when the BeforeSuite and AfterSuite pass", func() {
		BeforeEach(func() {
			pathToTest = tmpPath("suite_setup")
			copyIn("passing_suite_setup", pathToTest)
		})

		It("should run the BeforeSuite once, then run all the tests", func() {
			output, err := runGinkgo(pathToTest, "--noColor")
			Ω(err).ShouldNot(HaveOccurred())
			Ω(strings.Count(output, "BEFORE SUITE")).Should(Equal(1))
			Ω(strings.Count(output, "AFTER SUITE")).Should(Equal(1))
		})

		It("should run the BeforeSuite once per parallel node, then run all the tests", func() {
			output, err := runGinkgo(pathToTest, "--noColor", "--nodes=2")
			Ω(err).ShouldNot(HaveOccurred())
			Ω(strings.Count(output, "BEFORE SUITE")).Should(Equal(2))
			Ω(strings.Count(output, "AFTER SUITE")).Should(Equal(2))
		})
	})

	Context("when the BeforeSuite fails", func() {
		BeforeEach(func() {
			pathToTest = tmpPath("suite_setup")
			copyIn("failing_before_suite", pathToTest)
		})

		It("should run the BeforeSuite once, none of the tests, but it should run the AfterSuite", func() {
			output, err := runGinkgo(pathToTest, "--noColor")
			Ω(err).Should(HaveOccurred())
			Ω(strings.Count(output, "BEFORE SUITE")).Should(Equal(1))
			Ω(strings.Count(output, "Test Panicked")).Should(Equal(1))
			Ω(strings.Count(output, "AFTER SUITE")).Should(Equal(1))
			Ω(output).ShouldNot(ContainSubstring("NEVER SEE THIS"))
		})

		It("should run the BeforeSuite once per parallel node, none of the tests, but it should run the AfterSuite for each node", func() {
			output, err := runGinkgo(pathToTest, "--noColor", "--nodes=2")
			Ω(err).Should(HaveOccurred())
			Ω(strings.Count(output, "BEFORE SUITE")).Should(Equal(2))
			Ω(strings.Count(output, "Test Panicked")).Should(Equal(2))
			Ω(strings.Count(output, "AFTER SUITE")).Should(Equal(2))
			Ω(output).ShouldNot(ContainSubstring("NEVER SEE THIS"))
		})
	})

	Context("when the AfterSuite fails", func() {
		BeforeEach(func() {
			pathToTest = tmpPath("suite_setup")
			copyIn("failing_after_suite", pathToTest)
		})

		It("should run the BeforeSuite once, none of the tests, but it should run the AfterSuite", func() {
			output, err := runGinkgo(pathToTest, "--noColor")
			Ω(err).Should(HaveOccurred())
			Ω(strings.Count(output, "BEFORE SUITE")).Should(Equal(1))
			Ω(strings.Count(output, "AFTER SUITE")).Should(Equal(1))
			Ω(strings.Count(output, "Test Panicked")).Should(Equal(1))
			Ω(strings.Count(output, "A TEST")).Should(Equal(2))
		})

		It("should run the BeforeSuite once per parallel node, none of the tests, but it should run the AfterSuite for each node", func() {
			output, err := runGinkgo(pathToTest, "--noColor", "--nodes=2")
			Ω(err).Should(HaveOccurred())
			Ω(strings.Count(output, "BEFORE SUITE")).Should(Equal(2))
			Ω(strings.Count(output, "AFTER SUITE")).Should(Equal(2))
			Ω(strings.Count(output, "Test Panicked")).Should(Equal(2))
			Ω(strings.Count(output, "A TEST")).Should(Equal(2))
		})
	})
})
