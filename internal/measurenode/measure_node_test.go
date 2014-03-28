package measurenode_test

import (
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/internal/codelocation"
	"github.com/onsi/ginkgo/internal/failer"
	. "github.com/onsi/ginkgo/internal/measurenode"
	"github.com/onsi/ginkgo/internal/types"
	"github.com/onsi/ginkgo/types"
	. "github.com/onsi/gomega"
	"time"
)

var _ = Describe("MeasureNode", func() {
	var (
		measure *MeasureNode
		i       int
		fail    *failer.Failer
		outcome types.ExampleState
		failure types.ExampleFailure

		codeLocation      types.CodeLocation
		innerCodeLocation types.CodeLocation
	)

	BeforeEach(func() {
		fail = failer.New()
		i = 0
		codeLocation = codelocation.New(0)
		measure = New("foo", func(b Benchmarker) {
			b.RecordValue("bar", float64(i))
			i += 1
		}, internaltypes.FlagTypeFocused, codeLocation, 10, fail, 3)
	})

	It("should report on itself accurately", func() {
		Ω(measure.Text()).Should(Equal("foo"))
		Ω(measure.Flag()).Should(Equal(internaltypes.FlagTypeFocused))
		Ω(measure.CodeLocation()).Should(Equal(codeLocation))
		Ω(measure.Type()).Should(Equal(types.ExampleComponentTypeMeasure))
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
			Ω(outcome).Should(Equal(types.ExampleStatePassed))
			Ω(failureData).Should(BeZero())
		})
	})

	Context("when run, and a failure occurs", func() {
		BeforeEach(func() {
			measure = New("foo", func(Benchmarker) {
				innerCodeLocation = codelocation.New(0)
				fail.Fail("oops!", innerCodeLocation)
				panic("not a problem")
			}, internaltypes.FlagTypeFocused, codeLocation, 10, fail, 3)

			outcome, failure = measure.Run()
		})

		It("should return said failure", func() {
			Ω(outcome).Should(Equal(types.ExampleStateFailed))

			Ω(failure.Message).Should(Equal("oops!"))
			Ω(failure.Location).Should(Equal(innerCodeLocation))
			Ω(failure.ForwardedPanic).Should(BeNil())
			Ω(failure.ComponentIndex).Should(Equal(3))
			Ω(failure.ComponentType).Should(Equal(types.ExampleComponentTypeMeasure))
			Ω(failure.ComponentCodeLocation).Should(Equal(codeLocation))
		})
	})

	Context("when run, and the function panics", func() {
		BeforeEach(func() {
			measure = New("foo", func(Benchmarker) {
				innerCodeLocation = codelocation.New(0)
				panic("kaboom")
			}, internaltypes.FlagTypeFocused, codeLocation, 10, fail, 3)
			outcome, failure = measure.Run()
		})

		It("should return a failure representing the panic", func() {
			Ω(outcome).Should(Equal(types.ExampleStatePanicked))

			Ω(failure.Message).Should(Equal("Test Panicked"))
			Ω(failure.Location).Should(Equal(codeLocation))
			Ω(failure.ForwardedPanic).Should(Equal("kaboom"))
			Ω(failure.ComponentIndex).Should(Equal(3))
			Ω(failure.ComponentType).Should(Equal(types.ExampleComponentTypeMeasure))
			Ω(failure.ComponentCodeLocation).Should(Equal(codeLocation))
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
				}, internaltypes.FlagTypeFocused, codeLocation, 1, fail, 3)
				Ω(measure.Run()).Should(Equal(types.ExampleStatePassed))
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
				}, internaltypes.FlagTypeFocused, codeLocation, 1, fail, 3)
				Ω(measure.Run()).Should(Equal(types.ExampleStatePassed))
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
