package integration_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"strings"
)

var _ = Describe("BeforeSuite", func() {
	var pathToTest string

	Context("when the BeforeSuite passes", func() {
		BeforeEach(func() {
			pathToTest = tmpPath("before_suite")
			copyIn("passing_before_suite", pathToTest)
		})

		It("should run the BeforeSuite once, then run all the tests", func() {
			output, err := runGinkgo(pathToTest, "--noColor")
			Ω(err).ShouldNot(HaveOccurred())
			Ω(strings.Count(output, "BEFORE SUITE")).Should(Equal(1))
		})

		It("should run the BeforeSuite once per parallel node, then run all the tests", func() {
			output, err := runGinkgo(pathToTest, "--noColor", "--nodes=2")
			Ω(err).ShouldNot(HaveOccurred())
			Ω(strings.Count(output, "BEFORE SUITE")).Should(Equal(2))
		})
	})

	Context("when the BeforeSuite fails", func() {
		BeforeEach(func() {
			pathToTest = tmpPath("before_suite")
			copyIn("failing_before_suite", pathToTest)
		})

		It("should run the BeforeSuite once, then run all the tests", func() {
			output, err := runGinkgo(pathToTest, "--noColor")
			Ω(err).Should(HaveOccurred())
			Ω(strings.Count(output, "BEFORE SUITE")).Should(Equal(1))
			Ω(strings.Count(output, "Test Panicked")).Should(Equal(1))
			Ω(output).ShouldNot(ContainSubstring("NEVER SEE THIS"))
		})

		It("should run the BeforeSuite once per parallel node, then run all the tests", func() {
			output, err := runGinkgo(pathToTest, "--noColor", "--nodes=2")
			Ω(err).Should(HaveOccurred())
			Ω(strings.Count(output, "BEFORE SUITE")).Should(Equal(2))
			Ω(strings.Count(output, "Test Panicked")).Should(Equal(2))
			Ω(output).ShouldNot(ContainSubstring("NEVER SEE THIS"))
		})
	})
})
