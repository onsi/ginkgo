package synchronized_setup_tests_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/ginkgo/config"

	"fmt"
	"testing"
	"time"
)

func TestSynchronized_setup_tests(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Synchronized_setup_tests Suite")
}

var beforeData string

var _ = SynchronizedBeforeSuite(func() []byte {
	fmt.Printf("BEFORE_A_%d\n", config.GinkgoConfig.ParallelNode)
	time.Sleep(100 * time.Millisecond)
	return []byte("DATA")
}, func(data []byte) {
	fmt.Printf("BEFORE_B_%d: %s\n", config.GinkgoConfig.ParallelNode, string(data))
	beforeData += string(data) + "OTHER"
})

var _ = SynchronizedAfterSuite(func() {
	fmt.Printf("\nAFTER_A_%d\n", config.GinkgoConfig.ParallelNode)
	time.Sleep(100 * time.Millisecond)
}, func() {
	fmt.Printf("AFTER_B_%d\n", config.GinkgoConfig.ParallelNode)
})

var _ = Describe("Synchronized Setup", func() {
	It("should run the before suite once", func() {
		Ω(beforeData).Should(Equal("DATAOTHER"))
	})

	It("should run the before suite once", func() {
		Ω(beforeData).Should(Equal("DATAOTHER"))
	})
})
