package synchronized_setup_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"fmt"
	"os"
	"testing"
)

func TestSynchronized_setup(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Synchronized_setup_tests Suite")
}

var beforeData string

var _ = SynchronizedBeforeSuite(func() []byte {
	fmt.Printf("BEFORE_A_%d\n", GinkgoParallelProcess())
	os.Exit(1)
	return []byte("WHAT EVZ")
}, func(data []byte) {
	println("NEVER SEE THIS")
})

var _ = Describe("Synchronized Setup", func() {
	It("should do nothing", func() {
		Ω(true).Should(BeTrue())
	})

	It("should do nothing", func() {
		Ω(true).Should(BeTrue())
	})
})
