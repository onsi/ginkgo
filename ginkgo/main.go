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
	"regexp"
	"strings"
)

type Command struct {
	Name                      string
	AltName                   string
	FlagSet                   *flag.FlagSet
	Usage                     []string
	UsageCommand              string
	Command                   func(args []string)
	SuppressFlagDocumentation bool
	FlagDocSubstitute         []string
}

func (c *Command) Matches(name string) bool {
	return c.Name == name || (c.AltName != "" && c.AltName == name)
}

func (c *Command) Run(args []string) {
	c.FlagSet.Parse(args)
	c.Command(c.FlagSet.Args())
}

var DefaultCommand *Command
var Commands []*Command

func init() {
	DefaultCommand = BuildRunCommand()
	Commands = append(Commands, BuildWatchCommand())
	Commands = append(Commands, BuildBootstrapCommand())
	Commands = append(Commands, BuildGenerateCommand())
	Commands = append(Commands, BuildConvertCommand())
	Commands = append(Commands, BuildUnfocusCommand())
	Commands = append(Commands, BuildVersionCommand())
	Commands = append(Commands, BuildHelpCommand())
}

func main() {
	if len(os.Args) > 1 {
		commandToRun, found := commandMatching(os.Args[1])
		if found {
			commandToRun.Run(os.Args[2:])
			return
		}
	}

	DefaultCommand.Run(os.Args[1:])
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
	fmt.Fprintf(os.Stderr, "%s\n", command.UsageCommand)
	fmt.Fprintf(os.Stderr, "  %s\n", strings.Join(command.Usage, "\n  "))
	if command.SuppressFlagDocumentation && !longForm {
		fmt.Fprintf(os.Stderr, "  %s\n", strings.Join(command.FlagDocSubstitute, "\n  "))
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
		for _, dir := range args {
			suites = append(suites, testsuite.SuitesInDir(dir, recurse)...)
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
