package internal_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/ginkgo/internal"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("Writer", func() {
	var writer *internal.Writer
	var out *gbytes.Buffer

	BeforeEach(func() {
		out = gbytes.NewBuffer()
		writer = internal.NewWriter(out)
	})

	Context("when configured to stream (the default)", func() {
		It("should stream directly to the passed in writer by default", func() {
			writer.Write([]byte("foo"))
			Ω(out).Should(gbytes.Say("foo"))
		})

		It("does not store the bytes", func() {
			writer.Write([]byte("foo"))
			Ω(out).Should(gbytes.Say("foo"))
			Ω(writer.Bytes()).Should(BeEmpty())
		})
	})

	Context("when told not to stream", func() {
		BeforeEach(func() {
			writer.SetStream(false)
		})

		It("should not write to the passed in writer", func() {
			writer.Write([]byte("foo"))
			Ω(out).ShouldNot(gbytes.Say("foo"))
		})

		Describe("Bytes()", func() {
			BeforeEach(func() {
				writer.Write([]byte("foo"))
			})

			It("returns all that's been written so far", func() {
				Ω(writer.Bytes()).Should(Equal([]byte("foo")))
			})

			It("clears when told to truncate", func() {
				writer.Truncate()
				Ω(writer.Bytes()).Should(BeEmpty())
				writer.Write([]byte("bar"))
				Ω(writer.Bytes()).Should(Equal([]byte("bar")))
			})
		})
	})
})
