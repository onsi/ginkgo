package internal_integration_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/onsi/ginkgo/v2/internal/test_helpers"
	"github.com/onsi/ginkgo/v2/types"
)

var _ = Describe("handling test failures", func() {
	Describe("when BeforeSuite fails", func() {
		BeforeEach(func() {
			success, _ := RunFixture("failed beforesuite", func() {
				BeforeSuite(rt.T("before-suite", func() {
					writer.Write([]byte("before-suite"))
					F("fail", cl)
				}))
				It("A", rt.T("A"))
				It("B", rt.T("B"))
				AfterSuite(rt.T("after-suite"))
			})
			Ω(success).Should(BeFalse())
		})

		It("reports a suite failure", func() {
			Ω(reporter.End).Should(BeASuiteSummary(false, NSpecs(2), NSkipped(0)))
		})

		It("reports a failure for the BeforeSuite", func() {
			Ω(reporter.Did.FindByLeafNodeType(types.NodeTypeBeforeSuite)).Should(HaveFailed("fail", cl, CapturedGinkgoWriterOutput("before-suite")))
			Ω(reporter.Did.FindByLeafNodeType(types.NodeTypeAfterSuite)).Should(HavePassed())
		})

		It("does not run any of the Its", func() {
			Ω(rt).ShouldNot(HaveRun("A"))
			Ω(rt).ShouldNot(HaveRun("B"))
		})

		It("does run the AfterSuite", func() {
			Ω(rt).Should(HaveTracked("before-suite", "after-suite"))
		})
	})

	Describe("when BeforeSuite panics", func() {
		BeforeEach(func() {
			success, _ := RunFixture("panicked beforesuite", func() {
				BeforeSuite(rt.T("before-suite", func() {
					writer.Write([]byte("before-suite"))
					panic("boom")
				}))
				It("A", rt.T("A"))
				It("B", rt.T("B"))
				AfterSuite(rt.T("after-suite"))
			})
			Ω(success).Should(BeFalse())
		})

		It("reports a suite failure", func() {
			Ω(reporter.End).Should(BeASuiteSummary(false, NSpecs(2), NSkipped(0)))
		})

		It("reports a failure for the BeforeSuite", func() {
			Ω(reporter.Did.FindByLeafNodeType(types.NodeTypeBeforeSuite)).Should(HavePanicked("boom", CapturedGinkgoWriterOutput("before-suite")))
		})

		It("does not run any of the Its", func() {
			Ω(rt).ShouldNot(HaveRun("A"))
			Ω(rt).ShouldNot(HaveRun("B"))
		})

		It("does run the AfterSuite", func() {
			Ω(rt).Should(HaveTracked("before-suite", "after-suite"))
		})
	})

	Describe("when AfterSuite fails/panics", func() {
		BeforeEach(func() {
			success, _ := RunFixture("failed aftersuite", func() {
				BeforeSuite(rt.T("before-suite"))
				Describe("top-level", func() {
					It("A", rt.T("A"))
					It("B", rt.T("B"))
				})
				AfterSuite(rt.T("after-suite", func() {
					writer.Write([]byte("after-suite"))
					F("fail", cl)
				}))
			})
			Ω(success).Should(BeFalse())
		})

		It("reports a suite failure", func() {
			Ω(reporter.End).Should(BeASuiteSummary(false, NSpecs(2), NPassed(2)))
		})

		It("runs and reports on all the tests and reports a failure for the AfterSuite", func() {
			Ω(rt).Should(HaveTracked("before-suite", "A", "B", "after-suite"))
			Ω(reporter.Did.FindByLeafNodeType(types.NodeTypeBeforeSuite)).Should(HavePassed())
			Ω(reporter.Did.Find("A")).Should(HavePassed())
			Ω(reporter.Did.Find("B")).Should(HavePassed())
			Ω(reporter.Did.FindByLeafNodeType(types.NodeTypeAfterSuite)).Should(HaveFailed("fail", cl, CapturedGinkgoWriterOutput("after-suite")))
		})
	})

	Describe("individual test falures", func() {
		Describe("when an It fails", func() {
			BeforeEach(func() {
				success, _ := RunFixture("failed it", func() {
					BeforeSuite(rt.T("before-suite"))
					Describe("top-level", func() {
						It("A", rt.T("A", func() {
							writer.Write([]byte("running A"))
						}))
						It("B", rt.T("B", func() {
							writer.Write([]byte("running B"))
							F("fail", cl)
						}))
						It("C", rt.T("C"))
					})
					AfterEach(rt.T("after-each"))
					AfterSuite(rt.T("after-suite"))
				})
				Ω(success).Should(BeFalse())
			})

			It("reports a suite failure", func() {
				Ω(reporter.End).Should(BeASuiteSummary(false, NSpecs(3), NPassed(2), NFailed(1)))
			})

			It("runs other Its, the AfterEach, and the AfterSuite", func() {
				Ω(rt).Should(HaveTracked("before-suite", "A", "after-each", "B", "after-each", "C", "after-each", "after-suite"))
			})

			It("reports the It's failure", func() {
				Ω(reporter.Did.Find("A")).Should(HavePassed(CapturedGinkgoWriterOutput("running A")))
				Ω(reporter.Did.Find("B")).Should(HaveFailed("fail", cl, CapturedGinkgoWriterOutput("running B")))
				Ω(reporter.Did.Find("C")).Should(HavePassed())
			})

			It("sets up the failure node location correctly", func() {
				report := reporter.Did.Find("B")
				Ω(report.Failure.FailureNodeContext).Should(Equal(types.FailureNodeIsLeafNode))
				Ω(report.Failure.FailureNodeType).Should(Equal(types.NodeTypeIt))
				Ω(report.Failure.FailureNodeLocation).Should(Equal(report.LeafNodeLocation))
			})
		})

		Describe("when an It panics", func() {
			BeforeEach(func() {
				success, _ := RunFixture("panicked it", func() {
					BeforeSuite(rt.T("before-suite"))
					Describe("top-level", func() {
						It("A", rt.T("A", func() {
							writer.Write([]byte("running A"))
						}))
						It("B", rt.T("B", func() {
							writer.Write([]byte("running B"))
							panic("boom")
						}))
						It("C", rt.T("C"))
					})
					AfterSuite(rt.T("after-suite"))
				})
				Ω(success).Should(BeFalse())
			})

			It("reports a suite failure", func() {
				Ω(reporter.End).Should(BeASuiteSummary(false, NSpecs(3), NPassed(2), NFailed(1)))
			})

			It("runs other Its and the AfterSuite", func() {
				Ω(rt).Should(HaveTracked("before-suite", "A", "B", "C", "after-suite"))
			})

			It("reports the It's failure", func() {
				Ω(reporter.Did.Find("A")).Should(HavePassed(CapturedGinkgoWriterOutput("running A")))
				Ω(reporter.Did.Find("B")).Should(HavePanicked("boom", CapturedGinkgoWriterOutput("running B")))
				Ω(reporter.Did.Find("C")).Should(HavePassed())
			})

			It("sets up the failure node location correctly", func() {
				report := reporter.Did.Find("B")
				Ω(report.Failure.FailureNodeContext).Should(Equal(types.FailureNodeIsLeafNode))
				Ω(report.Failure.FailureNodeType).Should(Equal(types.NodeTypeIt))
				Ω(report.Failure.FailureNodeLocation).Should(Equal(report.LeafNodeLocation))
			})
		})

		Describe("when a BeforeEach fails/panics", func() {
			BeforeEach(func() {
				success, _ := RunFixture("failed before each", func() {
					BeforeEach(rt.T("bef-1"))
					JustBeforeEach(rt.T("jus-bef-1"))
					Describe("top-level", func() {
						BeforeEach(rt.T("bef-2", func() {
							writer.Write([]byte("bef-2 is running"))
							F("fail", cl)
						}))
						JustBeforeEach(rt.T("jus-bef-2"))
						Describe("nested", func() {
							BeforeEach(rt.T("bef-3"))
							JustBeforeEach(rt.T("jus-bef-3"))
							It("the test", rt.T("it"))
							JustAfterEach(rt.T("jus-aft-3"))
							AfterEach(rt.T("aft-3"))
						})
						JustAfterEach(rt.T("jus-aft-2"))
						AfterEach(rt.T("aft-2", func() {
							writer.Write([]byte("aft-2 is running"))
						}))
					})
					JustAfterEach(rt.T("jus-aft-1"))
					AfterEach(rt.T("aft-1"))
				})
				Ω(success).Should(BeFalse())
			})

			It("reports a suite failure and a spec failure", func() {
				Ω(reporter.End).Should(BeASuiteSummary(false, NSpecs(1), NPassed(0), NFailed(1)))
				specReport := reporter.Did.Find("the test")
				Ω(specReport).Should(HaveFailed("fail", cl), CapturedGinkgoWriterOutput("bef-2 is runningaft-2 is running"))
			})

			It("sets up the failure node location correctly", func() {
				report := reporter.Did.Find("the test")
				Ω(report.Failure.FailureNodeContext).Should(Equal(types.FailureNodeInContainer))
				Ω(report.Failure.FailureNodeType).Should(Equal(types.NodeTypeBeforeEach))
				Ω(report.Failure.FailureNodeContainerIndex).Should(Equal(0))
			})

			It("runs the JustAfterEaches and AfterEaches at the same or lesser nesting level", func() {
				Ω(rt).Should(HaveTracked("bef-1", "bef-2", "jus-aft-2", "jus-aft-1", "aft-2", "aft-1"))
			})
		})

		Describe("when a top-level BeforeEach fails/panics", func() {
			BeforeEach(func() {
				success, _ := RunFixture("failed before each", func() {
					BeforeEach(rt.T("bef-1", func() {
						F("fail", cl)
					}))
					It("the test", rt.T("it"))
				})
				Ω(success).Should(BeFalse())
			})

			It("sets up the failure node location correctly", func() {
				report := reporter.Did.Find("the test")
				Ω(report.Failure.FailureNodeContext).Should(Equal(types.FailureNodeAtTopLevel))
				Ω(report.Failure.FailureNodeType).Should(Equal(types.NodeTypeBeforeEach))
			})
		})

		Describe("when an AfterEach fails/panics", func() {
			BeforeEach(func() {
				success, _ := RunFixture("failed after each", func() {
					BeforeEach(rt.T("bef-1"))
					JustBeforeEach(rt.T("jus-bef-1"))
					Describe("top-level", func() {
						BeforeEach(rt.T("bef-2"))
						Describe("nested", func() {
							BeforeEach(rt.T("bef-3"))
							It("the test", rt.T("it"))
							JustAfterEach(rt.T("jus-aft-3"))
							AfterEach(rt.T("aft-3", func() {
								F("fail", types.NewCodeLocation(0))
							}))
						})
						JustAfterEach(rt.T("jus-aft-2"))
						AfterEach(rt.T("aft-2", func() {
							writer.Write([]byte("aft-2 is running"))
						}))
					})
					JustAfterEach(rt.T("jus-aft-1"))
					AfterEach(rt.T("aft-1"))
				})
				Ω(success).Should(BeFalse())
			})

			It("reports a suite failure and a spec failure", func() {
				Ω(reporter.End).Should(BeASuiteSummary(false, NSpecs(1), NPassed(0), NFailed(1)))
				specReport := reporter.Did.Find("the test")
				Ω(specReport).Should(HaveFailed("fail"), CapturedGinkgoWriterOutput("aft-2 is running"))
			})

			It("sets up the failure node location correctly", func() {
				report := reporter.Did.Find("the test")
				Ω(report.Failure.FailureNodeContext).Should(Equal(types.FailureNodeInContainer))
				Ω(report.Failure.FailureNodeType).Should(Equal(types.NodeTypeAfterEach))
				location := report.Failure.Location
				location.LineNumber = location.LineNumber - 1
				Ω(report.Failure.FailureNodeLocation).Should(Equal(location))
				Ω(report.Failure.FailureNodeContainerIndex).Should(Equal(1))
			})

			It("runs the subsequent after eaches", func() {
				Ω(rt).Should(HaveTracked("bef-1", "bef-2", "bef-3", "jus-bef-1", "it", "jus-aft-3", "jus-aft-2", "jus-aft-1", "aft-3", "aft-2", "aft-1"))
			})
		})

		Describe("when multiple nodes within a given test run and fail", func() {
			var clA, clB types.CodeLocation
			BeforeEach(func() {
				clA = types.CodeLocation{FileName: "A"}
				clB = types.CodeLocation{FileName: "B"}
				success, _ := RunFixture("failed after each", func() {
					BeforeEach(rt.T("bef-1", func() {
						writer.Write([]byte("run A"))
						F("fail-A", clA)
					}))
					It("the test", rt.T("it"))
					AfterEach(rt.T("aft-1", func() {
						writer.Write([]byte("run B"))
						F("fail-B", clB)
					}))
				})
				Ω(success).Should(BeFalse())
			})

			It("reports a suite failure and a spec failure and only tracks the first failure", func() {
				Ω(reporter.End).Should(BeASuiteSummary(false, NSpecs(1), NPassed(0), NFailed(1)))
				specReport := reporter.Did.Find("the test")
				Ω(specReport).Should(HaveFailed("fail-A", clA), CapturedGinkgoWriterOutput("run Arun B"))
				Ω(specReport.Failure.FailureNodeType).Should(Equal(types.NodeTypeBeforeEach))
				Ω(rt).Should(HaveTracked("bef-1", "aft-1"))
			})
		})
	})

	Describe("when there are multiple tests that fail", func() {
		BeforeEach(func() {
			success, _ := RunFixture("failed after each", func() {
				It("A", func() { F() })
				It("B", func() { F() })
				It("C", func() {})
				It("D", func() { F() })
				It("E", func() {})
				It("F", func() { panic("boom") })
			})
			Ω(success).Should(BeFalse())
		})

		It("reports the correct number of failures", func() {
			Ω(reporter.End).Should(BeASuiteSummary(false, NSpecs(6), NPassed(2), NFailed(4)))
			Ω(reporter.Did.WithState(types.SpecStatePassed).Names()).Should(ConsistOf("C", "E"))
			Ω(reporter.Did.WithState(types.SpecStateFailed).Names()).Should(ConsistOf("A", "B", "D"))
			Ω(reporter.Did.WithState(types.SpecStatePanicked).Names()).Should(ConsistOf("F"))
		})
	})
})
