package main

import (
	"fmt"
	"runtime"

	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/ginkgo/testrunner"
	"github.com/onsi/ginkgo/ginkgo/testsuite"
)

type SuiteRunner struct {
	notifier         *Notifier
	commandFlags     *RunAndWatchCommandFlags
	interruptHandler *InterruptHandler
}

type compiler struct {
	runner           *testrunner.TestRunner
	suite            testsuite.TestSuite
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

func NewSuiteRunner(notifier *Notifier, commandFlags *RunAndWatchCommandFlags, interruptHandler *InterruptHandler) *SuiteRunner {
	return &SuiteRunner{
		notifier:         notifier,
		commandFlags:     commandFlags,
		interruptHandler: interruptHandler,
	}
}

func (r *SuiteRunner) RunSuites(suites []testsuite.TestSuite, additionalArgs []string, keepGoing bool, willCompile func(suite testsuite.TestSuite)) (testrunner.RunResult, int) {
	runResult := testrunner.PassingRunResult()

	suiteCompilers := make([]*compiler, len(suites))
	for i, suite := range suites {
		runner := testrunner.New(suite, r.commandFlags.NumCPU, r.commandFlags.ParallelStream, r.commandFlags.Race, r.commandFlags.Cover, r.commandFlags.Tags, additionalArgs)
		suiteCompilers[i] = &compiler{
			runner:           runner,
			suite:            suite,
			compilationError: make(chan error, 1),
		}
	}

	compilerChannel := make(chan *compiler)
	numCompilers := runtime.NumCPU()
	for i := 0; i < numCompilers; i++ {
		go func() {
			for compiler := range compilerChannel {
				if willCompile != nil {
					willCompile(compiler.suite)
				}
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
	suitesThatFailed := []testsuite.TestSuite{}
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
			if !keepGoing {
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

	if keepGoing && !runResult.Passed {
		r.listFailedSuites(suitesThatFailed)
	}

	return runResult, numSuitesThatRan
}

func (r *SuiteRunner) listFailedSuites(suitesThatFailed []testsuite.TestSuite) {
	fmt.Println("")
	fmt.Println("There were failures detected in the following suites:")

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
