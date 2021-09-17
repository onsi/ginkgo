package ordered_fixture_test

import (
	"flag"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var noOrdered *bool

func init() {
	noOrdered = flag.CommandLine.Bool("no-ordered", false, "set to turn off ordered decoration")
}

var OrderedDecoration = []interface{}{Ordered}

func TestOrderedFixture(t *testing.T) {
	RegisterFailHandler(Fail)
	if *noOrdered {
		OrderedDecoration = []interface{}{}
	}

	RunSpecs(t, "OrderedFixture Suite")
}

var _ = Describe("tests", func() {
	for i := 0; i < 10; i += 1 {
		Context("ordered", OrderedDecoration, func() {
			terribleSharedCounter := 0
			var parallelNode int

			It("increments the counter", func() {
				parallelNode = GinkgoParallelNode()
				terribleSharedCounter++
				Ω(terribleSharedCounter).Should(Equal(1))
			})

			It("increments the shared counter", func() {
				Ω(parallelNode).Should(Equal(GinkgoParallelNode()))
				terribleSharedCounter++
				Ω(terribleSharedCounter).Should(Equal(2))
			})

			It("increments the terrible shared counter", func() {
				Ω(parallelNode).Should(Equal(GinkgoParallelNode()))
				terribleSharedCounter++
				Ω(terribleSharedCounter).Should(Equal(3))
			})

			It("increments the counter again", func() {
				Ω(parallelNode).Should(Equal(GinkgoParallelNode()))
				terribleSharedCounter++
				Ω(terribleSharedCounter).Should(Equal(4))
			})

			It("increments the counter and again", func() {
				Ω(parallelNode).Should(Equal(GinkgoParallelNode()))
				terribleSharedCounter++
				Ω(terribleSharedCounter).Should(Equal(5))
			})
		})
	}
})
