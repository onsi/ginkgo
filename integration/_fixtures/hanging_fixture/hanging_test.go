package hanging_suite_test

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
)

var _ = AfterSuite(func() {
	fmt.Println("Heading Out After Suite")
})

var _ = Describe("HangingSuite", func() {
	BeforeEach(func() {
		fmt.Fprintln(GinkgoWriter, "Just beginning")
	})

	Context("inner context", func() {
		BeforeEach(func() {
			fmt.Fprintln(GinkgoWriter, "Almost there...")
		})

		It("should hang out for a while", func() {
			fmt.Fprintln(GinkgoWriter, "Hanging Out")
			fmt.Println("Sleeping...")
			time.Sleep(time.Hour)
		})

		AfterEach(func() {
			fmt.Fprintln(GinkgoWriter, "Cleaning up once...")
		})
	})

	AfterEach(func() {
		fmt.Fprintln(GinkgoWriter, "Cleaning up twice...")
		fmt.Println("Sleeping again...")
		time.Sleep(time.Hour)
	})

	AfterEach(func() {
		fmt.Fprintln(GinkgoWriter, "Cleaning up thrice...")
	})
})
