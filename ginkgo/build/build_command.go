package build

import (
	"fmt"

	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/ginkgo/command"
	"github.com/onsi/ginkgo/ginkgo/internal"
)

func BuildBuildCommand() command.Command {
	var cliConfig = config.NewDefaultGinkgoCLIConfig()
	var goFlagsConfig = config.NewDefaultGoFlagsConfig()

	flags, err := config.BuildBuildCommandFlagSet(&cliConfig, &goFlagsConfig)
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
			cliConfig, goFlagsConfig, errors = config.VetAndInitializeCLIAndGoConfig(cliConfig, goFlagsConfig)
			command.AbortIfErrors("Ginkgo detected configuraiotn issues:", errors)

			buildSpecs(args, cliConfig, goFlagsConfig)
		},
	}
}

func buildSpecs(args []string, cliConfig config.GinkgoCLIConfigType, goFlagsConfig config.GoFlagsConfigType) {
	suites, _ := internal.FindSuites(args, cliConfig, false)
	if len(suites) == 0 {
		command.AbortWith("Found no test suites")
	}

	passed := true
	for _, suite := range suites {
		fmt.Printf("Compiling %s...\n", suite.PackageName)
		suite = internal.CompileSuite(suite, goFlagsConfig)
		if suite.CompilationError != nil {
			fmt.Println(suite.CompilationError.Error())
			passed = false
		} else {
			fmt.Printf("  compiled %s.test\n", suite.PackageName)
		}
	}

	if !passed {
		command.AbortWith("Failed to compile all tests")
	}
}
