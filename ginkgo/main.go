/*
The Ginkgo CLI

The Ginkgo CLI is fully documented [here](http://onsi.github.io/ginkgo/#the_ginkgo_cli)

To install:

	go install github.com/onsi/ginkgo/ginkgo

To run tests:

	ginkgo

To run tests in all subdirectories:

	ginkgo -r

To run tests in particular packages:

	ginkgo <flags> /path/to/package /path/to/another/package

To run tests in parallel

	ginkgo -nodes=N

where N is the number of nodes.  By default the Ginkgo CLI will spin up a server that the individual
test processes stream test output to.  The CLI then aggregates these streams into one coherent stream of output.
An alternative is to have the parallel nodes run and then present the resulting, final, output in one monolithic chunk - you can opt into this if streaming is giving you trouble:

	ginkgo -nodes=N -stream=false

On windows, the default value for stream is false.

By default, when running multiple tests (with -r or a list of packages) Ginkgo will abort when a test fails.  To have Ginkgo run subsequent test suites instead you can:

	ginkgo -keepGoing

To monitor packages and rerun tests when changes occur:

	ginkgo -watch <-r> </path/to/package>

passing `ginkgo -watch` the `-r` flag will recursively detect all test suites under the current directory and monitor them.
`-watch` does not detect *new* packages. Moreover, changes in package X only rerun the tests for package X, tests for packages
that depend on X are not rerun.

[OSX only] To receive (desktop) notifications when a test run completes:

	ginkgo -notify

this is particularly useful with `ginkgo -watch`.  Notifications are currently only supported on OS X and require that you `brew install terminal-notifier`

Sometimes (to suss out race conditions/flakey tests, for example) you want to keep running a test suite until it fails.  You can do this with:

	ginkgo -untilItFails

To bootstrap a test suite:

	ginkgo bootstrap

To generate a test file:

	ginkgo generate <test_file_name>

To unfocus tests:

	ginkgo unfocus

To print out Ginkgo's version:

	ginkgo version

To get more help:

	ginkgo help
*/
package main

import (
	"flag"
	"fmt"
	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/ginkgo/testsuite"
	"os"
	"os/signal"
	"runtime"
	"time"
)

var numCPU int
var parallelStream bool
var recurse bool
var runMagicI bool
var race bool
var cover bool
var watch bool
var notify bool
var keepGoing bool
var untilItFails bool

func init() {
	onWindows := (runtime.GOOS == "windows")
	onOSX := (runtime.GOOS == "darwin")

	config.Flags("", false)

	flag.IntVar(&(numCPU), "nodes", 1, "The number of parallel test nodes to run")
	flag.BoolVar(&(parallelStream), "stream", !onWindows, "Aggregate parallel test output into one coherent stream (default: true)")
	flag.BoolVar(&(recurse), "r", false, "Find and run test suites under the current directory recursively")
	flag.BoolVar(&(runMagicI), "i", false, "[DEPRECATED] Run go test -i first, then run the test suite")
	flag.BoolVar(&(race), "race", false, "Run tests with race detection enabled")
	flag.BoolVar(&(cover), "cover", false, "Run tests with coverage analysis, will generate coverage profiles with the package name in the current directory")
	flag.BoolVar(&(watch), "watch", false, "Monitor the target packages for changes, then run tests when changes are detected")
	if onOSX {
		flag.BoolVar(&(notify), "notify", false, "Send desktop notifications when a test run completes")
	}
	flag.BoolVar(&(keepGoing), "keepGoing", false, "When true, failures from earlier test suites do not prevent later test suites from running")
	flag.BoolVar(&(untilItFails), "untilItFails", false, "When true, Ginkgo will keep rerunning tests until a failure occurs")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of ginkgo:\n\n")
		fmt.Fprintf(os.Stderr, "ginkgo <FLAGS> <DIRECTORY> ...\n  Run the tests in the passed in <DIRECTORY> (or the current directory if left blank).\n  ginkgo accepts the following flags:\n")
		fmt.Fprintf(os.Stderr, "\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "ginkgo bootstrap\n  Bootstrap a test suite for the current package.\n\n")
		fmt.Fprintf(os.Stderr, "ginkgo generate <SUBJECT>\n  Generate a test file for SUBJECT, the file will be named SUBJECT_test.go\n  If omitted, a file named after the package will be created.\n\n")
		fmt.Fprintf(os.Stderr, "ginkgo unfocus\n  Unfocuses any focused tests.\n\n")
		fmt.Fprintf(os.Stderr, "ginkgo version\n  Print ginkgo's version.\n\n")
		fmt.Fprintf(os.Stderr, "ginkgo help\n  Print this usage information.\n")
	}

	flag.Parse()
}

func main() {
	if runMagicI {
		fmt.Printf("[DEPRECATION]\n  -i is deprecated.\n  Ginkgo now always runs with -i -- the -i flag will go away soon, so stop using it!\n")
	}

	if flag.NArg() > 0 {
		args := flag.Args()
		handled := handleSubcommands(args)
		if handled {
			os.Exit(0)
		}
	}

	if notify {
		verifyNotificationsAreAvailable()
	}

	runner := newTestRunner(numCPU, parallelStream, runMagicI, race, cover)

	registerSignalHandler(runner)

	if watch {
		watchTests(runner)
	} else {
		runTests(runner)
	}
}

func handleSubcommands(args []string) bool {
	switch args[0] {
	case "bootstrap":
		generateBootstrap()
	case "convert":
		convertPackage()
	case "generate":
		subject := ""
		if len(args) > 1 {
			subject = args[1]
		}
		generateSpec(subject)
	case "unfocus", "blur":
		unfocusSpecs()
	case "help":
		flag.Usage()
	case "version":
		fmt.Printf("Ginkgo V%s\n", config.VERSION)
	default:
		return false
	}

	return true
}

func findSuites() []*testsuite.TestSuite {
	suites := []*testsuite.TestSuite{}

	if flag.NArg() > 0 {
		for _, dir := range flag.Args() {
			suites = append(suites, testsuite.SuitesInDir(dir, recurse)...)
		}
	} else {
		suites = testsuite.SuitesInDir(".", recurse)
	}

	if len(suites) == 0 {
		fmt.Printf("Found no test suites.\nFor usage instructions:\n\tginkgo help\n")
		os.Exit(1)
	}

	return suites
}

func runTests(runner *testRunner) {
	t := time.Now()

	passed := true
	if untilItFails {
		iteration := 0
		for {
			passed = runTestSuites(runner)
			iteration++

			if passed {
				fmt.Printf("\nAll tests passed...\nWill keep running them until they fail.\nThis was attempt #%d\n\n", iteration)
			} else {
				fmt.Printf("\nTests failed on attempt #%d\n\n", iteration)
				break
			}
		}
	} else {
		passed = runTestSuites(runner)
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

func runTestSuites(runner *testRunner) bool {
	passed := true

	suites := findSuites()
	suitesThatFailed := []*testsuite.TestSuite{}

	for _, suite := range suites {
		suitePassed := runner.runSuite(suite)
		sendSuiteCompletionNotification(suite, suitePassed)

		if !suitePassed {
			passed = false
			suitesThatFailed = append(suitesThatFailed, suite)
			if !keepGoing {
				break
			}
		}
	}

	if keepGoing && !passed {
		fmt.Println("\nThere were failures detected in the following suites:")
		for _, suite := range suitesThatFailed {
			fmt.Printf("\t%s\n", suite.PackageName)
		}
	}

	return passed
}

func watchTests(runner *testRunner) {
	suites := findSuites()

	modifiedSuite := make(chan *testsuite.TestSuite)
	for _, suite := range suites {
		go suite.Watch(modifiedSuite)
	}

	if !recurse {
		suitePassed := runner.runSuite(suites[0])
		sendSuiteCompletionNotification(suites[0], suitePassed)
	}

	for {
		suite := <-modifiedSuite
		sendNotification("Ginkgo", fmt.Sprintf(`Detected change in "%s"...`, suite.PackageName))

		fmt.Printf("\n\nDetected change in %s\n\n", suite.PackageName)
		suitePassed := runner.runSuite(suite)

		sendSuiteCompletionNotification(suite, suitePassed)
	}
}

func registerSignalHandler(runner *testRunner) {
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, os.Kill)

		select {
		case sig := <-c:
			runner.abort(sig)
			os.Exit(1)
		}
	}()
}
