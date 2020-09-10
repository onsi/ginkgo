/*
The Ginkgo CLI

The Ginkgo CLI is fully documented [here](http://onsi.github.io/ginkgo/#the_ginkgo_cli)

You can also learn more by running:

	ginkgo help

Here are some of the more commonly used commands:

To install:

	go install github.com/onsi/ginkgo/ginkgo

To run tests:

	ginkgo

To run tests in all subdirectories:

	ginkgo -r

To run tests in particular packages:

	ginkgo <flags> /path/to/package /path/to/another/package

To pass arguments/flags to your tests:

	ginkgo <flags> <packages> -- <pass-throughs>

To run tests in parallel

	ginkgo -p

this will automatically detect the optimal number of nodes to use.  Alternatively, you can specify the number of nodes with:

	ginkgo -nodes=N

(note that you don't need to provide -p in this case).

By default the Ginkgo CLI will spin up a server that the individual test processes send test output to.  The CLI aggregates this output and then presents coherent test output, one test at a time, as each test completes.
An alternative is to have the parallel nodes run and stream interleaved output back.  This useful for debugging, particularly in contexts where tests hang/fail to start.  To get this interleaved output:

	ginkgo -nodes=N -stream=true

On windows, the default value for stream is true.

By default, when running multiple tests (with -r or a list of packages) Ginkgo will abort when a test fails.  To have Ginkgo run subsequent test suites instead you can:

	ginkgo -keepGoing

To fail if there are ginkgo tests in a directory but no test suite (missing `RunSpecs`)

	ginkgo -requireSuite

To monitor packages and rerun tests when changes occur:

	ginkgo watch <-r> </path/to/package>

passing `ginkgo watch` the `-r` flag will recursively detect all test suites under the current directory and monitor them.
`watch` does not detect *new* packages. Moreover, changes in package X only rerun the tests for package X, tests for packages
that depend on X are not rerun.

Sometimes (to suss out race conditions/flakey tests, for example) you want to keep running a test suite until it fails.  You can do this with:

	ginkgo -untilItFails

To bootstrap a test suite:

	ginkgo bootstrap

To generate a test file:

	ginkgo generate <test_file_name>

To bootstrap/generate test files without using "." imports:

	ginkgo bootstrap --nodot
	ginkgo generate --nodot

this will explicitly export all the identifiers in Ginkgo and Gomega allowing you to rename them to avoid collisions.  When you pull to the latest Ginkgo/Gomega you'll want to run

	ginkgo nodot

to refresh this list and pull in any new identifiers.  In particular, this will pull in any new Gomega matchers that get added.

To unfocus tests:

	ginkgo unfocus

or

	ginkgo blur

To compile a test suite:

	ginkgo build <path-to-package>

will output an executable file named `package.test`.  This can be run directly or by invoking

	ginkgo <path-to-package.test>


To print an outline of Ginkgo specs and containers in a file:

	gingko outline <filename>

To print out Ginkgo's version:

	ginkgo version

To get more help:

	ginkgo help
*/
package main

import (
	"fmt"
	"os"

	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/ginkgo/build"
	"github.com/onsi/ginkgo/ginkgo/command"
	"github.com/onsi/ginkgo/ginkgo/generators"
	"github.com/onsi/ginkgo/ginkgo/nodot"
	"github.com/onsi/ginkgo/ginkgo/outline"
	"github.com/onsi/ginkgo/ginkgo/run"
	"github.com/onsi/ginkgo/ginkgo/unfocus"
	"github.com/onsi/ginkgo/ginkgo/watch"
	"github.com/onsi/ginkgo/types"
)

var program command.Program

func GenerateCommands() []command.Command {
	return []command.Command{
		watch.BuildWatchCommand(),
		build.BuildBuildCommand(),
		generators.BuildBootstrapCommand(),
		generators.BuildGenerateCommand(),
		nodot.BuildNodotCommand(),
		outline.BuildOutlineCommand(),
		unfocus.BuildUnfocusCommand(),
		BuildVersionCommand(),
	}
}

func main() {
	program = command.Program{
		Name:           "ginkgo",
		Heading:        fmt.Sprintf("Ginkgo Version %s", config.VERSION),
		Commands:       GenerateCommands(),
		DefaultCommand: run.BuildRunCommand(),
		DeprecatedCommands: []command.DeprecatedCommand{
			{Name: "convert", Deprecation: types.Deprecations.Convert()},
		},
	}

	program.RunAndExit(os.Args)
}

func BuildVersionCommand() command.Command {
	return command.Command{
		Name:     "version",
		Usage:    "ginkgo version",
		ShortDoc: "Print Ginkgo's version",
		Command: func(_ []string, _ []string) {
			fmt.Printf("Ginkgo Version %s\n", config.VERSION)
		},
	}
}
