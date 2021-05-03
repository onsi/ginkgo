package build

import (
	"fmt"

	"github.com/onsi/ginkgo/ginkgo/command"
	"github.com/onsi/ginkgo/ginkgo/internal"
	"github.com/onsi/ginkgo/types"
)

func BuildBuildCommand() command.Command {
	var cliConfig = types.NewDefaultCLIConfig()
	var goFlagsConfig = types.NewDefaultGoFlagsConfig()

	flags, err := types.BuildBuildCommandFlagSet(&cliConfig, &goFlagsConfig)
	if err != nil {
		panic(err)
	}

	return command.Command{
		Name:     "build",
		Flags:    flags,
		Usage:    "ginkgo build <FLAGS> <PACKAGES>",
		ShortDoc: "Build the passed in <PACKAGES> (or the package in the current directory if left blank).",
		DocLink:  "precompiling-tests",
		Command: func(args []string, _ []string) {
			var errors []error
			cliConfig, goFlagsConfig, errors = types.VetAndInitializeCLIAndGoConfig(cliConfig, goFlagsConfig)
			command.AbortIfErrors("Ginkgo detected configuration issues:", errors)

			buildSpecs(args, cliConfig, goFlagsConfig)
		},
	}
}

func buildSpecs(args []string, cliConfig types.CLIConfig, goFlagsConfig types.GoFlagsConfig) {
	suites := internal.FindSuites(args, cliConfig, false).WithoutState(internal.TestSuiteStateSkippedByFilter)
	if len(suites) == 0 {
		command.AbortWith("Found no test suites")
	}

	for idx := range suites {
		fmt.Printf("Compiling %s...\n", suites[idx].PackageName)
		suites[idx] = internal.CompileSuite(suites[idx], goFlagsConfig)
		if suites[idx].State.Is(internal.TestSuiteStateFailedToCompile) {
			fmt.Println(suites[idx].CompilationError.Error())
		} else {
			fmt.Printf("  compiled %s.test\n", suites[idx].PackageName)
		}
	}

	if suites.CountWithState(internal.TestSuiteStateFailedToCompile) > 0 {
		command.AbortWith("Failed to compile all tests")
	}
}
