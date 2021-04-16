package parallel_support_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/ginkgo/internal/parallel_support"
	. "github.com/onsi/ginkgo/internal/test_helpers"
	"github.com/onsi/ginkgo/types"
)

var _ = Describe("The Parallel Support Client & Server", func() {
	var (
		server   *parallel_support.Server
		client   parallel_support.Client
		reporter *FakeReporter
	)

	BeforeEach(func() {
		var err error
		reporter = &FakeReporter{}
		server, err = parallel_support.NewServer(3, reporter)
		Ω(err).ShouldNot(HaveOccurred())
		server.Start()

		client = parallel_support.NewClient(server.Address())
		Eventually(client.CheckServerUp).Should(BeTrue())
	})

	AfterEach(func() {
		server.Close()
	})

	It("should make its address available", func() {
		Ω(server.Address()).Should(MatchRegexp(`http://127.0.0.1:\d{2,}`))
	})

	Describe("Reporting endpoints", func() {
		var beginReport, thirdBeginReport types.Report
		var endReport1, endReport2, endReport3 types.Report
		var specReportA, specReportB, specReportC types.SpecReport

		var t time.Time

		BeforeEach(func() {
			beginReport = types.Report{SuiteDescription: "my sweet suite"}
			thirdBeginReport = types.Report{SuiteDescription: "last one in gets forwarded"}

			specReportA = types.SpecReport{NodeTexts: []string{"A"}}
			specReportB = types.SpecReport{NodeTexts: []string{"B"}}
			specReportC = types.SpecReport{NodeTexts: []string{"C"}}

			t = time.Now()

			endReport1 = types.Report{StartTime: t.Add(-time.Second), EndTime: t.Add(time.Second), SuiteSucceeded: true, SpecReports: types.SpecReports{specReportA}}
			endReport2 = types.Report{StartTime: t.Add(-2 * time.Second), EndTime: t.Add(time.Second), SuiteSucceeded: true, SpecReports: types.SpecReports{specReportB}}
			endReport3 = types.Report{StartTime: t.Add(-time.Second), EndTime: t.Add(2 * time.Second), SuiteSucceeded: false, SpecReports: types.SpecReports{specReportC}}
		})

		Context("before all nodes have reported SuiteWillBegin", func() {
			BeforeEach(func() {
				client.PostSuiteWillBegin(beginReport)
				client.PostDidRun(specReportA)
				client.PostSuiteWillBegin(beginReport)
				client.PostDidRun(specReportB)
			})

			It("should not forward anything to the attached reporter", func() {
				Ω(reporter.Begin).Should(BeZero())
				Ω(reporter.Will).Should(BeEmpty())
				Ω(reporter.Did).Should(BeEmpty())
			})

			Context("when the final node reports SuiteWillBegin", func() {
				BeforeEach(func() {
					client.PostSuiteWillBegin(thirdBeginReport)
				})

				It("forwards to SuiteWillBegin and catches up on any received summaries", func() {
					Ω(reporter.Begin).Should(Equal(thirdBeginReport))
					Ω(reporter.Will.Names()).Should(ConsistOf("A", "B"))
					Ω(reporter.Did.Names()).Should(ConsistOf("A", "B"))
				})

				Context("any subsequent summaries", func() {
					BeforeEach(func() {
						client.PostDidRun(specReportC)
					})

					It("are forwarded immediately", func() {
						Ω(reporter.Will.Names()).Should(ConsistOf("A", "B", "C"))
						Ω(reporter.Did.Names()).Should(ConsistOf("A", "B", "C"))
					})
				})

				Context("when SuiteDidEnd start arriving", func() {
					BeforeEach(func() {
						client.PostSuiteDidEnd(endReport1)
						client.PostSuiteDidEnd(endReport2)
					})

					It("does not forward them yet...", func() {
						Ω(reporter.End).Should(BeZero())
					})

					It("doesn't signal it's done", func() {
						Ω(server.Done).ShouldNot(BeClosed())
					})

					Context("when the final SuiteDidEnd arrive", func() {
						BeforeEach(func() {
							client.PostSuiteDidEnd(endReport3)
						})

						It("forwards the aggregation of all received end summaries", func() {
							Ω(reporter.End.StartTime.Unix()).Should(BeNumerically("~", t.Add(-2*time.Second).Unix()))
							Ω(reporter.End.EndTime.Unix()).Should(BeNumerically("~", t.Add(2*time.Second).Unix()))
							Ω(reporter.End.RunTime).Should(BeNumerically("~", 4*time.Second))
							Ω(reporter.End.SuiteSucceeded).Should(BeFalse())
							Ω(reporter.End.SpecReports).Should(ConsistOf(specReportA, specReportB, specReportC))
						})

						It("should signal it's done", func() {
							Ω(server.Done).Should(BeClosed())
						})
					})
				})
			})
		})
	})

	Describe("Synchronization endpoints", func() {
		var node1Exited, node2Exited, node3Exited chan interface{}
		BeforeEach(func() {
			node1Exited, node2Exited, node3Exited = make(chan interface{}), make(chan interface{}), make(chan interface{})
			aliveFunc := func(c chan interface{}) func() bool {
				return func() bool {
					select {
					case <-c:
						return false
					default:
						return true
					}
				}
			}
			server.RegisterAlive(1, aliveFunc(node1Exited))
			server.RegisterAlive(2, aliveFunc(node2Exited))
			server.RegisterAlive(3, aliveFunc(node3Exited))
		})

		Describe("Managing SynchronizedBeforeSuite synchronization", func() {
			Context("when node 1 succeeds and returns data", func() {
				It("passes that data along to other nodes", func() {
					Ω(client.PostSynchronizedBeforeSuiteSucceeded([]byte("hello there"))).Should(Succeed())
					Ω(client.BlockUntilSynchronizedBeforeSuiteData()).Should(Equal([]byte("hello there")))
				})
			})

			Context("when node 1 succeeds and the data happens to be nil", func() {
				It("passes reports success and returns nil", func() {
					Ω(client.PostSynchronizedBeforeSuiteSucceeded(nil)).Should(Succeed())
					Ω(client.BlockUntilSynchronizedBeforeSuiteData()).Should(BeNil())
				})
			})

			Context("when node 1 fails", func() {
				It("returns a meaningful error", func() {
					Ω(client.PostSynchronizedBeforeSuiteFailed()).Should(Succeed())
					data, err := client.BlockUntilSynchronizedBeforeSuiteData()
					Ω(data).Should(BeNil())
					Ω(err).Should(MatchError(types.GinkgoErrors.SynchronizedBeforeSuiteFailedOnNode1()))
				})
			})

			Context("when node 1 disappears before reporting back", func() {
				It("returns a meaningful error", func() {
					close(node1Exited)
					data, err := client.BlockUntilSynchronizedBeforeSuiteData()
					Ω(data).Should(BeNil())
					Ω(err).Should(MatchError(types.GinkgoErrors.SynchronizedBeforeSuiteDisappearedOnNode1()))
				})
			})

			Context("when node 1 hasn't responded yet", func() {
				It("blocks until it does", func() {
					done := make(chan interface{})
					go func() {
						defer GinkgoRecover()
						Ω(client.BlockUntilSynchronizedBeforeSuiteData()).Should(Equal([]byte("hello there")))
						close(done)
					}()
					Consistently(done).ShouldNot(BeClosed())
					Ω(client.PostSynchronizedBeforeSuiteSucceeded([]byte("hello there"))).Should(Succeed())
					Eventually(done).Should(BeClosed())
				})
			})
		})

		Describe("BlockUntilNonprimaryNodesHaveFinished", func() {
			It("blocks until non-primary nodes exit", func() {
				done := make(chan interface{})
				go func() {
					defer GinkgoRecover()
					Ω(client.BlockUntilNonprimaryNodesHaveFinished()).Should(Succeed())
					close(done)
				}()
				Consistently(done).ShouldNot(BeClosed())
				close(node2Exited)
				Consistently(done).ShouldNot(BeClosed())
				close(node3Exited)
				Eventually(done).Should(BeClosed())
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

	})
})
