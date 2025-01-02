package parallel_support_test

import (
	"fmt"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"github.com/onsi/ginkgo/v2/internal"
	"github.com/onsi/ginkgo/v2/internal/parallel_support"
	. "github.com/onsi/ginkgo/v2/internal/test_helpers"
	"github.com/onsi/ginkgo/v2/types"
)

type ColorableStringerStruct struct {
	Label string
	Count int
}

func (s ColorableStringerStruct) String() string {
	return fmt.Sprintf("%s %d", s.Label, s.Count)
}

func (s ColorableStringerStruct) ColorableString() string {
	return fmt.Sprintf("{{red}}%s {{green}}%d{{/}}", s.Label, s.Count)
}

var _ = Describe("The Parallel Support Client & Server", func() {
	for _, protocol := range []string{"RPC", "HTTP"} {
		protocol := protocol
		Describe(fmt.Sprintf("The %s protocol", protocol), Label(protocol), func() {
			var (
				server   parallel_support.Server
				client   parallel_support.Client
				reporter *FakeReporter
				buffer   *gbytes.Buffer
			)

			BeforeEach(func() {
				GinkgoT().Setenv("GINKGO_PARALLEL_PROTOCOL", protocol)

				var err error
				reporter = NewFakeReporter()
				server, err = parallel_support.NewServer(3, reporter)
				Ω(err).ShouldNot(HaveOccurred())
				server.Start()

				buffer = gbytes.NewBuffer()
				server.SetOutputDestination(buffer)

				client = parallel_support.NewClient(server.Address())
				Eventually(client.Connect).Should(BeTrue())

				DeferCleanup(server.Close)
				DeferCleanup(client.Close)
			})

			Describe("Reporting endpoints", func() {
				var beginReport, thirdBeginReport types.Report
				var endReport1, endReport2, endReport3 types.Report
				var specReportA, specReportB, specReportC types.SpecReport

				var t time.Time

				BeforeEach(func() {
					beginReport = types.Report{SuiteDescription: "my sweet suite"}
					thirdBeginReport = types.Report{SuiteDescription: "last one in gets forwarded"}

					specReportA = types.SpecReport{LeafNodeText: "A"}
					specReportB = types.SpecReport{LeafNodeText: "B"}
					specReportC = types.SpecReport{LeafNodeText: "C"}

					t = time.Now()

					endReport1 = types.Report{StartTime: t.Add(-time.Second), EndTime: t.Add(time.Second), SuiteSucceeded: true, SpecReports: types.SpecReports{specReportA}}
					endReport2 = types.Report{StartTime: t.Add(-2 * time.Second), EndTime: t.Add(time.Second), SuiteSucceeded: true, SpecReports: types.SpecReports{specReportB}}
					endReport3 = types.Report{StartTime: t.Add(-time.Second), EndTime: t.Add(2 * time.Second), SuiteSucceeded: false, SpecReports: types.SpecReports{specReportC}}
				})

				Context("before all procs have reported SuiteWillBegin", func() {
					BeforeEach(func() {
						Ω(client.PostSuiteWillBegin(beginReport)).Should(Succeed())
						Ω(client.PostDidRun(specReportA)).Should(Succeed())
						Ω(client.PostSuiteWillBegin(beginReport)).Should(Succeed())
						Ω(client.PostDidRun(specReportB)).Should(Succeed())
					})

					It("should not forward anything to the attached reporter", func() {
						Ω(reporter.Begin).Should(BeZero())
						Ω(reporter.Will).Should(BeEmpty())
						Ω(reporter.Did).Should(BeEmpty())
					})

					Context("when the final proc reports SuiteWillBegin", func() {
						BeforeEach(func() {
							Ω(client.PostSuiteWillBegin(thirdBeginReport)).Should(Succeed())
						})

						It("forwards to SuiteWillBegin and catches up on any received summaries", func() {
							Ω(reporter.Begin).Should(Equal(thirdBeginReport))
							Ω(reporter.Will.Names()).Should(ConsistOf("A", "B"))
							Ω(reporter.Did.Names()).Should(ConsistOf("A", "B"))
						})

						Context("any subsequent summaries", func() {
							BeforeEach(func() {
								Ω(client.PostDidRun(specReportC)).Should(Succeed())
							})

							It("are forwarded immediately", func() {
								Ω(reporter.Will.Names()).Should(ConsistOf("A", "B", "C"))
								Ω(reporter.Did.Names()).Should(ConsistOf("A", "B", "C"))
							})
						})

						Context("when SuiteDidEnd start arriving", func() {
							BeforeEach(func() {
								Ω(client.PostSuiteDidEnd(endReport1)).Should(Succeed())
								Ω(client.PostSuiteDidEnd(endReport2)).Should(Succeed())
							})

							It("does not forward them yet...", func() {
								Ω(reporter.End).Should(BeZero())
							})

							It("doesn't signal it's done", func() {
								Ω(server.GetSuiteDone()).ShouldNot(BeClosed())
							})

							Context("when the final SuiteDidEnd arrive", func() {
								BeforeEach(func() {
									Ω(client.PostSuiteDidEnd(endReport3)).Should(Succeed())
								})

								It("forwards the aggregation of all received end summaries", func() {
									Ω(reporter.End.StartTime.Unix()).Should(BeNumerically("~", t.Add(-2*time.Second).Unix()))
									Ω(reporter.End.EndTime.Unix()).Should(BeNumerically("~", t.Add(2*time.Second).Unix()))
									Ω(reporter.End.RunTime).Should(BeNumerically("~", 4*time.Second))
									Ω(reporter.End.SuiteSucceeded).Should(BeFalse())
									Ω(reporter.End.SpecReports).Should(ConsistOf(specReportA, specReportB, specReportC))
								})

								It("should signal it's done", func() {
									Ω(server.GetSuiteDone()).Should(BeClosed())
								})
							})
						})
					})
				})
			})

			Describe("supporting ReportEntries (which RPC struggled with when I first implemented it)", func() {
				BeforeEach(func() {
					Ω(client.PostSuiteWillBegin(types.Report{SuiteDescription: "my sweet suite"})).Should(Succeed())
					Ω(client.PostSuiteWillBegin(types.Report{SuiteDescription: "my sweet suite"})).Should(Succeed())
					Ω(client.PostSuiteWillBegin(types.Report{SuiteDescription: "my sweet suite"})).Should(Succeed())
				})
				It("can pass in ReportEntries that include custom types", func() {
					cl := types.NewCodeLocation(0)
					entry, err := internal.NewReportEntry("No Value Entry", cl)
					Ω(err).ShouldNot(HaveOccurred())
					Ω(client.PostDidRun(types.SpecReport{
						LeafNodeText:  "no-value",
						ReportEntries: types.ReportEntries{entry},
					})).Should(Succeed())

					entry, err = internal.NewReportEntry("String Value Entry", cl, "The String")
					Ω(err).ShouldNot(HaveOccurred())
					Ω(client.PostDidRun(types.SpecReport{
						LeafNodeText:  "string-value",
						ReportEntries: types.ReportEntries{entry},
					})).Should(Succeed())

					entry, err = internal.NewReportEntry("Custom Type Value Entry", cl, ColorableStringerStruct{Label: "apples", Count: 17})
					Ω(err).ShouldNot(HaveOccurred())
					Ω(client.PostDidRun(types.SpecReport{
						LeafNodeText:  "custom-value",
						ReportEntries: types.ReportEntries{entry},
					})).Should(Succeed())

					Ω(reporter.Did.Find("no-value").ReportEntries[0].Name).Should(Equal("No Value Entry"))
					Ω(reporter.Did.Find("no-value").ReportEntries[0].StringRepresentation()).Should(Equal(""))

					Ω(reporter.Did.Find("string-value").ReportEntries[0].Name).Should(Equal("String Value Entry"))
					Ω(reporter.Did.Find("string-value").ReportEntries[0].StringRepresentation()).Should(Equal("The String"))

					Ω(reporter.Did.Find("custom-value").ReportEntries[0].Name).Should(Equal("Custom Type Value Entry"))
					Ω(reporter.Did.Find("custom-value").ReportEntries[0].StringRepresentation()).Should(Equal("{{red}}apples {{green}}17{{/}}"))
				})
			})

			Describe("Streaming output", func() {
				It("is configured to stream to stdout", func() {
					server, err := parallel_support.NewServer(3, reporter)
					Ω(err).ShouldNot(HaveOccurred())
					Ω(server.GetOutputDestination().(*os.File).Fd()).Should(Equal(uintptr(1)))
				})

				It("streams output to the provided buffer", func() {
					n, err := client.Write([]byte("hello"))
					Ω(n).Should(Equal(5))
					Ω(err).ShouldNot(HaveOccurred())
					Ω(buffer).Should(gbytes.Say("hello"))
				})
			})

			Describe("progress reports", func() {
				It("can emit progress reports", func() {
					pr := types.ProgressReport{LeafNodeText: "hola"}
					Ω(client.PostEmitProgressReport(pr)).Should(Succeed())
					Ω(reporter.ProgressReports).Should(ConsistOf(pr))
				})
			})

			Describe("Synchronization endpoints", func() {
				var proc1Exited, proc2Exited, proc3Exited chan any
				BeforeEach(func() {
					proc1Exited, proc2Exited, proc3Exited = make(chan any), make(chan any), make(chan any)
					aliveFunc := func(c chan any) func() bool {
						return func() bool {
							select {
							case <-c:
								return false
							default:
								return true
							}
						}
					}
					server.RegisterAlive(1, aliveFunc(proc1Exited))
					server.RegisterAlive(2, aliveFunc(proc2Exited))
					server.RegisterAlive(3, aliveFunc(proc3Exited))
				})

				Describe("Managing ReportBeforeSuite synchronization", func() {
					Context("when proc 1 succeeds", func() {
						It("passes that success along to other procs", func() {
							Ω(client.PostReportBeforeSuiteCompleted(types.SpecStatePassed)).Should(Succeed())
							state, err := client.BlockUntilReportBeforeSuiteCompleted()
							Ω(state).Should(Equal(types.SpecStatePassed))
							Ω(err).ShouldNot(HaveOccurred())
						})
					})

					Context("when proc 1 fails", func() {
						It("passes that state information along to the other procs", func() {
							Ω(client.PostReportBeforeSuiteCompleted(types.SpecStateFailed)).Should(Succeed())
							state, err := client.BlockUntilReportBeforeSuiteCompleted()
							Ω(state).Should(Equal(types.SpecStateFailed))
							Ω(err).ShouldNot(HaveOccurred())
						})
					})

					Context("when proc 1 disappears before reporting back", func() {
						It("returns a meaningful error", func() {
							close(proc1Exited)
							state, err := client.BlockUntilReportBeforeSuiteCompleted()
							Ω(state).Should(Equal(types.SpecStateFailed))
							Ω(err).ShouldNot(HaveOccurred())
						})
					})

					Context("when proc 1 hasn't responded yet", func() {
						It("blocks until it does", func() {
							done := make(chan any)
							go func() {
								defer GinkgoRecover()
								state, err := client.BlockUntilReportBeforeSuiteCompleted()
								Ω(state).Should(Equal(types.SpecStatePassed))
								Ω(err).ShouldNot(HaveOccurred())
								close(done)
							}()
							Consistently(done).ShouldNot(BeClosed())
							Ω(client.PostReportBeforeSuiteCompleted(types.SpecStatePassed)).Should(Succeed())
							Eventually(done).Should(BeClosed())
						})
					})
				})

				Describe("Managing SynchronizedBeforeSuite synchronization", func() {
					Context("when proc 1 succeeds and returns data", func() {
						It("passes that data along to other procs", func() {
							Ω(client.PostSynchronizedBeforeSuiteCompleted(types.SpecStatePassed, []byte("hello there"))).Should(Succeed())
							state, data, err := client.BlockUntilSynchronizedBeforeSuiteData()
							Ω(state).Should(Equal(types.SpecStatePassed))
							Ω(data).Should(Equal([]byte("hello there")))
							Ω(err).ShouldNot(HaveOccurred())
						})
					})

					Context("when proc 1 succeeds and the data happens to be nil", func() {
						It("passes reports success and returns nil", func() {
							Ω(client.PostSynchronizedBeforeSuiteCompleted(types.SpecStatePassed, nil)).Should(Succeed())
							state, data, err := client.BlockUntilSynchronizedBeforeSuiteData()
							Ω(state).Should(Equal(types.SpecStatePassed))
							Ω(data).Should(BeNil())
							Ω(err).ShouldNot(HaveOccurred())
						})
					})

					Context("when proc 1 is skipped", func() {
						It("passes that state information along to the other procs", func() {
							Ω(client.PostSynchronizedBeforeSuiteCompleted(types.SpecStateSkipped, nil)).Should(Succeed())
							state, data, err := client.BlockUntilSynchronizedBeforeSuiteData()
							Ω(state).Should(Equal(types.SpecStateSkipped))
							Ω(data).Should(BeNil())
							Ω(err).ShouldNot(HaveOccurred())
						})
					})

					Context("when proc 1 fails", func() {
						It("passes that state information along to the other procs", func() {
							Ω(client.PostSynchronizedBeforeSuiteCompleted(types.SpecStateFailed, nil)).Should(Succeed())
							state, data, err := client.BlockUntilSynchronizedBeforeSuiteData()
							Ω(state).Should(Equal(types.SpecStateFailed))
							Ω(data).Should(BeNil())
							Ω(err).ShouldNot(HaveOccurred())
						})
					})

					Context("when proc 1 disappears before reporting back", func() {
						It("returns a meaningful error", func() {
							close(proc1Exited)
							state, data, err := client.BlockUntilSynchronizedBeforeSuiteData()
							Ω(state).Should(Equal(types.SpecStateInvalid))
							Ω(data).Should(BeNil())
							Ω(err).Should(MatchError(types.GinkgoErrors.SynchronizedBeforeSuiteDisappearedOnProc1()))
						})
					})

					Context("when proc 1 hasn't responded yet", func() {
						It("blocks until it does", func() {
							done := make(chan any)
							go func() {
								defer GinkgoRecover()
								state, data, err := client.BlockUntilSynchronizedBeforeSuiteData()
								Ω(state).Should(Equal(types.SpecStatePassed))
								Ω(data).Should(Equal([]byte("hello there")))
								Ω(err).ShouldNot(HaveOccurred())
								close(done)
							}()
							Consistently(done).ShouldNot(BeClosed())
							Ω(client.PostSynchronizedBeforeSuiteCompleted(types.SpecStatePassed, []byte("hello there"))).Should(Succeed())
							Eventually(done).Should(BeClosed())
						})
					})
				})

				Describe("BlockUntilNonprimaryProcsHaveFinished", func() {
					It("blocks until non-primary procs exit", func() {
						done := make(chan any)
						go func() {
							defer GinkgoRecover()
							Ω(client.BlockUntilNonprimaryProcsHaveFinished()).Should(Succeed())
							close(done)
						}()
						Consistently(done).ShouldNot(BeClosed())
						close(proc2Exited)
						Consistently(done).ShouldNot(BeClosed())
						close(proc3Exited)
						Eventually(done).Should(BeClosed())
					})
				})

				Describe("BlockUntilAggregatedNonprimaryProcsReport", func() {
					var specReportA, specReportB types.SpecReport
					var endReport2, endReport3 types.Report

					BeforeEach(func() {
						specReportA = types.SpecReport{LeafNodeText: "A"}
						specReportB = types.SpecReport{LeafNodeText: "B"}
						endReport2 = types.Report{SpecReports: types.SpecReports{specReportA}}
						endReport3 = types.Report{SpecReports: types.SpecReports{specReportB}}
					})

					It("blocks until all non-primary procs exit, then returns the aggregated report", func() {
						done := make(chan any)
						go func() {
							defer GinkgoRecover()
							report, err := client.BlockUntilAggregatedNonprimaryProcsReport()
							Ω(err).ShouldNot(HaveOccurred())
							Ω(report.SpecReports).Should(ConsistOf(specReportA, specReportB))
							close(done)
						}()
						Consistently(done).ShouldNot(BeClosed())

						Ω(client.PostSuiteDidEnd(endReport2)).Should(Succeed())
						close(proc2Exited)
						Consistently(done).ShouldNot(BeClosed())

						Ω(client.PostSuiteDidEnd(endReport3)).Should(Succeed())
						close(proc3Exited)
						Eventually(done).Should(BeClosed())
					})

					Context("when a non-primary proc disappears without reporting back", func() {
						It("blocks returns an appropriate error", func() {
							done := make(chan any)
							go func() {
								defer GinkgoRecover()
								report, err := client.BlockUntilAggregatedNonprimaryProcsReport()
								Ω(err).Should(Equal(types.GinkgoErrors.AggregatedReportUnavailableDueToNodeDisappearing()))
								Ω(report).Should(BeZero())
								close(done)
							}()
							Consistently(done).ShouldNot(BeClosed())

							Ω(client.PostSuiteDidEnd(endReport2)).Should(Succeed())
							close(proc2Exited)
							Consistently(done).ShouldNot(BeClosed())

							close(proc3Exited)
							Eventually(done).Should(BeClosed())
						})
					})
				})

				Describe("Fetching counters", func() {
					It("returns ascending counters", func() {
						Ω(client.FetchNextCounter()).Should(Equal(0))
						Ω(client.FetchNextCounter()).Should(Equal(1))
						Ω(client.FetchNextCounter()).Should(Equal(2))
						Ω(client.FetchNextCounter()).Should(Equal(3))
					})
				})

				Describe("Aborting", func() {
					It("should not abort by default", func() {
						Ω(client.ShouldAbort()).Should(BeFalse())
					})

					Context("when told to abort", func() {
						BeforeEach(func() {
							Ω(client.PostAbort()).Should(Succeed())
						})

						It("should abort", func() {
							Ω(client.ShouldAbort()).Should(BeTrue())
						})
					})
				})

			})
		})
	}
})
