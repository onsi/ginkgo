package internal_integration_test

import (
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/types"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("when config.EmitSpecProgress is enabled", func() {
	BeforeEach(func() {
		writer.SetStream(true)
		conf.EmitSpecProgress = true
	})

	It("emits progress to the writer as it goes", func() {
		l := types.NewCodeLocation(0)
		RunFixture("emitting spec progress", func() {
			BeforeSuite(func() {
				Ω(writerBuffer).Should(gbytes.Say(`\[BeforeSuite\] TOP-LEVEL`))
				Ω(writerBuffer).Should(gbytes.Say(`%s:%d`, l.FileName, l.LineNumber+2))
			})
			Describe("a container", func() {
				BeforeEach(func() {
					Ω(writerBuffer).Should(gbytes.Say(`\[BeforeEach\] a container`))
					Ω(writerBuffer).Should(gbytes.Say(`%s:\d+`, l.FileName))
				})
				It("A", func() {
					Ω(writerBuffer).Should(gbytes.Say(`\[It\] A`))
					Ω(writerBuffer).Should(gbytes.Say(`%s:\d+`, l.FileName))
				})
				It("B", func() {
					Ω(writerBuffer).Should(gbytes.Say(`\[It\] B`))
					Ω(writerBuffer).Should(gbytes.Say(`%s:\d+`, l.FileName))
				})
				AfterEach(func() {
					Ω(writerBuffer).Should(gbytes.Say(`\[AfterEach\] a container`))
					Ω(writerBuffer).Should(gbytes.Say(`%s:\d+`, l.FileName))
				})
			})
			AfterSuite(func() {
				Ω(writerBuffer).Should(gbytes.Say(`\[AfterSuite\] TOP-LEVEL`))
				Ω(writerBuffer).Should(gbytes.Say(`%s:\d+`, l.FileName))
			})
		})

	})
})
