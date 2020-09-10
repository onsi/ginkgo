package config

import (
	"os"
	"runtime"
	"time"
)

type GinkgoCLIConfigType struct {
	//for build, run, and watch
	Recurse      bool
	SkipPackage  string
	RequireSuite bool
	NumCompilers int

	//for run and watch only
	Nodes                     int
	Parallel                  bool
	AfterRunHook              string
	Timeout                   time.Duration
	OutputDir                 string
	KeepSeparateCoverprofiles bool

	//for run only
	KeepGoing       bool
	UntilItFails    bool
	RandomizeSuites bool

	//for watch only
	Depth       int
	WatchRegExp string
}

func (g GinkgoCLIConfigType) ComputedNodes() int {
	if g.Nodes > 0 {
		return g.Nodes
	}

	n := 1
	if g.Parallel {
		n = runtime.NumCPU()
		if n > 4 {
			n = n - 1
		}
	}
	return n
}

func NewDefaultGinkgoCLIConfig() GinkgoCLIConfigType {
	return GinkgoCLIConfigType{
		Timeout:     time.Hour * 24,
		Depth:       1,
		WatchRegExp: `\.go$`,
	}
}

type deprecatedGinkgoCLIConfig struct {
	Stream bool
	Notify bool
}

var GinkgoCLISharedFlags = GinkgoFlags{
	{KeyPath: "C.Recurse", Name: "r", SectionKey: "multiple-suites",
		Usage: "If set, ginkgo finds and runs test suites under the current directory recursively."},
	{KeyPath: "C.SkipPackage", Name: "skip-package", SectionKey: "multiple-suites", DeprecatedName: "skipPackage", DeprecatedDocLink: "changed-command-line-flags",
		UsageArgument: "comma-separated list of packages",
		Usage:         "A comma-separated list of package names to be skipped.  If any part of the package's path matches, that package is ignored."},
	{KeyPath: "C.RequireSuite", Name: "require-suite", SectionKey: "failure", DeprecatedName: "requireSuite", DeprecatedDocLink: "changed-command-line-flags",
		Usage: "If set, Ginkgo fails if there are ginkgo tests in a directory but no invocation of RunSpecs."},
	{KeyPath: "C.NumCompilers", Name: "compilers", SectionKey: "multiple-suites", UsageDefaultValue: "0 (will autodetect)",
		Usage: "When running multiple packages, the number of concurrent compilations to perform."},
}

var GinkgoCLIRunAndWatchFlags = GinkgoFlags{
	{KeyPath: "C.Nodes", Name: "nodes", SectionKey: "parallel", UsageDefaultValue: "1 (run in series)",
		Usage: "The number of parallel test nodes to run."},
	{KeyPath: "C.Parallel", Name: "p", SectionKey: "parallel",
		Usage: "If set, ginkgo will run in parallel with an auto-detected number of nodes."},
	{KeyPath: "C.AfterRunHook", Name: "after-run-hook", SectionKey: "misc", DeprecatedName: "afterSuiteHook", DeprecatedDocLink: "changed-command-line-flags",
		Usage: "Command to run when a test suite completes."},
	{KeyPath: "C.Timeout", Name: "timeout", SectionKey: "debug", UsageDefaultValue: "24h",
		Usage: "Test suite fails if it does not complete within the specified timeout."},
	{KeyPath: "C.OutputDir", Name: "output-dir", SectionKey: "output", UsageArgument: "directory", DeprecatedName: "outputdir", DeprecatedDocLink: "changed-profiling-support",
		Usage: "A location to place all generated profiles and reports."},
	{KeyPath: "C.KeepSeparateCoverprofiles", Name: "keep-separate-coverprofiles", SectionKey: "code-and-coverage-analysis",
		Usage: "If set, Ginkgo does not merge coverprofiles into one monolithic coverprofile.  The coverprofiles will remain in their respective package direcotries or in -output-dir if set."},

	{KeyPath: "Dcli.Stream", DeprecatedName: "stream", DeprecatedDocLink: "removed--stream"},
	{KeyPath: "Dcli.Notify", DeprecatedName: "notify", DeprecatedDocLink: "removed--notify"},
}

var GinkgoCLIRunFlags = GinkgoFlags{
	{KeyPath: "C.KeepGoing", Name: "keep-going", SectionKey: "multiple-suites", DeprecatedName: "keepGoing", DeprecatedDocLink: "changed-command-line-flags",
		Usage: "If set, failures from earlier test suites do not prevent later test suites from running."},
	{KeyPath: "C.UntilItFails", Name: "until-it-fails", SectionKey: "debug", DeprecatedName: "untilItFails", DeprecatedDocLink: "changed-command-line-flags",
		Usage: "If set, ginkgo will keep rerunning test suites until a failure occurs."},
	{KeyPath: "C.RandomizeSuites", Name: "randomize-suites", SectionKey: "order", DeprecatedName: "randomizeSuites", DeprecatedDocLink: "changed-command-line-flags",
		Usage: "If set, ginkgo will randomize the order in which test suites run."},
}

var GinkgoCLIWatchFlags = GinkgoFlags{
	{KeyPath: "C.Depth", Name: "depth", SectionKey: "watch",
		Usage: "Ginkgo will watch dependencies down to this depth in the dependency tree."},
	{KeyPath: "C.WatchRegExp", Name: "watch-regexp", SectionKey: "watch", DeprecatedName: "watchRegExp", DeprecatedDocLink: "changed-command-line-flags",
		UsageArgument:     "Regular Expression",
		UsageDefaultValue: `\.go$`,
		Usage:             "Only files matching this regular expression will be watched for changes."},
}

func VetAndInitializeCLIAndGoConfig(cliConfig GinkgoCLIConfigType, goFlagsConfig GoFlagsConfigType) (GinkgoCLIConfigType, GoFlagsConfigType, []error) {
	errors := []error{}

	//initialize the output directory
	if cliConfig.OutputDir != "" {
		err := os.MkdirAll(cliConfig.OutputDir, 0777)
		if err != nil {
			errors = append(errors, err)
		}
	}

	//ensure cover mode is configured appropriately
	if goFlagsConfig.CoverMode != "" || goFlagsConfig.CoverPkg != "" || goFlagsConfig.CoverProfile != "" {
		goFlagsConfig.Cover = true
	}
	if goFlagsConfig.Cover && goFlagsConfig.CoverProfile == "" {
		goFlagsConfig.CoverProfile = "coverprofile.out"
	}

	return cliConfig, goFlagsConfig, errors
}

func GenerateTestRunArgs(ginkgoConfig GinkgoConfigType, reporterConfig DefaultReporterConfigType, goFlagsConfig GoFlagsConfigType) ([]string, error) {
	var flags GinkgoFlags
	flags = GinkgoConfigFlags.WithPrefix("ginkgo")
	flags = flags.CopyAppend(GinkgoParallelConfigFlags.WithPrefix("ginkgo")...)
	flags = flags.CopyAppend(ReporterConfigFlags.WithPrefix("ginkgo")...)
	flags = flags.CopyAppend(GoRunFlags.WithPrefix("test")...)
	bindings := map[string]interface{}{
		"G":  &ginkgoConfig,
		"R":  &reporterConfig,
		"Go": &goFlagsConfig,
	}

	return GenerateFlagArgs(flags, bindings)
}

func BuildRunCommandFlagSet(ginkgoConfig *GinkgoConfigType, reporterConfig *DefaultReporterConfigType, cliConfig *GinkgoCLIConfigType, goFlagsConfig *GoFlagsConfigType) (GinkgoFlagSet, error) {
	flags := GinkgoConfigFlags
	flags = flags.CopyAppend(ReporterConfigFlags...)
	flags = flags.CopyAppend(GinkgoCLISharedFlags...)
	flags = flags.CopyAppend(GinkgoCLIRunAndWatchFlags...)
	flags = flags.CopyAppend(GinkgoCLIRunFlags...)
	flags = flags.CopyAppend(GoBuildFlags...)
	flags = flags.CopyAppend(GoRunFlags...)

	bindings := map[string]interface{}{
		"G":    ginkgoConfig,
		"R":    reporterConfig,
		"C":    cliConfig,
		"Go":   goFlagsConfig,
		"D":    &deprecatedConfigsType{},
		"Dcli": &deprecatedGinkgoCLIConfig{},
	}

	return NewGinkgoFlagSet(flags, bindings, FlagSections)
}

func BuildWatchCommandFlagSet(ginkgoConfig *GinkgoConfigType, reporterConfig *DefaultReporterConfigType, cliConfig *GinkgoCLIConfigType, goFlagsConfig *GoFlagsConfigType) (GinkgoFlagSet, error) {
	flags := GinkgoConfigFlags
	flags = flags.CopyAppend(ReporterConfigFlags...)
	flags = flags.CopyAppend(GinkgoCLISharedFlags...)
	flags = flags.CopyAppend(GinkgoCLIRunAndWatchFlags...)
	flags = flags.CopyAppend(GinkgoCLIWatchFlags...)
	flags = flags.CopyAppend(GoBuildFlags...)
	flags = flags.CopyAppend(GoRunFlags...)

	bindings := map[string]interface{}{
		"G":    ginkgoConfig,
		"R":    reporterConfig,
		"C":    cliConfig,
		"Go":   goFlagsConfig,
		"D":    &deprecatedConfigsType{},
		"Dcli": &deprecatedGinkgoCLIConfig{},
	}

	return NewGinkgoFlagSet(flags, bindings, FlagSections)
}

func BuildBuildCommandFlagSet(cliConfig *GinkgoCLIConfigType, goFlagsConfig *GoFlagsConfigType) (GinkgoFlagSet, error) {
	flags := GinkgoCLISharedFlags
	flags = flags.CopyAppend(GoBuildFlags...)

	bindings := map[string]interface{}{
		"C":    cliConfig,
		"Go":   goFlagsConfig,
		"D":    &deprecatedConfigsType{},
		"DCli": &deprecatedGinkgoCLIConfig{},
	}

	flagSections := make(GinkgoFlagSections, len(FlagSections))
	copy(flagSections, FlagSections)
	for i := range flagSections {
		if flagSections[i].Key == "multiple-suites" {
			flagSections[i].Heading = "Building Multiple Suites"
		}
		if flagSections[i].Key == "go-build" {
			flagSections[i] = GinkgoFlagSection{Key: "go-build", Style: "{{/}}", Heading: "Go Build Flags",
				Description: "These flags are inherited from go build."}
		}
	}

	return NewGinkgoFlagSet(flags, bindings, flagSections)
}
