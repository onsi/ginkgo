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

To monitor packages and rerun tests when changes occur:

	ginkgo -watch <-r> </path/to/package>

passing `ginkgo -watch` the `-r` flag will recursively detect all test suites under the current directory and monitor them.
`-watch` does not detect *new* packages. Moreover, changes in package X only rerun the tests for package X, tests for packages
that depend on X are not rerun.

To run tests in parallel

	ginkgo -nodes=N

where N is the number of nodes.  By default the Ginkgo CLI will spin up a server that the individual
test processes stream test output to.  The CLI then aggregates these streams into one coherent stream of output.
An alternative is to have the parallel nodes run and then present the resulting, final, output in one monolithic chunk - you can opt into this if streaming is giving you trouble:

	ginkgo -nodes=N -stream=false

On windows, the default value for stream is false.

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

func init() {
	onWindows := (runtime.GOOS == "windows")

	config.Flags("", false)

	flag.IntVar(&(numCPU), "nodes", 1, "The number of parallel test nodes to run")
	flag.BoolVar(&(parallelStream), "stream", !onWindows, "Aggregate parallel test output into one coherent stream (default: true)")
	flag.BoolVar(&(recurse), "r", false, "Find and run test suites under the current directory recursively")
	flag.BoolVar(&(runMagicI), "i", false, "Run go test -i first, then run the test suite")
	flag.BoolVar(&(race), "race", false, "Run tests with race detection enabled")
	flag.BoolVar(&(cover), "cover", false, "Run tests with coverage analysis, will generate coverage profiles with the package name in the current directory")
	flag.BoolVar(&(watch), "watch", false, "Monitor the target packages for changes, then run tests when changes are detected")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of ginkgo:\n\n")
		fmt.Fprintf(os.Stderr, "ginkgo <FLAGS> <DIRECTORY> ...\n  Run the tests in the passed in <DIRECTORY> (or the current directory if left blank).\n  ginkgo accepts the following flags:\n")
		flag.PrintDefaults()
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
	if flag.NArg() > 0 {
		args := flag.Args()
		handled := handleSubcommands(args)
		if handled {
			os.Exit(0)
		}
	}

	if watch {
		watchTests()
	} else {
		runTests()
	}
}

func handleSubcommands(args []string) bool {
	switch args[0] {
	case "bootstrap":
		generateBootstrap()
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

func runTests() {
	t := time.Now()

	suites := findSuites()

	runner := newTestRunner(numCPU, parallelStream, runMagicI, race, cover)
	passed := runner.run(suites)
	fmt.Printf("\nGinkgo ran in %s\n", time.Since(t))

	if passed {
		fmt.Printf("Test Suite Passed\n")
		os.Exit(0)
	} else {
		fmt.Printf("Test Suite Failed\n")
		os.Exit(1)
	}
}

func watchTests() {
	suites := findSuites()

	runner := newTestRunner(numCPU, parallelStream, runMagicI, race, cover)
	modifiedSuite := make(chan *testsuite.TestSuite)
	for _, suite := range suites {
		go suite.Watch(modifiedSuite)
	}

	if !recurse {
		runner.runSuite(suites[0])
	}

	for {
		suite := <-modifiedSuite
		fmt.Printf("\n\nDetected change in %s\n\n", suite.PackageName)
		runner.runSuite(suite)
	}
}
