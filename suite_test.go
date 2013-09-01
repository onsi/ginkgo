package ginkgo

import (
	. "github.com/onsi/gomega"
)

func init() {
	Describe("Suite", func() {
		var (
			specSuite *suite
			fakeT     *fakeTestingT
			fakeR     *fakeReporter
		)

		BeforeEach(func() {
			fakeT = &fakeTestingT{}
			fakeR = &fakeReporter{}
			specSuite = newSuite()
		})

		Describe("running a suite", func() {
			var (
				runOrder          []string
				randomizeAllSpecs bool
				randomSeed        int64
			)

			var f = func(runText string) func() {
				return func() {
					runOrder = append(runOrder, runText)
				}
			}

			BeforeEach(func() {
				randomizeAllSpecs = false
				randomSeed = 22

				runOrder = make([]string, 0)
				specSuite.pushBeforeEachNode(f("top BE"), generateCodeLocation(0), 0)
				specSuite.pushJustBeforeEachNode(f("top JBE"), generateCodeLocation(0), 0)
				specSuite.pushAfterEachNode(f("top AE"), generateCodeLocation(0), 0)

				specSuite.pushContainerNode("container", func() {
					specSuite.pushBeforeEachNode(f("BE"), generateCodeLocation(0), 0)
					specSuite.pushJustBeforeEachNode(f("JBE"), generateCodeLocation(0), 0)
					specSuite.pushAfterEachNode(f("AE"), generateCodeLocation(0), 0)
					specSuite.pushItNode("it", f("IT"), flagTypeNone, generateCodeLocation(0), 0)

					specSuite.pushContainerNode("inner container", func() {
						specSuite.pushItNode("inner it", f("inner IT"), flagTypeNone, generateCodeLocation(0), 0)
					}, flagTypeNone, generateCodeLocation(0))
				}, flagTypeNone, generateCodeLocation(0))

				specSuite.pushContainerNode("container 2", func() {
					specSuite.pushBeforeEachNode(f("BE 2"), generateCodeLocation(0), 0)
					specSuite.pushItNode("it 2", f("IT 2"), flagTypeNone, generateCodeLocation(0), 0)
				}, flagTypeNone, generateCodeLocation(0))

				specSuite.pushItNode("top level it", f("top IT"), flagTypeNone, generateCodeLocation(0), 0)
			})

			JustBeforeEach(func() {
				specSuite.run(fakeT, "suite description", fakeR, GinkoConfigType{
					RandomSeed:        &randomSeed,
					RandomizeAllSpecs: &randomizeAllSpecs,
				})
			})

			It("reports the randomization strategy to the reporter", func() {
				Ω(fakeR.randomSeed).Should(Equal(int64(22)))
				Ω(fakeR.randomizeAllExamples).Should(BeFalse())
				Ω(fakeR.beginSummary.SuiteDescription).Should(Equal("suite description"))
			})

			It("creates a node hierarchy, converts it to an example collection, and runs it", func() {
				Ω(runOrder).Should(Equal([]string{
					"top BE", "BE", "top JBE", "JBE", "IT", "AE", "top AE",
					"top BE", "BE", "top JBE", "JBE", "inner IT", "AE", "top AE",
					"top BE", "top JBE", "top IT", "top AE",
					"top BE", "BE 2", "top JBE", "IT 2", "top AE",
				}), "Note this was randomized at the container level")
			})

			Context("when told to randomize all examples", func() {
				BeforeEach(func() {
					randomizeAllSpecs = true
				})

				It("does", func() {
					Ω(runOrder).Should(Equal([]string{
						"top BE", "BE", "top JBE", "JBE", "inner IT", "AE", "top AE",
						"top BE", "BE", "top JBE", "JBE", "IT", "AE", "top AE",
						"top BE", "BE 2", "top JBE", "IT 2", "top AE",
						"top BE", "top JBE", "top IT", "top AE",
					}))
				})
			})

			Context("when the specs pass", func() {
				It("doesn't report a failure", func() {
					Ω(fakeT.didFail).Should(BeFalse())
				})
			})

			Context("when a spec fails", func() {
				var location CodeLocation
				BeforeEach(func() {
					specSuite.pushItNode("top level it", func() {
						location = generateCodeLocation(0)
						func() { specSuite.fail("oops!", 0) }()
					}, flagTypeNone, generateCodeLocation(0), 0)
				})

				It("reports a failure", func() {
					Ω(fakeT.didFail).Should(BeTrue())
				})

				It("generates the correct failure data", func() {
					Ω(fakeR.exampleSummaries[4].Failure.Message).Should(Equal("oops!"))
					Ω(fakeR.exampleSummaries[4].Failure.Location.FileName).Should(Equal(location.FileName))
					Ω(fakeR.exampleSummaries[4].Failure.Location.LineNumber).Should(Equal(location.LineNumber + 1))
				})
			})
		})
	})
}
