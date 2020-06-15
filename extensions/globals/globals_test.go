package globals_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/extensions/globals"
	. "github.com/onsi/gomega"
)

func TestGlobals(t *testing.T) {
	RegisterFailHandler(Fail)

	// define some vars to store how many times a test has been run
	var (
		testI  = 0
		testII = 0
	)

	// Define a simple gingko test I
	var _ = Describe("ginkgo test I", func() {
		It("build tests I", func() {
			testI++
			Ω(testI).Should(Equal(1))
		})
	})

	RunSpecs(t, "Test Runner Suite I")

	// reset the global state of ginkgo. test I should now be removed, and it
	// won't run twice.
	globals.Reset()

	// Define a simple gingko test II
	var _ = Describe("ginkgo test II", func() {
		It("build tests II", func() {
			testII++
			Ω(testII).Should(Equal(1))
		})
	})

	RunSpecs(t, "Test Runner Suite II")
}
