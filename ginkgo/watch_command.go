package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/ginkgo/testrunner"
	"github.com/onsi/ginkgo/ginkgo/testsuite"
)

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

	suites := findSuites(args, w.commandFlags.Recurse, w.commandFlags.SkipPackage)
	w.WatchSuites(suites, additionalArgs)
}

func (w *SpecWatcher) WatchSuites(suites []*testsuite.TestSuite, additionalArgs []string) {
	modifiedSuite := make(chan *testsuite.TestSuite)
	for _, suite := range suites {
		go suite.Watch(modifiedSuite)
	}

	if len(suites) == 1 {
		w.RunSuite(suites[0], additionalArgs)
	}

	for {
		select {
		case suite := <-modifiedSuite:
			w.notifier.SendNotification("Ginkgo", fmt.Sprintf(`Detected change in "%s"...`, suite.PackageName))

			fmt.Printf("\n\nDetected change in %s\n\n", suite.PackageName)
			w.RunSuite(suite, additionalArgs)
		case <-w.interruptHandler.C:
			return
		}
	}
}

func (w *SpecWatcher) RunSuite(suite *testsuite.TestSuite, additionalArgs []string) {
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
