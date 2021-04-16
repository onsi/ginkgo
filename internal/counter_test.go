package internal_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/ginkgo/internal"
	"github.com/onsi/ginkgo/internal/parallel_support"
	"github.com/onsi/ginkgo/reporters"
	"github.com/onsi/ginkgo/types"
)

var _ = Describe("Counter", func() {
	var counter func() (int, error)
	var conf types.SuiteConfig

	BeforeEach(func() {
		conf = types.SuiteConfig{}
	})

	JustBeforeEach(func() {
		counter = internal.MakeNextIndexCounter(conf)
	})

	Context("when running in series", func() {
		BeforeEach(func() {
			conf.ParallelTotal = 1
		})

		It("returns a counter that grows by one with each invocation", func() {
			for i := 0; i < 10; i += 1 {
				Ω(counter()).Should(Equal(i))
			}
		})
	})

	Context("when running in parallel", func() {
		var server *parallel_support.Server
		BeforeEach(func() {
			var err error
			conf.ParallelTotal = 2
			server, err = parallel_support.NewServer(2, reporters.NoopReporter{})
			Ω(err).ShouldNot(HaveOccurred())
			server.Start()
			conf.ParallelHost = server.Address()
			client := parallel_support.NewClient(server.Address())
			Eventually(client.CheckServerUp).Should(BeTrue())
		})

		AfterEach(func() {
			server.Close()
		})

		It("returns a counter that grows by one with each invocation", func() {
			for i := 0; i < 10; i += 1 {
				Ω(counter()).Should(Equal(i))
			}
		})
	})
})
