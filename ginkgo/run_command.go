package main

import (
	"flag"
	"fmt"
	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/ginkgo/testrunner"
	"github.com/onsi/ginkgo/ginkgo/testsuite"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"
)

var numCPU int
var parallelStream bool
var recurse bool
var race bool
var cover bool
var watch bool
var notify bool
var keepGoing bool
var untilItFails bool

var activeRunners []*testrunner.TestRunner
var runnerLock *sync.Mutex

var runFlagSet *flag.FlagSet

func BuildRunCommand() *Command {
	runFlagSet = flag.NewFlagSet("ginkgo", flag.ExitOnError)

	onWindows := (runtime.GOOS == "windows")
	onOSX := (runtime.GOOS == "darwin")

	config.Flags(runFlagSet, "", false)

	runFlagSet.IntVar(&(numCPU), "nodes", 1, "The number of parallel test nodes to run")
	runFlagSet.BoolVar(&(parallelStream), "stream", onWindows, "stream parallel test output in real time: less coherent, but useful for debugging")
	runFlagSet.BoolVar(&(recurse), "r", false, "Find and run test suites under the current directory recursively")
	runFlagSet.BoolVar(&(race), "race", false, "Run tests with race detection enabled")
	runFlagSet.BoolVar(&(cover), "cover", false, "Run tests with coverage analysis, will generate coverage profiles with the package name in the current directory")
	runFlagSet.BoolVar(&(watch), "watch", false, "Monitor the target packages for changes, then run tests when changes are detected")
	if onOSX {
		runFlagSet.BoolVar(&(notify), "notify", false, "Send desktop notifications when a test run completes")
	}
	runFlagSet.BoolVar(&(keepGoing), "keepGoing", false, "When true, failures from earlier test suites do not prevent later test suites from running")
	runFlagSet.BoolVar(&(untilItFails), "untilItFails", false, "When true, Ginkgo will keep rerunning tests until a failure occurs")

	runnerLock = &sync.Mutex{}

	return &Command{
		Name:         "",
		FlagSet:      runFlagSet,
		UsageCommand: "ginkgo <FLAGS> <PACKAGES>...",
		Usage: []string{
			"Run the tests in the passed in <PACKAGES> (or the package in the current directory if left blank).",
			"Accepts the following flags:",
		},
		Command: runSpecs,
	}
}

func runSpecs(args []string) {
	if notify {
		verifyNotificationsAreAvailable()
	}

	registerSignalHandler()

	suites := findSuites(args)
	computeSuccinctMode(len(suites))
	if watch {
		watchTests(suites)
	} else {
		runTests(suites)
	}
}

func findSuites(args []string) []*testsuite.TestSuite {
	suites := []*testsuite.TestSuite{}

	if len(args) > 0 {
		for _, dir := range args {
			suites = append(suites, testsuite.SuitesInDir(dir, recurse)...)
		}
	} else {
		suites = testsuite.SuitesInDir(".", recurse)
	}

	if len(suites) == 0 {
		complainAndQuit("Found no test suites")
	}

	return suites
}

func computeSuccinctMode(numSuites int) {
	if config.DefaultReporterConfig.Verbose {
		config.DefaultReporterConfig.Succinct = false
		return
	}

	if numSuites == 1 {
		return
	}

	didSetSuccinct := false
	runFlagSet.Visit(func(f *flag.Flag) {
		if f.Name == "succinct" {
			didSetSuccinct = true
		}
	})

	if numSuites > 1 && !didSetSuccinct {
		config.DefaultReporterConfig.Succinct = true
	}
}

func runTests(suites []*testsuite.TestSuite) {
	t := time.Now()

	passed := true
	if untilItFails {
		iteration := 0
		for {
			passed = runTestSuites(suites)
			iteration++

			if passed {
				fmt.Printf("\nAll tests passed...\nWill keep running them until they fail.\nThis was attempt #%d\n\n", iteration)
			} else {
				fmt.Printf("\nTests failed on attempt #%d\n\n", iteration)
				break
			}
		}
	} else {
		passed = runTestSuites(suites)
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

func runTestSuites(suites []*testsuite.TestSuite) bool {
	passed := true

	suiteCompilers := make([]*compiler, len(suites))
	for i, suite := range suites {
		runner := makeAndRegisterTestRunner(suite)
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
		compilationError := <-suiteCompilers[i].compilationError
		if compilationError != nil {
			fmt.Print(compilationError.Error())
		}
		suitePassed := (compilationError == nil) && suiteCompilers[i].runner.Run()
		sendSuiteCompletionNotification(suite, suitePassed)

		if !suitePassed {
			passed = false
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
		unregisterRunner(suiteCompilers[i].runner)
	}

	if keepGoing && !passed {
		fmt.Println("There were failures detected in the following suites:")
		for _, suite := range suitesThatFailed {
			fmt.Printf("\t%s\n", suite.PackageName)
		}
	}

	return passed
}

func watchTests(suites []*testsuite.TestSuite) {
	modifiedSuite := make(chan *testsuite.TestSuite)
	for _, suite := range suites {
		go suite.Watch(modifiedSuite)
	}

	if len(suites) == 1 {
		runTestForSuite(suites[0])
	}

	for {
		suite := <-modifiedSuite
		sendNotification("Ginkgo", fmt.Sprintf(`Detected change in "%s"...`, suite.PackageName))

		fmt.Printf("\n\nDetected change in %s\n\n", suite.PackageName)
		runTestForSuite(suite)
	}
}

func runTestForSuite(suite *testsuite.TestSuite) {
	runner := makeAndRegisterTestRunner(suite)
	err := runner.Compile()
	if err != nil {
		fmt.Print(err.Error())
	}
	suitePassed := (err == nil) && runner.Run()
	sendSuiteCompletionNotification(suite, suitePassed)
	runner.CleanUp()
	unregisterRunner(runner)
}

func makeAndRegisterTestRunner(suite *testsuite.TestSuite) *testrunner.TestRunner {
	runnerLock.Lock()
	defer runnerLock.Unlock()

	runner := testrunner.New(suite, numCPU, parallelStream, race, cover)
	activeRunners = append(activeRunners, runner)
	return runner
}

func unregisterRunner(runner *testrunner.TestRunner) {
	runnerLock.Lock()
	defer runnerLock.Unlock()

	for i, registeredRunner := range activeRunners {
		if registeredRunner == runner {
			activeRunners[i] = activeRunners[len(activeRunners)-1]
			activeRunners = activeRunners[0 : len(activeRunners)-1]
			break
		}
	}
}

func registerSignalHandler() {
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)

		select {
		case sig := <-c:
			runnerLock.Lock()
			for _, runner := range activeRunners {
				runner.CleanUp(sig)
			}
			runnerLock.Unlock()
			os.Exit(1)
		}
	}()
}
