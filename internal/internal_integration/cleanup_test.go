package internal_integration_test

import (
	"fmt"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/internal/interrupt_handler"
	. "github.com/onsi/ginkgo/v2/internal/test_helpers"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
)

var _ = Describe("Cleanup", func() {
	C := func(label string) func() {
		return func() {
			DeferCleanup(rt.Run, label)
		}
	}

	Context("the happy path", func() {
		BeforeEach(func() {
			success, _ := RunFixture("cleanup happy path", func() {
				BeforeSuite(rt.T("BS", C("C-BS")))
				AfterSuite(rt.T("AS", C("C-AS")))

				BeforeEach(rt.T("BE-outer", C("C-BE-outer")))
				AfterEach(rt.T("AE-outer", C("C-AE-outer")))

				Context("non-randomizing container", func() {
					Context("container", Ordered, func() {
						JustBeforeEach(rt.T("JBE", C("C-JBE")))
						It("A", rt.T("A", C("C-A")))
						JustAfterEach(rt.T("JAE", C("C-JAE")))
					})

					Context("ordered container", Ordered, func() {
						BeforeAll(rt.T("BA", C("C-BA")))
						BeforeEach(rt.T("BE-inner", C("C-BE-inner")))
						It("B", rt.T("B", C("C-B")))
						It("C", rt.T("C", C("C-C")))
						It("D", rt.T("D", C("C-D")))
						AfterEach(rt.T("AE-inner", C("C-AE-inner")))
						AfterAll(rt.T("AA", C("C-AA")))
					})
				})
			})
			Ω(success).Should(BeTrue())
		})

		It("runs all the things in the correct order", func() {
			Ω(rt).Should(HaveTracked(
				//before suite
				"BS",

				//container
				"BE-outer", "JBE", "A", "JAE", "AE-outer", "C-AE-outer", "C-JAE", "C-A", "C-JBE", "C-BE-outer",

				//ordered container
				"BE-outer", "BA", "BE-inner", "B", "AE-inner", "AE-outer", "C-AE-outer", "C-AE-inner", "C-B", "C-BE-inner", "C-BE-outer",
				"BE-outer", "BE-inner", "C", "AE-inner", "AE-outer", "C-AE-outer", "C-AE-inner", "C-C", "C-BE-inner", "C-BE-outer",
				"BE-outer", "BE-inner", "D", "AE-inner", "AA", "AE-outer", "C-AE-outer", "C-AE-inner", "C-D", "C-BE-inner", "C-BE-outer", "C-AA", "C-BA",

				//after suite
				"AS",
				"C-AS", "C-BS",
			))
		})
	})

	Context("when cleanup fails", func() {
		Context("because of a failed assertion", func() {
			BeforeEach(func() {
				success, _ := RunFixture("cleanup failure", func() {
					BeforeEach(rt.T("BE", func() {
						DeferCleanup(func() {
							rt.Run("C-BE")
							F("fail")
						})
					}))

					It("A", rt.T("A", C("C-A")))
				})
				Ω(success).Should(BeFalse())
			})

			It("reports a failure", func() {
				Ω(rt).Should(HaveTracked("BE", "A", "C-A", "C-BE"))
				Ω(reporter.Did.Find("A")).Should(HaveFailed("fail", FailureNodeType(types.NodeTypeCleanupAfterEach), types.FailureNodeAtTopLevel))
			})
		})

		Context("because of a returned error", func() {
			BeforeEach(func() {
				success, _ := RunFixture("cleanup failure", func() {
					BeforeEach(rt.T("BE", C("C-BE")))
					It("A", rt.T("A", func() {
						DeferCleanup(func() error {
							rt.Run("C-A")
							return fmt.Errorf("fail")
						})
					}))
				})
				Ω(success).Should(BeFalse())
			})

			It("reports a failure", func() {
				Ω(rt).Should(HaveTracked("BE", "A", "C-A", "C-BE"))
				Ω(reporter.Did.Find("A")).Should(HaveFailed("DeferCleanup callback returned error: fail", FailureNodeType(types.NodeTypeCleanupAfterEach), types.FailureNodeAtTopLevel))
			})
		})

		Context("because of a returned error, for a multi-return function", func() {
			BeforeEach(func() {
				success, _ := RunFixture("cleanup failure", func() {
					BeforeEach(rt.T("BE", C("C-BE")))
					It("A", rt.T("A", func() {
						DeferCleanup(func() (string, error) {
							rt.Run("C-A")
							return "ok", fmt.Errorf("fail")
						})
					}))
				})
				Ω(success).Should(BeFalse())
			})

			It("reports a failure", func() {
				Ω(rt).Should(HaveTracked("BE", "A", "C-A", "C-BE"))
				Ω(reporter.Did.Find("A")).Should(HaveFailed("DeferCleanup callback returned error: fail", FailureNodeType(types.NodeTypeCleanupAfterEach), types.FailureNodeAtTopLevel))
			})
		})

		Context("at the suite level", func() {
			BeforeEach(func() {
				success, _ := RunFixture("cleanup failure", func() {
					BeforeSuite(rt.T("BS", func() {
						DeferCleanup(func() {
							rt.Run("C-BS")
							F("fail")
						})
					}))
					Context("container", func() {
						It("A", rt.T("A"))
						It("B", rt.T("B"))
					})
				})
				Ω(success).Should(BeFalse())
			})

			It("marks the suite as failed", func() {
				Ω(rt).Should(HaveTracked("BS", "A", "B", "C-BS"))
				Ω(reporter.End).Should(BeASuiteSummary(false, NSpecs(2), NPassed(2)))
				Ω(reporter.Did.FindByLeafNodeType(types.NodeTypeBeforeSuite)).Should(HavePassed())
				Ω(reporter.Did.FindByLeafNodeType(types.NodeTypeCleanupAfterSuite)).Should(HaveFailed("fail", FailureNodeType(types.NodeTypeCleanupAfterSuite)))
			})
		})

		Context("when cleanup is interrupted", func() {
			BeforeEach(func() {
				success, _ := RunFixture("cleanup failure", func() {
					BeforeEach(rt.T("BE", C("C-BE")))
					It("A", rt.T("A", func() {
						DeferCleanup(func() {
							rt.Run("C-A")
							interruptHandler.Interrupt(interrupt_handler.InterruptCauseSignal)
							time.Sleep(time.Minute)
						})
					}))
				})
				Ω(success).Should(BeFalse())
			})
			It("runs subsequent cleanups and is marked as interrupted", func() {
				Ω(rt).Should(HaveTracked("BE", "A", "C-A", "C-BE"))

				Ω(reporter.Did.Find("A")).Should(HaveBeenInterrupted(interrupt_handler.InterruptCauseSignal))
			})
		})
	})

	Context("edge cases", func() {
		Context("cleanup is added in a SynchronizedBeforeSuite and SynchronizedAfterSuite", func() {
			Context("when running in serial", func() {
				BeforeEach(func() {
					success, _ := RunFixture("cleanup in synchronized suites", func() {
						SynchronizedBeforeSuite(func() []byte {
							rt.Run("BS1")
							DeferCleanup(rt.Run, "C-BS1")
							return nil
						}, func(_ []byte) {
							rt.Run("BS2")
							DeferCleanup(rt.Run, "C-BS2")
						})

						SynchronizedAfterSuite(func() {
							rt.Run("AS1")
							DeferCleanup(rt.Run, "C-AS1")
						}, func() {
							rt.Run("AS2")
							DeferCleanup(rt.Run, "C-AS2")
						})
						Context("ordering", func() {
							It("A", rt.T("A", C("C-A")))
							It("B", rt.T("B", C("C-B")))
						})
					})
					Ω(success).Should(BeTrue())

				})
				It("runs the cleanup at the appropriate time", func() {
					Ω(rt).Should(HaveTracked("BS1", "BS2", "A", "C-A", "B", "C-B", "AS1", "AS2", "C-AS2", "C-AS1", "C-BS2", "C-BS1"))
				})
			})

			Context("when running in parallel and there is no SynchronizedAfterSuite", func() {
				fixture := func() {
					SynchronizedBeforeSuite(func() []byte {
						rt.Run("BS1")
						DeferCleanup(rt.Run, "C-BS1")
						return nil
					}, func(_ []byte) {
						rt.Run("BS2")
						DeferCleanup(rt.Run, "C-BS2")
					})

					Context("ordering", func() {
						It("A", rt.T("A", C("C-A")))
						It("B", rt.T("B", C("C-B")))
					})
				}

				BeforeEach(func() {
					SetUpForParallel(2)
				})

				Context("as process #1", func() {
					It("runs the cleanup only _after_ the other processes have finished", func() {
						done := make(chan any)
						go func() {
							defer GinkgoRecover()
							success, _ := RunFixture("DeferCleanup on SBS in parallel on process 1", fixture)
							Ω(success).Should(BeTrue())
							close(done)
						}()

						Eventually(rt).Should(HaveTracked("BS1", "BS2", "A", "C-A", "B", "C-B"))
						Consistently(rt).Should(HaveTracked("BS1", "BS2", "A", "C-A", "B", "C-B"))
						close(exitChannels[2])
						Eventually(rt).Should(HaveTracked("BS1", "BS2", "A", "C-A", "B", "C-B", "C-BS2", "C-BS1"))
						Eventually(done).Should(BeClosed())
					})
				})

				Context("as process #2", func() {
					BeforeEach(func() {
						conf.ParallelProcess = 2
						client.PostSynchronizedBeforeSuiteCompleted(types.SpecStatePassed, []byte("hola hola"))
						success, _ := RunFixture("DeferCleanup on SBS in parallel on process 2", fixture)
						Ω(success).Should(BeTrue())
					})

					It("runs the cleanup at the appropriate time", func() {
						Ω(rt).Should(HaveTracked("BS2", "A", "C-A", "B", "C-B", "C-BS2"))
					})
				})
			})
		})

		Context("cleanup is added in an AfterAll that is called because an AfterEach has caused the non-final spec in an ordered group to fail", func() {
			BeforeEach(func() {
				success, _ := RunFixture("cleanup in hairy edge case", func() {
					Context("ordered", Ordered, func() {
						It("A", rt.T("A", C("C-A")))
						It("B", rt.T("B"))
						AfterEach(rt.T("AE", func() {
							DeferCleanup(rt.Run, "C-AE")
							F("fail")
						}))
						AfterAll(rt.T("AA", C("C-AA")))
					})
				})
				Ω(success).Should(BeFalse())
			})

			It("notes that a cleanup was registered in the AfterAll and runs it", func() {
				Ω(rt).Should(HaveTracked("A", "AE", "AA", "C-AE", "C-A", "C-AA"))
			})
		})

		Context("when cleanup is added in parallel in some goroutines", func() {
			BeforeEach(func() {
				success, _ := RunFixture("concurrent cleanup", func() {
					Context("ordered", Ordered, func() {
						It("A", func() {
							wg := &sync.WaitGroup{}
							wg.Add(5)
							for i := 0; i < 5; i++ {
								i := i
								go func() {
									DeferCleanup(rt.Run, fmt.Sprintf("dc-%d", i))
									wg.Done()
								}()
							}
							wg.Wait()
						})
					})
				})
				Ω(success).Should(BeTrue())
			})
			It("doesn't race", func() {
				Ω(rt.TrackedRuns()).Should(ConsistOf("dc-0", "dc-1", "dc-2", "dc-3", "dc-4"))
			})
		})
	})
})
