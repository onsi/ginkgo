package internal_integration_test

import (
	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"

	. "github.com/onsi/ginkgo/v2/internal/test_helpers"
)

var _ = Describe("Skip", func() {
	Context("When Skip() is called in individual subject and setup nodes", func() {
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

	Context("when Skip() is called in BeforeSuite", func() {
		BeforeEach(func() {
			success, _ := RunFixture("Skip() BeforeSuite", func() {
				BeforeSuite(func() {
					rt.Run("befs")
					Skip("skip please")
				})
				Describe("container to ensure order", func() {
					It("A", rt.T("A"))
					It("B", rt.T("B"))
					It("C", rt.T("C"))
				})
			})

			Ω(success).Should(BeTrue())
		})

		It("skips all the tsts", func() {
			Ω(rt).Should(HaveTracked("befs"))
			Ω(reporter.Did.FindByLeafNodeType(types.NodeTypeBeforeSuite)).Should(HaveBeenSkippedWithMessage("skip please"))
			Ω(reporter.Did.Find("A")).Should(HaveBeenSkipped())
			Ω(reporter.Did.Find("B")).Should(HaveBeenSkipped())
			Ω(reporter.Did.Find("C")).Should(HaveBeenSkipped())
		})

		It("report on the suite with accurate numbers", func() {
			Ω(reporter.End).Should(BeASuiteSummary(true, NPassed(0), NSkipped(3), NPending(0), NSpecs(3), NWillRun(3)))
			Ω(reporter.End.SpecialSuiteFailureReasons).Should(ContainElement("Suite skipped in BeforeSuite"))
		})
	})
})
