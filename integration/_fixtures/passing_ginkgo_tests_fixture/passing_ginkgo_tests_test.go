package passing_ginkgo_tests_test

import (
	"sync"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/ginkgo/v2/integration/_fixtures/passing_ginkgo_tests_fixture"
	. "github.com/onsi/gomega"
)

var _ = Describe("PassingGinkgoTests", func() {
	It("should proxy strings", func() {
		Ω(StringIdentity("foo")).Should(Equal("foo"))
	})

	It("should proxy integers", func() {
		Ω(IntegerIdentity(3)).Should(Equal(3))
	})

	It("should do it again", func() {
		Ω(StringIdentity("foo")).Should(Equal("foo"))
		Ω(IntegerIdentity(3)).Should(Equal(3))
	})

	It("should be able to run Bys", func() {
		By("emitting one By")
		Ω(3).Should(Equal(3))

		By("emitting another By")
		Ω(4).Should(Equal(4))
	})

	Context("when called within goroutines", func() {
		It("does not trigger the race detector", func() {
			wg := &sync.WaitGroup{}
			wg.Add(3)
			for i := 0; i < 3; i += 1 {
				go func() {
					By("avoiding the race detector")
					wg.Done()
				}()
			}
			wg.Wait()
		})
	})

})
