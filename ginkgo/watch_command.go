package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/ginkgo/testrunner"
	"github.com/onsi/ginkgo/ginkgo/testsuite"
	"github.com/onsi/ginkgo/ginkgo/watch"
)

const greenColor = "\x1b[32m"
const defaultStyle = "\x1b[0m"

func BuildWatchCommand() *Command {
	commandFlags := NewWatchCommandFlags(flag.NewFlagSet("watch", flag.ExitOnError))
	watcher := &SpecWatcher{
		commandFlags:     commandFlags,
		notifier:         NewNotifier(commandFlags),
		interruptHandler: NewInterruptHandler(),
	}

	return &Command{
		Name:         "watch",
		FlagSet:      commandFlags.FlagSet,
		UsageCommand: "ginkgo watch <FLAGS> <PACKAGES> -- <PASS-THROUGHS>",
		Usage: []string{
			"Watches the tests in the passed in <PACKAGES> and runs them when changes occur.",
			"Any arguments after -- will be passed to the test.",
		},
		Command:                   watcher.WatchSpecs,
		SuppressFlagDocumentation: true,
		FlagDocSubstitute: []string{
			"Accepts all the flags that the ginkgo command accepts except for --keepGoing and --untilItFails",
		},
	}
}

type SpecWatcher struct {
	commandFlags     *RunAndWatchCommandFlags
	notifier         *Notifier
	interruptHandler *InterruptHandler
}

func (w *SpecWatcher) WatchSpecs(args []string, additionalArgs []string) {
	w.commandFlags.computeNodes()
	w.notifier.VerifyNotificationsAreAvailable()

	w.WatchSuites(args, additionalArgs)
}

func pluralizedWord(singular, plural string, count int) string {
	if count == 1 {
		return singular
	}
	return plural
}

func (w *SpecWatcher) WatchSuites(args []string, additionalArgs []string) {
	suites, _ := findSuites(args, w.commandFlags.Recurse, w.commandFlags.SkipPackage)
	if len(suites) == 1 {
		w.RunSuite(suites[0], additionalArgs)
	}
	if len(suites) == 0 {
		complainAndQuit("Found no test suites")
	}

	fmt.Printf("Identified %d test %s.  Locating dependencies to a depth of %d (this may take a while)...\n", len(suites), pluralizedWord("suite", "suites", len(suites)), w.commandFlags.Depth)
	deltaTracker := watch.NewDeltaTracker(w.commandFlags.Depth, w.commandFlags.DepFilter)
	delta, errors := deltaTracker.Delta(suites)

	fmt.Printf("Watching %d %s:\n", len(delta.NewSuites), pluralizedWord("suite", "suites", len(delta.NewSuites)))
	for _, suite := range delta.NewSuites {
		fmt.Println("  " + suite.Description())
	}

	for suite, err := range errors {
		fmt.Printf("Failed to watch %s: %s\n"+suite.PackageName, err)
	}

	ticker := time.NewTicker(time.Second)

	for {
		select {
		case <-ticker.C:
			suites, _ := findSuites(args, w.commandFlags.Recurse, w.commandFlags.SkipPackage)
			delta, _ := deltaTracker.Delta(suites)

			suitesToRun := []testsuite.TestSuite{}

			if len(delta.NewSuites) > 0 {
				fmt.Printf(greenColor+"Detected %d new %s:\n"+defaultStyle, len(delta.NewSuites), pluralizedWord("suite", "suites", len(delta.NewSuites)))
				for _, suite := range delta.NewSuites {
					suitesToRun = append(suitesToRun, suite.Suite)
					fmt.Println("  " + suite.Description())
				}
			}

			modifiedSuites := delta.ModifiedSuites()
			if len(modifiedSuites) > 0 {
				fmt.Println(greenColor + "\nDetected changes in:" + defaultStyle)
				for _, pkg := range delta.ModifiedPackages {
					fmt.Println("  " + pkg)
				}
				fmt.Printf(greenColor+"Will run %d %s:\n"+defaultStyle, len(modifiedSuites), pluralizedWord("suite", "suites", len(modifiedSuites)))
				for _, suite := range modifiedSuites {
					suitesToRun = append(suitesToRun, suite.Suite)
					fmt.Println("  " + suite.Description())
				}
				fmt.Println("")
			}

			if len(suitesToRun) > 0 {
				w.ComputeSuccinctMode(len(suitesToRun))
				for _, suite := range suitesToRun {
					deltaTracker.WillRun(suite)
					w.RunSuite(suite, additionalArgs)
				}
				fmt.Println(greenColor + "\nDone.  Resuming watch..." + defaultStyle)
			}

		case <-w.interruptHandler.C:
			return
		}
	}
}

func (w *SpecWatcher) ComputeSuccinctMode(numSuites int) {
	if config.DefaultReporterConfig.Verbose {
		config.DefaultReporterConfig.Succinct = false
		return
	}

	if w.commandFlags.wasSet("succinct") {
		return
	}

	if numSuites == 1 {
		config.DefaultReporterConfig.Succinct = false
	}

	if numSuites > 1 {
		config.DefaultReporterConfig.Succinct = true
	}
}

func (w *SpecWatcher) RunSuite(suite testsuite.TestSuite, additionalArgs []string) {
	w.notifier.SendNotification("Ginkgo", fmt.Sprintf(`Detected change in "%s"...`, suite.PackageName))
	w.UpdateSeed()
	runner := testrunner.New(suite, w.commandFlags.NumCPU, w.commandFlags.ParallelStream, w.commandFlags.Race, w.commandFlags.Cover, w.commandFlags.Tags, additionalArgs)
	err := runner.Compile()
	if err != nil {
		fmt.Print(err.Error())
	}
	runResult := testrunner.FailingRunResult()
	if err == nil {
		runResult = runner.Run()
	}
	w.notifier.SendSuiteCompletionNotification(suite, runResult.Passed)
	runner.CleanUp()
}

func (w *SpecWatcher) UpdateSeed() {
	if !w.commandFlags.wasSet("seed") {
		config.GinkgoConfig.RandomSeed = time.Now().Unix()
	}
}
