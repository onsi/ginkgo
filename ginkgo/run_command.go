package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/ginkgo/testrunner"
	"github.com/onsi/ginkgo/ginkgo/testsuite"
	"github.com/onsi/ginkgo/types"
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
	r.commandFlags.computeNodes()
	r.notifier.VerifyNotificationsAreAvailable()

	suites := findSuites(args, r.commandFlags.Recurse, r.commandFlags.SkipPackage)
	r.ComputeSuccinctMode(len(suites))

	t := time.Now()

	numSuites := 0
	runResult := testrunner.PassingRunResult()
	if r.commandFlags.UntilItFails {
		iteration := 0
		for {
			r.UpdateSeed()
			runResult, numSuites = r.RunSuites(suites, additionalArgs)
			iteration++

			if r.interruptHandler.WasInterrupted() {
				break
			}

			if runResult.Passed {
				fmt.Printf("\nAll tests passed...\nWill keep running them until they fail.\nThis was attempt #%d\n%s\n", iteration, orcMessage(iteration))
			} else {
				fmt.Printf("\nTests failed on attempt #%d\n\n", iteration)
				break
			}
		}
	} else {
		runResult, numSuites = r.RunSuites(suites, additionalArgs)
	}

	noun := "suites"
	if numSuites == 1 {
		noun = "suite"
	}

	fmt.Printf("\nGinkgo ran %d %s in %s\n", numSuites, noun, time.Since(t))

	if runResult.Passed {
		if runResult.HasProgrammaticFocus {
			fmt.Printf("Test Suite Passed\n")
			fmt.Printf("Detected Programmatic Focus - setting exit status to %d\n", types.GINKGO_FOCUS_EXIT_CODE)
			os.Exit(types.GINKGO_FOCUS_EXIT_CODE)
		} else {
			fmt.Printf("Test Suite Passed\n")
			os.Exit(0)
		}
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

	if numSuites > 1 && !r.commandFlags.wasSet("succinct") {
		config.DefaultReporterConfig.Succinct = true
	}
}

func (r *SpecRunner) UpdateSeed() {
	if !r.commandFlags.wasSet("seed") {
		config.GinkgoConfig.RandomSeed = time.Now().Unix()
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

func (r *SpecRunner) RunSuites(suites []*testsuite.TestSuite, additionalArgs []string) (testrunner.RunResult, int) {
	runResult := testrunner.PassingRunResult()

	suiteCompilers := make([]*compiler, len(suites))
	for i, suite := range suites {
		runner := testrunner.New(suite, r.commandFlags.NumCPU, r.commandFlags.ParallelStream, r.commandFlags.Race, r.commandFlags.Cover, r.commandFlags.Tags, additionalArgs)
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

	numSuitesThatRan := 0
	suitesThatFailed := []*testsuite.TestSuite{}
	for i, suite := range suites {
		if r.interruptHandler.WasInterrupted() {
			break
		}

		compilationError := <-suiteCompilers[i].compilationError
		if compilationError != nil {
			fmt.Print(compilationError.Error())
		}
		numSuitesThatRan++
		suiteRunResult := testrunner.FailingRunResult()
		if compilationError == nil {
			suiteRunResult = suiteCompilers[i].runner.Run()
		}
		r.notifier.SendSuiteCompletionNotification(suite, suiteRunResult.Passed)
		runResult = runResult.Merge(suiteRunResult)
		if !suiteRunResult.Passed {
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

	if r.commandFlags.KeepGoing && !runResult.Passed {
		r.listFailedSuites(suitesThatFailed)
	}

	return runResult, numSuitesThatRan
}

func (r *SpecRunner) listFailedSuites(suitesThatFailed []*testsuite.TestSuite) {
	fmt.Println("")
	fmt.Println("There were failures detected in the following suites:")

	redColor := "\x1b[91m"
	defaultStyle := "\x1b[0m"
	lightGrayColor := "\x1b[37m"

	maxPackageNameLength := 0
	for _, suite := range suitesThatFailed {
		if len(suite.PackageName) > maxPackageNameLength {
			maxPackageNameLength = len(suite.PackageName)
		}
	}

	packageNameFormatter := fmt.Sprintf("%%%ds", maxPackageNameLength)

	for _, suite := range suitesThatFailed {
		if config.DefaultReporterConfig.NoColor {
			fmt.Printf("\t"+packageNameFormatter+" %s\n", suite.PackageName, suite.Path)
		} else {
			fmt.Printf("\t%s"+packageNameFormatter+"%s %s%s%s\n", redColor, suite.PackageName, defaultStyle, lightGrayColor, suite.Path, defaultStyle)
		}
	}
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
			"Here we go again...",
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
		}[iteration-10]
	} else {
		return "No, seriously... you can probably stop now."
	}
}
