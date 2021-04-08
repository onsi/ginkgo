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

	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/formatter"
	"github.com/onsi/ginkgo/ginkgo/command"
	"github.com/onsi/ginkgo/internal/parallel_support"
	"github.com/onsi/ginkgo/reporters"
	"github.com/onsi/ginkgo/types"
)

func RunCompiledSuite(suite TestSuite, ginkgoConfig config.GinkgoConfigType, reporterConfig config.DefaultReporterConfigType, cliConfig config.GinkgoCLIConfigType, goFlagsConfig config.GoFlagsConfigType, additionalArgs []string) TestSuite {
	suite.Passed = false
	suite.HasProgrammaticFocus = false

	if suite.PathToCompiledTest == "" {
		return suite
	}

	if suite.IsGinkgo && cliConfig.ComputedNodes() > 1 {
		suite = runParallel(suite, ginkgoConfig, reporterConfig, cliConfig, goFlagsConfig, additionalArgs)
	} else if suite.IsGinkgo {
		suite = runSerial(suite, ginkgoConfig, reporterConfig, cliConfig, goFlagsConfig, additionalArgs)
	} else {
		suite = runGoTest(suite, cliConfig)
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

func runGoTest(suite TestSuite, cliConfig config.GinkgoCLIConfigType) TestSuite {
	cmd, buf := buildAndStartCommand(suite, []string{"-test.v"}, true)

	cmd.Wait()

	exitStatus := cmd.ProcessState.Sys().(syscall.WaitStatus).ExitStatus()
	suite.Passed = (exitStatus == 0) || (exitStatus == types.GINKGO_FOCUS_EXIT_CODE)
	suite.Passed = !(checkForNoTestsWarning(buf) && cliConfig.RequireSuite) && suite.Passed

	return suite
}

func runSerial(suite TestSuite, ginkgoConfig config.GinkgoConfigType, reporterConfig config.DefaultReporterConfigType, cliConfig config.GinkgoCLIConfigType, goFlagsConfig config.GoFlagsConfigType, additionalArgs []string) TestSuite {
	args, err := config.GenerateTestRunArgs(ginkgoConfig, reporterConfig, goFlagsConfig)
	command.AbortIfError("Failed to generate test run arguments", err)
	args = append([]string{"--test.timeout=0"}, args...)
	args = append(args, additionalArgs...)

	cmd, buf := buildAndStartCommand(suite, args, true)

	cmd.Wait()

	exitStatus := cmd.ProcessState.Sys().(syscall.WaitStatus).ExitStatus()
	suite.HasProgrammaticFocus = (exitStatus == types.GINKGO_FOCUS_EXIT_CODE)
	suite.Passed = (exitStatus == 0) || (exitStatus == types.GINKGO_FOCUS_EXIT_CODE)
	suite.Passed = !(checkForNoTestsWarning(buf) && cliConfig.RequireSuite) && suite.Passed

	return suite
}

func runParallel(suite TestSuite, ginkgoConfig config.GinkgoConfigType, reporterConfig config.DefaultReporterConfigType, cliConfig config.GinkgoCLIConfigType, goFlagsConfig config.GoFlagsConfigType, additionalArgs []string) TestSuite {
	type nodeResult struct {
		passed               bool
		hasProgrammaticFocus bool
	}

	numNodes := cliConfig.ComputedNodes()
	nodeOutput := make([]*bytes.Buffer, numNodes)
	coverProfiles := []string{}
	nodeResults := make(chan nodeResult)

	server, err := parallel_support.NewServer(numNodes, reporters.NewDefaultReporter(reporterConfig, formatter.ColorableStdOut))
	command.AbortIfError("Failed to start parallel spec server", err)
	server.Start()
	defer server.Close()

	for node := 1; node <= numNodes; node++ {
		nodeGinkgoConfig := ginkgoConfig
		nodeGinkgoConfig.ParallelNode, nodeGinkgoConfig.ParallelTotal, nodeGinkgoConfig.ParallelHost = node, numNodes, server.Address()

		nodeGoFlagsConfig := goFlagsConfig
		if goFlagsConfig.Cover {
			nodeGoFlagsConfig.CoverProfile = fmt.Sprintf("%s.%d", goFlagsConfig.CoverProfile, node)
			coverProfiles = append(coverProfiles, filepath.Join(suite.Path, nodeGoFlagsConfig.CoverProfile))
		}

		args, err := config.GenerateTestRunArgs(nodeGinkgoConfig, reporterConfig, nodeGoFlagsConfig)
		command.AbortIfError("Failed to generate test run argumnets", err)
		args = append([]string{"--test.timeout=0"}, args...)
		args = append(args, additionalArgs...)

		cmd, buf := buildAndStartCommand(suite, args, false)
		nodeOutput[node-1] = buf
		server.RegisterAlive(node, func() bool { return cmd.ProcessState == nil || !cmd.ProcessState.Exited() })

		go func() {
			cmd.Wait()
			exitStatus := cmd.ProcessState.Sys().(syscall.WaitStatus).ExitStatus()
			nodeResults <- nodeResult{
				passed:               (exitStatus == 0) || (exitStatus == types.GINKGO_FOCUS_EXIT_CODE),
				hasProgrammaticFocus: exitStatus == types.GINKGO_FOCUS_EXIT_CODE,
			}
		}()
	}

	suite.Passed = true
	for node := 1; node <= cliConfig.ComputedNodes(); node++ {
		result := <-nodeResults
		suite.Passed = suite.Passed && result.passed
		suite.HasProgrammaticFocus = suite.HasProgrammaticFocus || result.hasProgrammaticFocus
	}

	select {
	case <-server.Done:
		fmt.Println("")
	case <-time.After(time.Second):
		//the serve never got back to us.  Something must have gone wrong.
		fmt.Fprintln(os.Stderr, "** Ginkgo timed out waiting for all parallel nodes to report back. **")
		fmt.Fprintf(os.Stderr, "%s (%s)\n", suite.PackageName, suite.Path)
		for node := 1; node <= cliConfig.ComputedNodes(); node++ {
			fmt.Fprintf(os.Stderr, "Output from node %d\n:", node)
			fmt.Fprintln(os.Stderr, formatter.Fi(1, "%s", nodeOutput[node-1].String()))
		}
	}

	for node := 1; node <= cliConfig.ComputedNodes(); node++ {
		output := nodeOutput[node-1].String()
		if node == 1 {
			suite.Passed = !(checkForNoTestsWarning(nodeOutput[node-1]) && cliConfig.RequireSuite) && suite.Passed
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

	return suite
}

func runAfterRunHook(command string, noColor bool, suite TestSuite) {
	if command == "" {
		return
	}
	f := formatter.NewWithNoColorBool(noColor)

	// Allow for string replacement to pass input to the command
	passed := "[FAIL]"
	if suite.Passed {
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
