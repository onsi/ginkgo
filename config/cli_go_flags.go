package config

// A subset of Go flags are exposed by Ginkgo.  Some are avaiable at compile time (e.g. ginkgo build) and others only at run time (e.g. ginkgo run - which has both build and run time flags).
// More details can be found at:
// https://docs.google.com/spreadsheets/d/1zkp-DS4hU4sAJl5eHh1UmgwxCPQhf3s5a8fbiOI8tJU/

type GoFlagsConfigType struct {
	//build-time flags for code-and-performance analysis
	Race      bool
	Cover     bool
	CoverMode string
	CoverPkg  string
	Vet       string

	//run-time flags for code-and-performance analysis
	BlockProfile         string
	BlockProfileRate     int
	CoverProfile         string
	CPUProfile           string
	MemProfile           string
	MemProfileRate       int
	MutexProfile         string
	MutexProfileFraction int
	Trace                string

	//build-time flags for building
	A             bool
	ASMFlags      string
	BuildMode     string
	Compiler      string
	GCCGoFlags    string
	GCFlags       string
	InstallSuffix string
	LDFlags       string
	LinkShared    bool
	Mod           bool
	N             bool
	ModFile       string
	ModCacheRW    bool
	MSan          bool
	PkgDir        string
	Tags          string
	TrimPath      bool
	ToolExec      string
	Work          bool
	X             bool
}

func NewDefaultGoFlagsConfig() GoFlagsConfigType {
	return GoFlagsConfigType{}
}

func (g GoFlagsConfigType) BinaryMustBePreserved() bool {
	return g.BlockProfile != "" || g.CPUProfile != "" || g.MemProfile != "" || g.MutexProfile != ""
}

var GoBuildFlags = GinkgoFlags{
	{KeyPath: "Go.Race", Name: "race", SectionKey: "code-and-coverage-analysis",
		Usage: "enable data race detection. Supported only on linux/amd64, freebsd/amd64, darwin/amd64, windows/amd64, linux/ppc64le and linux/arm64 (only for 48-bit VMA)."},
	{KeyPath: "Go.Vet", Name: "vet", UsageArgument: "list", SectionKey: "code-and-coverage-analysis",
		Usage: `Configure the invocation of "go vet" during "go test" to use the comma-separated list of vet checks.  If list is empty, "go test" runs "go vet" with a curated list of checks believed to be always worth addressing.  If list is "off", "go test" does not run "go vet" at all.  Available checks can be found by running 'go doc cmd/vet'`},
	{KeyPath: "Go.Cover", Name: "cover", SectionKey: "code-and-coverage-analysis",
		Usage: "Enable coverage analysis.	Note that because coverage works by annotating the source code before compilation, compilation and test failures with coverage enabled may report line numbers that don't correspond to the original sources."},
	{KeyPath: "Go.CoverMode", Name: "covermode", UsageArgument: "set,count,atomic", SectionKey: "code-and-coverage-analysis",
		Usage: `Set the mode for coverage analysis for the package[s] being tested. 'set': does this statement run? 'count': how many times does this statement run? 'atomic': like count, but correct in multithreaded tests and more expensive (must use atomic with -race). Sets -cover`},
	{KeyPath: "Go.CoverPkg", Name: "coverpkg", UsageArgument: "pattern1,pattern2,pattern3", SectionKey: "code-and-coverage-analysis",
		Usage: "Apply coverage analysis in each test to packages matching the patterns. 	The default is for each test to analyze only the package being tested. See 'go help packages' for a description of package patterns. Sets -cover."},

	{KeyPath: "Go.A", Name: "a", SectionKey: "go-build",
		Usage: "force rebuilding of packages that are already up-to-date."},
	{KeyPath: "Go.ASMFlags", Name: "asmflags", UsageArgument: "'[pattern=]arg list'", SectionKey: "go-build",
		Usage: "arguments to pass on each go tool asm invocation."},
	{KeyPath: "Go.BuildMode", Name: "buildmode", UsageArgument: "mode", SectionKey: "go-build",
		Usage: "build mode to use. See 'go help buildmode' for more."},
	{KeyPath: "Go.Compiler", Name: "compiler", UsageArgument: "name", SectionKey: "go-build",
		Usage: "name of compiler to use, as in runtime.Compiler (gccgo or gc)."},
	{KeyPath: "Go.GCCGoFlags", Name: "gccgoflags", UsageArgument: "'[pattern=]arg list'", SectionKey: "go-build",
		Usage: "arguments to pass on each gccgo compiler/linker invocation."},
	{KeyPath: "Go.GCFlags", Name: "gcflags", UsageArgument: "'[pattern=]arg list'", SectionKey: "go-build",
		Usage: "arguments to pass on each go tool compile invocation."},
	{KeyPath: "Go.InstallSuffix", Name: "installsuffix", SectionKey: "go-build",
		Usage: "a suffix to use in the name of the package installation directory, in order to keep output separate from default builds. If using the -race flag, the install suffix is automatically set to raceor, if set explicitly, has _race appended to it. Likewise for the -msan flag.  Using a -buildmode option that requires non-default compile flags has a similar effect."},
	{KeyPath: "Go.LDFlags", Name: "ldflags", UsageArgument: "'[pattern=]arg list'", SectionKey: "go-build",
		Usage: "arguments to pass on each go tool link invocation."},
	{KeyPath: "Go.LinkShared", Name: "linkshared", SectionKey: "go-build",
		Usage: "build code that will be linked against shared libraries previously created with -buildmode=shared."},
	{KeyPath: "Go.Mod", Name: "mod", UsageArgument: "mode (readonly, vender, or mod)", SectionKey: "go-build",
		Usage: "module download mode to use: readonly, vendor, or mod.  See 'go help modules' for more."},
	{KeyPath: "Go.ModCacheRW", Name: "modcacherw", SectionKey: "go-build",
		Usage: "leave newly-created directories in the module cache read-write instead of making them read-only."},
	{KeyPath: "Go.ModFile", Name: "modfile", UsageArgument: "file", SectionKey: "go-build",
		Usage: `in module aware mode, read (and possibly write) an alternate go.mod file instead of the one in the module root directory. A file named go.mod must still be present in order to determine the module root directory, but it is not accessed. When -modfile is specified, an alternate go.sum file is also used: its path is derived from the -modfile flag by trimming the ".mod" extension and appending ".sum".`},
	{KeyPath: "Go.MSan", Name: "msan", SectionKey: "go-build",
		Usage: "enable interoperation with memory sanitizer. Supported only on linux/amd64, linux/arm64 and only with Clang/LLVM as the host C compiler. On linux/arm64, pie build mode will be used."},
	{KeyPath: "Go.N", Name: "n", SectionKey: "go-build",
		Usage: "print the commands but do not run them."},
	{KeyPath: "Go.PkgDir", Name: "pkgdir", UsageArgument: "dir", SectionKey: "go-build",
		Usage: "install and load all packages from dir instead of the usual locations. For example, when building with a non-standard configuration, use -pkgdir to keep generated packages in a separate location."},
	{KeyPath: "Go.Tags", Name: "tags", UsageArgument: "tag,list", SectionKey: "go-build",
		Usage: "a comma-separated list of build tags to consider satisfied during the build. For more information about build tags, see the description of build constraints in the documentation for the go/build package. (Earlier versions of Go used a space-separated list, and that form is deprecated but still recognized.)"},
	{KeyPath: "Go.TrimPath", Name: "trimpath", SectionKey: "go-build",
		Usage: `remove all file system paths from the resulting executable. Instead of absolute file system paths, the recorded file names will begin with either "go" (for the standard library), or a module path@version (when using modules), or a plain import path (when using GOPATH).`},
	{KeyPath: "Go.ToolExec", Name: "toolexec", UsageArgument: "'cmd args'", SectionKey: "go-build",
		Usage: "a program to use to invoke toolchain programs like vet and asm. For example, instead of running asm, the go command will run cmd args /path/to/asm <arguments for asm>'."},
	{KeyPath: "Go.Work", Name: "work", SectionKey: "go-build",
		Usage: "print the name of the temporary work directory and do not delete it when exiting."},
	{KeyPath: "Go.X", Name: "x", SectionKey: "go-build",
		Usage: "print the commands."},
}

var GoRunFlags = GinkgoFlags{
	{KeyPath: "Go.CoverProfile", Name: "coverprofile", UsageArgument: "file", SectionKey: "code-and-coverage-analysis",
		Usage: `Write a coverage profile to the file after all tests have passed. Sets -cover.`},
	{KeyPath: "Go.BlockProfile", Name: "blockprofile", UsageArgument: "rile", SectionKey: "performance-analysis",
		Usage: `Write a goroutine blocking profile to the specified file when all tests are complete. Preserves test binary.`},
	{KeyPath: "Go.BlockProfileRate", Name: "blockprofilerate", UsageArgument: "rate", SectionKey: "performance-analysis",
		Usage: `Control the detail provided in goroutine blocking profiles by calling runtime.SetBlockProfileRate with rate. See 'go doc runtime.SetBlockProfileRate'. The profiler aims to sample, on average, one blocking event every n nanoseconds the program spends blocked. By default, if -test.blockprofile is set without this flag, all blocking events are recorded, equivalent to -test.blockprofilerate=1.`},
	{KeyPath: "Go.CPUProfile", Name: "cpuprofile", UsageArgument: "file", SectionKey: "performance-analysis",
		Usage: `Write a CPU profile to the specified file before exiting. Preserves test binary.`},
	{KeyPath: "Go.MemProfile", Name: "memprofile", UsageArgument: "file", SectionKey: "performance-analysis",
		Usage: `Write an allocation profile to the file after all tests have passed. Preserves test binary.`},
	{KeyPath: "Go.MemProfileRate", Name: "memprofilerate", UsageArgument: "rate", SectionKey: "performance-analysis",
		Usage: `Enable more precise (and expensive) memory allocation profiles by setting runtime.MemProfileRate. See 'go doc runtime.MemProfileRate'. To profile all memory allocations, use -test.memprofilerate=1.`},
	{KeyPath: "Go.MutexProfile", Name: "mutexprofile", UsageArgument: "file", SectionKey: "performance-analysis",
		Usage: `Write a mutex contention profile to the specified file when all tests are complete. Preserves test binary.`},
	{KeyPath: "Go.MutexProfileFraction", Name: "mutexprofilefraction", UsageArgument: "n", SectionKey: "performance-analysis",
		Usage: `if >= 0, calls runtime.SetMutexProfileFraction()	Sample 1 in n stack traces of goroutines holding a contended mutex.`},
	{KeyPath: "Go.Trace", Name: "execution-trace", UsageArgument: "file", ExportAs: "trace", SectionKey: "performance-analysis",
		Usage: `Write an execution trace to the specified file before exiting.`},
}

func GenerateGoTestCompileArgs(goFlagsConfig GoFlagsConfigType, destination string, packageToBuild string) ([]string, error) {
	// if the user has set the CoverProfile run-time flag make sure to set the build-time cover flag to make sure
	// the built test binary can generate a coverprofile
	if goFlagsConfig.CoverProfile != "" {
		goFlagsConfig.Cover = true
	}

	args := []string{"test", "-c", "-o", destination, packageToBuild}
	goArgs, err := GenerateFlagArgs(
		GoBuildFlags,
		map[string]interface{}{
			"Go": &goFlagsConfig,
		},
	)

	if err != nil {
		return []string{}, err
	}
	args = append(args, goArgs...)
	return args, nil
}
