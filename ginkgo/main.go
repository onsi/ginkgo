/*
The Ginkgo CLI

The Ginkgo CLI is fully documented [here](http://onsi.github.io/ginkgo/#the_ginkgo_cli)

To install:

	go install github.com/onsi/ginkgo/ginkgo

To run tests:

	ginkgo

To run tests in all subdirectories:

	ginkgo -r

To bootstrap a test suite:

	ginkgo bootstrap

To generate a test file:

	ginkgo generate <test_file_name>

To unfocus tests:

	ginkgo unfocs

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
		fmt.Fprintf(os.Stderr, "ginkgo\n  Run the tests in the current directory.  The following flags are available:\n")
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
		handleSubcommands(args)
		fmt.Printf("Unkown command %s\n\n", args[0])
		flag.Usage()

		os.Exit(1)
	}

	t := time.Now()
	runner := newTestRunner(numCPU, recurse, runMagicI, race, cover)
	passed := runner.run()
	fmt.Printf("\nGinkgo ran in %s\n", time.Since(t))

	if passed {
		fmt.Printf("Test Suite Passed\n")
		os.Exit(0)
	} else {
		fmt.Printf("Test Suite Failed\n")
		os.Exit(1)
	}
}

func handleSubcommands(args []string) {
	if args[0] == "bootstrap" {
		generateBootstrap()
		os.Exit(0)
	} else if args[0] == "generate" {
		subject := ""
		if len(args) > 1 {
			subject = args[1]
		}
		generateSpec(subject)
		os.Exit(0)
	} else if args[0] == "unfocus" {
		unfocusSpecs()
		os.Exit(0)
	} else if args[0] == "help" {
		flag.Usage()
		os.Exit(0)
	} else if args[0] == "version" {
		fmt.Printf("Ginkgo V%s\n", config.VERSION)
		os.Exit(0)
	}
}
