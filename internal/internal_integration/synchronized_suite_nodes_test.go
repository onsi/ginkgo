package internal_integration_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/ginkgo/v2/internal/test_helpers"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("Synchronized Suite Nodes", func() {
	var failInBeforeSuiteProc1, failInBeforeSuiteAllProcs, failInAfterSuiteAllProcs, failInAfterSuiteProc1 bool
	var fixture func()

	BeforeEach(func() {
		failInBeforeSuiteProc1, failInBeforeSuiteAllProcs, failInAfterSuiteAllProcs, failInAfterSuiteProc1 = false, false, false, false
		fixture = func() {
			SynchronizedBeforeSuite(func() []byte {
				outputInterceptor.AppendInterceptedOutput("before-suite-proc-1")
				rt.Run("before-suite-proc-1")
				if failInBeforeSuiteProc1 {
					F("fail-in-before-suite-proc-1", cl)
				}
				return []byte("hey there")
			}, func(data []byte) {
				outputInterceptor.AppendInterceptedOutput("before-suite-all-procs")
				rt.RunWithData("before-suite-all-procs", "data", string(data))
				if failInBeforeSuiteAllProcs {
					F("fail-in-before-suite-all-procs", cl)
				}
			})
			It("test", rt.T("test"))
			SynchronizedAfterSuite(func() {
				outputInterceptor.AppendInterceptedOutput("after-suite-all-procs")
				rt.Run("after-suite-all-procs")
				if failInAfterSuiteAllProcs {
					F("fail-in-after-suite-all-procs", cl)
				}
			}, func() {
				outputInterceptor.AppendInterceptedOutput("after-suite-proc-1")
				rt.Run("after-suite-proc-1")
				if failInAfterSuiteProc1 {
					F("fail-in-after-suite-proc-1", cl)
				}
			})
		}
	})

	Describe("when running in series", func() {
		BeforeEach(func() {
			conf.ParallelTotal = 1
			conf.ParallelProcess = 1
		})

		Describe("happy path", func() {
			BeforeEach(func() {
				success, _ := RunFixture("happy-path", fixture)
				Ω(success).Should(BeTrue())
			})

			It("runs all the functions", func() {
				Ω(rt).Should(HaveTracked(
					"before-suite-proc-1", "before-suite-all-procs",
					"test",
					"after-suite-all-procs", "after-suite-proc-1",
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
				Ω(rt).Should(HaveRunWithData("before-suite-all-procs", "data", "hey there"))
			})
		})

		Describe("when the SynchronizedBeforeSuite proc1 function fails", func() {
			BeforeEach(func() {
				failInBeforeSuiteProc1 = true
				success, _ := RunFixture("fail in SynchronizedBeforeSuite proc1", fixture)
				Ω(success).Should(BeFalse())
			})

			It("doesn't run the allProcs function or any of the tests", func() {
				Ω(rt).Should(HaveTracked(
					"before-suite-proc-1",
					"after-suite-all-procs", "after-suite-proc-1",
				))
			})

			It("reports on the SynchronizedBeforeSuite and SynchronizedAfterSuite correctly", func() {
				Ω(reporter.Did.FindByLeafNodeType(types.NodeTypeSynchronizedBeforeSuite)).Should(HaveFailed("fail-in-before-suite-proc-1"))
				Ω(reporter.Did.FindByLeafNodeType(types.NodeTypeSynchronizedAfterSuite)).Should(HavePassed())
			})
		})

		Describe("when the SynchronizedBeforeSuite allProcs function fails", func() {
			BeforeEach(func() {
				failInBeforeSuiteAllProcs = true
				success, _ := RunFixture("fail in SynchronizedBeforeSuite allProcs", fixture)
				Ω(success).Should(BeFalse())
			})

			It("doesn't run the tests", func() {
				Ω(rt).Should(HaveTracked(
					"before-suite-proc-1", "before-suite-all-procs",
					"after-suite-all-procs", "after-suite-proc-1",
				))
				Ω(rt).Should(HaveRunWithData("before-suite-all-procs", "data", "hey there"))
			})

			It("reports on the SynchronizedBeforeSuite and SynchronizedAfterSuite correctly", func() {
				Ω(reporter.Did.FindByLeafNodeType(types.NodeTypeSynchronizedBeforeSuite)).Should(HaveFailed("fail-in-before-suite-all-procs"))
				Ω(reporter.Did.FindByLeafNodeType(types.NodeTypeSynchronizedAfterSuite)).Should(HavePassed())
			})
		})

		Describe("when the SynchronizedAfterSuite allProcs function fails", func() {
			BeforeEach(func() {
				failInAfterSuiteAllProcs = true
				success, _ := RunFixture("fail in SynchronizedAfterSuite allProcs", fixture)
				Ω(success).Should(BeFalse())
			})

			It("nonetheless runs the proc-1 function", func() {
				Ω(rt).Should(HaveTracked(
					"before-suite-proc-1", "before-suite-all-procs",
					"test",
					"after-suite-all-procs", "after-suite-proc-1",
				))
				Ω(rt).Should(HaveRunWithData("before-suite-all-procs", "data", "hey there"))
			})

			It("reports on the SynchronizedBeforeSuite and SynchronizedAfterSuite correctly", func() {
				Ω(reporter.Did.FindByLeafNodeType(types.NodeTypeSynchronizedBeforeSuite)).Should(HavePassed())
				Ω(reporter.Did.FindByLeafNodeType(types.NodeTypeSynchronizedAfterSuite)).Should(HaveFailed("fail-in-after-suite-all-procs"))
			})
		})

		Describe("when the SynchronizedAfterSuite proc1 function fails", func() {
			BeforeEach(func() {
				failInAfterSuiteProc1 = true
				success, _ := RunFixture("fail in SynchronizedAfterSuite proc1", fixture)
				Ω(success).Should(BeFalse())
			})

			It("will have run everything", func() {
				Ω(rt).Should(HaveTracked(
					"before-suite-proc-1", "before-suite-all-procs",
					"test",
					"after-suite-all-procs", "after-suite-proc-1",
				))
				Ω(rt).Should(HaveRunWithData("before-suite-all-procs", "data", "hey there"))
			})

			It("reports on the SynchronizedBeforeSuite and SynchronizedAfterSuite correctly", func() {
				Ω(reporter.Did.FindByLeafNodeType(types.NodeTypeSynchronizedBeforeSuite)).Should(HavePassed())
				Ω(reporter.Did.FindByLeafNodeType(types.NodeTypeSynchronizedAfterSuite)).Should(HaveFailed("fail-in-after-suite-proc-1"))
			})
		})
	})

	Describe("when running in parallel", func() {
		var serverOutputBuffer *gbytes.Buffer

		BeforeEach(func() {
			SetUpForParallel(2)
			serverOutputBuffer = gbytes.NewBuffer()
			server.SetOutputDestination(serverOutputBuffer)
		})

		Describe("when running as proc 1", func() {
			BeforeEach(func() {
				conf.ParallelProcess = 1
			})

			Describe("happy path", func() {
				BeforeEach(func() {
					close(exitChannels[2]) //trigger proc 2 exiting so the proc1 after suite runs
					success, _ := RunFixture("happy-path", fixture)
					Ω(success).Should(BeTrue())
				})

				It("runs all the functions", func() {
					Ω(rt).Should(HaveTracked(
						"before-suite-proc-1", "before-suite-all-procs",
						"test",
						"after-suite-all-procs", "after-suite-proc-1",
					))
				})

				It("reports on the SynchronizedBeforeSuite and SynchronizedAfterSuite as having passed", func() {
					Ω(reporter.Did.FindByLeafNodeType(types.NodeTypeSynchronizedBeforeSuite)).Should(HavePassed())
					Ω(reporter.Did.FindByLeafNodeType(types.NodeTypeSynchronizedAfterSuite)).Should(HavePassed())
				})

				It("passes data between the two SynchronizedBeforeSuite functions and up to the server", func() {
					Ω(rt).Should(HaveRunWithData("before-suite-all-procs", "data", "hey there"))
					state, data, err := client.BlockUntilSynchronizedBeforeSuiteData()
					Ω(state).Should(Equal(types.SpecStatePassed))
					Ω(data).Should(Equal([]byte("hey there")))
					Ω(err).ShouldNot(HaveOccurred())
				})

				It("emits the output of the proc-1 BeforeSuite function and the proc-1 AfterSuite function", func() {
					Ω(string(serverOutputBuffer.Contents())).Should(Equal("before-suite-proc-1after-suite-proc-1"))
					Ω(reporter.Did.FindByLeafNodeType(types.NodeTypeSynchronizedBeforeSuite)).Should(HavePassed(CapturedStdOutput("before-suite-proc-1before-suite-all-procs")))
					Ω(reporter.Did.FindByLeafNodeType(types.NodeTypeSynchronizedAfterSuite)).Should(HavePassed(CapturedStdOutput("after-suite-all-procsafter-suite-proc-1")))
				})
			})

			Describe("when the BeforeSuite proc1 function fails", func() {
				BeforeEach(func() {
					close(exitChannels[2]) //trigger proc 2 exiting so the proc1 after suite runs
					failInBeforeSuiteProc1 = true
					success, _ := RunFixture("happy-path", fixture)
					Ω(success).Should(BeFalse())
				})

				It("tells the server", func() {
					state, data, err := client.BlockUntilSynchronizedBeforeSuiteData()
					Ω(state).Should(Equal(types.SpecStateFailed))
					Ω(data).Should(BeNil())
					Ω(err).ShouldNot(HaveOccurred())
				})
			})

			Describe("waiting for all procs to finish before running the AfterSuite proc 1 function", func() {
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

		Describe("when running as another proc", func() {
			BeforeEach(func() {
				conf.ParallelProcess = 2
			})

			Describe("happy path", func() {
				BeforeEach(func() {
					client.PostSynchronizedBeforeSuiteCompleted(types.SpecStatePassed, []byte("hola hola"))
					success, _ := RunFixture("happy-path", fixture)
					Ω(success).Should(BeTrue())
				})

				It("runs all the all-procs functions", func() {
					Ω(rt).Should(HaveTracked(
						"before-suite-all-procs",
						"test",
						"after-suite-all-procs",
					))
				})

				It("reports on the SynchronizedBeforeSuite and SynchronizedAfterSuite as having passed", func() {
					Ω(reporter.Did.FindByLeafNodeType(types.NodeTypeSynchronizedBeforeSuite)).Should(HavePassed())
					Ω(reporter.Did.FindByLeafNodeType(types.NodeTypeSynchronizedAfterSuite)).Should(HavePassed())
				})

				It("gets data for the SynchronizedBeforeSuite all procs function from the server", func() {
					Ω(rt).Should(HaveRunWithData("before-suite-all-procs", "data", "hola hola"))
				})
			})

			Describe("waiting for the data from proc 1", func() {
				It("waits for the server to give it the data", func() {
					done := make(chan interface{})
					go func() {
						defer GinkgoRecover()
						success, _ := RunFixture("happy-path", fixture)
						Ω(success).Should(BeTrue())
						close(done)
					}()
					Consistently(done).ShouldNot(BeClosed())
					client.PostSynchronizedBeforeSuiteCompleted(types.SpecStatePassed, []byte("hola hola"))
					Eventually(done).Should(BeClosed())
					Ω(rt).Should(HaveRunWithData("before-suite-all-procs", "data", "hola hola"))
				})
			})

			Describe("when proc 1 fails the SynchronizedBeforeSuite proc1 function", func() {
				It("fails and only runs the after suite", func() {
					done := make(chan interface{})
					go func() {
						defer GinkgoRecover()
						success, _ := RunFixture("happy-path", fixture)
						Ω(success).Should(BeFalse())
						close(done)
					}()
					Consistently(done).ShouldNot(BeClosed())
					client.PostSynchronizedBeforeSuiteCompleted(types.SpecStateFailed, nil)
					Eventually(done).Should(BeClosed())

					Ω(rt).Should(HaveTracked("after-suite-all-procs"))

					Ω(reporter.Did.FindByLeafNodeType(types.NodeTypeSynchronizedBeforeSuite)).Should(HaveFailed(types.GinkgoErrors.SynchronizedBeforeSuiteFailedOnProc1().Error()))
				})
			})

			Describe("when the proc1 SynchronizedBeforeSuite function Skips()", func() {
				It("fails and only runs the after suite", func() {
					done := make(chan interface{})
					go func() {
						defer GinkgoRecover()
						success, _ := RunFixture("happy-path", fixture)
						Ω(success).Should(BeTrue())
						close(done)
					}()
					Consistently(done).ShouldNot(BeClosed())
					client.PostSynchronizedBeforeSuiteCompleted(types.SpecStateSkipped, nil)
					Eventually(done).Should(BeClosed())

					Ω(rt).Should(HaveTracked("after-suite-all-procs"))

					Ω(reporter.Did.FindByLeafNodeType(types.NodeTypeSynchronizedBeforeSuite)).Should(HaveBeenSkipped())
				})
			})

			Describe("when proc 1 disappears before the proc 1 function returns", func() {
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

					Ω(rt).Should(HaveTracked("after-suite-all-procs"))

					Ω(reporter.Did.FindByLeafNodeType(types.NodeTypeSynchronizedBeforeSuite)).Should(HaveFailed(types.GinkgoErrors.SynchronizedBeforeSuiteDisappearedOnProc1().Error()))
				})
			})
		})
	})
})
