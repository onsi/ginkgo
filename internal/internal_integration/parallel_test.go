package internal_integration_test

import (
	"net/http"
	"sync"
	"time"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/internal"
	"github.com/onsi/ginkgo/internal/parallel_support"
	. "github.com/onsi/ginkgo/internal/test_helpers"
	. "github.com/onsi/gomega"
)

var _ = Describe("Running tests in parallel", func() {
	var conf2 config.GinkgoConfigType
	var reporter2 *FakeReporter
	var rt2 *RunTracker

	var fixture = func(rt *RunTracker) {
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

		SynchronizedAfterSuite(rt.T("after-suite-1"), rt.T("after-suite-2"))
	}

	BeforeEach(func() {
		//set up configuration for node 1 and node 2
		conf.ParallelTotal = 2
		conf.ParallelNode = 1
		conf2 = config.GinkgoConfigType{
			ParallelTotal: 2,
			ParallelNode:  2,
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
			fixture(rt)
			Ω(suite1.BuildTree()).Should(Succeed())
		})

		//now construct suite 2...
		suite2 := internal.NewSuite()
		rt2 = NewRunTracker()
		WithSuite(suite2, func() {
			fixture(rt2)
			Ω(suite2.BuildTree()).Should(Succeed())
		})

		finished := make(chan bool)
		//now launch suite 1...
		go func() {
			success, _ := suite1.Run("node 1", failer, reporter, writer, outputInterceptor, interruptHandler, conf)
			finished <- success
			aliveState.Store(1, false)
		}()

		//and launch suite 2...
		reporter2 = &FakeReporter{}
		go func() {
			success, _ := suite2.Run("node 2", internal.NewFailer(), reporter2, writer, outputInterceptor, interruptHandler, conf2)
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
			"A", "B", "C", "D", "E", "F", //all ran
		))

		Ω(reporter.Did.Names()).ShouldNot(BeEmpty())
		Ω(reporter2.Did.Names()).ShouldNot(BeEmpty())
		names := append(reporter.Did.Names(), reporter2.Did.Names()...)
		Ω(names).Should(ConsistOf("A", "B", "C", "D", "E", "F"))
	})

	It("reports the correct statistics", func() {
		Ω(reporter.End.NumberOfTotalSpecs).Should(Equal(6))
		Ω(reporter2.End.NumberOfTotalSpecs).Should(Equal(6))
		Ω(reporter.End.NumberOfSpecsThatWillBeRun).Should(Equal(6))
		Ω(reporter2.End.NumberOfSpecsThatWillBeRun).Should(Equal(6))

		Ω(reporter.End.NumberOfPassedSpecs + reporter2.End.NumberOfPassedSpecs).Should(Equal(6))
	})
})
