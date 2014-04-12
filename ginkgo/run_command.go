package main

import (
	"flag"
	"fmt"
	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/ginkgo/testrunner"
	"github.com/onsi/ginkgo/ginkgo/testsuite"
	"os"
	"runtime"
	"time"
)

func BuildRunCommand() *Command {
	commandFlags := NewRunCommandFlags(flag.NewFlagSet("ginkgo", flag.ExitOnError))
	runner := &SpecRunner{
		commandFlags:     commandFlags,
		notifier:         NewNotifier(commandFlags),
		interruptHandler: NewInterruptHandler(),
	}

	return &Command{
		Name:         "",
		FlagSet:      commandFlags.FlagSet,
		UsageCommand: "ginkgo <FLAGS> <PACKAGES> -- <PASS-THROUGHS>",
		Usage: []string{
			"Run the tests in the passed in <PACKAGES> (or the package in the current directory if left blank).",
			"Any arguments after -- will be passed to the test.",
			"Accepts the following flags:",
		},
		Command: runner.RunSpecs,
	}
}

type SpecRunner struct {
	commandFlags     *RunAndWatchCommandFlags
	notifier         *Notifier
	interruptHandler *InterruptHandler
}

func (r *SpecRunner) RunSpecs(args []string, additionalArgs []string) {
	r.notifier.VerifyNotificationsAreAvailable()

	suites := findSuites(args, r.commandFlags.Recurse, r.commandFlags.SkipPackage)
	r.ComputeSuccinctMode(len(suites))

	t := time.Now()

	passed := true
	if r.commandFlags.UntilItFails {
		iteration := 0
		for {
			passed = r.RunSuites(suites, additionalArgs)
			iteration++

			if r.interruptHandler.WasInterrupted() {
				break
			}

			if passed {
				fmt.Printf("\nAll tests passed...\nWill keep running them until they fail.\nThis was attempt #%d\n\n", iteration)
			} else {
				fmt.Printf("\nTests failed on attempt #%d\n\n", iteration)
				break
			}
		}
	} else {
		passed = r.RunSuites(suites, additionalArgs)
	}

	fmt.Printf("\nGinkgo ran in %s\n", time.Since(t))

	if passed {
		fmt.Printf("Test Suite Passed\n")
		os.Exit(0)
	} else {
		fmt.Printf("Test Suite Failed\n")
		os.Exit(1)
	}
}

func (r *SpecRunner) ComputeSuccinctMode(numSuites int) {
	if config.DefaultReporterConfig.Verbose {
		config.DefaultReporterConfig.Succinct = false
		return
	}

	if numSuites == 1 {
		return
	}

	didSetSuccinct := false
	r.commandFlags.FlagSet.Visit(func(f *flag.Flag) {
		if f.Name == "succinct" {
			didSetSuccinct = true
		}
	})

	if numSuites > 1 && !didSetSuccinct {
		config.DefaultReporterConfig.Succinct = true
	}
}

type compiler struct {
	runner           *testrunner.TestRunner
	compilationError chan error
}

func (c *compiler) compile() {
	retries := 0

	err := c.runner.Compile()
	for err != nil && retries < 5 { //We retry because Go sometimes steps on itself when multiple compiles happen in parallel.  This is ugly, but should help resolve flakiness...
		err = c.runner.Compile()
		retries++
	}

	c.compilationError <- err
}

func (r *SpecRunner) RunSuites(suites []*testsuite.TestSuite, additionalArgs []string) bool {
	passed := true

	suiteCompilers := make([]*compiler, len(suites))
	for i, suite := range suites {
		runner := testrunner.New(suite, r.commandFlags.NumCPU, r.commandFlags.ParallelStream, r.commandFlags.Race, r.commandFlags.Cover, additionalArgs)
		suiteCompilers[i] = &compiler{
			runner:           runner,
			compilationError: make(chan error, 1),
		}
	}

	compilerChannel := make(chan *compiler)
	numCompilers := runtime.NumCPU()
	for i := 0; i < numCompilers; i++ {
		go func() {
			for compiler := range compilerChannel {
				compiler.compile()
			}
		}()
	}
	go func() {
		for _, compiler := range suiteCompilers {
			compilerChannel <- compiler
		}
		close(compilerChannel)
	}()

	suitesThatFailed := []*testsuite.TestSuite{}
	for i, suite := range suites {
		if r.interruptHandler.WasInterrupted() {
			break
		}

		compilationError := <-suiteCompilers[i].compilationError
		if compilationError != nil {
			fmt.Print(compilationError.Error())
		}
		suitePassed := (compilationError == nil) && suiteCompilers[i].runner.Run()
		r.notifier.SendSuiteCompletionNotification(suite, suitePassed)

		if !suitePassed {
			passed = false
			suitesThatFailed = append(suitesThatFailed, suite)
			if !r.commandFlags.KeepGoing {
				break
			}
		}
		if i < len(suites)-1 && !config.DefaultReporterConfig.Succinct {
			fmt.Println("")
		}
	}

	for i := range suites {
		suiteCompilers[i].runner.CleanUp()
	}

	if r.commandFlags.KeepGoing && !passed {
		fmt.Println("There were failures detected in the following suites:")
		for _, suite := range suitesThatFailed {
			fmt.Printf("\t%s\n", suite.PackageName)
		}
	}

	return passed
}
