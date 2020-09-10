package parallel_support_test

import (
	"fmt"
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	. "github.com/onsi/ginkgo/internal/parallel_support"
	"github.com/onsi/ginkgo/types"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("ForwardingReporter", func() {
	var (
		server      *ghttp.Server
		reporter    *ForwardingReporter
		interceptor *fakeOutputInterceptor
		debugFile   string
	)

	DebugFileContents := func() *gbytes.Buffer {
		content, err := ioutil.ReadFile(debugFile)
		Ω(err).ShouldNot(HaveOccurred())
		return gbytes.BufferWithBytes(content)
	}

	BeforeEach(func() {
		debugFile = fmt.Sprintf("ginkgo-debug-%d", GinkgoParallelNode())
		server = ghttp.NewServer()

		interceptor = &fakeOutputInterceptor{
			InterceptedOutput: "The intercepted output!",
		}

		reporter = NewForwardingReporter(config.DefaultReporterConfigType{}, server.URL(), interceptor, nil, debugFile)
	})

	AfterEach(func() {
		os.Remove(debugFile)
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

		It("should start intercepting output", func() {
			Ω(interceptor.DidStartInterceptingOutput).Should(BeTrue())
		})

		It("should POST the SuiteSummary and Ginkgo Config to the Ginkgo server", func() {
			Ω(server.ReceivedRequests()).Should(HaveLen(1))
		})

		It("should report to the debugFile", func() {
			debug := DebugFileContents()
			Ω(debug).Should(gbytes.Say("Running Suite: My Test Suite"))
		})
	})

	Context("When a spec will run", func() {
		BeforeEach(func() {
			reporter.WillRun(types.Summary{
				State:         types.SpecStatePassed,
				NodeTexts:     []string{"My test"},
				NodeLocations: []types.CodeLocation{types.NewCodeLocation(0)},
			})
		})

		It("should not send anything to the server", func() {
			Ω(server.ReceivedRequests()).Should(BeEmpty())
		})

		It("should report to the debugFile", func() {
			debug := DebugFileContents()
			Ω(debug).Should(gbytes.Say("My test"))
			Ω(debug).Should(gbytes.Say(`forwarding_reporter_test.go:\d+`))
		})
	})

	Context("When a spec completes", func() {
		BeforeEach(func() {
			cls := []types.CodeLocation{types.NewCodeLocation(0)}
			server.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest("POST", "/DidRun"),
				ghttp.VerifyJSONRepresenting(types.Summary{
					State:             types.SpecStatePassed,
					NodeTexts:         []string{"My test"},
					NodeLocations:     cls,
					CapturedStdOutErr: interceptor.InterceptedOutput,
				}),
			))

			reporter.DidRun(types.Summary{
				State:         types.SpecStatePassed,
				NodeTexts:     []string{"My test"},
				NodeLocations: cls,
			})
		})

		It("should POST the SpecSummary to the Ginkgo server and include any intercepted output", func() {
			Ω(server.ReceivedRequests()).Should(HaveLen(1))
		})

		It("should stop, then start intercepting output", func() {
			Ω(interceptor.DidStopInterceptingOutput).Should(BeTrue())
			Ω(interceptor.DidStartInterceptingOutput).Should(BeTrue())
		})

		It("should report to the debugFile", func() {
			debug := DebugFileContents()
			Ω(debug).Should(gbytes.Say("My test"))
			Ω(debug).Should(gbytes.Say(`forwarding_reporter_test.go:\d+`))
			Ω(debug).Should(gbytes.Say("Begin Captured StdOut/StdErr Output >>"))
			Ω(debug).Should(gbytes.Say("The intercepted output!"))
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

		It("should report to the debugFile", func() {
			debug := DebugFileContents()
			Ω(debug).Should(gbytes.Say("SUCCESS!"))
			Ω(debug).Should(gbytes.Say("10 Passed |"))
		})
	})
})
