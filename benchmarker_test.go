package ginkgo

import (
	. "github.com/onsi/gomega"
	"time"
)

func init() {
	var benchmarker *benchmarker

	Describe("Benchmarker", func() {
		BeforeEach(func() {
			benchmarker = newBenchmarker()
		})

		Describe("Value", func() {
			BeforeEach(func() {
				benchmarker.RecordValue("foo", 7, "info!")
				benchmarker.RecordValue("foo", 2)
				benchmarker.RecordValue("foo", 3)
				benchmarker.RecordValue("bar", 0.3)
				benchmarker.RecordValue("bar", 0.1)
				benchmarker.RecordValue("bar", 0.5)
				benchmarker.RecordValue("bar", 0.7)
			})

			It("records passed in values and reports on them", func() {
				report := benchmarker.measurementsReport()
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
				benchmarker.Time("foo", func() {
					time.Sleep(100 * time.Millisecond)
				}, "info!")
				benchmarker.Time("foo", func() {
					time.Sleep(200 * time.Millisecond)
				})
				benchmarker.Time("foo", func() {
					time.Sleep(170 * time.Millisecond)
				})
			})

			It("records passed in values and reports on them", func() {
				report := benchmarker.measurementsReport()
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
}
