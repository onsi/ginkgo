package integration_test

import (
	"os"
	"os/exec"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Coverage Specs", func() {
	AfterEach(func() {
		os.RemoveAll("./_fixtures/coverage_fixture/coverage_fixture.coverprofile")
	})

	It("runs coverage analysis in series and in parallel", func() {
		output, err := runGinkgo("./_fixtures/coverage_fixture", "-cover")
		Ω(err).ShouldNot(HaveOccurred())
		Ω(output).Should(ContainSubstring("coverage: 80.0% of statements"))

		serialCoverProfileOutput, err := exec.Command("go", "tool", "cover", "-func=./_fixtures/coverage_fixture/coverage_fixture.coverprofile").CombinedOutput()
		Ω(err).ShouldNot(HaveOccurred())

		os.RemoveAll("./_fixtures/coverage_fixture/coverage_fixture.coverprofile")

		output, err = runGinkgo("./_fixtures/coverage_fixture", "-cover", "-nodes=4")
		Ω(err).ShouldNot(HaveOccurred())

		parallelCoverProfileOutput, err := exec.Command("go", "tool", "cover", "-func=./_fixtures/coverage_fixture/coverage_fixture.coverprofile").CombinedOutput()
		Ω(err).ShouldNot(HaveOccurred())

		Ω(parallelCoverProfileOutput).Should(Equal(serialCoverProfileOutput))
	})
})
