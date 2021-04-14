package parallel_support_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/internal/parallel_support"
	. "github.com/onsi/gomega"

	"bytes"
	"encoding/json"
	"net/http"

	. "github.com/onsi/ginkgo/internal/test_helpers"
	"github.com/onsi/ginkgo/types"
)

var _ = Describe("Server", func() {
	var (
		server             *Server
		reporter           *FakeReporter
		forwardingReporter *ForwardingReporter
	)

	BeforeEach(func() {
		var err error
		reporter = &FakeReporter{}
		server, err = NewServer(3, reporter)
		Ω(err).ShouldNot(HaveOccurred())

		server.Start()
		Eventually(StatusCodePoller(server.Address() + "/up")).Should(Equal(http.StatusOK))

		forwardingReporter = NewForwardingReporter(types.ReporterConfig{}, server.Address(), nil)
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("Streaming endpoints", func() {
		var beginSummary, thirdBeginSummary types.SuiteSummary
		var endSummary1, endSummary2, endSummary3 types.SuiteSummary
		var reportA, reportB, reportC types.SpecReport

		BeforeEach(func() {
			beginSummary = types.SuiteSummary{SuiteDescription: "my sweet suite"}
			thirdBeginSummary = types.SuiteSummary{SuiteDescription: "laste one in gets forwarded"}

			reportA = types.SpecReport{NodeTexts: []string{"A"}}
			reportB = types.SpecReport{NodeTexts: []string{"B"}}
			reportC = types.SpecReport{NodeTexts: []string{"C"}}

			endSummary1 = types.SuiteSummary{NumberOfPassedSpecs: 2, RunTime: time.Second, SuiteSucceeded: true}
			endSummary2 = types.SuiteSummary{NumberOfPassedSpecs: 3, RunTime: time.Minute, NumberOfSkippedSpecs: 2, SuiteSucceeded: true}
			endSummary3 = types.SuiteSummary{NumberOfPassedSpecs: 1, RunTime: time.Second, NumberOfFailedSpecs: 3, SuiteSucceeded: false}
		})

		It("should make its address available", func() {
			Ω(server.Address()).Should(MatchRegexp(`http://127.0.0.1:\d{2,}`))
		})

		Context("before all nodes have reported SpecSuiteWillBegin", func() {
			BeforeEach(func() {
				forwardingReporter.SpecSuiteWillBegin(types.SuiteConfig{RandomSeed: 17}, beginSummary)
				forwardingReporter.DidRun(reportA)
				forwardingReporter.SpecSuiteWillBegin(types.SuiteConfig{RandomSeed: 17}, beginSummary)
				forwardingReporter.DidRun(reportB)
			})

			It("should not forward anything to the attached reporter", func() {
				Ω(reporter.Config).Should(BeZero())
				Ω(reporter.Begin).Should(BeZero())
				Ω(reporter.Will).Should(BeEmpty())
				Ω(reporter.Did).Should(BeEmpty())
			})

			Context("when the final node reports SpecSuiteWillBegin", func() {
				BeforeEach(func() {
					forwardingReporter.SpecSuiteWillBegin(types.SuiteConfig{RandomSeed: 3}, thirdBeginSummary)
				})

				It("forwards to SpecSuiteWillBegin and catches up on any received summareis", func() {
					Ω(reporter.Config).Should(Equal(types.SuiteConfig{RandomSeed: 3}))
					Ω(reporter.Begin).Should(Equal(thirdBeginSummary))
					Ω(reporter.Will.Names()).Should(ConsistOf("A", "B"))
					Ω(reporter.Did.Names()).Should(ConsistOf("A", "B"))
				})

				Context("any subsequent summaries", func() {
					BeforeEach(func() {
						forwardingReporter.DidRun(reportC)
					})

					It("are forwarded immediately", func() {
						Ω(reporter.Will.Names()).Should(ConsistOf("A", "B", "C"))
						Ω(reporter.Did.Names()).Should(ConsistOf("A", "B", "C"))
					})
				})

				Context("when SpecSuiteDidEnd start arriving", func() {
					BeforeEach(func() {
						forwardingReporter.SpecSuiteDidEnd(endSummary1)
						forwardingReporter.SpecSuiteDidEnd(endSummary2)
					})

					It("does not forward them yet...", func() {
						Ω(reporter.End).Should(BeZero())
					})

					It("doesn't signal it's done", func() {
						Ω(server.Done).ShouldNot(BeClosed())
					})

					Context("when the final SpecSuiteDidEnd arrive", func() {
						BeforeEach(func() {
							forwardingReporter.SpecSuiteDidEnd(endSummary3)
						})

						It("forwards the aggregation of all received end summaries", func() {
							Ω(reporter.End).Should(Equal(types.SuiteSummary{
								SuiteSucceeded:       false,
								NumberOfPassedSpecs:  6,
								NumberOfSkippedSpecs: 2,
								NumberOfFailedSpecs:  3,
								RunTime:              time.Minute,
							}))
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
		Describe("GETting and POSTing BeforeSuiteState", func() {
			getBeforeSuite := func() types.RemoteBeforeSuiteData {
				resp, err := http.Get(server.Address() + "/BeforeSuiteState")
				Ω(err).ShouldNot(HaveOccurred())
				Ω(resp.StatusCode).Should(Equal(http.StatusOK))

				r := types.RemoteBeforeSuiteData{}
				decoder := json.NewDecoder(resp.Body)
				err = decoder.Decode(&r)
				Ω(err).ShouldNot(HaveOccurred())

				return r
			}

			postBeforeSuite := func(r types.RemoteBeforeSuiteData) {
				resp, err := http.Post(server.Address()+"/BeforeSuiteState", "application/json", bytes.NewReader(r.ToJSON()))
				Ω(err).ShouldNot(HaveOccurred())
				Ω(resp.StatusCode).Should(Equal(http.StatusOK))
			}

			Context("when the first node's Alive has not been registered yet", func() {
				It("should return pending", func() {
					state := getBeforeSuite()
					Ω(state).Should(Equal(types.RemoteBeforeSuiteData{Data: nil, State: types.RemoteBeforeSuiteStatePending}))

					state = getBeforeSuite()
					Ω(state).Should(Equal(types.RemoteBeforeSuiteData{Data: nil, State: types.RemoteBeforeSuiteStatePending}))
				})
			})

			Context("when the first node is Alive but has not responded yet", func() {
				BeforeEach(func() {
					server.RegisterAlive(1, func() bool {
						return true
					})
				})

				It("should return pending", func() {
					state := getBeforeSuite()
					Ω(state).Should(Equal(types.RemoteBeforeSuiteData{Data: nil, State: types.RemoteBeforeSuiteStatePending}))

					state = getBeforeSuite()
					Ω(state).Should(Equal(types.RemoteBeforeSuiteData{Data: nil, State: types.RemoteBeforeSuiteStatePending}))
				})
			})

			Context("when the first node has responded", func() {
				var state types.RemoteBeforeSuiteData
				BeforeEach(func() {
					server.RegisterAlive(1, func() bool {
						return false
					})

					state = types.RemoteBeforeSuiteData{
						Data:  []byte("my data"),
						State: types.RemoteBeforeSuiteStatePassed,
					}
					postBeforeSuite(state)
				})

				It("should return the passed in state", func() {
					returnedState := getBeforeSuite()
					Ω(returnedState).Should(Equal(state))
				})
			})

			Context("when the first node is no longer Alive and has not responded yet", func() {
				BeforeEach(func() {
					server.RegisterAlive(1, func() bool {
						return false
					})
				})

				It("should return disappeared", func() {
					state := getBeforeSuite()
					Ω(state).Should(Equal(types.RemoteBeforeSuiteData{Data: nil, State: types.RemoteBeforeSuiteStateDisappeared}))

					state = getBeforeSuite()
					Ω(state).Should(Equal(types.RemoteBeforeSuiteData{Data: nil, State: types.RemoteBeforeSuiteStateDisappeared}))
				})
			})
		})

		Describe("GETting RemoteAfterSuiteData", func() {
			getRemoteAfterSuiteData := func() bool {
				resp, err := http.Get(server.Address() + "/AfterSuiteState")
				Ω(err).ShouldNot(HaveOccurred())
				Ω(resp.StatusCode).Should(Equal(http.StatusOK))

				a := types.RemoteAfterSuiteData{}
				decoder := json.NewDecoder(resp.Body)
				err = decoder.Decode(&a)
				Ω(err).ShouldNot(HaveOccurred())

				return a.CanRun
			}

			Context("when there are unregistered nodes", func() {
				BeforeEach(func() {
					server.RegisterAlive(2, func() bool {
						return false
					})
				})

				It("should return false", func() {
					Ω(getRemoteAfterSuiteData()).Should(BeFalse())
				})
			})

			Context("when all none-node-1 nodes are still running", func() {
				BeforeEach(func() {
					server.RegisterAlive(2, func() bool {
						return true
					})

					server.RegisterAlive(3, func() bool {
						return false
					})
				})

				It("should return false", func() {
					Ω(getRemoteAfterSuiteData()).Should(BeFalse())
				})
			})

			Context("when all none-1 nodes are done", func() {
				BeforeEach(func() {
					server.RegisterAlive(2, func() bool {
						return false
					})

					server.RegisterAlive(3, func() bool {
						return false
					})
				})

				It("should return true", func() {
					Ω(getRemoteAfterSuiteData()).Should(BeTrue())
				})

			})
		})
	})
})
