package measurenode_test

import (
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/internal/codelocation"
	. "github.com/onsi/ginkgo/internal/measurenode"
	"github.com/onsi/ginkgo/internal/types"
	"github.com/onsi/ginkgo/types"
	. "github.com/onsi/gomega"
	"time"
)

var _ = Describe("MeasureNode", func() {
	var measure *MeasureNode
	var i int
	var codeLocation types.CodeLocation

	BeforeEach(func() {
		i = 0
		codeLocation = codelocation.New(0)
		measure = New("foo", func(b Benchmarker) {
			b.RecordValue("bar", float64(i))
			i += 1
		}, internaltypes.FlagTypeFocused, codeLocation, 10)
	})

	It("should report on itself accurately", func() {
		Ω(measure.Text()).Should(Equal("foo"))
		Ω(measure.Flag()).Should(Equal(internaltypes.FlagTypeFocused))
		Ω(measure.CodeLocation()).Should(Equal(codeLocation))
		Ω(measure.Type()).Should(Equal(internaltypes.NodeTypeMeasure))
		Ω(measure.Samples()).Should(Equal(10))
	})

	Context("when run", func() {
		It("should provide the body function with a benchmarker and be able to aggregate reports", func() {
			measure.Run()
			measure.Run()
			measure.Run()
			measure.Run()

			report := measure.MeasurementsReport()
			Ω(report).Should(HaveLen(1))
			Ω(report["bar"].Name).Should(Equal("bar"))
			Ω(report["bar"].Results).Should(Equal([]float64{0, 1, 2, 3}))
		})

		It("should report success", func() {
			outcome, failureData := measure.Run()
			Ω(outcome).Should(Equal(internaltypes.OutcomeCompleted))
			Ω(failureData).Should(BeZero())
		})
	})

	Context("when run, and the function panics", func() {
		var (
			innerCodeLocation types.CodeLocation
			outcome           internaltypes.Outcome
			failure           internaltypes.FailureData
		)

		BeforeEach(func() {
			measure = New("foo", func(Benchmarker) {
				innerCodeLocation = codelocation.New(0)
				panic("kaboom")
			}, internaltypes.FlagTypeFocused, innerCodeLocation, 10)
			outcome, failure = measure.Run()
		})

		It("should run the function and report a runOutcomePanicked", func() {
			Ω(outcome).Should(Equal(internaltypes.OutcomePanicked))
			Ω(failure.Message).Should(Equal("Test Panicked"))
		})

		It("should include the code location of the panic itself", func() {
			Ω(failure.CodeLocation.FileName).Should(Equal(innerCodeLocation.FileName))
			Ω(failure.CodeLocation.LineNumber).Should(Equal(innerCodeLocation.LineNumber + 1))
		})

		It("should include the panic data", func() {
			Ω(failure.ForwardedPanic).Should(Equal("kaboom"))
		})
	})

	Describe("the benchmarker", func() {
		Describe("Value", func() {
			BeforeEach(func() {
				measure = New("the measurement", func(b Benchmarker) {
					b.RecordValue("foo", 7, "info!")
					b.RecordValue("foo", 2)
					b.RecordValue("foo", 3)
					b.RecordValue("bar", 0.3)
					b.RecordValue("bar", 0.1)
					b.RecordValue("bar", 0.5)
					b.RecordValue("bar", 0.7)
				}, internaltypes.FlagTypeFocused, codeLocation, 1)
				Ω(measure.Run()).Should(Equal(internaltypes.OutcomeCompleted))
			})

			It("records passed in values and reports on them", func() {
				report := measure.MeasurementsReport()
				Ω(report).Should(HaveLen(2))
				Ω(report["foo"].Name).Should(Equal("foo"))
				Ω(report["foo"].Info).Should(Equal("info!"))
				Ω(report["foo"].SmallestLabel).Should(Equal("Smallest"))
				Ω(report["foo"].LargestLabel).Should(Equal(" Largest"))
				Ω(report["foo"].AverageLabel).Should(Equal(" Average"))
				Ω(report["foo"].Units).Should(Equal(""))
				Ω(report["foo"].Results).Should(Equal([]float64{7, 2, 3}))
				Ω(report["foo"].Smallest).Should(BeNumerically("==", 2))
				Ω(report["foo"].Largest).Should(BeNumerically("==", 7))
				Ω(report["foo"].Average).Should(BeNumerically("==", 4))
				Ω(report["foo"].StdDeviation).Should(BeNumerically("~", 2.16, 0.01))

				Ω(report["bar"].Name).Should(Equal("bar"))
				Ω(report["bar"].Info).Should(BeNil())
				Ω(report["bar"].SmallestLabel).Should(Equal("Smallest"))
				Ω(report["bar"].LargestLabel).Should(Equal(" Largest"))
				Ω(report["bar"].AverageLabel).Should(Equal(" Average"))
				Ω(report["foo"].Units).Should(Equal(""))
				Ω(report["bar"].Results).Should(Equal([]float64{0.3, 0.1, 0.5, 0.7}))
				Ω(report["bar"].Smallest).Should(BeNumerically("==", 0.1))
				Ω(report["bar"].Largest).Should(BeNumerically("==", 0.7))
				Ω(report["bar"].Average).Should(BeNumerically("==", 0.4))
				Ω(report["bar"].StdDeviation).Should(BeNumerically("~", 0.22, 0.01))
			})
		})

		Describe("Time", func() {
			BeforeEach(func() {
				measure = New("the measurement", func(b Benchmarker) {
					b.Time("foo", func() {
						time.Sleep(100 * time.Millisecond)
					}, "info!")
					b.Time("foo", func() {
						time.Sleep(200 * time.Millisecond)
					})
					b.Time("foo", func() {
						time.Sleep(170 * time.Millisecond)
					})
				}, internaltypes.FlagTypeFocused, codeLocation, 1)
				Ω(measure.Run()).Should(Equal(internaltypes.OutcomeCompleted))
			})

			It("records passed in values and reports on them", func() {
				report := measure.MeasurementsReport()
				Ω(report).Should(HaveLen(1))
				Ω(report["foo"].Name).Should(Equal("foo"))
				Ω(report["foo"].Info).Should(Equal("info!"))
				Ω(report["foo"].SmallestLabel).Should(Equal("Fastest Time"))
				Ω(report["foo"].LargestLabel).Should(Equal("Slowest Time"))
				Ω(report["foo"].AverageLabel).Should(Equal("Average Time"))
				Ω(report["foo"].Units).Should(Equal("s"))
				Ω(report["foo"].Results).Should(HaveLen(3))
				Ω(report["foo"].Results[0]).Should(BeNumerically("~", 0.1, 0.01))
				Ω(report["foo"].Results[1]).Should(BeNumerically("~", 0.2, 0.01))
				Ω(report["foo"].Results[2]).Should(BeNumerically("~", 0.17, 0.01))
				Ω(report["foo"].Smallest).Should(BeNumerically("~", 0.1, 0.01))
				Ω(report["foo"].Largest).Should(BeNumerically("~", 0.2, 0.01))
				Ω(report["foo"].Average).Should(BeNumerically("~", 0.16, 0.01))
				Ω(report["foo"].StdDeviation).Should(BeNumerically("~", 0.04, 0.01))
			})
		})
	})
})
