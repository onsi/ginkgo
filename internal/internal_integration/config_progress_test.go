package internal_integration_test

import (
	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("when config.EmitSpecProgress is enabled", func() {
	var buffer *gbytes.Buffer
	BeforeEach(func() {
		buffer = gbytes.NewBuffer()
		writer.TeeTo(buffer)
		conf.EmitSpecProgress = true
	})

	It("emits progress to the writer as it goes", func() {
		l := types.NewCodeLocation(0)
		RunFixture("emitting spec progress", func() {
			BeforeSuite(func() {
				Ω(buffer).Should(gbytes.Say(`\[BeforeSuite\] TOP-LEVEL`))
				Ω(buffer).Should(gbytes.Say(`%s:%d`, l.FileName, l.LineNumber+2))
			})
			Describe("a container", func() {
				BeforeEach(func() {
					Ω(buffer).Should(gbytes.Say(`\[BeforeEach\] a container`))
					Ω(buffer).Should(gbytes.Say(`%s:\d+`, l.FileName))
				})
				It("A", func() {
					Ω(buffer).Should(gbytes.Say(`\[It\] A`))
					Ω(buffer).Should(gbytes.Say(`%s:\d+`, l.FileName))
				})
				It("B", func() {
					Ω(buffer).Should(gbytes.Say(`\[It\] B`))
					Ω(buffer).Should(gbytes.Say(`%s:\d+`, l.FileName))
				})
				AfterEach(func() {
					Ω(buffer).Should(gbytes.Say(`\[AfterEach\] a container`))
					Ω(buffer).Should(gbytes.Say(`%s:\d+`, l.FileName))
				})
			})
			AfterSuite(func() {
				Ω(buffer).Should(gbytes.Say(`\[AfterSuite\] TOP-LEVEL`))
				Ω(buffer).Should(gbytes.Say(`%s:\d+`, l.FileName))
			})
		})

	})
})
