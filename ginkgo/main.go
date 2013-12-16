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
	"os"
	"time"
)

var numCPU int
var recurse bool
var runMagicI bool
var race bool
var cover bool

func init() {
	config.Flags("", false)

	flag.IntVar(&(numCPU), "nodes", 1, "The number of parallel test nodes to run")
	flag.BoolVar(&(recurse), "r", false, "Find and run test suites under the current directory recursively")
	flag.BoolVar(&(runMagicI), "i", false, "Run go test -i first, then run the test suite")
	flag.BoolVar(&(race), "race", false, "Run tests with race detection enabled")
	flag.BoolVar(&(cover), "cover", false, "Run tests with coverage analysis, will generate coverage profiles with the package name in the current directory")

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

	runTests()
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

func runTests() {
	t := time.Now()

	suites := []testSuite{}

	if flag.NArg() > 0 {
		for _, dir := range flag.Args() {
			suites = append(suites, suitesInDir(dir, recurse)...)
		}
	} else {
		suites = suitesInDir(".", recurse)
	}

	if len(suites) == 0 {
		fmt.Printf("Found no test suites.\nFor usage instructions:\n\tginkgo help\n")
		os.Exit(1)
	}

	runner := newTestRunner(numCPU, runMagicI, race, cover)
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
