package internal_integration_test

import (
	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"

	. "github.com/onsi/ginkgo/v2/internal/test_helpers"
)

var _ = Describe("Focus", func() {
	Describe("when a suite has pending tests", func() {
		fixture := func() {
			It("A", rt.T("A"))
			It("B", rt.T("B"))
			It("C", rt.T("C"), Pending)
			Describe("container", func() {
				It("D", rt.T("D"))
			})
			PDescribe("pending container", func() {
				It("E", rt.T("E"))
				It("F", rt.T("F"))
			})
		}
		Context("without config.FailOnPending", func() {
			BeforeEach(func() {
				success, hPF := RunFixture("pending tests", fixture)
				Ω(success).Should(BeTrue())
				Ω(hPF).Should(BeFalse())
			})

			It("should not report that the suite hasProgrammaticFocus", func() {
				Ω(reporter.Begin.SuiteHasProgrammaticFocus).Should(BeFalse())
				Ω(reporter.End.SuiteHasProgrammaticFocus).Should(BeFalse())
			})

			It("does not run the pending tests", func() {
				Ω(rt.TrackedRuns()).Should(ConsistOf("A", "B", "D"))
			})

			It("reports on the pending tests", func() {
				Ω(reporter.Did.WithState(types.SpecStatePassed).Names()).Should(ConsistOf("A", "B", "D"))
				Ω(reporter.Did.WithState(types.SpecStatePending).Names()).Should(ConsistOf("C", "E", "F"))
			})

			It("reports on the suite with accurate numbers", func() {
				Ω(reporter.End).Should(BeASuiteSummary(true, NSpecs(6), NPassed(3), NPending(3), NWillRun(3), NSkipped(0)))
			})

			It("does not include a special suite failure reason", func() {
				Ω(reporter.End.SpecialSuiteFailureReasons).Should(BeEmpty())
			})
		})

		Context("with config.FailOnPending", func() {
			BeforeEach(func() {
				conf.FailOnPending = true
				success, hPF := RunFixture("pending tests", fixture)
				Ω(success).Should(BeFalse())
				Ω(hPF).Should(BeFalse())
			})

			It("reports on the suite with accurate numbers", func() {
				Ω(reporter.End).Should(BeASuiteSummary(false, NPassed(3), NSpecs(6), NPending(3), NWillRun(3), NSkipped(0)))
			})

			It("includes a special suite failure reason", func() {
				Ω(reporter.End.SpecialSuiteFailureReasons).Should(ContainElement("Detected pending specs and --fail-on-pending is set"))
			})
		})
	})

	Describe("with programmatic focus", func() {
		var success bool
		var hasProgrammaticFocus bool
		BeforeEach(func() {
			success, hasProgrammaticFocus = RunFixture("focused tests", func() {
				It("A", rt.T("A"))
				It("B", rt.T("B"))
				FDescribe("focused container", func() {
					It("C", rt.T("C"))
					It("D", rt.T("D"))
					PIt("E", rt.T("E"))
				})
				FDescribe("focused container with focused child", func() {
					It("F", rt.T("F"))
					It("G", Focus, rt.T("G"))
				})
				Describe("container", func() {
					It("H", rt.T("H"))
				})
				FIt("I", rt.T("I"))
			})
			Ω(success).Should(BeTrue())
		})

		It("should return true for hasProgrammaticFocus", func() {
			Ω(hasProgrammaticFocus).Should(BeTrue())
		})

		It("should report that the suite hasProgrammaticFocus", func() {
			Ω(reporter.Begin.SuiteHasProgrammaticFocus).Should(BeTrue())
			Ω(reporter.End.SuiteHasProgrammaticFocus).Should(BeTrue())
		})

		It("should run the focused tests, honoring the nested focus policy", func() {
			Ω(rt.TrackedRuns()).Should(ConsistOf("C", "D", "G", "I"))
		})

		It("should report on the tests correctly", func() {
			Ω(reporter.Did.WithState(types.SpecStateSkipped).Names()).Should(ConsistOf("A", "B", "F", "H"))
			Ω(reporter.Did.WithState(types.SpecStatePending).Names()).Should(ConsistOf("E"))
			Ω(reporter.Did.WithState(types.SpecStatePassed).Names()).Should(ConsistOf("C", "D", "G", "I"))
		})

		It("report on the suite with accurate numbers", func() {
			Ω(reporter.End).Should(BeASuiteSummary(true, NPassed(4), NSkipped(4), NPending(1), NSpecs(9), NWillRun(4)))
		})
	})

	Describe("with config.FocusStrings and config.SkipStrings", func() {
		BeforeEach(func() {
			conf.FocusStrings = []string{"blue", "green"}
			conf.SkipStrings = []string{"red"}
			success, _ := RunFixture("cli focus tests", func() {
				It("blue.1", rt.T("blue.1"))
				It("blue.2", rt.T("blue.2"))
				Describe("blue.container", func() {
					It("yellow.1", rt.T("yellow.1"))
					It("red.1", rt.T("red.1"))
					PIt("blue.3", rt.T("blue.3"))
				})
				Describe("green.container", func() {
					It("yellow.2", rt.T("yellow.2"))
					It("green.1", rt.T("green.1"))
				})
				Describe("red.2", func() {
					It("green.2", rt.T("green.2"))
				})
				FIt("red.3", rt.T("red.3"))
			})
			Ω(success).Should(BeTrue())
		})

		It("should run tests that match", func() {
			Ω(rt.TrackedRuns()).Should(ConsistOf("blue.1", "blue.2", "yellow.1", "yellow.2", "green.1"))
		})

		It("should report on the tests correctly", func() {
			Ω(reporter.Did.WithState(types.SpecStateSkipped).Names()).Should(ConsistOf("red.1", "green.2", "red.3"))
			Ω(reporter.Did.WithState(types.SpecStatePending).Names()).Should(ConsistOf("blue.3"))
			Ω(reporter.Did.WithState(types.SpecStatePassed).Names()).Should(ConsistOf("blue.1", "blue.2", "yellow.1", "yellow.2", "green.1"))
		})

		It("report on the suite with accurate numbers", func() {
			Ω(reporter.End).Should(BeASuiteSummary(true, NPassed(5), NSkipped(3), NPending(1), NSpecs(9), NWillRun(5)))
		})
	})

	Describe("when no tests will end up running", func() {
		BeforeEach(func() {
			conf.FocusStrings = []string{"red"}
			success, _ := RunFixture("cli focus tests", func() {
				BeforeSuite(rt.T("bef-suite"))
				AfterSuite(rt.T("aft-suite"))
				It("blue.1", rt.T("blue.1"))
				It("blue.2", rt.T("blue.2"))
			})
			Ω(success).Should(BeTrue())
		})

		It("does not run the BeforeSuite or the AfterSuite", func() {
			Ω(rt).Should(HaveTrackedNothing())
		})
	})

	Describe("Skip()", func() {
		BeforeEach(func() {
			success, _ := RunFixture("Skip() tests", func() {
				Describe("container to ensure order", func() {
					It("A", rt.T("A"))
					Describe("container", func() {
						BeforeEach(rt.T("bef", func() {
							failer.Skip("skip in Bef", cl)
							panic("boom") //simulates what Ginkgo DSL does
						}))
						It("B", rt.T("B"))
						It("C", rt.T("C"))
						AfterEach(rt.T("aft"))
					})
					It("D", rt.T("D", func() {
						failer.Skip("skip D", cl)
						panic("boom") //simulates what Ginkgo DSL does
					}))
				})
			})

			Ω(success).Should(BeTrue())
		})

		It("skips the tests that are Skipped()", func() {
			Ω(rt).Should(HaveTracked("A", "bef", "aft", "bef", "aft", "D"))
			Ω(reporter.Did.WithState(types.SpecStatePassed).Names()).Should(ConsistOf("A"))
			Ω(reporter.Did.WithState(types.SpecStateSkipped).Names()).Should(ConsistOf("B", "C", "D"))

			Ω(reporter.Did.Find("B").Failure.Message).Should(Equal("skip in Bef"))
			Ω(reporter.Did.Find("B").Failure.Location).Should(Equal(cl))

			Ω(reporter.Did.Find("D").Failure.Message).Should(Equal("skip D"))
			Ω(reporter.Did.Find("D").Failure.Location).Should(Equal(cl))
		})

		It("report on the suite with accurate numbers", func() {
			Ω(reporter.End).Should(BeASuiteSummary(true, NPassed(1), NSkipped(3), NPending(0), NSpecs(4), NWillRun(4)))
		})
	})
})
