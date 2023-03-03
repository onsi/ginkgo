package internal_integration_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/ginkgo/v2/internal/test_helpers"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
)

var _ = Describe("Progress Reporting", func() {
	BeforeEach(func() {
		conf.PollProgressAfter = 100 * time.Millisecond
		conf.PollProgressInterval = 50 * time.Millisecond
	})

	AfterEach(func() {
		Ω(triggerProgressSignal).Should(BeNil(), "It should have unregistered the progress signal handler")
	})

	Context("when progress is reported in a BeforeSuite", func() {
		BeforeEach(func() {
			success, _ := RunFixture("emitting spec progress in a BeforeSuite", func() {
				BeforeSuite(func() {
					cl = types.NewCodeLocation(0)
					writer.Print("ginkgo-writer-content")
					triggerProgressSignal()
				})
				It("runs", func() {})
			})
			Ω(success).Should(BeTrue())
		})

		It("emits progress when asked and includes source code", func() {
			Ω(reporter.ProgressReports).Should(HaveLen(1))
			pr := reporter.ProgressReports[0]
			Ω(pr.Message).Should(Equal("{{bold}}You've requested a progress report:{{/}}"))
			Ω(pr.CurrentNodeType).Should(Equal(types.NodeTypeBeforeSuite))
			Ω(pr.LeafNodeLocation).Should(Equal(clLine(-1)))
			Ω(pr.CurrentStepText).Should(Equal(""))
			Ω(pr.CapturedGinkgoWriterOutput).Should(Equal("ginkgo-writer-content"))
			Ω(pr.SpecGoroutine().State).Should(Equal("running"))
			Ω(pr.SpecGoroutine().Stack).Should(HaveHighlightedStackLine(clLine(2)))

			By("validating that source code is extracted")
			targetCl := clLine(2)
			var functionCall types.FunctionCall
			for _, call := range pr.SpecGoroutine().Stack {
				if call.Filename == targetCl.FileName && call.Line == targetCl.LineNumber {
					functionCall = call
					break
				}
			}
			Ω(functionCall).ShouldNot(BeZero())
			//note - these are the lines from up above
			Ω(functionCall.Source).Should(Equal([]string{
				"\t\t\t\t\tcl = types.NewCodeLocation(0)",
				"\t\t\t\t\twriter.Print(\"ginkgo-writer-content\")",
				"\t\t\t\t\ttriggerProgressSignal()",
				"\t\t\t\t})",
				"\t\t\t\tIt(\"runs\", func() {})",
			}))
			Ω(functionCall.SourceHighlight).Should(Equal(2))
		})
	})

	Context("when progress is emitted in an It", func() {
		BeforeEach(func() {
			success, _ := RunFixture("emitting spec progress", func() {
				Describe("a container", func() {
					It("A", func() {
						cl = types.NewCodeLocation(0)
						triggerProgressSignal()
					})
				})
			})
			Ω(success).Should(BeTrue())
		})

		It("emits progress when asked ", func() {
			Ω(reporter.ProgressReports).Should(HaveLen(1))
			pr := reporter.ProgressReports[0]

			Ω(pr.ContainerHierarchyTexts).Should(ConsistOf("a container"))
			Ω(pr.LeafNodeLocation).Should(Equal(clLine(-1)))
			Ω(pr.LeafNodeText).Should(Equal("A"))

			Ω(pr.CurrentNodeType).Should(Equal(types.NodeTypeIt))
			Ω(pr.CurrentNodeLocation).Should(Equal(clLine(-1)))

			Ω(pr.CurrentStepText).Should(Equal(""))

			Ω(pr.SpecGoroutine().State).Should(Equal("running"))
			Ω(pr.SpecGoroutine().Stack).Should(HaveHighlightedStackLine(clLine(1)))
		})
	})

	Context("when progress is emitted in a setup node", func() {
		BeforeEach(func() {
			success, _ := RunFixture("emitting spec progress", func() {
				Describe("a container", func() {
					BeforeEach(func() {
						cl = types.NewCodeLocation(0)
						triggerProgressSignal()
					})

					It("A", func() {})
				})
			})
			Ω(success).Should(BeTrue())
		})

		It("emits progress when asked ", func() {
			Ω(reporter.ProgressReports).Should(HaveLen(1))
			pr := reporter.ProgressReports[0]

			Ω(pr.ContainerHierarchyTexts).Should(ConsistOf("a container"))
			Ω(pr.LeafNodeLocation).Should(Equal(clLine(4)))
			Ω(pr.LeafNodeText).Should(Equal("A"))

			Ω(pr.CurrentNodeType).Should(Equal(types.NodeTypeBeforeEach))
			Ω(pr.CurrentNodeLocation).Should(Equal(clLine(-1)))

			Ω(pr.CurrentStepText).Should(Equal(""))

			Ω(pr.SpecGoroutine().State).Should(Equal("running"))
			Ω(pr.SpecGoroutine().Stack).Should(HaveHighlightedStackLine(clLine(1)))
		})
	})

	Context("when progress is emitted in a By", func() {
		BeforeEach(func() {
			success, _ := RunFixture("emitting spec progress", func() {
				Describe("a container", func() {
					It("A", func() {
						cl = types.NewCodeLocation(0)
						By("B")
						By("C", func() {
							triggerProgressSignal()
						})
						By("D")
					})
				})
			})
			Ω(success).Should(BeTrue())
		})

		It("emits progress when asked and includes the By step", func() {
			Ω(reporter.ProgressReports).Should(HaveLen(1))
			pr := reporter.ProgressReports[0]

			Ω(pr.ContainerHierarchyTexts).Should(ConsistOf("a container"))
			Ω(pr.LeafNodeLocation).Should(Equal(clLine(-1)))
			Ω(pr.LeafNodeText).Should(Equal("A"))

			Ω(pr.CurrentNodeType).Should(Equal(types.NodeTypeIt))
			Ω(pr.CurrentNodeLocation).Should(Equal(clLine(-1)))

			Ω(pr.CurrentStepText).Should(Equal("C"))
			Ω(pr.CurrentStepLocation).Should(Equal(clLine(2)))

			Ω(pr.SpecGoroutine().State).Should(Equal("running"))
			Ω(pr.SpecGoroutine().Stack).Should(HaveHighlightedStackLine(clLine(3)))
		})
	})

	Context("when progress is emitted after a By", func() {
		BeforeEach(func() {
			success, _ := RunFixture("emitting spec progress", func() {
				Describe("a container", func() {
					It("A", func() {
						cl = types.NewCodeLocation(0)
						By("B")
						By("C")
						triggerProgressSignal()
						By("D")
					})
				})
			})
			Ω(success).Should(BeTrue())
		})

		It("emits progress when asked and includes the last run By step", func() {
			Ω(reporter.ProgressReports).Should(HaveLen(1))
			pr := reporter.ProgressReports[0]

			Ω(pr.ContainerHierarchyTexts).Should(ConsistOf("a container"))
			Ω(pr.LeafNodeLocation).Should(Equal(clLine(-1)))
			Ω(pr.LeafNodeText).Should(Equal("A"))

			Ω(pr.CurrentNodeType).Should(Equal(types.NodeTypeIt))
			Ω(pr.CurrentNodeLocation).Should(Equal(clLine(-1)))

			Ω(pr.CurrentStepText).Should(Equal("C"))
			Ω(pr.CurrentStepLocation).Should(Equal(clLine(2)))

			Ω(pr.SpecGoroutine().State).Should(Equal("running"))
			Ω(pr.SpecGoroutine().Stack).Should(HaveHighlightedStackLine(clLine(3)))
		})
	})

	Context("when progress is emitted in a node with no By", func() {
		BeforeEach(func() {
			success, _ := RunFixture("emitting spec progress", func() {
				Describe("a container", func() {
					BeforeEach(func() {
						By("B")
					})

					It("A", func() {
						cl = types.NewCodeLocation(0)
						triggerProgressSignal()
					})
				})
			})
			Ω(success).Should(BeTrue())
		})

		It("emits progress when asked and does not include the By step", func() {
			Ω(reporter.ProgressReports).Should(HaveLen(1))
			pr := reporter.ProgressReports[0]

			Ω(pr.ContainerHierarchyTexts).Should(ConsistOf("a container"))
			Ω(pr.LeafNodeLocation).Should(Equal(clLine(-1)))
			Ω(pr.LeafNodeText).Should(Equal("A"))

			Ω(pr.CurrentNodeType).Should(Equal(types.NodeTypeIt))
			Ω(pr.CurrentNodeLocation).Should(Equal(clLine(-1)))

			Ω(pr.CurrentStepText).Should(Equal(""))

			Ω(pr.SpecGoroutine().State).Should(Equal("running"))
			Ω(pr.SpecGoroutine().Stack).Should(HaveHighlightedStackLine(clLine(1)))
		})
	})

	Context("when a goroutine is launched by the spec", func() {
		BeforeEach(func() {
			success, _ := RunFixture("emitting spec progress", func() {
				Describe("a container", func() {
					It("A", func() {
						cl = types.NewCodeLocation(0)
						c := make(chan bool, 0)
						go func() {
							c <- true
							<-c
						}()
						<-c
						triggerProgressSignal()
						c <- true
					})
				})
			})
			Ω(success).Should(BeTrue())
		})

		It("lists the goroutine as a line of interest", func() {
			Ω(reporter.ProgressReports).Should(HaveLen(1))
			pr := reporter.ProgressReports[0]

			Ω(pr.ContainerHierarchyTexts).Should(ConsistOf("a container"))
			Ω(pr.LeafNodeLocation).Should(Equal(clLine(-1)))
			Ω(pr.LeafNodeText).Should(Equal("A"))

			Ω(pr.CurrentNodeType).Should(Equal(types.NodeTypeIt))
			Ω(pr.CurrentNodeLocation).Should(Equal(clLine(-1)))

			Ω(pr.CurrentStepText).Should(Equal(""))

			Ω(pr.SpecGoroutine().State).Should(Equal("running"))
			Ω(pr.SpecGoroutine().Stack).Should(HaveHighlightedStackLine(clLine(7)))

			var waitingGoroutine types.Goroutine
			Ω(pr.HighlightedGoroutines()).Should(ContainElement(HaveField("State", "chan receive"), &waitingGoroutine))
			Ω(waitingGoroutine.Stack).Should(HaveHighlightedStackLine(clLine(2)), "the goroutine invocation")
			Ω(waitingGoroutine.Stack).Should(HaveHighlightedStackLine(clLine(4)), "the <-c line in the goroutine")
		})
	})

	Context("when a test takes longer then the configured PollProgressAfter", func() {
		BeforeEach(func() {
			success, _ := RunFixture("emitting spec progress", func() {
				Describe("a container", func() {
					It("A", func() {
						cl = types.NewCodeLocation(0)
						time.Sleep(300 * time.Millisecond)
					})
				})
			})
			Ω(success).Should(BeTrue())
		})

		It("emits progress periodically", func() {
			Ω(len(reporter.ProgressReports)).Should(BeNumerically(">", 1))

			for _, pr := range reporter.ProgressReports {
				Ω(pr.Message).Should(Equal("{{bold}}Automatically polling progress:{{/}}"))
				Ω(pr.ContainerHierarchyTexts).Should(ConsistOf("a container"))
				Ω(pr.LeafNodeLocation).Should(Equal(clLine(-1)))
				Ω(pr.LeafNodeText).Should(Equal("A"))

				Ω(pr.CurrentNodeType).Should(Equal(types.NodeTypeIt))
				Ω(pr.CurrentNodeLocation).Should(Equal(clLine(-1)))
			}
		})
	})

	Context("when a test takes longer then the overridden PollProgressAfter", func() {
		BeforeEach(func() {
			success, _ := RunFixture("emitting spec progress", func() {
				Describe("a container", func() {
					It("A", func() {
						cl = types.NewCodeLocation(0)
						time.Sleep(50 * time.Millisecond)
					}, PollProgressAfter(20*time.Millisecond), PollProgressInterval(10*time.Millisecond))
				})
			})
			Ω(success).Should(BeTrue())
		})

		It("emits progress periodically", func() {
			Ω(len(reporter.ProgressReports)).Should(BeNumerically(">", 1))

			for _, pr := range reporter.ProgressReports {
				Ω(pr.ContainerHierarchyTexts).Should(ConsistOf("a container"))
				Ω(pr.LeafNodeLocation).Should(Equal(clLine(-1)))
				Ω(pr.LeafNodeText).Should(Equal("A"))

				Ω(pr.CurrentNodeType).Should(Equal(types.NodeTypeIt))
				Ω(pr.CurrentNodeLocation).Should(Equal(clLine(-1)))
			}
		})
	})

	Context("SynchronizedBeforeSuite can also be decorated", func() {
		BeforeEach(func() {
			success, _ := RunFixture("emitting spec progress", func() {
				SynchronizedBeforeSuite(func() []byte {
					cl = types.NewCodeLocation(0)
					return []byte("hello")
				}, func(_ []byte) {
					time.Sleep(50 * time.Millisecond)
				}, PollProgressAfter(20*time.Millisecond), PollProgressInterval(10*time.Millisecond))
				Describe("a container", func() {
					It("A", func() {})
				})
			})
			Ω(success).Should(BeTrue())
		})

		It("emits progress periodically", func() {
			Ω(len(reporter.ProgressReports)).Should(BeNumerically(">", 1))

			for _, pr := range reporter.ProgressReports {
				Ω(pr.CurrentNodeType).Should(Equal(types.NodeTypeSynchronizedBeforeSuite))
				Ω(pr.CurrentNodeLocation).Should(Equal(clLine(-1)))
			}
		})
	})

	Context("DeferCleanup can also be decorated", func() {
		BeforeEach(func() {
			success, _ := RunFixture("emitting spec progress", func() {
				Describe("a container", func() {
					It("A", func() {
						cl = types.NewCodeLocation(0)
						DeferCleanup(func() {
							time.Sleep(50 * time.Millisecond)
						}, PollProgressAfter(20*time.Millisecond), PollProgressInterval(10*time.Millisecond))
					})
				})
			})
			Ω(success).Should(BeTrue())
		})

		It("emits progress periodically", func() {
			Ω(len(reporter.ProgressReports)).Should(BeNumerically(">", 1))

			for _, pr := range reporter.ProgressReports {
				Ω(pr.CurrentNodeType).Should(Equal(types.NodeTypeCleanupAfterEach))
				Ω(pr.CurrentNodeLocation).Should(Equal(clLine(1)))
			}
		})
	})

	Context("when an additional progress report provider has been registered with the current context", func() {
		BeforeEach(func() {
			success, _ := RunFixture("emitting spec progress", func() {
				Describe("a container", func() {
					It("A", func(ctx SpecContext) {
						cancel := ctx.AttachProgressReporter(func() string { return "Some Additional Information" })
						cl = types.NewCodeLocation(0)
						triggerProgressSignal()
						cancel()
						ctx.AttachProgressReporter(func() string { return "Some Different Information (never cancelled)" })
						triggerProgressSignal()
						cancel = ctx.AttachProgressReporter(func() string { return "Yet More Information" })
						triggerProgressSignal()
						cancel()
						triggerProgressSignal()
					})

					AfterEach(func() {
						triggerProgressSignal()
					})
				})
			})
			Ω(success).Should(BeTrue())
		})

		It("includes information from that progress report provider", func() {
			Ω(reporter.ProgressReports).Should(HaveLen(5))
			pr := reporter.ProgressReports[0]

			Ω(pr.ContainerHierarchyTexts).Should(ConsistOf("a container"))
			Ω(pr.LeafNodeLocation).Should(Equal(clLine(-2)))
			Ω(pr.LeafNodeText).Should(Equal("A"))

			Ω(pr.CurrentNodeType).Should(Equal(types.NodeTypeIt))
			Ω(pr.CurrentNodeLocation).Should(Equal(clLine(-2)))

			Ω(pr.CurrentStepText).Should(Equal(""))

			Ω(pr.SpecGoroutine().State).Should(Equal("running"))
			Ω(pr.SpecGoroutine().Stack).Should(HaveHighlightedStackLine(clLine(1)))
			Ω(pr.AdditionalReports).Should(Equal([]string{"Some Additional Information"}))

			pr = reporter.ProgressReports[1]
			Ω(pr.AdditionalReports).Should(Equal([]string{"Some Different Information (never cancelled)"}))

			pr = reporter.ProgressReports[2]
			Ω(pr.AdditionalReports).Should(Equal([]string{"Some Different Information (never cancelled)", "Yet More Information"}))

			pr = reporter.ProgressReports[3]
			Ω(pr.AdditionalReports).Should(Equal([]string{"Some Different Information (never cancelled)"}))

			pr = reporter.ProgressReports[4]
			Ω(pr.CurrentNodeType).Should(Equal(types.NodeTypeAfterEach))
			Ω(pr.AdditionalReports).Should(BeEmpty())
		})
	})

	Context("when a global progress report provider has been registered", func() {
		BeforeEach(func() {
			success, _ := RunFixture("emitting spec progress", func() {
				Describe("a container", func() {
					It("A", func(ctx SpecContext) {
						cancelGlobal := AttachProgressReporter(func() string { return "Some Global Information" })
						AttachProgressReporter(func() string { return "Some More (Never Cancelled) Global Information" })
						ctx.AttachProgressReporter(func() string { return "Some Additional Information" })
						cl = types.NewCodeLocation(0)
						triggerProgressSignal()
						cancelGlobal()
						triggerProgressSignal()
					})

					It("B", func() {
						triggerProgressSignal()
					})
				})
			})
			Ω(success).Should(BeTrue())
		})

		It("includes information from that progress report provider", func() {
			Ω(reporter.ProgressReports).Should(HaveLen(3))
			pr := reporter.ProgressReports[0]

			Ω(pr.LeafNodeText).Should(Equal("A"))
			Ω(pr.AdditionalReports).Should(Equal([]string{"Some Additional Information", "Some Global Information", "Some More (Never Cancelled) Global Information"}))

			pr = reporter.ProgressReports[1]
			Ω(pr.LeafNodeText).Should(Equal("A"))
			Ω(pr.AdditionalReports).Should(Equal([]string{"Some Additional Information", "Some More (Never Cancelled) Global Information"}))

			pr = reporter.ProgressReports[2]
			Ω(pr.LeafNodeText).Should(Equal("B"))
			Ω(pr.AdditionalReports).Should(Equal([]string{"Some More (Never Cancelled) Global Information"}))
		})
	})

	Context("when a global progress reporter fails", func() {
		BeforeEach(func() {
			success, _ := RunFixture("emitting spec progress", func() {
				Describe("a container", func() {
					It("A", func(ctx SpecContext) {
						AttachProgressReporter(func() string {
							F("bam")
							return "Some Global Information"
						})
						triggerProgressSignal()
					})
				})
			})
			Ω(success).Should(BeFalse())
		})

		It("marks the spec as failed", func() {
			Ω(reporter.Did.Find("A")).Should(HaveFailed("bam"))
		})
	})

})
