package main

import (
	"bytes"
	"fmt"
	"github.com/onsi/ginkgo/config"
	"io"
	"os"
	"os/exec"
	"os/signal"
)

type testRunner struct {
	numCPU           int
	runMagicI        bool
	race             bool
	cover            bool
	executedCommands []*exec.Cmd
	reports          []*bytes.Buffer
}

func newTestRunner(numCPU int, runMagicI bool, race bool, cover bool) *testRunner {
	return &testRunner{
		numCPU:           numCPU,
		runMagicI:        runMagicI,
		race:             race,
		cover:            cover,
		executedCommands: []*exec.Cmd{},
		reports:          []*bytes.Buffer{},
	}
}

func (t *testRunner) run(suites []testSuite) bool {
	t.registerSignalHandler()

	for _, suite := range suites {
		if !t.runSuite(suite) {
			return false
		}
	}

	return true
}

func (t *testRunner) runSuite(suite testSuite) bool {
	if t.runMagicI {
		t.runGoI(suite)
	}

	if suite.isGinkgo {
		if t.numCPU > 1 {
			return t.runParallelGinkgoSuite(suite)
		} else {
			return t.runSerialGinkgoSuite(suite)
		}
	} else {
		return t.runGoTestSuite(suite)
	}
}

func (t *testRunner) runGoI(suite testSuite) {
	args := []string{"test", "-i"}
	if t.race {
		args = append(args, "-race")
	}
	args = append(args, suite.path)
	cmd := exec.Command("go", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("go test -i %s failed with:\n\n%s", suite.path, output)
		os.Exit(1)
	}
}

func (t *testRunner) runParallelGinkgoSuite(suite testSuite) bool {
	completions := make(chan bool)
	for cpu := 0; cpu < t.numCPU; cpu++ {
		config.GinkgoConfig.ParallelNode = cpu + 1
		config.GinkgoConfig.ParallelTotal = t.numCPU

		args := config.BuildFlagArgs("ginkgo", config.GinkgoConfig, config.DefaultReporterConfig)
		args = append(args, t.commonArgs(suite)...)

		buffer := new(bytes.Buffer)
		t.reports = append(t.reports, buffer)

		go t.runCommand(suite.path, args, buffer, completions)
	}

	passed := true

	for cpu := 0; cpu < t.numCPU; cpu++ {
		passed = <-completions && passed
	}

	for _, report := range t.reports {
		fmt.Print(report.String())
	}
	os.Stdout.Sync()

	return passed
}

func (t *testRunner) runSerialGinkgoSuite(suite testSuite) bool {
	args := config.BuildFlagArgs("ginkgo", config.GinkgoConfig, config.DefaultReporterConfig)
	args = append(args, t.commonArgs(suite)...)
	return t.runCommand(suite.path, args, os.Stdout, nil)
}

func (t *testRunner) runGoTestSuite(suite testSuite) bool {
	args := t.commonArgs(suite)
	return t.runCommand(suite.path, args, os.Stdout, nil)
}

func (t *testRunner) commonArgs(suite testSuite) []string {
	args := []string{}
	if t.race {
		args = append(args, "--race")
	}
	if t.cover {
		args = append([]string{"--cover", "--coverprofile=" + suite.packageName + ".coverprofile"})
	}
	return args
}

func (t *testRunner) runCommand(path string, args []string, stream io.Writer, completions chan bool) bool {
	args = append([]string{"test", "-v", "-timeout=24h", path}, args...)

	cmd := exec.Command("go", args...)
	t.executedCommands = append(t.executedCommands, cmd)

	doneStreaming := make(chan bool, 2)
	streamPipe := func(pipe io.ReadCloser) {
		io.Copy(stream, pipe)
		doneStreaming <- true
	}

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	go streamPipe(stdout)
	go streamPipe(stderr)

	err := cmd.Start()
	if err != nil {
		os.Exit(1)
	}

	<-doneStreaming
	<-doneStreaming

	err = cmd.Wait()
	if completions != nil {
		completions <- (err == nil)
	}
	return err == nil
}

func (t *testRunner) registerSignalHandler() {
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, os.Kill)

		select {
		case sig := <-c:
			for _, cmd := range t.executedCommands {
				cmd.Process.Signal(sig)
			}
			os.Exit(1)
		}
	}()
}
