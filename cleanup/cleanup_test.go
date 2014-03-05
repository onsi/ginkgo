package cleanup_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/cleanup"
	. "github.com/onsi/gomega"
)

var _ = Describe("Cleanup", func() {
	Context("with cleanup functions registered", func() {
		var calls []string
		BeforeEach(func() {
			Register(func() {
				calls = append(calls, "A")
			})
			Register(func() {
				calls = append(calls, "B")
			})
		})

		It("should call the functions when told to clean up", func() {
			Cleanup()
			Ω(calls).Should(Equal([]string{"A", "B"}))
		})
	})
})
