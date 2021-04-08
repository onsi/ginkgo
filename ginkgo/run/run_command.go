package run

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/formatter"
	"github.com/onsi/ginkgo/ginkgo/command"
	"github.com/onsi/ginkgo/ginkgo/internal"
	"github.com/onsi/ginkgo/ginkgo/interrupthandler"
	"github.com/onsi/ginkgo/types"
)

func BuildRunCommand() command.Command {
	var ginkgoConfig = config.NewDefaultGinkgoConfig()
	var reporterConfig = config.NewDefaultReporterConfig()
	var cliConfig = config.NewDefaultGinkgoCLIConfig()
	var goFlagsConfig = config.NewDefaultGoFlagsConfig()

	flags, err := config.BuildRunCommandFlagSet(&ginkgoConfig, &reporterConfig, &cliConfig, &goFlagsConfig)
	if err != nil {
		panic(err)
	}

	interruptHandler := interrupthandler.NewInterruptHandler()

	return command.Command{
		Name:          "run",
		Flags:         flags,
		Usage:         "ginkgo run <FLAGS> <PACKAGES> -- <PASS-THROUGHS>",
		ShortDoc:      "Run the tests in the passed in <PACKAGES> (or the package in the current directory if left blank)",
		Documentation: "Any arguments after -- will be passed to the test.",
		DocLink:       "running-tests",
		Command: func(args []string, additionalArgs []string) {
			var errors []error
			cliConfig, goFlagsConfig, errors = config.VetAndInitializeCLIAndGoConfig(cliConfig, goFlagsConfig)
			command.AbortIfErrors("Ginkgo detected configuraiotn issues:", errors)

			runner := &SpecRunner{
				cliConfig:      cliConfig,
				goFlagsConfig:  goFlagsConfig,
				ginkgoConfig:   ginkgoConfig,
				reporterConfig: reporterConfig,
				flags:          flags,

				interruptHandler: interruptHandler,
			}

			runner.RunSpecs(args, additionalArgs)
		},
	}
}

type SpecRunner struct {
	ginkgoConfig   config.GinkgoConfigType
	reporterConfig config.DefaultReporterConfigType
	cliConfig      config.GinkgoCLIConfigType
	goFlagsConfig  config.GoFlagsConfigType
	flags          config.GinkgoFlagSet

	interruptHandler *interrupthandler.InterruptHandler
}

//TODO - THIS IS INTENTIONALLY SLOWER THAN GINKGO V1 FOR NOW
//COMPILATION AND RUNNING AND SERIALIZED.  IN TIME WE'LL RUN EXPERIMENTS TO FIND THE OPTIMAL WAY TO
//IMPROVE PERFORMANCE
func (r *SpecRunner) RunSpecs(args []string, additionalArgs []string) {
	suites, skippedPackages := internal.FindSuites(args, r.cliConfig, true)
	if len(skippedPackages) > 0 {
		fmt.Println("Will skip:")
		for _, skippedPackage := range skippedPackages {
			fmt.Println("  " + skippedPackage)
		}
	}

	if len(skippedPackages) > 0 && len(suites) == 0 {
		command.Abort(command.AbortDetails{
			ExitCode: 0,
			Error:    fmt.Errorf("All tests skipped! Exiting..."),
		})
	}

	if len(suites) == 0 {
		command.AbortWith("Found no test suites")
	}

	if len(suites) > 1 && !r.flags.WasSet("succinct") && !r.reporterConfig.Verbose {
		r.reporterConfig.Succinct = true
	}

	t := time.Now()
	iteration := 0

	var failedSuites []internal.TestSuite
	var hasProgrammaticFocus = false

OUTER_LOOP:
	for {
		if !r.flags.WasSet("seed") {
			r.ginkgoConfig.RandomSeed = time.Now().Unix()
		}
		if r.cliConfig.RandomizeSuites && len(suites) > 1 {
			suites = r.randomize(suites)
		}
		failedSuites = []internal.TestSuite{}
		hasProgrammaticFocus = false

	SUITE_LOOP:
		for suiteIdx := range suites {
			if r.interruptHandler.WasInterrupted() {
				break OUTER_LOOP
			}

			suites[suiteIdx] = internal.CompileSuite(suites[suiteIdx], r.goFlagsConfig)
			if suites[suiteIdx].CompilationError != nil {
				fmt.Println(suites[suiteIdx].CompilationError.Error())
				failedSuites = append(failedSuites, suites[suiteIdx])
				if r.cliConfig.KeepGoing {
					continue SUITE_LOOP
				} else {
					break SUITE_LOOP
				}
			}

			if r.interruptHandler.WasInterrupted() {
				break OUTER_LOOP
			}

			suites[suiteIdx] = internal.RunCompiledSuite(suites[suiteIdx], r.ginkgoConfig, r.reporterConfig, r.cliConfig, r.goFlagsConfig, additionalArgs)
			hasProgrammaticFocus = hasProgrammaticFocus || suites[suiteIdx].HasProgrammaticFocus
			if !suites[suiteIdx].Passed {
				failedSuites = append(failedSuites, suites[suiteIdx])
				if !r.cliConfig.KeepGoing {
					break SUITE_LOOP
				}
			}
		}

		if !r.cliConfig.UntilItFails || len(failedSuites) > 0 {
			if iteration > 0 {
				fmt.Printf("\nTests failed on attempt #%d\n\n", iteration+1)
			}
			break OUTER_LOOP
		}

		fmt.Printf("\nAll tests passed...\nWill keep running them until they fail.\nThis was attempt #%d\n%s\n", iteration+1, orcMessage(iteration+1))
		iteration += 1
	}

	internal.Cleanup(suites...)

	messages, err := internal.FinalizeProfilesForSuites(suites, r.cliConfig, r.goFlagsConfig)
	command.AbortIfError("could not finalize profiles:", err)
	for _, message := range messages {
		fmt.Println(message)
	}

	fmt.Printf("\nGinkgo ran %d %s in %s\n", len(suites), internal.PluralizedWord("suite", "suites", len(suites)), time.Since(t))

	if len(failedSuites) == 0 {
		if hasProgrammaticFocus && strings.TrimSpace(os.Getenv("GINKGO_EDITOR_INTEGRATION")) == "" {
			fmt.Printf("Test Suite Passed\n")
			fmt.Printf("Detected Programmatic Focus - setting exit status to %d\n", types.GINKGO_FOCUS_EXIT_CODE)
			command.Abort(command.AbortDetails{ExitCode: types.GINKGO_FOCUS_EXIT_CODE})
		} else {
			fmt.Printf("Test Suite Passed\n")
			command.Abort(command.AbortDetails{})
		}
	} else {
		fmt.Fprintln(formatter.ColorableStdOut, "")
		if len(failedSuites) > 1 {
			fmt.Fprintln(formatter.ColorableStdOut,
				internal.FailedSuitesReport(failedSuites, formatter.NewWithNoColorBool(r.reporterConfig.NoColor)))
		}
		fmt.Printf("Test Suite Failed\n")
		command.Abort(command.AbortDetails{ExitCode: 1})
	}
}

func (r *SpecRunner) randomize(suites []internal.TestSuite) []internal.TestSuite {
	randomized := make([]internal.TestSuite, len(suites))
	randomizer := rand.New(rand.NewSource(r.ginkgoConfig.RandomSeed))
	permutation := randomizer.Perm(len(suites))
	for i, j := range permutation {
		randomized[i] = suites[j]
	}
	return randomized
}

func orcMessage(iteration int) string {
	if iteration < 10 {
		return ""
	} else if iteration < 30 {
		return []string{
			"If at first you succeed...",
			"...try, try again.",
			"Looking good!",
			"Still good...",
			"I think your tests are fine....",
			"Yep, still passing",
			"Oh boy, here I go testin' again!",
			"Even the gophers are getting bored",
			"Did you try -race?",
			"Maybe you should stop now?",
			"I'm getting tired...",
			"What if I just made you a sandwich?",
			"Hit ^C, hit ^C, please hit ^C",
			"Make it stop. Please!",
			"Come on!  Enough is enough!",
			"Dave, this conversation can serve no purpose anymore. Goodbye.",
			"Just what do you think you're doing, Dave? ",
			"I, Sisyphus",
			"Insanity: doing the same thing over and over again and expecting different results. -Einstein",
			"I guess Einstein never tried to churn butter",
		}[iteration-10] + "\n"
	} else {
		return "No, seriously... you can probably stop now.\n"
	}
}
