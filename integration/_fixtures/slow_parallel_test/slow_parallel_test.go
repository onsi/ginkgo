package slow_parallel_suite_test

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
)

var _ = Describe("SlowParallelSuite", func() {
	It("should hang out for a while", func() {
		fmt.Fprintln(GinkgoWriter, "Slow Test is Hanging Out")
		fmt.Println("Slow Test is Sleeping...")
		time.Sleep(5 * time.Second)
	})
})
