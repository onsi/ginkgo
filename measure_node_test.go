package ginkgo

import (
	"github.com/onsi/ginkgo/types"
	. "github.com/onsi/gomega"
)

func init() {
	Describe("MeasureNode", func() {
		var measure *measureNode
		var i int
		var codeLocation types.CodeLocation

		BeforeEach(func() {
			i = 0
			codeLocation = types.GenerateCodeLocation(0)
			measure = newMeasureNode("foo", func(b Benchmarker) {
				b.RecordValue("bar", float64(i))
				i += 1
			}, flagTypeFocused, codeLocation, 10)
		})

		It("should report on itself accurately", func() {
			Ω(measure.getText()).Should(Equal("foo"))
			Ω(measure.getFlag()).Should(Equal(flagTypeFocused))
			Ω(measure.getCodeLocation()).Should(Equal(codeLocation))
			Ω(measure.nodeType()).Should(Equal(nodeTypeMeasure))
			Ω(measure.samples).Should(Equal(10))
		})

		Context("when run", func() {
			It("should provide the body function with a benchmarker and be able to aggregate reports", func() {
				measure.run()
				measure.run()
				measure.run()
				measure.run()

				report := measure.measurementsReport()
				Ω(report).Should(HaveLen(1))
				Ω(report["bar"].Name).Should(Equal("bar"))
				Ω(report["bar"].Results).Should(Equal([]float64{0, 1, 2, 3}))
			})
		})
	})
}
