package ginkgo

import (
	"bytes"
	. "github.com/onsi/gomega"
)

func init() {
	Describe("GinkgoWriter", func() {
		var writer *ginkgoWriter
		var out *bytes.Buffer

		BeforeEach(func() {
			out = &bytes.Buffer{}
			writer = newGinkgoWriter(out)
		})

		It("should write to the outbuffer by default", func() {
			writer.Write([]byte("foo"))
			Ω(out.String()).Should(Equal("foo"))
		})

		Context("when told not to direct to stdout", func() {
			BeforeEach(func() {
				writer.setDirectToStdout(false)
			})

			It("should only write to the buffer when told to DumpOut", func() {
				writer.Write([]byte("foo"))
				Ω(out.String()).Should(BeEmpty())
				writer.DumpOut()
				Ω(out.String()).Should(Equal("foo"))
			})

			It("should truncate the internal buffer when told to truncate", func() {
				writer.Write([]byte("foo"))
				writer.Truncate()
				writer.DumpOut()
				Ω(out.String()).Should(BeEmpty())

				writer.Write([]byte("foo"))
				writer.DumpOut()
				Ω(out.String()).Should(Equal("foo"))
			})
		})
	})
}
