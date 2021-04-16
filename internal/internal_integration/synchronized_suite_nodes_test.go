package internal_integration_test

import (
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/internal/parallel_support"
	. "github.com/onsi/ginkgo/internal/test_helpers"
	"github.com/onsi/ginkgo/types"
	. "github.com/onsi/gomega"
)

var _ = Describe("Synchronized Suite Nodes", func() {
	var failInBeforeSuiteNode1, failInBeforeSuiteAllNodes, failInAfterSuiteAllNodes, failInAfterSuiteNode1 bool
	var fixture func()

	BeforeEach(func() {
		failInBeforeSuiteNode1, failInBeforeSuiteAllNodes, failInAfterSuiteAllNodes, failInAfterSuiteNode1 = false, false, false, false
		fixture = func() {
			SynchronizedBeforeSuite(func() []byte {
				rt.Run("before-suite-node-1")
				if failInBeforeSuiteNode1 {
					F("fail-in-before-suite-node-1", cl)
				}
				return []byte("hey there")
			}, func(data []byte) {
				rt.RunWithData("before-suite-all-nodes", "data", string(data))
				if failInBeforeSuiteAllNodes {
					F("fail-in-before-suite-all-nodes", cl)
				}
			})
			It("test", rt.T("test"))
			SynchronizedAfterSuite(func() {
				rt.Run("after-suite-all-nodes")
				if failInAfterSuiteAllNodes {
					F("fail-in-after-suite-all-nodes", cl)
				}
			}, func() {
				rt.Run("after-suite-node-1")
				if failInAfterSuiteNode1 {
					F("fail-in-after-suite-node-1", cl)
				}
			})
		}
	})

	Describe("when running in series", func() {
		BeforeEach(func() {
			conf.ParallelTotal = 1
			conf.ParallelNode = 1
		})

		Describe("happy path", func() {
			BeforeEach(func() {
				success, _ := RunFixture("happy-path", fixture)
				Ω(success).Should(BeTrue())
			})

			It("runs all the functions", func() {
				Ω(rt).Should(HaveTracked(
					"before-suite-node-1", "before-suite-all-nodes",
					"test",
					"after-suite-all-nodes", "after-suite-node-1",
				))
			})

			It("reports on the SynchronizedBeforeSuite and SynchronizedAfterSuite as having passed", func() {
				befReports := reporter.Did.WithLeafNodeType(types.NodeTypeSynchronizedBeforeSuite)
				Ω(befReports).Should(HaveLen(1))
				Ω(befReports[0]).Should(HavePassed())

				aftReports := reporter.Did.WithLeafNodeType(types.NodeTypeSynchronizedAfterSuite)
				Ω(aftReports).Should(HaveLen(1))
				Ω(aftReports[0]).Should(HavePassed())
			})

			It("passes data between the two SynchronizedBeforeSuite functions", func() {
				Ω(rt).Should(HaveRunWithData("before-suite-all-nodes", "data", "hey there"))
			})
		})

		Describe("when the SynchronizedBeforeSuite node1 function fails", func() {
			BeforeEach(func() {
				failInBeforeSuiteNode1 = true
				success, _ := RunFixture("fail in SynchronizedBeforeSuite node1", fixture)
				Ω(success).Should(BeFalse())
			})

			It("doens't run the allNodes function or any of the tests", func() {
				Ω(rt).Should(HaveTracked(
					"before-suite-node-1",
					"after-suite-all-nodes", "after-suite-node-1",
				))
			})

			It("reports on the SynchronizedBeforeSuite and SynchronizedAfterSuite correctly", func() {
				Ω(reporter.Did.FindByLeafNodeType(types.NodeTypeSynchronizedBeforeSuite)).Should(HaveFailed("fail-in-before-suite-node-1"))
				Ω(reporter.Did.FindByLeafNodeType(types.NodeTypeSynchronizedAfterSuite)).Should(HavePassed())
			})
		})

		Describe("when the SynchronizedBeforeSuite allNodes function fails", func() {
			BeforeEach(func() {
				failInBeforeSuiteAllNodes = true
				success, _ := RunFixture("fail in SynchronizedBeforeSuite allNodes", fixture)
				Ω(success).Should(BeFalse())
			})

			It("doesn't run the tests", func() {
				Ω(rt).Should(HaveTracked(
					"before-suite-node-1", "before-suite-all-nodes",
					"after-suite-all-nodes", "after-suite-node-1",
				))
				Ω(rt).Should(HaveRunWithData("before-suite-all-nodes", "data", "hey there"))
			})

			It("reports on the SynchronizedBeforeSuite and SynchronizedAfterSuite correctly", func() {
				Ω(reporter.Did.FindByLeafNodeType(types.NodeTypeSynchronizedBeforeSuite)).Should(HaveFailed("fail-in-before-suite-all-nodes"))
				Ω(reporter.Did.FindByLeafNodeType(types.NodeTypeSynchronizedAfterSuite)).Should(HavePassed())
			})
		})

		Describe("when the SynchronizedAfterSuite allNodes function fails", func() {
			BeforeEach(func() {
				failInAfterSuiteAllNodes = true
				success, _ := RunFixture("fail in SynchronizedAfterSuite allNodes", fixture)
				Ω(success).Should(BeFalse())
			})

			It("nonetheless runs the node-1 function", func() {
				Ω(rt).Should(HaveTracked(
					"before-suite-node-1", "before-suite-all-nodes",
					"test",
					"after-suite-all-nodes", "after-suite-node-1",
				))
				Ω(rt).Should(HaveRunWithData("before-suite-all-nodes", "data", "hey there"))
			})

			It("reports on the SynchronizedBeforeSuite and SynchronizedAfterSuite correctly", func() {
				Ω(reporter.Did.FindByLeafNodeType(types.NodeTypeSynchronizedBeforeSuite)).Should(HavePassed())
				Ω(reporter.Did.FindByLeafNodeType(types.NodeTypeSynchronizedAfterSuite)).Should(HaveFailed("fail-in-after-suite-all-nodes"))
			})
		})

		Describe("when the SynchronizedAfterSuite node1 function fails", func() {
			BeforeEach(func() {
				failInAfterSuiteNode1 = true
				success, _ := RunFixture("fail in SynchronizedAfterSuite node1", fixture)
				Ω(success).Should(BeFalse())
			})

			It("will have run everything", func() {
				Ω(rt).Should(HaveTracked(
					"before-suite-node-1", "before-suite-all-nodes",
					"test",
					"after-suite-all-nodes", "after-suite-node-1",
				))
				Ω(rt).Should(HaveRunWithData("before-suite-all-nodes", "data", "hey there"))
			})

			It("reports on the SynchronizedBeforeSuite and SynchronizedAfterSuite correctly", func() {
				Ω(reporter.Did.FindByLeafNodeType(types.NodeTypeSynchronizedBeforeSuite)).Should(HavePassed())
				Ω(reporter.Did.FindByLeafNodeType(types.NodeTypeSynchronizedAfterSuite)).Should(HaveFailed("fail-in-after-suite-node-1"))
			})
		})
	})

	Describe("when running in parallel", func() {
		var server *parallel_support.Server
		var client parallel_support.Client
		var exitChannels map[int]chan interface{}

		BeforeEach(func() {
			conf.ParallelTotal = 2
			server, client, exitChannels = SetUpServerAndClient(conf.ParallelTotal)
			conf.ParallelHost = server.Address()
		})

		AfterEach(func() {
			server.Close()
		})

		Describe("when running as node 1", func() {
			BeforeEach(func() {
				conf.ParallelNode = 1
			})

			Describe("happy path", func() {
				BeforeEach(func() {
					close(exitChannels[2]) //trigger node 2 exiting so the node1 after suite runs
					success, _ := RunFixture("happy-path", fixture)
					Ω(success).Should(BeTrue())
				})

				It("runs all the functions", func() {
					Ω(rt).Should(HaveTracked(
						"before-suite-node-1", "before-suite-all-nodes",
						"test",
						"after-suite-all-nodes", "after-suite-node-1",
					))
				})

				It("reports on the SynchronizedBeforeSuite and SynchronizedAfterSuite as having passed", func() {
					Ω(reporter.Did.FindByLeafNodeType(types.NodeTypeSynchronizedBeforeSuite)).Should(HavePassed())
					Ω(reporter.Did.FindByLeafNodeType(types.NodeTypeSynchronizedAfterSuite)).Should(HavePassed())
				})

				It("passes data between the two SynchronizedBeforeSuite functions and up to the server", func() {
					Ω(rt).Should(HaveRunWithData("before-suite-all-nodes", "data", "hey there"))
					Ω(client.BlockUntilSynchronizedBeforeSuiteData()).Should(Equal([]byte("hey there")))
				})
			})

			Describe("when the BeforeSuite node1 function fails", func() {
				BeforeEach(func() {
					close(exitChannels[2]) //trigger node 2 exiting so the node1 after suite runs
					failInBeforeSuiteNode1 = true
					success, _ := RunFixture("happy-path", fixture)
					Ω(success).Should(BeFalse())
				})

				It("tells the server", func() {
					data, err := client.BlockUntilSynchronizedBeforeSuiteData()
					Ω(data).Should(BeNil())
					Ω(err).Should(MatchError(types.GinkgoErrors.SynchronizedBeforeSuiteFailedOnNode1()))
				})
			})

			Describe("waiting for all nodes to finish before running the AfterSuite node 1 function", func() {
				It("waits for the server to give it the all clear", func() {
					done := make(chan interface{})
					go func() {
						defer GinkgoRecover()
						success, _ := RunFixture("happy-path", fixture)
						Ω(success).Should(BeTrue())
						close(done)
					}()
					Consistently(done).ShouldNot(BeClosed())
					close(exitChannels[2])
					Eventually(done).Should(BeClosed())
				})
			})
		})

		Describe("when running as another node", func() {
			BeforeEach(func() {
				conf.ParallelNode = 2
			})

			Describe("happy path", func() {
				BeforeEach(func() {
					client.PostSynchronizedBeforeSuiteSucceeded([]byte("hola hola"))
					success, _ := RunFixture("happy-path", fixture)
					Ω(success).Should(BeTrue())
				})

				It("runs all the all-nodes functions", func() {
					Ω(rt).Should(HaveTracked(
						"before-suite-all-nodes",
						"test",
						"after-suite-all-nodes",
					))
				})

				It("reports on the SynchronizedBeforeSuite and SynchronizedAfterSuite as having passed", func() {
					Ω(reporter.Did.FindByLeafNodeType(types.NodeTypeSynchronizedBeforeSuite)).Should(HavePassed())
					Ω(reporter.Did.FindByLeafNodeType(types.NodeTypeSynchronizedAfterSuite)).Should(HavePassed())
				})

				It("gets data for the SynchronizedBeforeSuite all nodes function from the server", func() {
					Ω(rt).Should(HaveRunWithData("before-suite-all-nodes", "data", "hola hola"))
				})
			})

			Describe("waiting for the data from node 1", func() {
				It("waits for the server to give it the data", func() {
					done := make(chan interface{})
					go func() {
						defer GinkgoRecover()
						success, _ := RunFixture("happy-path", fixture)
						Ω(success).Should(BeTrue())
						close(done)
					}()
					Consistently(done).ShouldNot(BeClosed())
					client.PostSynchronizedBeforeSuiteSucceeded([]byte("hola hola"))
					Eventually(done).Should(BeClosed())
					Ω(rt).Should(HaveRunWithData("before-suite-all-nodes", "data", "hola hola"))
				})
			})

			Describe("when node 1 fails the SynchronizedBeforeSuite node1 function", func() {
				It("fails and only runs the after suite", func() {
					done := make(chan interface{})
					go func() {
						defer GinkgoRecover()
						success, _ := RunFixture("happy-path", fixture)
						Ω(success).Should(BeFalse())
						close(done)
					}()
					Consistently(done).ShouldNot(BeClosed())
					client.PostSynchronizedBeforeSuiteFailed()
					Eventually(done).Should(BeClosed())

					Ω(rt).Should(HaveTracked("after-suite-all-nodes"))

					Ω(reporter.Did.FindByLeafNodeType(types.NodeTypeSynchronizedBeforeSuite)).Should(HaveFailed(types.GinkgoErrors.SynchronizedBeforeSuiteFailedOnNode1().Error()))
				})
			})

			Describe("when node 1 disappears before the node 1 function returns", func() {
				It("fails and only runs the after suite", func() {
					done := make(chan interface{})
					go func() {
						defer GinkgoRecover()
						success, _ := RunFixture("happy-path", fixture)
						Ω(success).Should(BeFalse())
						close(done)
					}()
					Consistently(done).ShouldNot(BeClosed())
					close(exitChannels[1])
					Eventually(done).Should(BeClosed())

					Ω(rt).Should(HaveTracked("after-suite-all-nodes"))

					Ω(reporter.Did.FindByLeafNodeType(types.NodeTypeSynchronizedBeforeSuite)).Should(HaveFailed(types.GinkgoErrors.SynchronizedBeforeSuiteDisappearedOnNode1().Error()))
				})
			})
		})
	})
})
