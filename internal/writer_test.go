package internal_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/onsi/ginkgo/v2/internal"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("Writer", func() {
	var writer *internal.Writer
	var out *gbytes.Buffer

	BeforeEach(func() {
		out = gbytes.NewBuffer()
		writer = internal.NewWriter(out)
	})

	Context("when configured to WriterModeStreamAndBuffer (the default setting)", func() {
		It("streams directly to the passed in writer, indenting content as it writes it", func() {
			writer.Write([]byte("foo"))
			Ω(out).Should(gbytes.Say("  foo"))
		})

		It("does also stores the bytes", func() {
			writer.Write([]byte("foo"))
			Ω(out).Should(gbytes.Say("foo"))
			Ω(string(writer.Bytes())).Should(Equal("foo"))
		})
	})

	Describe("indenting output", func() {
		It("handles all the vagaries of indenting a stream", func() {
			writer.Write([]byte("foo"))
			writer.Write([]byte(" bar\nbaz"))
			writer.Write([]byte("zle\n"))
			writer.Write([]byte("\nbloop"))
			writer.Write([]byte("floop\n"))
			writer.Write([]byte("flarp\r\n"))
			writer.Write([]byte("flan"))

			Ω(string(out.Contents())).Should(Equal("  foo bar\n  bazzle\n\n  bloopfloop\n  flarp\r\n  flan"))
		})
	})

	Context("when configured to WriterModeBufferOnly", func() {
		BeforeEach(func() {
			writer.SetMode(internal.WriterModeBufferOnly)
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

	Describe("Teeing to additional writers", func() {
		var tee1, tee2 *gbytes.Buffer
		BeforeEach(func() {
			tee1 = gbytes.NewBuffer()
			tee2 = gbytes.NewBuffer()
		})

		Context("when told to tee to additional writers", func() {
			BeforeEach(func() {
				writer.TeeTo(tee1)
				writer.TeeTo(tee2)
			})

			It("tees to all registered tee writers", func() {
				writer.Print("hello")
				Ω(string(tee1.Contents())).Should(Equal("hello"))
				Ω(string(tee2.Contents())).Should(Equal("hello"))
			})

			Context("when told to clear tee writers", func() {
				BeforeEach(func() {
					writer.ClearTeeWriters()
				})

				It("stops teeing to said writers", func() {
					writer.Print("goodbye")
					Ω(tee1.Contents()).Should(BeEmpty())
					Ω(tee2.Contents()).Should(BeEmpty())
				})
			})

		})
	})

	Describe("Convenience print methods", func() {
		It("can Print", func() {
			writer.Print("foo", "baz", " ", "bizzle")
			Ω(string(out.Contents())).Should(Equal("  foobaz bizzle"))
		})

		It("can Println", func() {
			writer.Println("foo", "baz", " ", "bizzle")
			Ω(string(out.Contents())).Should(Equal("  foo baz   bizzle\n"))
		})

		It("can Printf", func() {
			writer.Printf("%s%d - %s\n", "foo", 17, "bar")
			Ω(string(out.Contents())).Should(Equal("  foo17 - bar\n"))
		})
	})

	When("wrapped by logr", func() {
		It("can print Info logs", func() {
			log := internal.GinkgoLogrFunc(writer)
			log.Info("message", "key", 5)
			Ω(string(out.Contents())).Should(Equal("  \"level\"=0 \"msg\"=\"message\" \"key\"=5\n"))
		})

		It("can print Error logs", func() {
			log := internal.GinkgoLogrFunc(writer)
			log.Error(errors.New("cake"), "planned failure", "key", "banana")
			Ω(string(out.Contents())).Should(Equal("  \"msg\"=\"planned failure\" \"error\"=\"cake\" \"key\"=\"banana\"\n"))
		})

		It("can print multiple logs", func() {
			log := internal.GinkgoLogrFunc(writer)
			log.Info("message", "key", 5)
			log.Error(errors.New("cake"), "planned failure", "key", "banana")
			Ω(string(out.Contents())).Should(Equal("  \"level\"=0 \"msg\"=\"message\" \"key\"=5\n  \"msg\"=\"planned failure\" \"error\"=\"cake\" \"key\"=\"banana\"\n"))
		})

		It("can print the logr prefix", func() {
			log := internal.GinkgoLogrFunc(writer)
			log.WithName("berry").Info("message", "key", 5)
			Ω(string(out.Contents())).Should(Equal("  berry \"level\"=0 \"msg\"=\"message\" \"key\"=5\n"))
		})
	})
})
