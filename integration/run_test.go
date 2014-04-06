package integration_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"strings"
)

var _ = Describe("Running Specs", func() {
	var pathToTest string

	Context("when pointed at the current directory", func() {
		BeforeEach(func() {
			pathToTest = tmpPath("ginkgo")
			copyIn("passing_ginkgo_tests", pathToTest)
		})

		It("should run the tests in the working directory", func() {
			output, err := runGinkgo(pathToTest, "--noColor")

			Ω(err).ShouldNot(HaveOccurred())
			Ω(output).Should(ContainSubstring("Running Suite: Passing_ginkgo_tests Suite"))
			Ω(output).Should(ContainSubstring("•••"))
			Ω(output).Should(ContainSubstring("SUCCESS! -- 3 Passed"))
			Ω(output).Should(ContainSubstring("Test Suite Passed"))
		})
	})

	Context("when passed an explicit package to run", func() {
		BeforeEach(func() {
			pathToTest = tmpPath("ginkgo")
			copyIn("passing_ginkgo_tests", pathToTest)
		})

		It("should run the ginkgo style tests", func() {
			output, err := runGinkgo(tmpDir, "--noColor", pathToTest)

			Ω(err).ShouldNot(HaveOccurred())
			Ω(output).Should(ContainSubstring("Running Suite: Passing_ginkgo_tests Suite"))
			Ω(output).Should(ContainSubstring("•••"))
			Ω(output).Should(ContainSubstring("SUCCESS! -- 3 Passed"))
			Ω(output).Should(ContainSubstring("Test Suite Passed"))
		})
	})

	Context("when passed a number of packages to run", func() {
		BeforeEach(func() {
			pathToTest = tmpPath("ginkgo")
			otherPathToTest := tmpPath("other")
			copyIn("passing_ginkgo_tests", pathToTest)
			copyIn("more_ginkgo_tests", otherPathToTest)
		})

		It("should run the ginkgo style tests", func() {
			output, err := runGinkgo(tmpDir, "--noColor", "--succinct=false", "ginkgo", "./other")

			Ω(err).ShouldNot(HaveOccurred())
			Ω(output).Should(ContainSubstring("Running Suite: Passing_ginkgo_tests Suite"))
			Ω(output).Should(ContainSubstring("Running Suite: More_ginkgo_tests Suite"))
			Ω(output).Should(ContainSubstring("Test Suite Passed"))
		})
	})

	Context("when told to skipPackages", func() {
		BeforeEach(func() {
			pathToTest = tmpPath("ginkgo")
			otherPathToTest := tmpPath("other")
			copyIn("passing_ginkgo_tests", pathToTest)
			copyIn("more_ginkgo_tests", otherPathToTest)
		})

		It("should skip packages that match the regexp", func() {
			output, err := runGinkgo(tmpDir, "--noColor", "--skipPackage=other", "-r")

			Ω(err).ShouldNot(HaveOccurred())
			Ω(output).Should(ContainSubstring("Passing_ginkgo_tests Suite"))
			Ω(output).ShouldNot(ContainSubstring("More_ginkgo_tests Suite"))
			Ω(output).Should(ContainSubstring("Test Suite Passed"))
		})
	})

	Context("when pointed at a package with xunit style tests", func() {
		BeforeEach(func() {
			pathToTest = tmpPath("xunit")
			copyIn("xunit_tests", pathToTest)
		})

		It("should run the xunit style tests", func() {
			output, err := runGinkgo(pathToTest)

			Ω(err).ShouldNot(HaveOccurred())
			Ω(output).Should(ContainSubstring("--- PASS: TestAlwaysTrue"))
			Ω(output).Should(ContainSubstring("Test Suite Passed"))
		})
	})

	Context("when pointed at a package with no tests", func() {
		BeforeEach(func() {
			pathToTest = tmpPath("no_tests")
			copyIn("no_tests", pathToTest)
		})

		It("should fail", func() {
			output, err := runGinkgo(pathToTest, "--noColor")

			Ω(err).Should(HaveOccurred())
			Ω(output).Should(ContainSubstring("Found no test suites"))
		})
	})

	Context("when pointed at a package that fails to compile", func() {
		BeforeEach(func() {
			pathToTest = tmpPath("does_not_compile")
			copyIn("does_not_compile", pathToTest)
		})

		It("should fail", func() {
			output, err := runGinkgo(pathToTest, "--noColor")

			Ω(err).Should(HaveOccurred())
			Ω(output).Should(ContainSubstring("Failed to compile"))
		})
	})

	Context("when running in parallel", func() {
		BeforeEach(func() {
			pathToTest = tmpPath("ginkgo")
			copyIn("passing_ginkgo_tests", pathToTest)
		})

		It("should aggregate output", func() {
			output, err := runGinkgo(pathToTest, "--noColor", "-succinct", "-nodes=2")

			Ω(err).ShouldNot(HaveOccurred())
			Ω(output).Should(MatchRegexp(`\[\d+\] Passing_ginkgo_tests Suite - 3/3 specs - 2 nodes ••• SUCCESS! [\d.mus]+`))
			Ω(output).Should(ContainSubstring("Test Suite Passed"))
		})
	})

	Context("when streaming in parallel", func() {
		BeforeEach(func() {
			pathToTest = tmpPath("ginkgo")
			copyIn("passing_ginkgo_tests", pathToTest)
		})

		It("should print output in realtime", func() {
			output, err := runGinkgo(pathToTest, "--noColor", "-stream", "-nodes=2")

			Ω(err).ShouldNot(HaveOccurred())
			Ω(output).Should(ContainSubstring(`[1] Parallel test node 1/2.`))
			Ω(output).Should(ContainSubstring(`[2] Parallel test node 2/2.`))
			Ω(output).Should(ContainSubstring(`[1] SUCCESS!`))
			Ω(output).Should(ContainSubstring(`[2] SUCCESS!`))
			Ω(output).Should(ContainSubstring("Test Suite Passed"))
		})
	})

	Context("when running recursively", func() {
		BeforeEach(func() {
			passingTest := tmpPath("A")
			otherPassingTest := tmpPath("E")
			copyIn("passing_ginkgo_tests", passingTest)
			copyIn("more_ginkgo_tests", otherPassingTest)
		})

		Context("when all the tests pass", func() {
			It("should run all the tests (in succinct mode) and succeed", func() {
				output, err := runGinkgo(tmpDir, "--noColor", "-r")

				Ω(err).ShouldNot(HaveOccurred())
				outputLines := strings.Split(output, "\n")
				Ω(outputLines[0]).Should(MatchRegexp(`\[\d+\] Passing_ginkgo_tests Suite - 3/3 specs ••• SUCCESS! [\d.mus]+ PASS`))
				Ω(outputLines[1]).Should(MatchRegexp(`\[\d+\] More_ginkgo_tests Suite - 2/2 specs •• SUCCESS! [\d.mus]+ PASS`))
				Ω(output).Should(ContainSubstring("Test Suite Passed"))
			})
		})

		Context("when one of the packages has a failing tests", func() {
			BeforeEach(func() {
				failingTest := tmpPath("C")
				copyIn("failing_ginkgo_tests", failingTest)
			})

			It("should fail and stop running tests", func() {
				output, err := runGinkgo(tmpDir, "--noColor", "-r")

				Ω(err).Should(HaveOccurred())
				outputLines := strings.Split(output, "\n")
				Ω(outputLines[0]).Should(MatchRegexp(`\[\d+\] Passing_ginkgo_tests Suite - 3/3 specs ••• SUCCESS! [\d.mus]+ PASS`))
				Ω(outputLines[1]).Should(MatchRegexp(`\[\d+\] Failing_ginkgo_tests Suite - 2/2 specs`))
				Ω(output).Should(ContainSubstring("• Failure"))
				Ω(output).ShouldNot(ContainSubstring("More_ginkgo_tests Suite"))
				Ω(output).Should(ContainSubstring("Test Suite Failed"))
			})
		})

		Context("when one of the packages fails to compile", func() {
			BeforeEach(func() {
				doesNotCompileTest := tmpPath("C")
				copyIn("does_not_compile", doesNotCompileTest)
			})

			It("should fail and stop running tests", func() {
				output, err := runGinkgo(tmpDir, "--noColor", "-r")

				Ω(err).Should(HaveOccurred())
				outputLines := strings.Split(output, "\n")
				Ω(outputLines[0]).Should(MatchRegexp(`\[\d+\] Passing_ginkgo_tests Suite - 3/3 specs ••• SUCCESS! [\d.mus]+ PASS`))
				Ω(outputLines[1]).Should(ContainSubstring("Failed to compile C:"))
				Ω(output).ShouldNot(ContainSubstring("More_ginkgo_tests Suite"))
				Ω(output).Should(ContainSubstring("Test Suite Failed"))
			})
		})

		Context("when either is the case, but the keepGoing flag is set", func() {
			BeforeEach(func() {
				doesNotCompileTest := tmpPath("B")
				copyIn("does_not_compile", doesNotCompileTest)

				failingTest := tmpPath("C")
				copyIn("failing_ginkgo_tests", failingTest)
			})

			It("should soldier on", func() {
				output, err := runGinkgo(tmpDir, "--noColor", "-r", "-keepGoing")

				Ω(err).Should(HaveOccurred())
				outputLines := strings.Split(output, "\n")
				Ω(outputLines[0]).Should(MatchRegexp(`\[\d+\] Passing_ginkgo_tests Suite - 3/3 specs ••• SUCCESS! [\d.mus]+ PASS`))
				Ω(outputLines[1]).Should(ContainSubstring("Failed to compile B:"))
				Ω(output).Should(MatchRegexp(`\[\d+\] Failing_ginkgo_tests Suite - 2/2 specs`))
				Ω(output).Should(ContainSubstring("• Failure"))
				Ω(output).Should(MatchRegexp(`\[\d+\] More_ginkgo_tests Suite - 2/2 specs •• SUCCESS! [\d.mus]+ PASS`))
				Ω(output).Should(ContainSubstring("Test Suite Failed"))
			})
		})
	})
})
