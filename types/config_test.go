package types_test

import (
	"flag"
	"net/http"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/types"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("Config", func() {
	It("has valid deprecation doc links", func() {
		flags := types.SuiteConfigFlags
		flags = flags.CopyAppend(types.ParallelConfigFlags...)
		flags = flags.CopyAppend(types.ReporterConfigFlags...)
		flags = flags.CopyAppend(types.GinkgoCLISharedFlags...)
		flags = flags.CopyAppend(types.GinkgoCLIRunAndWatchFlags...)
		flags = flags.CopyAppend(types.GinkgoCLIRunFlags...)
		flags = flags.CopyAppend(types.GinkgoCLIWatchFlags...)
		flags = flags.CopyAppend(types.GoBuildFlags...)
		flags = flags.CopyAppend(types.GoRunFlags...)
		for _, flag := range flags {
			if flag.DeprecatedDocLink != "" {
				Ω(flag.DeprecatedDocLink).Should(BeElementOf(DEPRECATION_ANCHORS))
			}
		}
	})

	Describe("VetConfig", func() {
		var suiteConf types.SuiteConfig
		var repConf types.ReporterConfig
		var flagSet types.GinkgoFlagSet
		var goFlagSet *flag.FlagSet

		BeforeEach(func() {
			var err error
			goFlagSet = flag.NewFlagSet("test", flag.ContinueOnError)
			goFlagSet.Bool("count", false, "")
			goFlagSet.Int("parallel", 0, "")
			flagSet, err = types.NewAttachedGinkgoFlagSet(goFlagSet, types.GinkgoFlags{}, nil, types.GinkgoFlagSections{}, types.GinkgoFlagSection{})
			Ω(err).ShouldNot(HaveOccurred())

			suiteConf = types.NewDefaultSuiteConfig()
			repConf = types.NewDefaultReporterConfig()
		})

		Context("when all is well", func() {
			It("retuns no errors", func() {
				errors := types.VetConfig(flagSet, suiteConf, repConf)
				Ω(errors).Should(BeEmpty())
			})
		})

		Context("when unsupported go flags are parsed", func() {
			BeforeEach(func() {
				goFlagSet.Parse([]string{"-count", "-parallel=2"})
			})
			It("returns errors when unsupported go flags are set", func() {
				errors := types.VetConfig(flagSet, suiteConf, repConf)
				Ω(errors).Should(ConsistOf(types.GinkgoErrors.InvalidGoFlagCount(), types.GinkgoErrors.InvalidGoFlagParallel()))
			})
		})

		Describe("errors related to parallelism", func() {
			Context("when parallel total is less than one", func() {
				BeforeEach(func() {
					suiteConf.ParallelTotal = 0
				})

				It("errors", func() {
					errors := types.VetConfig(flagSet, suiteConf, repConf)
					Ω(errors).Should(ContainElement(types.GinkgoErrors.InvalidParallelTotalConfiguration()))
				})
			})

			Context("when parallel node is less than one", func() {
				BeforeEach(func() {
					suiteConf.ParallelNode = 0
				})

				It("errors", func() {
					errors := types.VetConfig(flagSet, suiteConf, repConf)
					Ω(errors).Should(ConsistOf(types.GinkgoErrors.InvalidParallelNodeConfiguration()))
				})
			})

			Context("when parallel node is greater than parallel total", func() {
				BeforeEach(func() {
					suiteConf.ParallelNode = suiteConf.ParallelTotal + 1
				})

				It("errors", func() {
					errors := types.VetConfig(flagSet, suiteConf, repConf)
					Ω(errors).Should(ConsistOf(types.GinkgoErrors.InvalidParallelNodeConfiguration()))
				})
			})

			Context("when running in parallel", func() {
				var server *ghttp.Server
				BeforeEach(func() {
					server = ghttp.NewServer()
					server.SetAllowUnhandledRequests(true)
					server.SetUnhandledRequestStatusCode(http.StatusOK)

					suiteConf.ParallelTotal = 2
					suiteConf.ParallelHost = server.URL()
				})

				AfterEach(func() {
					server.Close()
				})

				Context("and parallel host is not set", func() {
					BeforeEach(func() {
						suiteConf.ParallelHost = ""
					})
					It("errors", func() {
						errors := types.VetConfig(flagSet, suiteConf, repConf)
						Ω(errors).Should(ConsistOf(types.GinkgoErrors.MissingParallelHostConfiguration()))
					})
				})

				Context("and parallel host is set but fails", func() {
					BeforeEach(func() {
						server.SetUnhandledRequestStatusCode(http.StatusGone)
					})
					It("errors", func() {
						errors := types.VetConfig(flagSet, suiteConf, repConf)
						Ω(errors).Should(ConsistOf(types.GinkgoErrors.UnreachableParallelHost(server.URL())))
					})
				})

				Context("when trying to dry run in parallel", func() {
					BeforeEach(func() {
						suiteConf.DryRun = true
					})
					It("errors", func() {
						errors := types.VetConfig(flagSet, suiteConf, repConf)
						Ω(errors).Should(ConsistOf(types.GinkgoErrors.DryRunInParallelConfiguration()))
					})
				})
			})
		})

		Context("when succint and verbose are both set", func() {
			BeforeEach(func() {
				repConf.Succinct = true
				repConf.Verbose = true
			})
			It("errors", func() {
				errors := types.VetConfig(flagSet, suiteConf, repConf)
				Ω(errors).Should(ConsistOf(types.GinkgoErrors.ConflictingVerboseSuccinctConfiguration()))
			})
		})
	})
})
