package ginkgo

import (
	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/reporters"
	"github.com/onsi/ginkgo/types"
	. "github.com/onsi/gomega"
	"math/rand"
	"time"
)

func init() {
	Describe("Suite", func() {
		var (
			specSuite *suite
			fakeT     *fakeTestingT
			fakeR     *reporters.FakeReporter
		)

		BeforeEach(func() {
			fakeT = &fakeTestingT{}
			fakeR = reporters.NewFakeReporter()
			specSuite = newSuite()
		})

		Describe("running a suite", func() {
			var (
				runOrder          []string
				randomizeAllSpecs bool
				randomSeed        int64
				focusString       string
				parallelNode      int
				parallelTotal     int
				runResult         bool
			)

			var f = func(runText string) func() {
				return func() {
					runOrder = append(runOrder, runText)
				}
			}

			BeforeEach(func() {
				randomizeAllSpecs = false
				randomSeed = 11
				parallelNode = 1
				parallelTotal = 1
				focusString = ""

				runOrder = make([]string, 0)
				specSuite.pushBeforeEachNode(f("top BE"), types.GenerateCodeLocation(0), 0)
				specSuite.pushJustBeforeEachNode(f("top JBE"), types.GenerateCodeLocation(0), 0)
				specSuite.pushAfterEachNode(f("top AE"), types.GenerateCodeLocation(0), 0)

				specSuite.pushContainerNode("container", func() {
					specSuite.pushBeforeEachNode(f("BE"), types.GenerateCodeLocation(0), 0)
					specSuite.pushJustBeforeEachNode(f("JBE"), types.GenerateCodeLocation(0), 0)
					specSuite.pushAfterEachNode(f("AE"), types.GenerateCodeLocation(0), 0)
					specSuite.pushItNode("it", f("IT"), flagTypeNone, types.GenerateCodeLocation(0), 0)

					specSuite.pushContainerNode("inner container", func() {
						specSuite.pushItNode("inner it", f("inner IT"), flagTypeNone, types.GenerateCodeLocation(0), 0)
					}, flagTypeNone, types.GenerateCodeLocation(0))
				}, flagTypeNone, types.GenerateCodeLocation(0))

				specSuite.pushContainerNode("container 2", func() {
					specSuite.pushBeforeEachNode(f("BE 2"), types.GenerateCodeLocation(0), 0)
					specSuite.pushItNode("it 2", f("IT 2"), flagTypeNone, types.GenerateCodeLocation(0), 0)
				}, flagTypeNone, types.GenerateCodeLocation(0))

				specSuite.pushItNode("top level it", f("top IT"), flagTypeNone, types.GenerateCodeLocation(0), 0)
			})

			JustBeforeEach(func() {
				runResult = specSuite.run(fakeT, "suite description", []Reporter{fakeR}, config.GinkgoConfigType{
					RandomSeed:        randomSeed,
					RandomizeAllSpecs: randomizeAllSpecs,
					FocusString:       focusString,
					ParallelNode:      parallelNode,
					ParallelTotal:     parallelTotal,
				})
			})

			It("provides the config and suite description to the reporter", func() {
				Ω(fakeR.Config.RandomSeed).Should(Equal(int64(randomSeed)))
				Ω(fakeR.Config.RandomizeAllSpecs).Should(Equal(randomizeAllSpecs))
				Ω(fakeR.BeginSummary.SuiteDescription).Should(Equal("suite description"))
			})

			It("provides information about the current test", func() {
				description := CurrentGinkgoTestDescription()
				Ω(description.ComponentTexts).Should(Equal([]string{"Suite", "running a suite", "provides information about the current test"}))
				Ω(description.FullTestText).Should(Equal("Suite running a suite provides information about the current test"))
				Ω(description.TestText).Should(Equal("provides information about the current test"))
				Ω(description.IsMeasurement).Should(BeFalse())
				Ω(description.FileName).Should(ContainSubstring("suite_test.go"))
				Ω(description.LineNumber).Should(BeNumerically(">", 50))
				Ω(description.LineNumber).Should(BeNumerically("<", 150))
			})

			Measure("should run measurements", func(b Benchmarker) {
				r := rand.New(rand.NewSource(time.Now().UnixNano()))

				runtime := b.Time("sleeping", func() {
					sleepTime := time.Duration(r.Float64() * 0.01 * float64(time.Second))
					time.Sleep(sleepTime)
				})
				Ω(runtime.Seconds()).Should(BeNumerically("<=", 0.012))
				Ω(runtime.Seconds()).Should(BeNumerically(">=", 0))

				randomValue := r.Float64() * 10.0
				b.RecordValue("random value", randomValue)
				Ω(randomValue).Should(BeNumerically("<=", 10.0))
				Ω(randomValue).Should(BeNumerically(">=", 0.0))
			}, 10)

			It("creates a node hierarchy, converts it to an example collection, and runs it", func() {
				Ω(runOrder).Should(Equal([]string{
					"top BE", "BE", "top JBE", "JBE", "IT", "AE", "top AE",
					"top BE", "BE", "top JBE", "JBE", "inner IT", "AE", "top AE",
					"top BE", "BE 2", "top JBE", "IT 2", "top AE",
					"top BE", "top JBE", "top IT", "top AE",
				}))
			})

			Context("when told to randomize all examples", func() {
				BeforeEach(func() {
					randomizeAllSpecs = true
				})

				It("does", func() {
					Ω(runOrder).Should(Equal([]string{
						"top BE", "top JBE", "top IT", "top AE",
						"top BE", "BE", "top JBE", "JBE", "inner IT", "AE", "top AE",
						"top BE", "BE", "top JBE", "JBE", "IT", "AE", "top AE",
						"top BE", "BE 2", "top JBE", "IT 2", "top AE",
					}))
				})
			})

			Describe("with ginkgo.parallel.total > 1", func() {
				BeforeEach(func() {
					parallelTotal = 2
					randomizeAllSpecs = true
				})

				Context("for one worker", func() {
					BeforeEach(func() {
						parallelNode = 1
					})

					It("should run a subset of tests", func() {
						Ω(runOrder).Should(Equal([]string{
							"top BE", "top JBE", "top IT", "top AE",
							"top BE", "BE", "top JBE", "JBE", "inner IT", "AE", "top AE",
						}))
					})
				})

				Context("for another worker", func() {
					BeforeEach(func() {
						parallelNode = 2
					})

					It("should run a (different) subset of tests", func() {
						Ω(runOrder).Should(Equal([]string{
							"top BE", "BE", "top JBE", "JBE", "IT", "AE", "top AE",
							"top BE", "BE 2", "top JBE", "IT 2", "top AE",
						}))
					})
				})
			})

			Context("when provided with a filter", func() {
				BeforeEach(func() {
					focusString = `inner|\d`
				})

				It("converts the filter to a regular expression and uses it to filter the running examples", func() {
					Ω(runOrder).Should(Equal([]string{
						"top BE", "BE", "top JBE", "JBE", "inner IT", "AE", "top AE",
						"top BE", "BE 2", "top JBE", "IT 2", "top AE",
					}))
				})
			})

			Context("when the specs pass", func() {
				It("doesn't report a failure", func() {
					Ω(fakeT.didFail).Should(BeFalse())
				})

				It("should return true", func() {
					Ω(runResult).Should(BeTrue())
				})
			})

			Context("when a spec fails", func() {
				var location types.CodeLocation
				BeforeEach(func() {
					specSuite.pushItNode("top level it", func() {
						location = types.GenerateCodeLocation(0)
						func() { specSuite.fail("oops!", 0) }()
					}, flagTypeNone, types.GenerateCodeLocation(0), 0)
				})

				It("should return false", func() {
					Ω(runResult).Should(BeFalse())
				})

				It("reports a failure", func() {
					Ω(fakeT.didFail).Should(BeTrue())
				})

				It("generates the correct failure data", func() {
					Ω(fakeR.ExampleSummaries[0].Failure.Message).Should(Equal("oops!"))
					Ω(fakeR.ExampleSummaries[0].Failure.Location.FileName).Should(Equal(location.FileName))
					Ω(fakeR.ExampleSummaries[0].Failure.Location.LineNumber).Should(Equal(location.LineNumber + 1))
				})
			})
		})
	})
}
