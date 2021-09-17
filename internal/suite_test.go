package internal_test

import (
	"fmt"
	"io"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/internal"
	"github.com/onsi/ginkgo/internal/interrupt_handler"
	. "github.com/onsi/ginkgo/internal/test_helpers"
	"github.com/onsi/ginkgo/types"
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

	BeforeEach(func() {
		failer = internal.NewFailer()
		reporter = &FakeReporter{}
		writer = internal.NewWriter(io.Discard)
		outputInterceptor = NewFakeOutputInterceptor()
		interruptHandler = interrupt_handler.NewInterruptHandler(0, "")
		conf = types.SuiteConfig{
			ParallelTotal: 1,
			ParallelNode:  1,
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
				suite.Run("suite", "/path/to/suite", failer, reporter, writer, outputInterceptor, interruptHandler, conf)
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
				suite.Run("suite", "/path/to/suite", failer, reporter, writer, outputInterceptor, interruptHandler, conf)
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

			Context("when nesting ordered containers", func() {
				It("errors", func() {
					var errors = make([]error, 3)
					errors[0] = suite.PushNode(N(ntCon, "top-level-container", Ordered, func() {
						errors[1] = suite.PushNode(N(ntCon, "inner-container", Ordered, cl, func() {
							errors[2] = suite.PushNode(N(ntIt, "it", func() {}))
						}))
					}))
					Ω(errors[0]).ShouldNot(HaveOccurred())
					Ω(suite.BuildTree()).Should(Succeed())
					Ω(errors[1]).Should(MatchError(types.GinkgoErrors.InvalidNestedOrderedContainer(cl)))
					Ω(errors[2]).ShouldNot(HaveOccurred())
				})
			})

			Context("when pushing a suite node suring PhaseBuildTree", func() {
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

			Context("when pushing a suite node suring PhaseRun", func() {
				It("errors", func() {
					var pushSuiteNodeErr error
					err := suite.PushNode(N(ntIt, "top-level it", func() {
						pushSuiteNodeErr = suite.PushNode(N(types.NodeTypeBeforeSuite, cl))
					}))

					Ω(err).ShouldNot(HaveOccurred())
					Ω(suite.BuildTree()).Should(Succeed())
					suite.Run("suite", "/path/to/suite", failer, reporter, writer, outputInterceptor, interruptHandler, conf)
					Ω(pushSuiteNodeErr).Should(HaveOccurred())
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
