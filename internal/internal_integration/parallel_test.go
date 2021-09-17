package internal_integration_test

import (
	"net/http"
	"sync"
	"time"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/internal"
	"github.com/onsi/ginkgo/internal/parallel_support"
	. "github.com/onsi/ginkgo/internal/test_helpers"
	"github.com/onsi/ginkgo/types"
	. "github.com/onsi/gomega"
)

var _ = Describe("Running tests in parallel", func() {
	var conf2 types.SuiteConfig
	var reporter2 *FakeReporter
	var rt2 *RunTracker
	var serialValidator chan interface{}

	var fixture = func(rt *RunTracker, node int) {
		SynchronizedBeforeSuite(func() []byte {
			rt.Run("before-suite-1")
			return []byte("floop")
		}, func(node1Data []byte) {
			rt.Run("before-suite-2 " + string(node1Data))
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
			if node == 2 {
				close(serialValidator)
			}
		}), rt.T("after-suite-2"))
	}

	BeforeEach(func() {
		serialValidator = make(chan interface{})
		//set up configuration for node 1 and node 2
		conf.ParallelTotal = 2
		conf.ParallelNode = 1
		conf.RandomizeAllSpecs = true
		conf.RandomSeed = 17
		conf2 = types.SuiteConfig{
			ParallelTotal:     2,
			ParallelNode:      2,
			RandomSeed:        17,
			RandomizeAllSpecs: true,
		}

		// start up a remote server - we're using the real thing here, not a fake
		server, err := parallel_support.NewServer(2, &FakeReporter{})
		Ω(err).ShouldNot(HaveOccurred())
		server.Start()

		// we're using a SynchronizedAfterSuite and making sure it runs on the correct nodes
		// so we need to pass the server these "alive" callbacks.
		// in real life the ginkgo cli sets these up and monitors the running processes
		// here we do the same but are simply monitoring the running goroutines
		aliveState := &sync.Map{}
		for i := 1; i <= 2; i += 1 {
			node := i
			aliveState.Store(node, true)
			server.RegisterAlive(node, func() bool {
				alive, _ := aliveState.Load(node)
				return alive.(bool)
			})
		}

		// wait for the server to come up
		Eventually(StatusCodePoller(server.Address() + "/up")).Should(Equal(http.StatusOK))
		conf.ParallelHost = server.Address()
		conf2.ParallelHost = server.Address()

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
		//now launch suite 1...
		go func() {
			success, _ := suite1.Run("node 1", "/path/to/suite", failer, reporter, writer, outputInterceptor, interruptHandler, conf)
			finished <- success
			aliveState.Store(1, false)
		}()

		//and launch suite 2...
		reporter2 = &FakeReporter{}
		go func() {
			success, _ := suite2.Run("node 2", "/path/to/suite", internal.NewFailer(), reporter2, writer, outputInterceptor, interruptHandler, conf2)
			finished <- success
			aliveState.Store(2, false)
		}()

		// eventually both suites should finish (and succeed)...
		Eventually(finished).Should(Receive(Equal(true)))
		Eventually(finished).Should(Receive(Equal(true)))

		// ...so we can safely shut down the server
		server.Close()

		// and now we're ready to make asserts on the various run trackers and reporters
	})

	It("distributes tests across the parallel nodes and runs them", func() {
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

	It("only runs serial tests on node 1, after the other node has finished", func() {
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
