package internal_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gleak"

	"github.com/onsi/ginkgo/v2/internal"
)

var _ = Describe("ProgressReporterManager", func() {
	var manager *internal.ProgressReporterManager

	BeforeEach(func() {
		manager = internal.NewProgressReporterManager()
	})

	It("can attach and detach progress reporters", func() {
		Ω(manager.QueryProgressReporters(context.Background(), nil)).Should(BeEmpty())
		cancelA := manager.AttachProgressReporter(func() string { return "A" })
		Ω(manager.QueryProgressReporters(context.Background(), nil)).Should(Equal([]string{"A"}))
		cancelB := manager.AttachProgressReporter(func() string { return "B" })
		Ω(manager.QueryProgressReporters(context.Background(), nil)).Should(Equal([]string{"A", "B"}))
		cancelC := manager.AttachProgressReporter(func() string { return "C" })
		Ω(manager.QueryProgressReporters(context.Background(), nil)).Should(Equal([]string{"A", "B", "C"}))
		cancelB()
		Ω(manager.QueryProgressReporters(context.Background(), nil)).Should(Equal([]string{"A", "C"}))
		cancelA()
		Ω(manager.QueryProgressReporters(context.Background(), nil)).Should(Equal([]string{"C"}))
		cancelC()
		Ω(manager.QueryProgressReporters(context.Background(), nil)).Should(BeEmpty())
	})

	It("bails if a progress reporter takes longer than the passed-in context's deadline", func() {
		startingGoroutines := gleak.Goroutines()
		c := make(chan struct{})
		manager.AttachProgressReporter(func() string { return "A" })
		manager.AttachProgressReporter(func() string { return "B" })
		manager.AttachProgressReporter(func() string {
			<-c
			return "C"
		})
		manager.AttachProgressReporter(func() string { return "D" })
		context, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
		result := manager.QueryProgressReporters(context, nil)
		Ω(result).Should(Equal([]string{"A", "B"}))
		cancel()
		close(c)

		Eventually(gleak.Goroutines).ShouldNot(gleak.HaveLeaked(startingGoroutines))
	})

	It("ignores empty progress reports", func() {
		manager.AttachProgressReporter(func() string { return "A" })
		manager.AttachProgressReporter(func() string { return "" })
		manager.AttachProgressReporter(func() string { return "  " })
		manager.AttachProgressReporter(func() string { return "C" })
		result := manager.QueryProgressReporters(context.Background(), nil)
		Ω(result).Should(Equal([]string{"A", "C"}))
	})

	It("catches panics and reports them as failures", func() {
		manager.AttachProgressReporter(func() string {
			panic("bam")
		})
		manager.AttachProgressReporter(func() string { return "B" })
		failer := internal.NewFailer()
		result := manager.QueryProgressReporters(context.Background(), failer)
		Ω(result).Should(Equal([]string{"failed to query attached progress reporter", "B"}))
		state, failure := failer.Drain()
		Ω(state).Should(Equal(types.SpecStatePanicked))
		Ω(failure.Message).Should(Equal("Test Panicked"))
		Ω(failure.ForwardedPanic).Should(Equal("bam"))
	})
})
