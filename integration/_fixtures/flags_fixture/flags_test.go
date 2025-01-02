package flags_test

import (
	"flag"
	"fmt"
	remapped "math"
	_ "math/cmplx"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/ginkgo/v2/integration/_fixtures/flags_fixture"
	. "github.com/onsi/gomega"
)

var customFlag string

func init() {
	flag.StringVar(&customFlag, "customFlag", "default", "custom flag!")
}

var _ = Describe("Testing various flags", func() {
	It("should honor -cover", func() {
		Ω(Tested()).Should(Equal("tested"))
	})

	It("should allow gcflags", func() {
		fmt.Printf("NaN returns %T\n", remapped.NaN())
	})

	PIt("should honor -failOnPending and -noisyPendings")

	Describe("smores", func() {
		It("should honor -skip: marshmallow", func() {
			println("marshmallow")
		})

		It("should honor -focus: chocolate", func() {
			println("chocolate")
		})
	})

	It("should detect races", func() {
		var a string
			c := make(chan any, 0)
		go func() {
			a = "now you don't"
			close(c)
		}()
		a = "now you see me"
		println(a)
		Eventually(c).Should(BeClosed())
	})

	It("should randomize A", func() {
		println("RANDOM_A")
	})

	It("should randomize B", func() {
		println("RANDOM_B")
	})

	It("should randomize C", func() {
		println("RANDOM_C")
	})

	It("should pass in additional arguments after '--' directly to the test process", func() {
		fmt.Printf("CUSTOM_FLAG: %s", customFlag)
	})

	It("should fail", func() {
		Ω(true).Should(Equal(false))
	})

	Describe("a flaky test", func() {
		runs := 0
		It("should only pass the second time it's run", func() {
			runs++
			Ω(runs).Should(BeNumerically("==", 2))
		})
	})
})
