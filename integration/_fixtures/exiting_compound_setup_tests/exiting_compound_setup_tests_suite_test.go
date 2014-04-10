package compound_setup_tests_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/ginkgo/config"

	"fmt"
	"os"
	"testing"
)

func TestCompound_setup_tests(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Compound_setup_tests Suite")
}

var beforeData string

var _ = CompoundBeforeSuite(func() []byte {
	fmt.Printf("BEFORE_A_%d\n", config.GinkgoConfig.ParallelNode)
	os.Exit(1)
	return []byte("WHAT EVZ")
}, func(data []byte) {
	println("NEVER SEE THIS")
})

var _ = Describe("Compound Setup", func() {
	It("should do nothing", func() {
		Ω(true).Should(BeTrue())
	})

	It("should do nothing", func() {
		Ω(true).Should(BeTrue())
	})
})
