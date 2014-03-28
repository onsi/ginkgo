package failer_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/internal/failer"
	. "github.com/onsi/gomega"
)

var _ = Describe("Failer", func() {
	It("should be tested", func() {
		panic("BAM!")
	})
})
