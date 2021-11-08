package internal

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/onsi/ginkgo/formatter"
	"github.com/onsi/ginkgo/ginkgo/command"
	"github.com/onsi/ginkgo/internal/parallel_support"
	"github.com/onsi/ginkgo/reporters"
	"github.com/onsi/ginkgo/types"
)

func RunCompiledSuite(suite TestSuite, ginkgoConfig types.SuiteConfig, reporterConfig types.ReporterConfig, cliConfig types.CLIConfig, goFlagsConfig types.GoFlagsConfig, additionalArgs []string) TestSuite {
	suite.State = TestSuiteStateFailed
	suite.HasProgrammaticFocus = false

	if suite.PathToCompiledTest == "" {
		return suite
	}

	if suite.IsGinkgo && cliConfig.ComputedProcs() > 1 {
		suite = runParallel(suite, ginkgoConfig, reporterConfig, cliConfig, goFlagsConfig, additionalArgs)
	} else if suite.IsGinkgo {
		suite = runSerial(suite, ginkgoConfig, reporterConfig, cliConfig, goFlagsConfig, additionalArgs)
	} else {
		suite = runGoTest(suite, cliConfig, goFlagsConfig)
	}
	runAfterRunHook(cliConfig.AfterRunHook, reporterConfig.NoColor, suite)
	return suite
}

func buildAndStartCommand(suite TestSuite, args []string, pipeToStdout bool) (*exec.Cmd, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	cmd := exec.Command(suite.PathToCompiledTest, args...)
	cmd.Dir = suite.Path
	if pipeToStdout {
		cmd.Stderr = io.MultiWriter(os.Stdout, buf)
		cmd.Stdout = os.Stdout
	} else {
		cmd.Stderr = buf
		cmd.Stdout = buf
	}
	err := cmd.Start()
	command.AbortIfError("Failed to start test suite", err)

	return cmd, buf
}

func checkForNoTestsWarning(buf *bytes.Buffer) bool {
	if strings.Contains(buf.String(), "warning: no tests to run") {
		fmt.Fprintf(os.Stderr, `Found no test suites, did you forget to run "ginkgo bootstrap"?`)
		return true
	}
	return false
}

func runGoTest(suite TestSuite, cliConfig types.CLIConfig, goFlagsConfig types.GoFlagsConfig) TestSuite {
	args, err := types.GenerateGoTestRunArgs(goFlagsConfig)
	command.AbortIfError("Failed to generate test run arguments", err)
	cmd, buf := buildAndStartCommand(suite, args, true)

	cmd.Wait()

	exitStatus := cmd.ProcessState.Sys().(syscall.WaitStatus).ExitStatus()
	passed := (exitStatus == 0) || (exitStatus == types.GINKGO_FOCUS_EXIT_CODE)
	passed = !(checkForNoTestsWarning(buf) && cliConfig.RequireSuite) && passed
	if passed {
		suite.State = TestSuiteStatePassed
	} else {
		suite.State = TestSuiteStateFailed
	}

	return suite
}

func runSerial(suite TestSuite, ginkgoConfig types.SuiteConfig, reporterConfig types.ReporterConfig, cliConfig types.CLIConfig, goFlagsConfig types.GoFlagsConfig, additionalArgs []string) TestSuite {
	args, err := types.GenerateGinkgoTestRunArgs(ginkgoConfig, reporterConfig, goFlagsConfig)
	command.AbortIfError("Failed to generate test run arguments", err)
	args = append([]string{"--test.timeout=0"}, args...)
	args = append(args, additionalArgs...)

	cmd, buf := buildAndStartCommand(suite, args, true)

	cmd.Wait()

	exitStatus := cmd.ProcessState.Sys().(syscall.WaitStatus).ExitStatus()
	suite.HasProgrammaticFocus = (exitStatus == types.GINKGO_FOCUS_EXIT_CODE)
	passed := (exitStatus == 0) || (exitStatus == types.GINKGO_FOCUS_EXIT_CODE)
	passed = !(checkForNoTestsWarning(buf) && cliConfig.RequireSuite) && passed
	if passed {
		suite.State = TestSuiteStatePassed
	} else {
		suite.State = TestSuiteStateFailed
	}

	return suite
}

func runParallel(suite TestSuite, ginkgoConfig types.SuiteConfig, reporterConfig types.ReporterConfig, cliConfig types.CLIConfig, goFlagsConfig types.GoFlagsConfig, additionalArgs []string) TestSuite {
	type procResult struct {
		passed               bool
		hasProgrammaticFocus bool
	}

	numProcs := cliConfig.ComputedProcs()
	procOutput := make([]*bytes.Buffer, numProcs)
	coverProfiles := []string{}

	blockProfiles := []string{}
	cpuProfiles := []string{}
	memProfiles := []string{}
	mutexProfiles := []string{}

	procResults := make(chan procResult)

	server, err := parallel_support.NewServer(numProcs, reporters.NewDefaultReporter(reporterConfig, formatter.ColorableStdOut))
	command.AbortIfError("Failed to start parallel spec server", err)
	server.Start()
	defer server.Close()

	for proc := 1; proc <= numProcs; proc++ {
		procGinkgoConfig := ginkgoConfig
		procGinkgoConfig.ParallelProcess, procGinkgoConfig.ParallelTotal, procGinkgoConfig.ParallelHost = proc, numProcs, server.Address()

		procGoFlagsConfig := goFlagsConfig
		if goFlagsConfig.Cover {
			procGoFlagsConfig.CoverProfile = fmt.Sprintf("%s.%d", goFlagsConfig.CoverProfile, proc)
			coverProfiles = append(coverProfiles, filepath.Join(suite.Path, procGoFlagsConfig.CoverProfile))
		}
		if goFlagsConfig.BlockProfile != "" {
			procGoFlagsConfig.BlockProfile = fmt.Sprintf("%s.%d", goFlagsConfig.BlockProfile, proc)
			blockProfiles = append(blockProfiles, filepath.Join(suite.Path, procGoFlagsConfig.BlockProfile))
		}
		if goFlagsConfig.CPUProfile != "" {
			procGoFlagsConfig.CPUProfile = fmt.Sprintf("%s.%d", goFlagsConfig.CPUProfile, proc)
			cpuProfiles = append(cpuProfiles, filepath.Join(suite.Path, procGoFlagsConfig.CPUProfile))
		}
		if goFlagsConfig.MemProfile != "" {
			procGoFlagsConfig.MemProfile = fmt.Sprintf("%s.%d", goFlagsConfig.MemProfile, proc)
			memProfiles = append(memProfiles, filepath.Join(suite.Path, procGoFlagsConfig.MemProfile))
		}
		if goFlagsConfig.MutexProfile != "" {
			procGoFlagsConfig.MutexProfile = fmt.Sprintf("%s.%d", goFlagsConfig.MutexProfile, proc)
			mutexProfiles = append(mutexProfiles, filepath.Join(suite.Path, procGoFlagsConfig.MutexProfile))
		}

		args, err := types.GenerateGinkgoTestRunArgs(procGinkgoConfig, reporterConfig, procGoFlagsConfig)
		command.AbortIfError("Failed to generate test run argumnets", err)
		args = append([]string{"--test.timeout=0"}, args...)
		args = append(args, additionalArgs...)

		cmd, buf := buildAndStartCommand(suite, args, false)
		procOutput[proc-1] = buf
		server.RegisterAlive(proc, func() bool { return cmd.ProcessState == nil || !cmd.ProcessState.Exited() })

		go func() {
			cmd.Wait()
			exitStatus := cmd.ProcessState.Sys().(syscall.WaitStatus).ExitStatus()
			procResults <- procResult{
				passed:               (exitStatus == 0) || (exitStatus == types.GINKGO_FOCUS_EXIT_CODE),
				hasProgrammaticFocus: exitStatus == types.GINKGO_FOCUS_EXIT_CODE,
			}
		}()
	}

	passed := true
	for proc := 1; proc <= cliConfig.ComputedProcs(); proc++ {
		result := <-procResults
		passed = passed && result.passed
		suite.HasProgrammaticFocus = suite.HasProgrammaticFocus || result.hasProgrammaticFocus
	}
	if passed {
		suite.State = TestSuiteStatePassed
	} else {
		suite.State = TestSuiteStateFailed
	}

	select {
	case <-server.GetSuiteDone():
		fmt.Println("")
	case <-time.After(time.Second):
		//the serve never got back to us.  Something must have gone wrong.
		fmt.Fprintln(os.Stderr, "** Ginkgo timed out waiting for all parallel procs to report back. **")
		fmt.Fprintf(os.Stderr, "%s (%s)\n", suite.PackageName, suite.Path)
		for proc := 1; proc <= cliConfig.ComputedProcs(); proc++ {
			fmt.Fprintf(os.Stderr, "Output from proc %d:\n", proc)
			fmt.Fprintln(os.Stderr, formatter.Fi(1, "%s", procOutput[proc-1].String()))
		}
		fmt.Fprintf(os.Stderr, "** End **")
	}

	for proc := 1; proc <= cliConfig.ComputedProcs(); proc++ {
		output := procOutput[proc-1].String()
		if proc == 1 && checkForNoTestsWarning(procOutput[0]) && cliConfig.RequireSuite {
			suite.State = TestSuiteStateFailed
		}
		if strings.Contains(output, "deprecated Ginkgo functionality") {
			fmt.Fprintln(os.Stderr, output)
		}
	}

	if len(coverProfiles) > 0 {
		coverProfile := filepath.Join(suite.Path, goFlagsConfig.CoverProfile)
		err := MergeAndCleanupCoverProfiles(coverProfiles, coverProfile)
		command.AbortIfError("Failed to combine cover profiles", err)

		coverage, err := GetCoverageFromCoverProfile(coverProfile)
		command.AbortIfError("Failed to compute coverage", err)
		if coverage == 0 {
			fmt.Fprintln(os.Stdout, "coverage: [no statements]")
		} else {
			fmt.Fprintf(os.Stdout, "coverage: %.1f%% of statements\n", coverage)
		}
	}
	if len(blockProfiles) > 0 {
		blockProfile := filepath.Join(suite.Path, goFlagsConfig.BlockProfile)
		err := MergeProfiles(blockProfiles, blockProfile)
		command.AbortIfError("Failed to combine blockprofiles", err)
	}
	if len(cpuProfiles) > 0 {
		cpuProfile := filepath.Join(suite.Path, goFlagsConfig.CPUProfile)
		err := MergeProfiles(cpuProfiles, cpuProfile)
		command.AbortIfError("Failed to combine cpuprofiles", err)
	}
	if len(memProfiles) > 0 {
		memProfile := filepath.Join(suite.Path, goFlagsConfig.MemProfile)
		err := MergeProfiles(memProfiles, memProfile)
		command.AbortIfError("Failed to combine memprofiles", err)
	}
	if len(mutexProfiles) > 0 {
		mutexProfile := filepath.Join(suite.Path, goFlagsConfig.MutexProfile)
		err := MergeProfiles(mutexProfiles, mutexProfile)
		command.AbortIfError("Failed to combine mutexprofiles", err)
	}

	return suite
}

func runAfterRunHook(command string, noColor bool, suite TestSuite) {
	if command == "" {
		return
	}
	f := formatter.NewWithNoColorBool(noColor)

	// Allow for string replacement to pass input to the command
	passed := "[FAIL]"
	if suite.State.Is(TestSuiteStatePassed) {
		passed = "[PASS]"
	}
	command = strings.Replace(command, "(ginkgo-suite-passed)", passed, -1)
	command = strings.Replace(command, "(ginkgo-suite-name)", suite.PackageName, -1)

	// Must break command into parts
	splitArgs := regexp.MustCompile(`'.+'|".+"|\S+`)
	parts := splitArgs.FindAllString(command, -1)

	output, err := exec.Command(parts[0], parts[1:]...).CombinedOutput()
	if err != nil {
		fmt.Fprintln(formatter.ColorableStdOut, f.Fi(0, "{{red}}{{bold}}After-run-hook failed:{{/}}"))
		fmt.Fprintln(formatter.ColorableStdOut, f.Fi(1, "{{red}}%s{{/}}", output))
	} else {
		fmt.Fprintln(formatter.ColorableStdOut, f.Fi(0, "{{green}}{{bold}}After-run-hook succeeded:{{/}}"))
		fmt.Fprintln(formatter.ColorableStdOut, f.Fi(1, "{{green}}%s{{/}}", output))
	}
}
