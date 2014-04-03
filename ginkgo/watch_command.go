package main

import (
	"flag"
	"fmt"
	"github.com/onsi/ginkgo/ginkgo/testsuite"
)

func BuildWatchCommand() *Command {
	commandFlags := NewWatchCommandFlags(flag.NewFlagSet("watch", flag.ExitOnError))
	watcher := &SpecWatcher{
		commandFlags: commandFlags,
		manager:      NewTestRunnerManager(commandFlags),
		notifier:     NewNotifier(commandFlags),
	}

	return &Command{
		Name:         "watch",
		FlagSet:      commandFlags.FlagSet,
		UsageCommand: "ginkgo watch <FLAGS> <PACKAGES>...",
		Usage: []string{
			"Watches the tests in the passed in <PACKAGES> and runs them when changes occur.",
		},
		Command:                   watcher.WatchSpecs,
		SuppressFlagDocumentation: true,
		FlagDocSubstitute: []string{
			"Accepts all the flags that the ginkgo command accepts except for --keepGoing and --untilItFails",
		},
	}
}

type SpecWatcher struct {
	commandFlags *RunAndWatchCommandFlags
	manager      *TestRunnerManager
	notifier     *Notifier
}

func (w *SpecWatcher) WatchSpecs(args []string) {
	w.notifier.VerifyNotificationsAreAvailable()

	suites := findSuites(args, w.commandFlags.Recurse)
	w.WatchSuites(suites)
}

func (w *SpecWatcher) WatchSuites(suites []*testsuite.TestSuite) {
	modifiedSuite := make(chan *testsuite.TestSuite)
	for _, suite := range suites {
		go suite.Watch(modifiedSuite)
	}

	if len(suites) == 1 {
		w.RunSuite(suites[0])
	}

	for {
		suite := <-modifiedSuite
		w.notifier.SendNotification("Ginkgo", fmt.Sprintf(`Detected change in "%s"...`, suite.PackageName))

		fmt.Printf("\n\nDetected change in %s\n\n", suite.PackageName)
		w.RunSuite(suite)
	}
}

func (w *SpecWatcher) RunSuite(suite *testsuite.TestSuite) {
	runner := w.manager.MakeAndRegisterTestRunner(suite)
	err := runner.Compile()
	if err != nil {
		fmt.Print(err.Error())
	}
	suitePassed := (err == nil) && runner.Run()
	w.notifier.SendSuiteCompletionNotification(suite, suitePassed)
	runner.CleanUp()
	w.manager.UnregisterRunner(runner)
}
