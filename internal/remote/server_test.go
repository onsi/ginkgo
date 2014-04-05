package remote_test

import (
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	. "github.com/onsi/ginkgo/internal/remote"
	"github.com/onsi/ginkgo/reporters"
	"github.com/onsi/ginkgo/types"
	. "github.com/onsi/gomega"
	"net/http"
)

var _ = Describe("Server", func() {
	var (
		server               *Server
		reporterA, reporterB *reporters.FakeReporter
		forwardingReporter   *ForwardingReporter

		suiteSummary *types.SuiteSummary
		setupSummary *types.SetupSummary
		specSummary  *types.SpecSummary
	)

	BeforeEach(func() {
		var err error
		server, err = NewServer()
		Ω(err).ShouldNot(HaveOccurred())
		reporterA = reporters.NewFakeReporter()
		reporterB = reporters.NewFakeReporter()

		server.RegisterReporters(reporterA, reporterB)

		forwardingReporter = NewForwardingReporter(server.Address(), &http.Client{}, &fakeOutputInterceptor{})

		suiteSummary = &types.SuiteSummary{
			SuiteDescription: "My Test Suite",
		}

		setupSummary = &types.SetupSummary{
			State: types.SpecStatePassed,
		}

		specSummary = &types.SpecSummary{
			ComponentTexts: []string{"My", "Spec"},
			State:          types.SpecStatePassed,
		}

		server.Start()
	})

	AfterEach(func() {
		server.Stop()
	})

	It("should make its address available", func() {
		Ω(server.Address()).Should(MatchRegexp(`127.0.0.1:\d{2,}`))
	})

	Describe("/SpecSuiteWillBegin", func() {
		It("should decode and forward the Ginkgo config and suite summary", func(done Done) {
			forwardingReporter.SpecSuiteWillBegin(config.GinkgoConfig, suiteSummary)
			Ω(reporterA.Config).Should(Equal(config.GinkgoConfig))
			Ω(reporterB.Config).Should(Equal(config.GinkgoConfig))
			Ω(reporterA.BeginSummary).Should(Equal(suiteSummary))
			Ω(reporterB.BeginSummary).Should(Equal(suiteSummary))
			close(done)
		})
	})

	Describe("/BeforeSuiteDidRun", func() {
		It("should decode and forward the setup summary", func() {
			forwardingReporter.BeforeSuiteDidRun(setupSummary)
			Ω(reporterA.BeforeSuiteSummary).Should(Equal(setupSummary))
			Ω(reporterB.BeforeSuiteSummary).Should(Equal(setupSummary))
		})
	})

	Describe("/SpecWillRun", func() {
		It("should decode and forward the spec summary", func(done Done) {
			forwardingReporter.SpecWillRun(specSummary)
			Ω(reporterA.SpecWillRunSummaries[0]).Should(Equal(specSummary))
			Ω(reporterB.SpecWillRunSummaries[0]).Should(Equal(specSummary))
			close(done)
		})
	})

	Describe("/SpecDidComplete", func() {
		It("should decode and forward the spec summary", func(done Done) {
			forwardingReporter.SpecDidComplete(specSummary)
			Ω(reporterA.SpecSummaries[0]).Should(Equal(specSummary))
			Ω(reporterB.SpecSummaries[0]).Should(Equal(specSummary))
			close(done)
		})
	})

	Describe("/SpecSuiteDidEnd", func() {
		It("should decode and forward the suite summary", func(done Done) {
			forwardingReporter.SpecSuiteDidEnd(suiteSummary)
			Ω(reporterA.EndSummary).Should(Equal(suiteSummary))
			Ω(reporterB.EndSummary).Should(Equal(suiteSummary))
			close(done)
		})
	})
})
