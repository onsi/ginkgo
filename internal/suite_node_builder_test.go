package internal_test

import (
	"net/http"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/internal"
	. "github.com/onsi/ginkgo/internal/test_helpers"
	"github.com/onsi/ginkgo/types"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("SuiteNodeBuilder", func() {
	var rt *RunTracker
	var builder internal.SuiteNodeBuilder
	var node Node
	var conf config.GinkgoConfigType
	var failer *internal.Failer

	BeforeEach(func() {
		rt = NewRunTracker()
		conf = config.GinkgoConfigType{}
		conf.ParallelTotal = 1
		conf.ParallelNode = 1
		failer = internal.NewFailer()
	})

	Describe("Building BeforeSuite Nodes", func() {
		BeforeEach(func() {
			builder = internal.SuiteNodeBuilder{
				NodeType:        types.NodeTypeBeforeSuite,
				CodeLocation:    cl,
				BeforeSuiteBody: rt.T("before suite"),
			}

			node = builder.BuildNode(conf, failer)
		})

		It("returns an appropriately configured node", func() {
			Ω(node.ID).ShouldNot(BeZero())
			Ω(node.NodeType).Should(Equal(types.NodeTypeBeforeSuite))
			Ω(node.CodeLocation).Should(Equal(cl))
		})

		It("configures the node with a body that simply calls the registered BeforeSuite body", func() {
			node.Body()
			Ω(rt).Should(HaveTracked("before suite"))
		})
	})

	Describe("Building AfterSuite Nodes", func() {
		BeforeEach(func() {
			builder = internal.SuiteNodeBuilder{
				NodeType:       types.NodeTypeAfterSuite,
				CodeLocation:   cl,
				AfterSuiteBody: rt.T("after suite"),
			}

			node = builder.BuildNode(conf, failer)
		})

		It("returns an appropriately configured node", func() {
			Ω(node.ID).ShouldNot(BeZero())
			Ω(node.NodeType).Should(Equal(types.NodeTypeAfterSuite))
			Ω(node.CodeLocation).Should(Equal(cl))
		})

		It("configures the node with a body that simply calls the registered AfterSuite body", func() {
			node.Body()
			Ω(rt).Should(HaveTracked("after suite"))
		})

	})

	Describe("Building SynchronizedBeforeSuite Nodes", func() {
		var failNode1 bool
		var failLocation types.CodeLocation
		BeforeEach(func() {
			failNode1 = false
			failLocation = CL("fail-location")

			builder = internal.SuiteNodeBuilder{
				NodeType:     types.NodeTypeSynchronizedBeforeSuite,
				CodeLocation: cl,
				SynchronizedBeforeSuiteNode1Body: func() []byte {
					rt.Run("node1body")
					if failNode1 {
						failer.Fail("node 1 failed", failLocation)
						panic("boom") // simulates Ginkgo DSL's Fail behavior
					}
					return []byte("snarfblat")
				},
				SynchronizedBeforeSuiteAllNodesBody: func(data []byte) {
					rt.Run("allnodesbody - " + string(data))
				},
			}
		})

		JustBeforeEach(func() {
			node = builder.BuildNode(conf, failer)
		})

		It("returns an appropriately configured node", func() {
			Ω(node.ID).ShouldNot(BeZero())
			Ω(node.NodeType).Should(Equal(types.NodeTypeSynchronizedBeforeSuite))
			Ω(node.CodeLocation).Should(Equal(cl))
		})

		Context("when running in series", func() {
			It("simply runs both functions, passing data from node1Body to allNodesBody", func() {
				node.Body()
				Ω(rt).Should(HaveTracked("node1body", "allnodesbody - snarfblat"))
				Ω(failer.GetState()).Should(Equal(types.SpecStatePassed))
			})
		})

		Context("when running in parallel", func() {
			var server *ghttp.Server
			BeforeEach(func() {
				conf.ParallelTotal = 2
				server = ghttp.NewServer()
				conf.ParallelHost = server.URL()
			})

			AfterEach(func() {
				server.Close()
			})

			Context("on Node 1", func() {
				var responseCode int
				var expectedRemoteBeforeSuiteData types.RemoteBeforeSuiteData
				BeforeEach(func() {
					conf.ParallelNode = 1
					expectedRemoteBeforeSuiteData = types.RemoteBeforeSuiteData{
						State: types.RemoteBeforeSuiteStatePassed,
						Data:  []byte("snarfblat"),
					}
					responseCode = http.StatusOK
				})

				JustBeforeEach(func() {
					server.AppendHandlers(ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/BeforeSuiteState"),
						ghttp.VerifyJSONRepresenting(expectedRemoteBeforeSuiteData),
						ghttp.RespondWith(responseCode, nil),
					))
				})

				It("runs node1Body, posts to the server, then runs allnodesBody", func() {
					node.Body()
					Ω(server.ReceivedRequests()).Should(HaveLen(1))
					Ω(rt).Should(HaveTracked("node1body", "allnodesbody - snarfblat"))
					Ω(failer.GetState()).Should(Equal(types.SpecStatePassed))
				})

				Context("when node1Body fails (and panics - as that is what Ginkgo's Fail does)", func() {
					BeforeEach(func() {
						failNode1 = true
						expectedRemoteBeforeSuiteData = types.RemoteBeforeSuiteData{
							State: types.RemoteBeforeSuiteStateFailed,
							Data:  nil,
						}
					})

					It("registers a failure and posts such to the server", func() {
						node.Body()
						Ω(server.ReceivedRequests()).Should(HaveLen(1))
						Ω(failer.GetState()).Should(Equal(types.SpecStateFailed))
						Ω(failer.GetFailure().Location).Should(Equal(failLocation))
					})

					It("does not run allNodesBody", func() {
						node.Body()
						Ω(rt).Should(HaveTracked("node1body"))
					})
				})

				Context("when the server post fails", func() {
					BeforeEach(func() {
						responseCode = http.StatusInternalServerError
					})

					It("sends out a failure", func() {
						node.Body()
						Ω(server.ReceivedRequests()).Should(HaveLen(1))
						Ω(failer.GetState()).Should(Equal(types.SpecStateFailed))
						Ω(failer.GetFailure().Message).Should(Equal("SynchronizedBeforeSuite failed to send data to other nodes"))
						Ω(failer.GetFailure().Location).Should(Equal(cl))
					})

					It("does not run allNodesBody", func() {
						node.Body()
						Ω(rt).Should(HaveTracked("node1body"))
					})
				})
			})

			Context("on other nodes", func() {
				var responseCode int
				var responseRemoteBeforeSuiteData types.RemoteBeforeSuiteData

				BeforeEach(func() {
					conf.ParallelNode = 2
					responseCode = http.StatusOK
					responseRemoteBeforeSuiteData = types.RemoteBeforeSuiteData{
						State: types.RemoteBeforeSuiteStatePassed,
						Data:  []byte("dinglehopper"),
					}
				})

				JustBeforeEach(func() {
					server.AppendHandlers(ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/BeforeSuiteState"),
						ghttp.RespondWithJSONEncoded(http.StatusOK, types.RemoteBeforeSuiteData{
							State: types.RemoteBeforeSuiteStatePending,
						}),
					), ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/BeforeSuiteState"),
						ghttp.RespondWithJSONEncoded(responseCode, responseRemoteBeforeSuiteData),
					))
				})

				It("never runs node1Body, but waits until the first node has finished to run allnodesBody passing in the correct data", func() {
					node.Body()
					Ω(server.ReceivedRequests()).Should(HaveLen(2))
					Ω(failer.GetState()).Should(Equal(types.SpecStatePassed))
					Ω(rt).Should(HaveTracked("allnodesbody - dinglehopper"))
				})

				sharedFailureBehavior := func(expectedFailureMessage string) {
					It("registers a failure", func() {
						node.Body()
						Ω(server.ReceivedRequests()).Should(HaveLen(2))
						Ω(failer.GetState()).Should(Equal(types.SpecStateFailed))
						Ω(failer.GetFailure().Message).Should(Equal(expectedFailureMessage))
						Ω(failer.GetFailure().Location).Should(Equal(cl))
					})

					It("does not run the allNodesBody", func() {
						node.Body()
						Ω(rt).Should(HaveTrackedNothing())
					})
				}

				Context("when the sync host returns an invalid status code", func() {
					BeforeEach(func() {
						responseCode = http.StatusInternalServerError
						responseRemoteBeforeSuiteData = types.RemoteBeforeSuiteData{}
					})
					sharedFailureBehavior("SynchronizedBeforeSuite Server Communication Issue:\nunexpected status code 500")
				})

				Context("when the first node fails", func() {
					BeforeEach(func() {
						responseRemoteBeforeSuiteData = types.RemoteBeforeSuiteData{
							State: types.RemoteBeforeSuiteStateFailed,
						}
					})

					sharedFailureBehavior("SynchronizedBeforeSuite on Node 1 failed")
				})

				Context("when the first node disappears", func() {
					BeforeEach(func() {
						responseRemoteBeforeSuiteData = types.RemoteBeforeSuiteData{
							State: types.RemoteBeforeSuiteStateDisappeared,
						}
					})

					sharedFailureBehavior("SynchronizedBeforeSuite on Node 1 disappeared before it could report back")
				})
			})
		})
	})

	Describe("Building SynchronizedAfterSuite Nodes", func() {
		BeforeEach(func() {
			builder = internal.SuiteNodeBuilder{
				NodeType:                           types.NodeTypeSynchronizedAfterSuite,
				CodeLocation:                       cl,
				SynchronizedAfterSuiteAllNodesBody: rt.T("allnodesbody"),
				SynchronizedAfterSuiteNode1Body:    rt.T("node1body"),
			}
		})

		JustBeforeEach(func() {
			node = builder.BuildNode(conf, failer)
		})

		It("returns an appropriately configured node", func() {
			Ω(node.ID).ShouldNot(BeZero())
			Ω(node.NodeType).Should(Equal(types.NodeTypeSynchronizedAfterSuite))
			Ω(node.CodeLocation).Should(Equal(cl))
		})

		Context("when running in series", func() {
			It("simply runs both allNodesBody followed by node1Body", func() {
				node.Body()
				Ω(rt).Should(HaveTracked("allnodesbody", "node1body"))
				Ω(failer.GetState()).Should(Equal(types.SpecStatePassed))
			})
		})

		Context("when running in parallel", func() {
			var server *ghttp.Server
			BeforeEach(func() {
				conf.ParallelTotal = 2
				server = ghttp.NewServer()
				conf.ParallelHost = server.URL()
			})

			AfterEach(func() {
				server.Close()
			})

			Context("on Node 1", func() {
				var responseCode int
				BeforeEach(func() {
					conf.ParallelNode = 1
					responseCode = http.StatusOK
				})

				JustBeforeEach(func() {
					server.AppendHandlers(ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/AfterSuiteState"),
						func(_ http.ResponseWriter, _ *http.Request) {
							rt.Run("made-request")
						},
						ghttp.RespondWithJSONEncoded(http.StatusOK, types.RemoteAfterSuiteData{
							CanRun: false,
						}),
					), ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/AfterSuiteState"),
						ghttp.RespondWithJSONEncoded(responseCode, types.RemoteAfterSuiteData{
							CanRun: true,
						}),
					))
				})

				It("runs allNodesBody, waits for the sync host to give it the all-clear, then runs node1Body", func() {
					node.Body()
					Ω(server.ReceivedRequests()).Should(HaveLen(2))
					Ω(rt).Should(HaveTracked("allnodesbody", "made-request", "node1body"))
					Ω(failer.GetState()).Should(Equal(types.SpecStatePassed))
				})

				Context("when the sync host returns an invalid status code", func() {
					BeforeEach(func() {
						responseCode = http.StatusInternalServerError
					})

					It("registers a failure", func() {
						node.Body()
						Ω(server.ReceivedRequests()).Should(HaveLen(2))
						Ω(failer.GetState()).Should(Equal(types.SpecStateFailed))
						Ω(failer.GetFailure().Message).Should(Equal("SynchronizedAfterSuite Server Communication Issue:\nunexpected status code 500"))
						Ω(failer.GetFailure().Location).Should(Equal(cl))
					})

					It("does still runs node1Body, to ensure everything is cleaned up", func() {
						node.Body()
						Ω(rt).Should(HaveTracked("allnodesbody", "made-request", "node1body"))
					})
				})
			})

			Context("on other nodes", func() {
				BeforeEach(func() {
					conf.ParallelNode = 2
				})

				It("only runs allNodesBody and never runs node1Body", func() {
					node.Body()
					Ω(server.ReceivedRequests()).Should(HaveLen(0))
					Ω(rt).Should(HaveTracked("allnodesbody"))
					Ω(failer.GetState()).Should(Equal(types.SpecStatePassed))
				})
			})
		})
	})
})
