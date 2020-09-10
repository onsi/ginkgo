package reporters_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/ginkgo/config"
	. "github.com/onsi/ginkgo/internal/test_helpers"
	"github.com/onsi/ginkgo/reporters"
	"github.com/onsi/ginkgo/types"
)

var _ = Describe("MultiReporter", func() {
	var fake1, fake2 *FakeReporter
	var multiReporter *reporters.MultiReporter

	BeforeEach(func() {
		fake1 = &FakeReporter{}
		fake2 = &FakeReporter{}
		multiReporter = reporters.NewMultiReporter(fake1, fake2)
	})

	It("forwards all calls to the registered reporters", func() {
		conf := config.GinkgoConfigType{ParallelNode: 3}
		suiteSummary := types.SuiteSummary{SuiteDescription: "foo"}

		multiReporter.SpecSuiteWillBegin(conf, suiteSummary)
		Ω(fake1.Config).Should(Equal(conf))
		Ω(fake1.Begin).Should(Equal(suiteSummary))
		Ω(fake2.Config).Should(Equal(conf))
		Ω(fake2.Begin).Should(Equal(suiteSummary))

		summary := types.Summary{CapturedGinkgoWriterOutput: "spec"}
		multiReporter.WillRun(summary)
		Ω(fake1.Will[0]).Should(Equal(summary))
		Ω(fake2.Will[0]).Should(Equal(summary))
		multiReporter.DidRun(summary)
		Ω(fake1.Did[0]).Should(Equal(summary))
		Ω(fake2.Did[0]).Should(Equal(summary))

		multiReporter.SpecSuiteDidEnd(suiteSummary)
		Ω(fake1.End).Should(Equal(suiteSummary))
		Ω(fake2.End).Should(Equal(suiteSummary))
	})

})
