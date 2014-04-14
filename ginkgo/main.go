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

	ginkgo -nodes=N

where N is the number of nodes.  By default the Ginkgo CLI will spin up a server that the individual
test processes send test output to.  The CLI aggregates this output and then presents coherent test output, one test at a time, as each test completes.
An alternative is to have the parallel nodes run and stream interleaved output back.  This useful for debugging, particularly in contexts where tests hang/fail to start.  To get this interleaved output:

	ginkgo -nodes=N -stream=true

On windows, the default value for stream is true.

By default, when running multiple tests (with -r or a list of packages) Ginkgo will abort when a test fails.  To have Ginkgo run subsequent test suites instead you can:

	ginkgo -keepGoing

To monitor packages and rerun tests when changes occur:

	ginkgo watch <-r> </path/to/package>

passing `ginkgo watch` the `-r` flag will recursively detect all test suites under the current directory and monitor them.
`watch` does not detect *new* packages. Moreover, changes in package X only rerun the tests for package X, tests for packages
that depend on X are not rerun.

[OSX only] To receive (desktop) notifications when a test run completes:

	ginkgo -notify

this is particularly useful with `ginkgo watch`.  Notifications are currently only supported on OS X and require that you `brew install terminal-notifier`

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

To convert an existing XUnit style test suite to a Ginkgo-style test suite:

	ginkgo convert .

To unfocus tests:

	ginkgo unfocus

or

	ginkgo blur

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
	"os/exec"
	"regexp"
	"strings"
)

type Command struct {
	Name                      string
	AltName                   string
	FlagSet                   *flag.FlagSet
	Usage                     []string
	UsageCommand              string
	Command                   func(args []string, additionalArgs []string)
	SuppressFlagDocumentation bool
	FlagDocSubstitute         []string
}

func (c *Command) Matches(name string) bool {
	return c.Name == name || (c.AltName != "" && c.AltName == name)
}

func (c *Command) Run(args []string, additionalArgs []string) {
	c.FlagSet.Parse(args)
	c.Command(c.FlagSet.Args(), additionalArgs)
}

var DefaultCommand *Command
var Commands []*Command

func init() {
	DefaultCommand = BuildRunCommand()
	Commands = append(Commands, BuildWatchCommand())
	Commands = append(Commands, BuildBootstrapCommand())
	Commands = append(Commands, BuildGenerateCommand())
	Commands = append(Commands, BuildNodotCommand())
	Commands = append(Commands, BuildConvertCommand())
	Commands = append(Commands, BuildUnfocusCommand())
	Commands = append(Commands, BuildVersionCommand())
	Commands = append(Commands, BuildHelpCommand())
}

func main() {
	args := []string{}
	additionalArgs := []string{}

	foundDelimiter := false

	for _, arg := range os.Args[1:] {
		if !foundDelimiter {
			if arg == "--" {
				foundDelimiter = true
				continue
			}
		}

		if foundDelimiter {
			additionalArgs = append(additionalArgs, arg)
		} else {
			args = append(args, arg)
		}
	}

	if len(args) > 0 {
		commandToRun, found := commandMatching(args[0])
		if found {
			commandToRun.Run(args[1:], additionalArgs)
			return
		}
	}

	DefaultCommand.Run(args, additionalArgs)
}

func commandMatching(name string) (*Command, bool) {
	for _, command := range Commands {
		if command.Matches(name) {
			return command, true
		}
	}
	return nil, false
}

func usage() {
	fmt.Fprintf(os.Stderr, "Ginkgo Version %s\n\n", config.VERSION)
	usageForCommand(DefaultCommand, false)
	for _, command := range Commands {
		fmt.Fprintf(os.Stderr, "\n")
		usageForCommand(command, false)
	}
}

func usageForCommand(command *Command, longForm bool) {
	fmt.Fprintf(os.Stderr, "%s\n%s\n", command.UsageCommand, strings.Repeat("-", len(command.UsageCommand)))
	fmt.Fprintf(os.Stderr, "%s\n", strings.Join(command.Usage, "\n"))
	if command.SuppressFlagDocumentation && !longForm {
		fmt.Fprintf(os.Stderr, "%s\n", strings.Join(command.FlagDocSubstitute, "\n  "))
	} else {
		command.FlagSet.PrintDefaults()
	}
}

func complainAndQuit(complaint string) {
	fmt.Fprintf(os.Stderr, "%s\nFor usage instructions:\n\tginkgo help\n", complaint)
	os.Exit(1)
}

func findSuites(args []string, recurse bool, skipPackage string) []*testsuite.TestSuite {
	suites := []*testsuite.TestSuite{}

	if len(args) > 0 {
		for _, arg := range args {
			suites = append(suites, testsuite.SuitesInDir(arg, recurse)...)
		}
	} else {
		suites = testsuite.SuitesInDir(".", recurse)
	}

	if skipPackage != "" {
		re := regexp.MustCompile(skipPackage)
		filteredSuites := []*testsuite.TestSuite{}
		skippedPackages := []string{}
		for _, suite := range suites {
			if re.Match([]byte(suite.PackageName)) {
				skippedPackages = append(skippedPackages, suite.PackageName)
			} else {
				filteredSuites = append(filteredSuites, suite)
			}
		}
		if len(skippedPackages) > 0 {
			fmt.Printf("Will skip %s\n", strings.Join(skippedPackages, ", "))
		}
		suites = filteredSuites
	}

	if len(suites) == 0 {
		complainAndQuit("Found no test suites")
	}

	return suites
}

func goFmt(path string) {
	err := exec.Command("go", "fmt", path).Run()
	if err != nil {
		complainAndQuit("Could not fmt: " + err.Error())
	}
}
