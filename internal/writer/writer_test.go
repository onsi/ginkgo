package writer_test

import (
	"bytes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/internal/writer"
	. "github.com/onsi/gomega"
)

var _ = Describe("Writer", func() {
	var writer *Writer
	var out *bytes.Buffer

	BeforeEach(func() {
		out = &bytes.Buffer{}
		writer = New(out)
	})

	It("should stream directly to the outbuffer by default", func() {
		writer.Write([]byte("foo"))
		Ω(out.String()).Should(Equal("foo"))
	})

	Context("when told not to stream", func() {
		BeforeEach(func() {
			writer.SetStream(false)
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
