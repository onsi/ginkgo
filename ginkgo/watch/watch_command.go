package watch

import (
	"fmt"
	"regexp"
	"time"

	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/formatter"
	"github.com/onsi/ginkgo/ginkgo/command"
	"github.com/onsi/ginkgo/ginkgo/internal"
	"github.com/onsi/ginkgo/ginkgo/interrupthandler"
)

func BuildWatchCommand() command.Command {
	var ginkgoConfig = config.NewDefaultGinkgoConfig()
	var reporterConfig = config.NewDefaultReporterConfig()
	var cliConfig = config.NewDefaultGinkgoCLIConfig()
	var goFlagsConfig = config.NewDefaultGoFlagsConfig()

	flags, err := config.BuildWatchCommandFlagSet(&ginkgoConfig, &reporterConfig, &cliConfig, &goFlagsConfig)
	if err != nil {
		panic(err)
	}
	interruptHandler := interrupthandler.NewInterruptHandler()

	return command.Command{
		Name:          "watch",
		Flags:         flags,
		Usage:         "ginkgo watch <FLAGS> <PACKAGES> -- <PASS-THROUGHS>",
		ShortDoc:      "Watch the passed in <PACKAGES> and runs their tests whenever changes occur.",
		Documentation: "Any arguments after -- will be passed to the test.",
		DocLink:       "watching-for-changes",
		Command: func(args []string, additionalArgs []string) {
			var errors []error
			cliConfig, goFlagsConfig, errors = config.VetAndInitializeCLIAndGoConfig(cliConfig, goFlagsConfig)
			command.AbortIfErrors("Ginkgo detected configuraiotn issues:", errors)

			watcher := &SpecWatcher{
				cliConfig:      cliConfig,
				goFlagsConfig:  goFlagsConfig,
				ginkgoConfig:   ginkgoConfig,
				reporterConfig: reporterConfig,
				flags:          flags,

				interruptHandler: interruptHandler,
			}

			watcher.WatchSpecs(args, additionalArgs)
		},
	}
}

type SpecWatcher struct {
	ginkgoConfig   config.GinkgoConfigType
	reporterConfig config.DefaultReporterConfigType
	cliConfig      config.GinkgoCLIConfigType
	goFlagsConfig  config.GoFlagsConfigType
	flags          config.GinkgoFlagSet

	interruptHandler *interrupthandler.InterruptHandler
}

func (w *SpecWatcher) WatchSpecs(args []string, additionalArgs []string) {
	suites, _ := internal.FindSuites(args, w.cliConfig, false)

	if len(suites) == 0 {
		command.AbortWith("Found no test suites")
	}

	fmt.Printf("Identified %d test %s.  Locating dependencies to a depth of %d (this may take a while)...\n", len(suites), internal.PluralizedWord("suite", "suites", len(suites)), w.cliConfig.Depth)
	deltaTracker := NewDeltaTracker(w.cliConfig.Depth, regexp.MustCompile(w.cliConfig.WatchRegExp))
	delta, errors := deltaTracker.Delta(suites)

	fmt.Printf("Watching %d %s:\n", len(delta.NewSuites), internal.PluralizedWord("suite", "suites", len(delta.NewSuites)))
	for _, suite := range delta.NewSuites {
		fmt.Println("  " + suite.Description())
	}

	for suite, err := range errors {
		fmt.Printf("Failed to watch %s: %s\n", suite.PackageName, err)
	}

	if len(suites) == 1 {
		w.updateSeed()
		w.compileAndRun(suites[0], additionalArgs)
	}

	ticker := time.NewTicker(time.Second)

	for {
		select {
		case <-ticker.C:
			suites, _ := internal.FindSuites(args, w.cliConfig, false)
			delta, _ := deltaTracker.Delta(suites)
			coloredStream := formatter.ColorableStdOut

			suites = []internal.TestSuite{}

			if len(delta.NewSuites) > 0 {
				fmt.Fprintln(coloredStream, formatter.F("{{green}}Detected %d new %s:{{/}}", len(delta.NewSuites), internal.PluralizedWord("suite", "suites", len(delta.NewSuites))))
				for _, suite := range delta.NewSuites {
					suites = append(suites, suite.Suite)
					fmt.Fprintln(coloredStream, formatter.Fi(1, "%s", suite.Description()))
				}
			}

			modifiedSuites := delta.ModifiedSuites()
			if len(modifiedSuites) > 0 {
				fmt.Fprintln(coloredStream, formatter.F("{{green}}Detected changes in:{{/}}"))
				for _, pkg := range delta.ModifiedPackages {
					fmt.Fprintln(coloredStream, formatter.Fi(1, "%s", pkg))
				}
				fmt.Fprintln(coloredStream, formatter.F("{{green}}Will run %d %s:{{/}}", len(modifiedSuites), internal.PluralizedWord("suite", "suites", len(modifiedSuites))))
				for _, suite := range modifiedSuites {
					suites = append(suites, suite.Suite)
					fmt.Fprintln(coloredStream, formatter.Fi(1, "%s", suite.Description()))
				}
				fmt.Fprintln(coloredStream, "")
			}

			if len(suites) == 0 {
				break
			}

			w.updateSeed()
			w.computeSuccinctMode(len(suites))
			passed := true
			for _, suite := range suites {
				if w.interruptHandler.WasInterrupted() {
					return
				}
				deltaTracker.WillRun(suite)
				passed = w.compileAndRun(suite, additionalArgs) && passed
			}
			color := "{{red}}"
			if passed {
				color = "{{green}}"
			}
			fmt.Fprintln(coloredStream, formatter.F(color+"\nDone.  Resuming watch...{{/}}"))

			err := internal.FinalizeProfilesForSuites(suites, w.cliConfig, w.goFlagsConfig)
			command.AbortIfError("could not finalize profiles:", err)
		case <-w.interruptHandler.InterruptChannel():
			return
		}
	}
}

func (w *SpecWatcher) compileAndRun(suite internal.TestSuite, additionalArgs []string) bool {
	suite = internal.CompileSuite(suite, w.goFlagsConfig)
	if suite.CompilationError != nil {
		fmt.Println(suite.CompilationError.Error())
		return false
	}
	if w.interruptHandler.WasInterrupted() {
		return false
	}
	suite = internal.RunCompiledSuite(suite, w.ginkgoConfig, w.reporterConfig, w.cliConfig, w.goFlagsConfig, additionalArgs)
	internal.Cleanup(suite)
	return suite.Passed
}

func (w *SpecWatcher) computeSuccinctMode(numSuites int) {
	if w.reporterConfig.Verbose {
		w.reporterConfig.Succinct = false
		return
	}

	if w.flags.WasSet("succinct") {
		return
	}

	if numSuites == 1 {
		w.reporterConfig.Succinct = false
	}

	if numSuites > 1 {
		w.reporterConfig.Succinct = true
	}
}

func (w *SpecWatcher) updateSeed() {
	if !w.flags.WasSet("seed") {
		w.ginkgoConfig.RandomSeed = time.Now().Unix()
	}
}
