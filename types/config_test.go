package types_test

import (
	"flag"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"
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
				Ω(anchors.DocAnchors["MIGRATING_TO_V2.md"]).Should(ContainElement(flag.DeprecatedDocLink))
			}
		}
	})

	Describe("ReporterConfig", func() {
		Describe("WillGenerateReport", func() {
			It("returns true if it will generate a report", func() {
				repConf := types.ReporterConfig{}
				Ω(repConf.WillGenerateReport()).Should(BeFalse())

				repConf = types.ReporterConfig{JSONReport: "foo"}
				Ω(repConf.WillGenerateReport()).Should(BeTrue())

				repConf = types.ReporterConfig{JUnitReport: "foo"}
				Ω(repConf.WillGenerateReport()).Should(BeTrue())

				repConf = types.ReporterConfig{TeamcityReport: "foo"}
				Ω(repConf.WillGenerateReport()).Should(BeTrue())
			})
		})

		Describe("Verbosity", func() {
			It("returns the appropriate verbosity level", func() {
				repConf := types.ReporterConfig{}
				Ω(repConf.Verbosity()).Should(Equal(types.VerbosityLevelNormal))

				repConf = types.ReporterConfig{Succinct: true}
				Ω(repConf.Verbosity()).Should(Equal(types.VerbosityLevelSuccinct))

				repConf = types.ReporterConfig{Verbose: true}
				Ω(repConf.Verbosity()).Should(Equal(types.VerbosityLevelVerbose))

				repConf = types.ReporterConfig{VeryVerbose: true}
				Ω(repConf.Verbosity()).Should(Equal(types.VerbosityLevelVeryVerbose))
			})

			It("can do verbosity math", func() {
				Ω(types.VerbosityLevelNormal.LT(types.VerbosityLevelVeryVerbose)).Should(BeTrue())
				Ω(types.VerbosityLevelNormal.LT(types.VerbosityLevelVerbose)).Should(BeTrue())
				Ω(types.VerbosityLevelNormal.LT(types.VerbosityLevelNormal)).Should(BeFalse())
				Ω(types.VerbosityLevelNormal.LT(types.VerbosityLevelSuccinct)).Should(BeFalse())

				Ω(types.VerbosityLevelNormal.LTE(types.VerbosityLevelVeryVerbose)).Should(BeTrue())
				Ω(types.VerbosityLevelNormal.LTE(types.VerbosityLevelVerbose)).Should(BeTrue())
				Ω(types.VerbosityLevelNormal.LTE(types.VerbosityLevelNormal)).Should(BeTrue())
				Ω(types.VerbosityLevelNormal.LTE(types.VerbosityLevelSuccinct)).Should(BeFalse())

				Ω(types.VerbosityLevelNormal.GT(types.VerbosityLevelVeryVerbose)).Should(BeFalse())
				Ω(types.VerbosityLevelNormal.GT(types.VerbosityLevelVerbose)).Should(BeFalse())
				Ω(types.VerbosityLevelNormal.GT(types.VerbosityLevelNormal)).Should(BeFalse())
				Ω(types.VerbosityLevelNormal.GT(types.VerbosityLevelSuccinct)).Should(BeTrue())

				Ω(types.VerbosityLevelNormal.GTE(types.VerbosityLevelVeryVerbose)).Should(BeFalse())
				Ω(types.VerbosityLevelNormal.GTE(types.VerbosityLevelVerbose)).Should(BeFalse())
				Ω(types.VerbosityLevelNormal.GTE(types.VerbosityLevelNormal)).Should(BeTrue())
				Ω(types.VerbosityLevelNormal.GTE(types.VerbosityLevelSuccinct)).Should(BeTrue())

				Ω(types.VerbosityLevelNormal.Is(types.VerbosityLevelVeryVerbose)).Should(BeFalse())
				Ω(types.VerbosityLevelNormal.Is(types.VerbosityLevelVerbose)).Should(BeFalse())
				Ω(types.VerbosityLevelNormal.Is(types.VerbosityLevelNormal)).Should(BeTrue())
				Ω(types.VerbosityLevelNormal.Is(types.VerbosityLevelSuccinct)).Should(BeFalse())
			})
		})
	})

	Describe("VetConfig", func() {
		var suiteConf types.SuiteConfig
		var repConf types.ReporterConfig
		var flagSet types.GinkgoFlagSet
		var goFlagSet *flag.FlagSet

		BeforeEach(func() {
			var err error
			goFlagSet = flag.NewFlagSet("test", flag.ContinueOnError)
			goFlagSet.Int("count", 1, "")
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
				goFlagSet.Parse([]string{"-count=2", "-parallel=2"})
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
					suiteConf.ParallelProcess = 0
				})

				It("errors", func() {
					errors := types.VetConfig(flagSet, suiteConf, repConf)
					Ω(errors).Should(ConsistOf(types.GinkgoErrors.InvalidParallelProcessConfiguration()))
				})
			})

			Context("when parallel node is greater than parallel total", func() {
				BeforeEach(func() {
					suiteConf.ParallelProcess = suiteConf.ParallelTotal + 1
				})

				It("errors", func() {
					errors := types.VetConfig(flagSet, suiteConf, repConf)
					Ω(errors).Should(ConsistOf(types.GinkgoErrors.InvalidParallelProcessConfiguration()))
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

		Describe("file filter errors", func() {
			Context("with an invalid --focus-file and/or --skip-file", func() {
				BeforeEach(func() {
					suiteConf.FocusFiles = append(suiteConf.FocusFiles, "bloop:123a")
					suiteConf.SkipFiles = append(suiteConf.SkipFiles, "bloop:123b")
				})

				It("errors", func() {
					errors := types.VetConfig(flagSet, suiteConf, repConf)
					Ω(errors).Should(ConsistOf(types.GinkgoErrors.InvalidFileFilter("bloop:123a"), types.GinkgoErrors.InvalidFileFilter("bloop:123b")))
				})
			})
		})

		Describe("validating --output-interceptor-mode", func() {
			It("errors if an invalid output interceptor mode is specified", func() {
				suiteConf.OutputInterceptorMode = "DURP"
				errors := types.VetConfig(flagSet, suiteConf, repConf)
				Ω(errors).Should(ConsistOf(types.GinkgoErrors.InvalidOutputInterceptorModeConfiguration("DURP")))

				for _, value := range []string{"", "dup", "DUP", "swap", "SWAP", "none", "NONE"} {
					suiteConf.OutputInterceptorMode = value
					errors = types.VetConfig(flagSet, suiteConf, repConf)
					Ω(errors).Should(BeEmpty())
				}
			})
		})

		Context("when more than one verbosity flag is set", func() {
			It("errors", func() {
				repConf.Succinct, repConf.Verbose, repConf.VeryVerbose = true, true, false
				errors := types.VetConfig(flagSet, suiteConf, repConf)
				Ω(errors).Should(ConsistOf(types.GinkgoErrors.ConflictingVerbosityConfiguration()))

				repConf.Succinct, repConf.Verbose, repConf.VeryVerbose = true, false, true
				errors = types.VetConfig(flagSet, suiteConf, repConf)
				Ω(errors).Should(ConsistOf(types.GinkgoErrors.ConflictingVerbosityConfiguration()))

				repConf.Succinct, repConf.Verbose, repConf.VeryVerbose = false, true, true
				errors = types.VetConfig(flagSet, suiteConf, repConf)
				Ω(errors).Should(ConsistOf(types.GinkgoErrors.ConflictingVerbosityConfiguration()))

				repConf.Succinct, repConf.Verbose, repConf.VeryVerbose = true, true, true
				errors = types.VetConfig(flagSet, suiteConf, repConf)
				Ω(errors).Should(ConsistOf(types.GinkgoErrors.ConflictingVerbosityConfiguration()))
			})
		})
	})
})
