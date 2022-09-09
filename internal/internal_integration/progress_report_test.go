package internal_integration_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("Progress Reporting", func() {
	var cl types.CodeLocation

	clLine := func(offset int) string {
		cl := cl
		cl.LineNumber += offset
		return cl.String()
	}

	progressShouldSay := func(sayings ...string) {
		buf := gbytes.NewBuffer()
		for _, line := range reporter.EmittedImmediately {
			buf.Write([]byte(line + "\n"))
		}
		for _, saying := range sayings {
			Expect(buf).WithOffset(1).To(gbytes.Say(saying))
		}
	}

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
					triggerProgressSignal()
				})
				It("runs", func() {})
			})
			Ω(success).Should(BeTrue())
		})

		It("emits progress when asked and includes source code", func() {
			progressShouldSay(
				`In {{bold}}{{orange}}\[BeforeSuite\]{{/}`,
				clLine(-1),
				`\n{{bold}}{{underline}}Spec Goroutine{{/}}`,
				`\n{{orange}}goroutine \d+ \[running\]{{/}}`,
				clLine(1),
				`\|\s*BeforeSuite\(func\(\) {`,
				`\|\s*cl = types.NewCodeLocation\(0\)`,
				`{{bold}}{{orange}}>\s*triggerProgressSignal\(\){{/}}`,
				`\|\s*}\)`,
				`\|\s*It\("runs",`,
			)
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
			progressShouldSay(
				`{{/}}a container{{/}} {{bold}}{{orange}}A{{/}}`,
				clLine(-1),
				`In {{bold}}{{orange}}\[It\]{{/}}`,
				clLine(-1),
				`{{bold}}{{underline}}Spec Goroutine{{/}}`,
				`{{orange}}goroutine \d+ \[running\]{{/}}`,
				clLine(1),
				`>\s*triggerProgressSignal\(\)`,
			)
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
			progressShouldSay(
				`{{/}}a container{{/}} {{bold}}{{orange}}A{{/}}`,
				clLine(4),
				`In {{bold}}{{orange}}\[BeforeEach\]{{/}}`,
				clLine(-1),
				`{{bold}}{{underline}}Spec Goroutine{{/}}`,
				`{{orange}}goroutine \d+ \[running\]{{/}}`,
				clLine(1),
				`>\s*triggerProgressSignal\(\)`,
			)
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
			progressShouldSay(
				`{{/}}a container{{/}} {{bold}}{{orange}}A{{/}}`,
				clLine(-1),
				`In {{bold}}{{orange}}\[It\]{{/}}`,
				clLine(-1),
				`At {{bold}}{{orange}}\[By Step\] C{{/}}`,
				clLine(2),
				`{{bold}}{{underline}}Spec Goroutine{{/}}`,
				`{{orange}}goroutine \d+ \[running\]{{/}}`,
				clLine(3),
				`>\s*triggerProgressSignal\(\)`,
			)
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
			progressShouldSay(
				`{{/}}a container{{/}} {{bold}}{{orange}}A{{/}}`,
				clLine(-1),
				`In {{bold}}{{orange}}\[It\]{{/}}`,
				clLine(-1),
				`At {{bold}}{{orange}}\[By Step\] C{{/}}`,
				clLine(2),
				`{{bold}}{{underline}}Spec Goroutine{{/}}`,
				`{{orange}}goroutine \d+ \[running\]{{/}}`,
				clLine(3),
				`>\s*triggerProgressSignal\(\)`,
			)
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
			progressShouldSay(
				`{{/}}a container{{/}} {{bold}}{{orange}}A{{/}}`,
				clLine(-1),
				`In {{bold}}{{orange}}\[It\]{{/}}`,
				clLine(-1),
				`{{bold}}{{underline}}Spec Goroutine{{/}}`,
				`{{orange}}goroutine \d+ \[running\]{{/}}`,
				clLine(1),
				`>\s*triggerProgressSignal\(\)`,
			)

			for _, line := range reporter.EmittedImmediately {
				Ω(line).ShouldNot(ContainSubstring(`By Step`))
			}
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
			progressShouldSay(
				`{{/}}a container{{/}} {{bold}}{{orange}}A{{/}}`,
				clLine(-1),
				`In {{bold}}{{orange}}\[It\]{{/}}`,
				clLine(-1),
				`{{bold}}{{underline}}Spec Goroutine{{/}}`,
				`{{orange}}goroutine \d+ \[running\]{{/}}`,
				clLine(7),
				`triggerProgressSignal\(\)`,
				`{{bold}}{{underline}}Goroutines of Interest{{/}}`,
				`{{orange}}goroutine \d+ \[chan receive\]{{/}}`,
				clLine(4),
				`>\s*<-c`,
				clLine(2),
				`>\s*go func\(\) {`,
			)
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
			progressShouldSay(
				//first hit - starts after 100ms
				`{{/}}a container{{/}} {{bold}}{{orange}}A{{/}} \(Spec Runtime: \d+\.\d+ms\)`,
				clLine(-1),
				`In {{bold}}{{orange}}\[It\]{{/}}`,
				clLine(-1),
				`{{orange}}goroutine \d+ \[sleep\]{{/}}`,
				clLine(1),
				`>\s*time.Sleep\(300 \* time\.Millisecond\)`,

				//subsequent hit - should happen again in the 200ms range
				`{{/}}a container{{/}} {{bold}}{{orange}}A{{/}} \(Spec Runtime: \d+\.\d+ms\)`,
				clLine(-1),
				`In {{bold}}{{orange}}\[It\]{{/}}`,
				clLine(-1),
				`{{orange}}goroutine \d+ \[sleep\]{{/}}`,
				clLine(1),
				`>\s*time\.Sleep\(300 \* time\.Millisecond\)`,
			)
		})
	})

	Context("when a test takes longer then the overriden PollProgressAfter", func() {
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
			progressShouldSay(
				`{{/}}a container{{/}} {{bold}}{{orange}}A{{/}} \(Spec Runtime: \d+\.\d+ms\)`,
				clLine(-1),
				`In {{bold}}{{orange}}\[It\]{{/}}`,
				clLine(-1),
				`{{orange}}goroutine \d+ \[sleep\]{{/}}`,
				clLine(1),
				`>\s*time.Sleep\(50 \* time\.Millisecond\)`,

				`{{/}}a container{{/}} {{bold}}{{orange}}A{{/}} \(Spec Runtime: \d+\.\d+ms\)`,
				clLine(-1),
				`In {{bold}}{{orange}}\[It\]{{/}}`,
				clLine(-1),
				`{{orange}}goroutine \d+ \[sleep\]{{/}}`,
				clLine(1),
				`>\s*time\.Sleep\(50 \* time\.Millisecond\)`,
			)
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
			progressShouldSay(
				`In {{bold}}{{orange}}\[SynchronizedBeforeSuite\]{{/}} \(Node Runtime: \d+\.\d+ms\)`,
				clLine(-1),
				`{{orange}}goroutine \d+ \[sleep\]{{/}}`,
				clLine(3),
				`>\s*time.Sleep\(50 \* time\.Millisecond\)`,

				`In {{bold}}{{orange}}\[SynchronizedBeforeSuite\]{{/}} \(Node Runtime: \d+\.\d+ms\)`,
				clLine(-1),
				`{{orange}}goroutine \d+ \[sleep\]{{/}}`,
				clLine(3),
				`>\s*time.Sleep\(50 \* time\.Millisecond\)`,
			)
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
			progressShouldSay(
				`{{/}}a container{{/}} {{bold}}{{orange}}A{{/}} \(Spec Runtime: \d+\.\d+ms\)`,
				clLine(-1),
				`In {{bold}}{{orange}}\[DeferCleanup\]{{/}}`,
				clLine(1),
				`{{orange}}goroutine \d+ \[sleep\]{{/}}`,
				clLine(2),
				`>\s*time.Sleep\(50 \* time\.Millisecond\)`,

				`{{/}}a container{{/}} {{bold}}{{orange}}A{{/}} \(Spec Runtime: \d+\.\d+ms\)`,
				clLine(-1),
				`In {{bold}}{{orange}}\[DeferCleanup\]{{/}}`,
				clLine(1),
				`{{orange}}goroutine \d+ \[sleep\]{{/}}`,
				clLine(2),
				`>\s*time\.Sleep\(50 \* time\.Millisecond\)`,
			)
		})
	})
})
