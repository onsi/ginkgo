package config_test

import (
	"flag"
	"net/http"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/internal/test_helpers"
	"github.com/onsi/ginkgo/types"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var DEPRECATION_ANCHORS = test_helpers.LoadMarkdownHeadingAnchors("../docs/MIGRATING_TO_V2.md")

var _ = Describe("Config", func() {
	It("has valid deprecation doc links", func() {
		flags := config.GinkgoConfigFlags.CopyAppend(config.GinkgoParallelConfigFlags...).CopyAppend(config.ReporterConfigFlags...)
		for _, flag := range flags {
			if flag.DeprecatedDocLink != "" {
				Ω(flag.DeprecatedDocLink).Should(BeElementOf(DEPRECATION_ANCHORS))
			}
		}
	})

	Describe("VetConfig", func() {
		var conf config.GinkgoConfigType
		var repConf config.DefaultReporterConfigType
		var flagSet config.GinkgoFlagSet
		var goFlagSet *flag.FlagSet

		BeforeEach(func() {
			var err error
			goFlagSet = flag.NewFlagSet("test", flag.ContinueOnError)
			goFlagSet.Bool("count", false, "")
			goFlagSet.Int("parallel", 0, "")
			flagSet, err = config.NewAttachedGinkgoFlagSet(goFlagSet, config.GinkgoFlags{}, nil, config.GinkgoFlagSections{}, config.GinkgoFlagSection{})
			Ω(err).ShouldNot(HaveOccurred())

			conf = config.NewDefaultGinkgoConfig()
			repConf = config.NewDefaultReporterConfig()
		})

		Context("when all is well", func() {
			It("retuns no errors", func() {
				errors := config.VetConfig(flagSet, conf, repConf)
				Ω(errors).Should(BeEmpty())
			})
		})

		Context("when unsupported go flags are parsed", func() {
			BeforeEach(func() {
				goFlagSet.Parse([]string{"-count", "-parallel=2"})
			})
			It("returns errors when unsupported go flags are set", func() {
				errors := config.VetConfig(flagSet, conf, repConf)
				Ω(errors).Should(ConsistOf(types.GinkgoErrors.InvalidGoFlagCount(), types.GinkgoErrors.InvalidGoFlagParallel()))
			})
		})

		Describe("errors related to parallelism", func() {
			Context("when parallel total is less than one", func() {
				BeforeEach(func() {
					conf.ParallelTotal = 0
				})

				It("errors", func() {
					errors := config.VetConfig(flagSet, conf, repConf)
					Ω(errors).Should(ContainElement(types.GinkgoErrors.InvalidParallelTotalConfiguration()))
				})
			})

			Context("when parallel node is less than one", func() {
				BeforeEach(func() {
					conf.ParallelNode = 0
				})

				It("errors", func() {
					errors := config.VetConfig(flagSet, conf, repConf)
					Ω(errors).Should(ConsistOf(types.GinkgoErrors.InvalidParallelNodeConfiguration()))
				})
			})

			Context("when parallel node is greater than parallel total", func() {
				BeforeEach(func() {
					conf.ParallelNode = conf.ParallelTotal + 1
				})

				It("errors", func() {
					errors := config.VetConfig(flagSet, conf, repConf)
					Ω(errors).Should(ConsistOf(types.GinkgoErrors.InvalidParallelNodeConfiguration()))
				})
			})

			Context("when running in parallel", func() {
				var server *ghttp.Server
				BeforeEach(func() {
					server = ghttp.NewServer()
					server.SetAllowUnhandledRequests(true)
					server.SetUnhandledRequestStatusCode(http.StatusOK)

					conf.ParallelTotal = 2
					conf.ParallelHost = server.URL()
				})

				AfterEach(func() {
					server.Close()
				})

				Context("and parallel host is not set", func() {
					BeforeEach(func() {
						conf.ParallelHost = ""
					})
					It("errors", func() {
						errors := config.VetConfig(flagSet, conf, repConf)
						Ω(errors).Should(ConsistOf(types.GinkgoErrors.MissingParallelHostConfiguration()))
					})
				})

				Context("and parallel host is set but fails", func() {
					BeforeEach(func() {
						server.SetUnhandledRequestStatusCode(http.StatusGone)
					})
					It("errors", func() {
						errors := config.VetConfig(flagSet, conf, repConf)
						Ω(errors).Should(ConsistOf(types.GinkgoErrors.UnreachableParallelHost(server.URL())))
					})
				})

				Context("when trying to dry run in parallel", func() {
					BeforeEach(func() {
						conf.DryRun = true
					})
					It("errors", func() {
						errors := config.VetConfig(flagSet, conf, repConf)
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
				errors := config.VetConfig(flagSet, conf, repConf)
				Ω(errors).Should(ConsistOf(types.GinkgoErrors.ConflictingVerboseSuccinctConfiguration()))
			})
		})
	})
})
