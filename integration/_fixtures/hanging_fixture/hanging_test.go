package hanging_suite_test

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
)

var _ = AfterSuite(func() {
	fmt.Println("Heading Out After Suite")
})

var _ = ReportAfterSuite("", func(r Report) {
	fmt.Println("Reporting at the end")
})

var _ = Describe("HangingSuite", func() {
	BeforeEach(func() {
		fmt.Fprintln(GinkgoWriter, "Just beginning")
	})

	Context("inner context", func() {
		BeforeEach(func() {
			fmt.Fprintln(GinkgoWriter, "Almost there...")
		})

		It("should hang out for a while", func(ctx SpecContext) {
			fmt.Fprintln(GinkgoWriter, "Hanging Out")
			fmt.Println("Sleeping...")
			select {
			case <-ctx.Done():
				fmt.Println("Got your signal, but still taking a nap")
				time.Sleep(time.Hour)
			case <-time.After(time.Hour):
			}
		})

		AfterEach(func() {
			fmt.Println("Cleaning up once...")
		})
	})

	AfterEach(func() {
		fmt.Println("Cleaning up twice...")
		fmt.Println("Sleeping again...")
		time.Sleep(time.Hour)
	})

	AfterEach(func() {
		fmt.Println("Cleaning up thrice...")
	})
})
