package integration_test

import (
	"os"
	"path/filepath"
	"strings"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Flags Specs", func() {
	var pathToTest string

	BeforeEach(func() {
		pathToTest = tmpPath("flags")
		copyIn("flags_tests", pathToTest)
	})

	getRandomOrders := func(output string) []int {
		return []int{strings.Index(output, "RANDOM_A"), strings.Index(output, "RANDOM_B"), strings.Index(output, "RANDOM_C")}
	}

	It("normally passes, runs measurements, prints out noisy pendings, does not randomize tests, and honors the programmatic focus", func() {
		output, err := runGinkgo(pathToTest, "--noColor")
		Ω(err).ShouldNot(HaveOccurred())
		Ω(output).Should(ContainSubstring("Ran 3 samples:"), "has a measurement")
		Ω(output).Should(ContainSubstring("10 Passed"))
		Ω(output).Should(ContainSubstring("0 Failed"))
		Ω(output).Should(ContainSubstring("1 Pending"))
		Ω(output).Should(ContainSubstring("2 Skipped"))
		Ω(output).Should(ContainSubstring("[PENDING]"))
		Ω(output).Should(ContainSubstring("marshmallow"))
		Ω(output).Should(ContainSubstring("chocolate"))
		Ω(output).Should(ContainSubstring("CUSTOM_FLAG: default"))
		Ω(output).ShouldNot(ContainSubstring("smores"))
		Ω(output).ShouldNot(ContainSubstring("SLOW TEST"))
		Ω(output).ShouldNot(ContainSubstring("should honor -slowSpecThreshold"))

		orders := getRandomOrders(output)
		Ω(orders[0]).Should(BeNumerically("<", orders[1]))
		Ω(orders[1]).Should(BeNumerically("<", orders[2]))
	})

	It("should run a coverprofile when passed -cover", func() {
		output, err := runGinkgo(pathToTest, "--noColor", "--cover")
		Ω(err).ShouldNot(HaveOccurred())
		_, err = os.Stat(filepath.Join(pathToTest, "flags.coverprofile"))
		Ω(err).ShouldNot(HaveOccurred())
		Ω(output).Should(ContainSubstring("coverage: "))
	})

	It("should fail when there are pending tests and it is passed --failOnPending", func() {
		_, err := runGinkgo(pathToTest, "--noColor", "--failOnPending")
		Ω(err).Should(HaveOccurred())
	})

	It("should not print out pendings when --noisyPendings=false", func() {
		output, err := runGinkgo(pathToTest, "--noColor", "--noisyPendings=false")
		Ω(err).ShouldNot(HaveOccurred())
		Ω(output).ShouldNot(ContainSubstring("[PENDING]"))
		Ω(output).Should(ContainSubstring("1 Pending"))
	})

	It("should override the programmatic focus when told to focus", func() {
		output, err := runGinkgo(pathToTest, "--noColor", "--focus=smores")
		Ω(err).ShouldNot(HaveOccurred())
		Ω(output).Should(ContainSubstring("marshmallow"))
		Ω(output).Should(ContainSubstring("chocolate"))
		Ω(output).Should(ContainSubstring("smores"))
		Ω(output).Should(ContainSubstring("3 Passed"))
		Ω(output).Should(ContainSubstring("0 Failed"))
		Ω(output).Should(ContainSubstring("0 Pending"))
		Ω(output).Should(ContainSubstring("10 Skipped"))
	})

	It("should override the programmatic focus when told to skip", func() {
		output, err := runGinkgo(pathToTest, "--noColor", "--skip=marshmallow|failing")
		Ω(err).ShouldNot(HaveOccurred())
		Ω(output).ShouldNot(ContainSubstring("marshmallow"))
		Ω(output).Should(ContainSubstring("chocolate"))
		Ω(output).Should(ContainSubstring("smores"))
		Ω(output).Should(ContainSubstring("10 Passed"))
		Ω(output).Should(ContainSubstring("0 Failed"))
		Ω(output).Should(ContainSubstring("1 Pending"))
		Ω(output).Should(ContainSubstring("2 Skipped"))
	})

	It("should run the race detector when told to", func() {
		output, err := runGinkgo(pathToTest, "--noColor", "--race")
		Ω(err).Should(HaveOccurred())
		Ω(output).Should(ContainSubstring("WARNING: DATA RACE"))
	})

	It("should randomize tests when told to", func() {
		output, err := runGinkgo(pathToTest, "--noColor", "--randomizeAllSpecs", "--seed=21")
		Ω(err).ShouldNot(HaveOccurred())
		orders := getRandomOrders(output)
		Ω(orders[0]).ShouldNot(BeNumerically("<", orders[1]))
	})

	It("should skip measurements when told to", func() {
		output, err := runGinkgo(pathToTest, "--skipMeasurements")
		Ω(err).ShouldNot(HaveOccurred())
		Ω(output).ShouldNot(ContainSubstring("Ran 3 samples:"), "has a measurement")
		Ω(output).Should(ContainSubstring("3 Skipped"))
	})

	It("should watch for slow specs", func() {
		output, err := runGinkgo(pathToTest, "--slowSpecThreshold=0.05")
		Ω(err).ShouldNot(HaveOccurred())
		Ω(output).Should(ContainSubstring("SLOW TEST"))
		Ω(output).Should(ContainSubstring("should honor -slowSpecThreshold"))
	})

	It("should pass additional arguments in", func() {
		output, err := runGinkgo(pathToTest, "--", "--customFlag=madagascar")
		Ω(err).ShouldNot(HaveOccurred())
		Ω(output).Should(ContainSubstring("CUSTOM_FLAG: madagascar"))
	})

	It("should print out full stack traces for failures when told to", func() {
		output, err := runGinkgo(pathToTest, "--focus=a failing test", "--trace")
		Ω(err).Should(HaveOccurred())
		Ω(output).Should(ContainSubstring("Full Stack Trace"))
	})
})
