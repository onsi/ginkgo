package C_test

import (
	. "github.com/onsi/ginkgo/v2/integration/_fixtures/watch_fixture/C"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("C", func() {
	It("should do it", func() {
		Ω(DoIt()).Should(Equal("done!"))
	})
})
