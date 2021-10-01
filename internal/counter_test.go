package internal_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/ginkgo/internal"
)

var _ = Describe("Counter", func() {
	It("counts.  plain and simple.", func() {
		counter := internal.MakeIncrementingIndexCounter()
		for i := 0; i < 10; i += 1 {
			Ω(counter()).Should(Equal(i))
		}
	})
})
