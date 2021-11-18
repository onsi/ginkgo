package internal_test

import (
	"fmt"
	"io"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/internal"
	"github.com/onsi/ginkgo/v2/internal/interrupt_handler"
	"github.com/onsi/ginkgo/v2/internal/parallel_support"
	. "github.com/onsi/ginkgo/v2/internal/test_helpers"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
)

var _ = Describe("Suite", func() {
	It("is heavily integration tested over in internal_integration", func() {
	})

	var suite *internal.Suite
	var failer *internal.Failer
	var reporter *FakeReporter
	var writer *internal.Writer
	var outputInterceptor *FakeOutputInterceptor
	var interruptHandler *interrupt_handler.InterruptHandler
	var conf types.SuiteConfig
	var rt *RunTracker
	var client parallel_support.Client

	BeforeEach(func() {
		failer = internal.NewFailer()
		reporter = &FakeReporter{}
		writer = internal.NewWriter(io.Discard)
		outputInterceptor = NewFakeOutputInterceptor()
		client = nil
		interruptHandler = interrupt_handler.NewInterruptHandler(0, client)
		DeferCleanup(interruptHandler.Stop)
		conf = types.SuiteConfig{
			ParallelTotal:   1,
			ParallelProcess: 1,
		}
		rt = NewRunTracker()
		suite = internal.NewSuite()
	})

	Describe("Constructing Trees", func() {
		Describe("PhaseBuildTopLevel vs PhaseBuildTree", func() {
			var err1, err2, err3 error
			BeforeEach(func() {
				err1 = suite.PushNode(N(ntCon, "a top-level container", func() {
					rt.Run("traversing outer")
					err2 = suite.PushNode(N(ntCon, "a nested container", func() {
						rt.Run("traversing nested")
						err3 = suite.PushNode(N(ntIt, "an it", rt.T("running it")))
					}))
				}))
			})

			It("only traverses top-level containers when told to BuildTree", func() {
				fmt.Fprintln(GinkgoWriter, "HELLO!")
				Ω(rt).Should(HaveTrackedNothing())
				Ω(suite.BuildTree()).Should(Succeed())
				Ω(rt).Should(HaveTracked("traversing outer", "traversing nested"))

				rt.Reset()
				suite.Run("suite", Labels{}, "/path/to/suite", failer, reporter, writer, outputInterceptor, interruptHandler, client, conf)
				Ω(rt).Should(HaveTracked("running it"))

				Ω(err1).ShouldNot(HaveOccurred())
				Ω(err2).ShouldNot(HaveOccurred())
				Ω(err3).ShouldNot(HaveOccurred())
			})
		})

		Context("when pushing nodes during PhaseRun", func() {
			var pushNodeErrDuringRun error

			BeforeEach(func() {
				err := suite.PushNode(N(ntCon, "a top-level container", func() {
					suite.PushNode(N(ntIt, "an it", func() {
						rt.Run("in it")
						pushNodeErrDuringRun = suite.PushNode(N(ntIt, "oops - illegal operation", cl, rt.T("illegal")))
					}))
				}))

				Ω(err).ShouldNot(HaveOccurred())
				Ω(suite.BuildTree()).Should(Succeed())
			})

			It("errors", func() {
				suite.Run("suite", Labels{}, "/path/to/suite", failer, reporter, writer, outputInterceptor, interruptHandler, client, conf)
				Ω(pushNodeErrDuringRun).Should(HaveOccurred())
				Ω(rt).Should(HaveTracked("in it"))
			})

		})

		Context("when the user attemps to fail during PhaseBuildTree", func() {
			BeforeEach(func() {
				suite.PushNode(N(ntCon, "a top-level container", func() {
					failer.Fail("boom", cl)
					panic("simulate ginkgo panic")
				}))
			})

			It("errors", func() {
				err := suite.BuildTree()
				Ω(err.Error()).Should(ContainSubstring(cl.String()))
				Ω(err.Error()).Should(ContainSubstring("simulate ginkgo panic"))
			})
		})

		Context("when the user panics during PhaseBuildTree", func() {
			BeforeEach(func() {
				suite.PushNode(N(ntCon, "a top-level container", func() {
					panic("boom")
				}))
			})

			It("errors", func() {
				err := suite.BuildTree()
				Ω(err).Should(HaveOccurred())
				Ω(err.Error()).Should(ContainSubstring("boom"))
			})
		})

		Describe("Suite Nodes", func() {
			Context("when pushing suite nodes at the top level", func() {
				BeforeEach(func() {
					err := suite.PushNode(N(types.NodeTypeBeforeSuite))
					Ω(err).ShouldNot(HaveOccurred())

					err = suite.PushNode(N(types.NodeTypeAfterSuite))
					Ω(err).ShouldNot(HaveOccurred())
				})

				Context("when pushing more than one BeforeSuite node", func() {
					It("errors", func() {
						err := suite.PushNode(N(types.NodeTypeBeforeSuite))
						Ω(err).Should(HaveOccurred())

						err = suite.PushNode(N(types.NodeTypeSynchronizedBeforeSuite))
						Ω(err).Should(HaveOccurred())
					})
				})

				Context("when pushing more than one AfterSuite node", func() {
					It("errors", func() {
						err := suite.PushNode(N(types.NodeTypeAfterSuite))
						Ω(err).Should(HaveOccurred())

						err = suite.PushNode(N(types.NodeTypeSynchronizedAfterSuite))
						Ω(err).Should(HaveOccurred())
					})
				})
			})

			Context("when pushing a serial node in an ordered container", func() {
				Context("when the outer-most ordered container is marked serial", func() {
					It("succeeds", func() {
						var errors = make([]error, 3)
						errors[0] = suite.PushNode(N(ntCon, "top-level-container", Ordered, Serial, func() {
							errors[1] = suite.PushNode(N(ntCon, "inner-container", func() {
								errors[2] = suite.PushNode(N(ntIt, "it", Serial, func() {}))
							}))
						}))
						Ω(errors[0]).ShouldNot(HaveOccurred())
						Ω(suite.BuildTree()).Should(Succeed())
						Ω(errors[1]).ShouldNot(HaveOccurred())
						Ω(errors[2]).ShouldNot(HaveOccurred())
					})
				})

				Context("when the outer-most ordered container is not marked serial", func() {
					It("errors", func() {
						var errors = make([]error, 3)
						errors[0] = suite.PushNode(N(ntCon, "top-level-container", Ordered, func() {
							errors[1] = suite.PushNode(N(ntCon, "inner-container", func() {
								errors[2] = suite.PushNode(N(ntIt, "it", Serial, cl, func() {}))
							}))
						}))
						Ω(errors[0]).ShouldNot(HaveOccurred())
						Ω(suite.BuildTree()).Should(Succeed())
						Ω(errors[1]).ShouldNot(HaveOccurred())
						Ω(errors[2]).Should(MatchError(types.GinkgoErrors.InvalidSerialNodeInNonSerialOrderedContainer(cl, ntIt)))
					})
				})
			})

			Context("when pushing BeforeAll and AfterAll nodes", func() {
				Context("in an ordered container", func() {
					It("succeeds", func() {
						var errors = make([]error, 3)
						errors[0] = suite.PushNode(N(ntCon, "top-level-container", Ordered, func() {
							errors[1] = suite.PushNode(N(types.NodeTypeBeforeAll, func() {}))
							errors[2] = suite.PushNode(N(types.NodeTypeAfterAll, func() {}))
						}))
						Ω(errors[0]).ShouldNot(HaveOccurred())
						Ω(suite.BuildTree()).Should(Succeed())
						Ω(errors[1]).ShouldNot(HaveOccurred())
						Ω(errors[2]).ShouldNot(HaveOccurred())
					})
				})

				Context("anywhere else", func() {
					It("errors", func() {
						var errors = make([]error, 3)
						errors[0] = suite.PushNode(N(ntCon, "top-level-container", func() {
							errors[1] = suite.PushNode(N(types.NodeTypeBeforeAll, cl, func() {}))
							errors[2] = suite.PushNode(N(types.NodeTypeAfterAll, cl, func() {}))
						}))
						Ω(errors[0]).ShouldNot(HaveOccurred())
						Ω(suite.BuildTree()).Should(Succeed())
						Ω(errors[1]).Should(MatchError(types.GinkgoErrors.SetupNodeNotInOrderedContainer(cl, types.NodeTypeBeforeAll)))
						Ω(errors[2]).Should(MatchError(types.GinkgoErrors.SetupNodeNotInOrderedContainer(cl, types.NodeTypeAfterAll)))
					})
				})
			})

			Context("when pushing a suite node during PhaseBuildTree", func() {
				It("errors", func() {
					var pushSuiteNodeErr error
					err := suite.PushNode(N(ntCon, "top-level-container", func() {
						pushSuiteNodeErr = suite.PushNode(N(types.NodeTypeBeforeSuite, cl))
					}))

					Ω(err).ShouldNot(HaveOccurred())
					Ω(suite.BuildTree()).Should(Succeed())
					Ω(pushSuiteNodeErr).Should(HaveOccurred())
				})
			})

			Context("when pushing a suite node during PhaseRun", func() {
				It("errors", func() {
					var pushSuiteNodeErr error
					err := suite.PushNode(N(ntIt, "top-level it", func() {
						pushSuiteNodeErr = suite.PushNode(N(types.NodeTypeBeforeSuite, cl))
					}))

					Ω(err).ShouldNot(HaveOccurred())
					Ω(suite.BuildTree()).Should(Succeed())
					suite.Run("suite", Labels{}, "/path/to/suite", failer, reporter, writer, outputInterceptor, interruptHandler, client, conf)
					Ω(pushSuiteNodeErr).Should(HaveOccurred())
				})
			})
		})

		Describe("Cleanup Nodes", func() {
			Context("when pushing a cleanup node during PhaseTopLevel", func() {
				It("errors", func() {
					err := suite.PushNode(N(types.NodeTypeCleanupInvalid, cl))
					Ω(err).Should(MatchError(types.GinkgoErrors.PushingCleanupNodeDuringTreeConstruction(cl)))
				})
			})

			Context("when pushing a cleanup node during PhaseBuildTree", func() {
				It("errors", func() {
					var errors = make([]error, 2)
					errors[0] = suite.PushNode(N(ntCon, "container", func() {
						errors[1] = suite.PushNode(N(types.NodeTypeCleanupInvalid, cl))
					}))
					Ω(errors[0]).ShouldNot(HaveOccurred())
					Ω(suite.BuildTree()).Should(Succeed())
					Ω(errors[1]).Should(MatchError(types.GinkgoErrors.PushingCleanupNodeDuringTreeConstruction(cl)))
				})
			})

			Context("when pushing a cleanup node in a ReportBeforeEach node", func() {
				It("errors", func() {
					var errors = make([]error, 4)
					reportBeforeEachNode, _ := internal.NewReportBeforeEachNode(func(_ types.SpecReport) {
						errors[3] = suite.PushNode(N(types.NodeTypeCleanupInvalid, cl))
					}, types.NewCodeLocation(0))

					errors[0] = suite.PushNode(N(ntCon, "container", func() {
						errors[1] = suite.PushNode(reportBeforeEachNode)
						errors[2] = suite.PushNode(N(ntIt, "test"))
					}))
					Ω(errors[0]).ShouldNot(HaveOccurred())

					Ω(suite.BuildTree()).Should(Succeed())
					Ω(errors[1]).ShouldNot(HaveOccurred())
					Ω(errors[2]).ShouldNot(HaveOccurred())

					suite.Run("suite", Labels{}, "/path/to/suite", failer, reporter, writer, outputInterceptor, interruptHandler, client, conf)
					Ω(errors[3]).Should(MatchError(types.GinkgoErrors.PushingCleanupInReportingNode(cl, types.NodeTypeReportBeforeEach)))
				})
			})

			Context("when pushing a cleanup node in a ReportAfterEach node", func() {
				It("errors", func() {
					var errors = make([]error, 4)
					reportAfterEachNode, _ := internal.NewReportAfterEachNode(func(_ types.SpecReport) {
						errors[3] = suite.PushNode(N(types.NodeTypeCleanupInvalid, cl))
					}, types.NewCodeLocation(0))

					errors[0] = suite.PushNode(N(ntCon, "container", func() {
						errors[1] = suite.PushNode(N(ntIt, "test"))
						errors[2] = suite.PushNode(reportAfterEachNode)
					}))
					Ω(errors[0]).ShouldNot(HaveOccurred())

					Ω(suite.BuildTree()).Should(Succeed())
					Ω(errors[1]).ShouldNot(HaveOccurred())
					Ω(errors[2]).ShouldNot(HaveOccurred())

					suite.Run("suite", Labels{}, "/path/to/suite", failer, reporter, writer, outputInterceptor, interruptHandler, client, conf)
					Ω(errors[3]).Should(MatchError(types.GinkgoErrors.PushingCleanupInReportingNode(cl, types.NodeTypeReportAfterEach)))
				})
			})

			Context("when pushing a cleanup node in a ReportAfterSuite node", func() {
				It("errors", func() {
					var errors = make([]error, 4)
					reportAfterSuiteNode, _ := internal.NewReportAfterSuiteNode("report", func(_ types.Report) {
						errors[3] = suite.PushNode(N(types.NodeTypeCleanupInvalid, cl))
					}, types.NewCodeLocation(0))

					errors[0] = suite.PushNode(N(ntCon, "container", func() {
						errors[2] = suite.PushNode(N(ntIt, "test"))
					}))
					errors[1] = suite.PushNode(reportAfterSuiteNode)
					Ω(errors[0]).ShouldNot(HaveOccurred())
					Ω(errors[1]).ShouldNot(HaveOccurred())

					Ω(suite.BuildTree()).Should(Succeed())
					Ω(errors[2]).ShouldNot(HaveOccurred())

					suite.Run("suite", Labels{}, "/path/to/suite", failer, reporter, writer, outputInterceptor, interruptHandler, client, conf)
					Ω(errors[3]).Should(MatchError(types.GinkgoErrors.PushingCleanupInReportingNode(cl, types.NodeTypeReportAfterSuite)))
				})
			})

			Context("when pushing a cleanup node within a cleanup node", func() {
				It("errors", func() {
					var errors = make([]error, 3)
					errors[0] = suite.PushNode(N(ntIt, "It", func() {
						cleanupNode, _ := internal.NewCleanupNode(nil, types.NewCustomCodeLocation("outerCleanup"), func() {
							innerCleanupNode, _ := internal.NewCleanupNode(nil, cl, func() {})
							errors[2] = suite.PushNode(innerCleanupNode)
						})
						errors[1] = suite.PushNode(cleanupNode)
					}))
					Ω(errors[0]).ShouldNot(HaveOccurred())
					Ω(suite.BuildTree()).Should(Succeed())
					suite.Run("suite", Labels{}, "/path/to/suite", failer, reporter, writer, outputInterceptor, interruptHandler, client, conf)
					Ω(errors[1]).ShouldNot(HaveOccurred())
					Ω(errors[2]).Should(MatchError(types.GinkgoErrors.PushingCleanupInCleanupNode(cl)))
				})
			})
		})

		Describe("ReportEntries", func() {
			Context("when adding a report entry outside of the run phase", func() {
				It("errors", func() {
					entry, err := internal.NewReportEntry("name", cl)
					Ω(err).ShouldNot(HaveOccurred())
					err = suite.AddReportEntry(entry)
					Ω(err).Should(MatchError(types.GinkgoErrors.AddReportEntryNotDuringRunPhase(cl)))
					suite.BuildTree()
					err = suite.AddReportEntry(entry)
					Ω(err).Should(MatchError(types.GinkgoErrors.AddReportEntryNotDuringRunPhase(cl)))
				})
			})
		})
	})
})
