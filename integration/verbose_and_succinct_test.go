package integration_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Verbose And Succinct Mode", func() {
	var pathToTest string
	var otherPathToTest string

	Context("when running one package", func() {
		BeforeEach(func() {
			pathToTest = tmpPath("ginkgo")
			copyIn("passing_ginkgo_tests", pathToTest)
		})

		It("should default to non-succinct mode", func() {
			output, err := runGinkgo(pathToTest, "--noColor")
			Ω(err).ShouldNot(HaveOccurred())
			Ω(output).Should(ContainSubstring("Running Suite: Passing_ginkgo_tests Suite"))
		})
	})

	Context("when running more than one package", func() {
		BeforeEach(func() {
			pathToTest = tmpPath("ginkgo")
			copyIn("passing_ginkgo_tests", pathToTest)
			otherPathToTest = tmpPath("more_ginkgo")
			copyIn("more_ginkgo_tests", otherPathToTest)
		})

		Context("with no flags set", func() {
			It("should default to succinct mode", func() {
				output, err := runGinkgo(pathToTest, "--noColor", pathToTest, otherPathToTest)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(output).Should(ContainSubstring("] Passing_ginkgo_tests Suite - 3/3 specs ••• SUCCESS!"))
				Ω(output).Should(ContainSubstring("] More_ginkgo_tests Suite - 2/2 specs •• SUCCESS!"))
			})
		})

		Context("with --succinct=false", func() {
			It("should not be in succinct mode", func() {
				output, err := runGinkgo(pathToTest, "--noColor", "--succinct=false", pathToTest, otherPathToTest)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(output).Should(ContainSubstring("Running Suite: Passing_ginkgo_tests Suite"))
				Ω(output).Should(ContainSubstring("Running Suite: More_ginkgo_tests Suite"))
			})
		})

		Context("with -v", func() {
			It("should not be in succinct mode, but should be verbose", func() {
				output, err := runGinkgo(pathToTest, "--noColor", "-v", pathToTest, otherPathToTest)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(output).Should(ContainSubstring("Running Suite: Passing_ginkgo_tests Suite"))
				Ω(output).Should(ContainSubstring("Running Suite: More_ginkgo_tests Suite"))
				Ω(output).Should(ContainSubstring("should proxy strings"))
				Ω(output).Should(ContainSubstring("should always pass"))
			})
		})
	})
})
