package internal_integration_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/internal"
	. "github.com/onsi/ginkgo/v2/internal/test_helpers"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
)

var _ = Describe("Running tests in parallel", func() {
	var conf2 types.SuiteConfig
	var reporter2 *FakeReporter
	var rt2 *RunTracker
	var serialValidator chan interface{}

	var fixture = func(rt *RunTracker, proc int) {
		SynchronizedBeforeSuite(func() []byte {
			rt.Run("before-suite-1")
			return []byte("floop")
		}, func(proc1Data []byte) {
			rt.Run("before-suite-2 " + string(proc1Data))
		})

		It("A", rt.T("A", func() {
			time.Sleep(10 * time.Millisecond)
		}))
		It("B", rt.T("B", func() {
			time.Sleep(10 * time.Millisecond)
		}))
		It("C", rt.T("C", func() {
			time.Sleep(10 * time.Millisecond)
		}))
		It("D", rt.T("D", func() {
			time.Sleep(10 * time.Millisecond)
		}))
		It("E", rt.T("E", func() {
			time.Sleep(10 * time.Millisecond)
		}))
		It("F", rt.T("F", func() {
			time.Sleep(10 * time.Millisecond)
		}))
		Context("Ordered", Ordered, func() {
			It("OA", rt.T("OA", func() {
				time.Sleep(10 * time.Millisecond)
			}))
			It("OB", rt.T("OB", func() {
				time.Sleep(10 * time.Millisecond)
			}))
			It("OC", rt.T("OC", func() {
				time.Sleep(10 * time.Millisecond)
			}))
		})
		It("G", Serial, rt.T("G", func() {
			Ω(serialValidator).Should(BeClosed())
			time.Sleep(10 * time.Millisecond)
		}))
		It("H", Serial, rt.T("H", func() {
			Ω(serialValidator).Should(BeClosed())
			time.Sleep(10 * time.Millisecond)
		}))
		It("I", Serial, rt.T("I", func() {
			Ω(serialValidator).Should(BeClosed())
			time.Sleep(10 * time.Millisecond)
		}))
		Context("Ordered and Serial", Ordered, Serial, func() {
			It("OSA", rt.T("OSA", func() {
				Ω(serialValidator).Should(BeClosed())
				time.Sleep(10 * time.Millisecond)
			}))
			It("OSB", rt.T("OSB", func() {
				Ω(serialValidator).Should(BeClosed())
				time.Sleep(10 * time.Millisecond)
			}))
		})

		SynchronizedAfterSuite(rt.T("after-suite-1", func() {
			if proc == 2 {
				close(serialValidator)
			}
		}), rt.T("after-suite-2"))
	}

	BeforeEach(func() {
		serialValidator = make(chan interface{})
		//set up configuration for proc 1 and proc 2

		//SetUpForParallel starts up a server, sets up a client, and sets up the exitChannels map - they're all cleaned up automatically after the test
		SetUpForParallel(2)

		conf.ParallelProcess = 1
		conf.RandomizeAllSpecs = true
		conf.RandomSeed = 17

		conf2 = conf //makes a copy
		conf2.ParallelProcess = 2

		// construct suite 1...
		suite1 := internal.NewSuite()
		WithSuite(suite1, func() {
			fixture(rt, 1)
			Ω(suite1.BuildTree()).Should(Succeed())
		})

		//now construct suite 2...
		suite2 := internal.NewSuite()
		rt2 = NewRunTracker()
		WithSuite(suite2, func() {
			fixture(rt2, 2)
			Ω(suite2.BuildTree()).Should(Succeed())
		})

		finished := make(chan bool)
		exit1 := exitChannels[1] //avoid a race around exitChannels access in a separate goroutine
		//now launch suite 1...
		go func() {
			success, _ := suite1.Run("proc 1", Label("TopLevelLabel"), "/path/to/suite", failer, reporter, writer, outputInterceptor, interruptHandler, client, conf)
			finished <- success
			close(exit1)
		}()

		//and launch suite 2...
		reporter2 = &FakeReporter{}
		exit2 := exitChannels[2] //avoid a race around exitChannels access in a separate goroutine
		go func() {
			success, _ := suite2.Run("proc 2", Label("TopLevelLabel"), "/path/to/suite", internal.NewFailer(), reporter2, writer, outputInterceptor, interruptHandler, client, conf2)
			finished <- success
			close(exit2)
		}()

		// eventually both suites should finish (and succeed)...
		Eventually(finished).Should(Receive(Equal(true)))
		Eventually(finished).Should(Receive(Equal(true)))
		// and now we're ready to make asserts on the various run trackers and reporters
	})

	It("distributes tests across the parallel procs and runs them", func() {
		Ω(rt).Should(HaveRun("before-suite-1"))
		Ω(rt).Should(HaveRun("before-suite-2 floop"))
		Ω(rt).Should(HaveRun("after-suite-1"))
		Ω(rt).Should(HaveRun("after-suite-2"))

		Ω(rt2).ShouldNot(HaveRun("before-suite-1"))
		Ω(rt2).Should(HaveRun("before-suite-2 floop"))
		Ω(rt2).Should(HaveRun("after-suite-1"))
		Ω(rt2).ShouldNot(HaveRun("after-suite-2"))

		allRuns := append(rt.TrackedRuns(), rt2.TrackedRuns()...)
		Ω(allRuns).Should(ConsistOf(
			"before-suite-1", "before-suite-2 floop", "after-suite-1", "after-suite-2", "before-suite-2 floop", "after-suite-1",
			"A", "B", "C", "D", "E", "F", "G", "H", "I", "OA", "OB", "OC", "OSA", "OSB", //all ran
		))

		Ω(reporter.Did.Names()).ShouldNot(BeEmpty())
		Ω(reporter2.Did.Names()).ShouldNot(BeEmpty())
		names := append(reporter.Did.Names(), reporter2.Did.Names()...)
		Ω(names).Should(ConsistOf("A", "B", "C", "D", "E", "F", "G", "H", "I", "OA", "OB", "OC", "OSA", "OSB"))
	})

	It("only runs serial tests on proc 1, after the other proc has finished", func() {
		names := reporter.Did.Names()
		Ω(names).Should(ContainElements("G", "H", "I", "OSA", "OSB"))
		for idx, name := range names {
			if name == "OSA" {
				Ω(names[idx+1]).Should(Equal("OSB"))
				break
			}
		}
		Ω(reporter2.Did.Names()).ShouldNot(ContainElements("G", "H", "I", "OSA", "OSB"))
	})

	It("it ensures specs in an ordered container run on the same process and are ordered", func() {
		names1 := reporter.Did.Names()
		names2 := reporter2.Did.Names()
		in1, _ := ContainElement("OA").Match(names1)
		winner := names1
		if !in1 {
			winner = names2
		}
		found := false
		for idx, name := range winner {
			if name == "OA" {
				found = true
				Ω(winner[idx+1]).Should(Equal("OB"))
				Ω(winner[idx+2]).Should(Equal("OC"))
				break
			}
		}
		Ω(found).Should(BeTrue())
	})

	It("reports the correct statistics", func() {
		Ω(reporter.End.PreRunStats.TotalSpecs).Should(Equal(14))
		Ω(reporter2.End.PreRunStats.TotalSpecs).Should(Equal(14))
		Ω(reporter.End.PreRunStats.SpecsThatWillRun).Should(Equal(14))
		Ω(reporter2.End.PreRunStats.SpecsThatWillRun).Should(Equal(14))

		Ω(reporter.End.SpecReports.WithLeafNodeType(types.NodeTypeIt).CountWithState(types.SpecStatePassed) +
			reporter2.End.SpecReports.WithLeafNodeType(types.NodeTypeIt).CountWithState(types.SpecStatePassed)).Should(Equal(14))
	})
})
