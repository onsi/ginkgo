package all_async_timeout_tests_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AllAsyncTimeoutTests", func() {
	It("should timeout", func() {
		time.Sleep(20 * time.Millisecond)
	}, 0.01)

	It("should pass", func() {
		Ω(false).Should(BeFalse())
	}, 0.01)
})
