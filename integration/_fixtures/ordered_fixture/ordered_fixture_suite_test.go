package ordered_fixture_test

import (
	"flag"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var noOrdered *bool

func init() {
	noOrdered = flag.CommandLine.Bool("no-ordered", false, "set to turn off ordered decoration")
}

var OrderedDecoration = []any{Ordered}

func TestOrderedFixture(t *testing.T) {
	RegisterFailHandler(Fail)
	if *noOrdered {
		OrderedDecoration = []any{}
	}

	RunSpecs(t, "OrderedFixture Suite")
}

var _ = Describe("tests", func() {
	for i := 0; i < 10; i += 1 {
		Context("ordered", OrderedDecoration, func() {
			var terribleSharedCounter int
			var parallelNode int

			BeforeAll(func() {
				terribleSharedCounter = 1
				parallelNode = GinkgoParallelProcess()
			})

			It("increments the counter", func() {
				Ω(parallelNode).Should(Equal(GinkgoParallelProcess()))
				terribleSharedCounter++
				Ω(terribleSharedCounter).Should(Equal(2))
			})

			It("increments the shared counter", func() {
				Ω(parallelNode).Should(Equal(GinkgoParallelProcess()))
				terribleSharedCounter++
				Ω(terribleSharedCounter).Should(Equal(3))
			})

			It("increments the terrible shared counter", func() {
				Ω(parallelNode).Should(Equal(GinkgoParallelProcess()))
				terribleSharedCounter++
				Ω(terribleSharedCounter).Should(Equal(4))
			})

			It("increments the counter again", func() {
				Ω(parallelNode).Should(Equal(GinkgoParallelProcess()))
				terribleSharedCounter++
				Ω(terribleSharedCounter).Should(Equal(5))
			})

			It("increments the counter and again", func() {
				Ω(parallelNode).Should(Equal(GinkgoParallelProcess()))
				terribleSharedCounter++
				Ω(terribleSharedCounter).Should(Equal(6))
			})

			AfterAll(func() {
				Ω(parallelNode).Should(Equal(GinkgoParallelProcess()))
				Ω(terribleSharedCounter).Should(Equal(6))
			})
		})
	}
})
