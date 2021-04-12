package parallel_support_test

import (
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	. "github.com/onsi/ginkgo/internal/parallel_support"
	"github.com/onsi/ginkgo/types"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("ForwardingReporter", func() {
	var (
		server   *ghttp.Server
		reporter *ForwardingReporter
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
		reporter = NewForwardingReporter(config.DefaultReporterConfigType{}, server.URL(), nil)
	})

	AfterEach(func() {
		server.Close()
	})

	Context("When a suite begins", func() {
		BeforeEach(func() {
			suiteSummary := types.SuiteSummary{
				SuiteDescription: "My Test Suite",
			}

			server.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest("POST", "/SpecSuiteWillBegin"),
				ghttp.VerifyJSONRepresenting(ConfigAndSummary{
					Config:  config.GinkgoConfig,
					Summary: suiteSummary,
				}),
			))

			reporter.SpecSuiteWillBegin(config.GinkgoConfig, suiteSummary)
		})

		It("should POST the SuiteSummary and Ginkgo Config to the Ginkgo server", func() {
			Ω(server.ReceivedRequests()).Should(HaveLen(1))
		})
	})

	Context("When a spec will run", func() {
		BeforeEach(func() {
			reporter.WillRun(types.SpecReport{
				State:         types.SpecStatePassed,
				NodeTexts:     []string{"My test"},
				NodeLocations: []types.CodeLocation{types.NewCodeLocation(0)},
			})
		})

		It("should not send anything to the server", func() {
			Ω(server.ReceivedRequests()).Should(BeEmpty())
		})
	})

	Context("When a spec completes", func() {
		BeforeEach(func() {
			cls := []types.CodeLocation{types.NewCodeLocation(0)}
			server.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest("POST", "/DidRun"),
				ghttp.VerifyJSONRepresenting(types.SpecReport{
					State:         types.SpecStatePassed,
					NodeTexts:     []string{"My test"},
					NodeLocations: cls,
				}),
			))

			reporter.DidRun(types.SpecReport{
				State:         types.SpecStatePassed,
				NodeTexts:     []string{"My test"},
				NodeLocations: cls,
			})
		})

		It("should POST the SpecSummary to the Ginkgo server", func() {
			Ω(server.ReceivedRequests()).Should(HaveLen(1))
		})
	})

	Context("When a suite ends", func() {
		BeforeEach(func() {
			suiteSummary := types.SuiteSummary{
				SuiteDescription:    "My Test Suite",
				SuiteSucceeded:      true,
				NumberOfPassedSpecs: 10,
			}

			server.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest("POST", "/SpecSuiteDidEnd"),
				ghttp.VerifyJSONRepresenting(suiteSummary),
			))

			reporter.SpecSuiteDidEnd(suiteSummary)
		})

		It("should POST the SuiteSummary to the Ginkgo server", func() {
			Ω(server.ReceivedRequests()).Should(HaveLen(1))
		})
	})
})
