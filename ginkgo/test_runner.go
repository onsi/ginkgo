package main

import (
	"bytes"
	"fmt"
	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/ginkgo/aggregator"
	"github.com/onsi/ginkgo/ginkgo/testsuite"
	"github.com/onsi/ginkgo/remote"
	"github.com/onsi/ginkgo/stenographer"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

type testRunner struct {
	numCPU           int
	parallelStream   bool
	race             bool
	cover            bool
	executedCommands []*exec.Cmd
	compiledArtifact string
	reports          []*bytes.Buffer

	lock *sync.Mutex
}

func newTestRunner(numCPU int, parallelStream bool, runMagicI bool, race bool, cover bool) *testRunner {
	return &testRunner{
		numCPU:           numCPU,
		parallelStream:   parallelStream,
		race:             race,
		cover:            cover,
		executedCommands: []*exec.Cmd{},
		reports:          []*bytes.Buffer{},
		lock:             &sync.Mutex{},
	}
}

func (t *testRunner) runSuite(suite *testsuite.TestSuite) bool {
	var success bool
	success = t.compileSuite(suite)
	if !success {
		return success
	}

	if suite.IsGinkgo {
		if t.numCPU > 1 {
			if t.parallelStream {
				success = t.runAndStreamParallelGinkgoSuite(suite)
			} else {
				success = t.runParallelGinkgoSuite(suite)
			}
		} else {
			success = t.runSerialGinkgoSuite(suite)
		}
	} else {
		success = t.runGoTestSuite(suite)
	}

	t.cleanUpCompiledSuite()
	return success
}

func (t *testRunner) compileSuite(suite *testsuite.TestSuite) bool {
	args := []string{"test", "-c", "-i"}
	if t.race {
		args = append(args, "-race")
	}
	if t.cover {
		args = append(args, "-cover", "-covermode=atomic")
	}
	cmd := exec.Command("go", args...)
	cmd.Dir = suite.Path
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Failed to compile %s:\n\n%s", suite.PackageName, output)
		t.compiledArtifact = ""
		return false
	}
	t.compiledArtifact, _ = filepath.Abs(filepath.Join(suite.Path, fmt.Sprintf("%s.test", suite.PackageName)))
	return true
}

func (t *testRunner) cleanUpCompiledSuite() {
	os.Remove(t.compiledArtifact)
}

func (t *testRunner) runSerialGinkgoSuite(suite *testsuite.TestSuite) bool {
	ginkgoArgs := config.BuildFlagArgs("ginkgo", config.GinkgoConfig, config.DefaultReporterConfig)
	return t.runCompiledSuite(suite, ginkgoArgs, nil, os.Stdout, nil)
}

func (t *testRunner) runGoTestSuite(suite *testsuite.TestSuite) bool {
	return t.runCompiledSuite(suite, []string{}, nil, os.Stdout, nil)
}

func (t *testRunner) runParallelGinkgoSuite(suite *testsuite.TestSuite) bool {
	completions := make(chan bool)
	for cpu := 0; cpu < t.numCPU; cpu++ {
		config.GinkgoConfig.ParallelNode = cpu + 1
		config.GinkgoConfig.ParallelTotal = t.numCPU

		ginkgoArgs := config.BuildFlagArgs("ginkgo", config.GinkgoConfig, config.DefaultReporterConfig)

		buffer := new(bytes.Buffer)
		t.reports = append(t.reports, buffer)

		go t.runCompiledSuite(suite, ginkgoArgs, nil, buffer, completions)
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

func (t *testRunner) runAndStreamParallelGinkgoSuite(suite *testsuite.TestSuite) bool {
	result := make(chan bool, 0)
	stenographer := stenographer.New(!config.DefaultReporterConfig.NoColor)
	aggregator := aggregator.NewAggregator(t.numCPU, result, config.DefaultReporterConfig, stenographer)

	server, err := remote.NewServer()
	if err != nil {
		panic("Failed to start parallel spec server")
	}

	server.RegisterReporters(aggregator)
	server.Start()
	defer server.Stop()
	serverAddress := server.Address()

	completions := make(chan bool)
	for cpu := 0; cpu < t.numCPU; cpu++ {
		config.GinkgoConfig.ParallelNode = cpu + 1
		config.GinkgoConfig.ParallelTotal = t.numCPU

		ginkgoArgs := config.BuildFlagArgs("ginkgo", config.GinkgoConfig, config.DefaultReporterConfig)

		env := os.Environ()
		env = append(env, fmt.Sprintf("GINKGO_REMOTE_REPORTING_SERVER=%s", serverAddress))

		buffer := new(bytes.Buffer)
		t.reports = append(t.reports, buffer)

		go t.runCompiledSuite(suite, ginkgoArgs, env, buffer, completions)
	}

	for cpu := 0; cpu < t.numCPU; cpu++ {
		<-completions
	}

	//all test processes are done, at this point
	//we should be able to wait for the aggregator to tell us that it's done

	var passed = false
	select {
	case passed = <-result:
		//the aggregator is done and can tell us whether or not the suite passed
	case <-time.After(time.Second):
		//the aggregator never got back to us!  something must have gone wrong
		fmt.Println("")
		fmt.Println("")
		fmt.Println("   ----------------------------------------------------------- ")
		fmt.Println("  |                                                           |")
		fmt.Println("  |  Ginkgo timed out waiting for all parallel nodes to end!  |")
		fmt.Println("  |  Here is some salvaged output:                            |")
		fmt.Println("  |                                                           |")
		fmt.Println("   ----------------------------------------------------------- ")
		fmt.Println("")
		fmt.Println("")

		os.Stdout.Sync()

		time.Sleep(time.Second)

		for _, report := range t.reports {
			fmt.Print(report.String())
		}

		os.Stdout.Sync()
	}

	return passed
}

func (t *testRunner) runCompiledSuite(suite *testsuite.TestSuite, ginkgoArgs []string, env []string, stream io.Writer, completions chan bool) bool {
	args := []string{"-test.v", "-test.timeout=24h"}
	if t.cover {
		args = append(args, "--test.coverprofile="+suite.PackageName+".coverprofile")
	}

	args = append(args, ginkgoArgs...)

	cmd := exec.Command(t.compiledArtifact, args...)
	cmd.Env = env
	cmd.Dir = suite.Path

	t.lock.Lock()
	t.executedCommands = append(t.executedCommands, cmd)
	t.lock.Unlock()

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
		fmt.Printf("Failed to run test suite!\n\t%s", err.Error())
		if completions != nil {
			completions <- false
		}
		return false
	}

	<-doneStreaming
	<-doneStreaming

	err = cmd.Wait()

	t.lock.Lock()
	//delete the command that just finished executing
	for commandIndex, executedCommand := range t.executedCommands {
		if executedCommand == cmd {
			t.executedCommands[commandIndex] = t.executedCommands[len(t.executedCommands)-1]
			t.executedCommands = t.executedCommands[0 : len(t.executedCommands)-1]
			break
		}
	}
	t.lock.Unlock()

	if completions != nil {
		completions <- (err == nil)
	}

	return err == nil
}

func (t *testRunner) abort(sig os.Signal) {
	t.lock.Lock()
	for _, cmd := range t.executedCommands {
		if cmd.Process != nil {
			cmd.Process.Signal(sig)
			cmd.Wait()
		}
	}
	t.lock.Unlock()
	t.cleanUpCompiledSuite()
}
