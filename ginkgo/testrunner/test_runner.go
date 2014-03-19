package testrunner

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
	"syscall"
	"time"
)

type TestRunner struct {
	suite *testsuite.TestSuite

	numCPU         int
	parallelStream bool
	race           bool
	cover          bool

	executedCommands []*exec.Cmd

	lock *sync.Mutex
}

func New(suite *testsuite.TestSuite, numCPU int, parallelStream bool, race bool, cover bool) *TestRunner {
	return &TestRunner{
		suite:            suite,
		numCPU:           numCPU,
		parallelStream:   parallelStream,
		race:             race,
		cover:            cover,
		executedCommands: []*exec.Cmd{},
		lock:             &sync.Mutex{},
	}
}

func (t *TestRunner) Compile() error {
	os.Remove(t.compiledArtifact())

	args := []string{"test", "-c", "-i"}
	if t.race {
		args = append(args, "-race")
	}
	if t.cover {
		args = append(args, "-cover", "-covermode=atomic")
	}

	cmd := exec.Command("go", args...)
	t.registerCmd(cmd)
	defer t.unregisterCmd(cmd)

	cmd.Dir = t.suite.Path

	output, err := cmd.CombinedOutput()

	if err != nil {
		if len(output) > 0 {
			return fmt.Errorf("Failed to compile %s:\n\n%s", t.suite.PackageName, output)
		}
		return fmt.Errorf("")
	}

	return nil
}

func (t *TestRunner) Run() bool {
	var success bool

	if t.suite.IsGinkgo {
		if t.numCPU > 1 {
			if t.parallelStream {
				success = t.runAndStreamParallelGinkgoSuite()
			} else {
				success = t.runParallelGinkgoSuite()
			}
		} else {
			success = t.runSerialGinkgoSuite()
		}
	} else {
		success = t.runGoTestSuite()
	}

	return success
}

func (t *TestRunner) CleanUp(signal ...os.Signal) {
	t.lock.Lock()
	for _, cmd := range t.executedCommands {
		if cmd.Process != nil {
			if len(signal) == 0 {
				cmd.Process.Signal(syscall.SIGINT)
			} else {
				cmd.Process.Signal(signal[0])
			}
			cmd.Wait()
		}
	}
	os.Remove(t.compiledArtifact())
	t.lock.Unlock()
}

func (t *TestRunner) compiledArtifact() string {
	compiledArtifact, _ := filepath.Abs(filepath.Join(t.suite.Path, fmt.Sprintf("%s.test", t.suite.PackageName)))
	return compiledArtifact
}

func (t *TestRunner) runSerialGinkgoSuite() bool {
	ginkgoArgs := config.BuildFlagArgs("ginkgo", config.GinkgoConfig, config.DefaultReporterConfig)
	return t.run(ginkgoArgs, nil, os.Stdout, nil)
}

func (t *TestRunner) runGoTestSuite() bool {
	return t.run([]string{"-test.v"}, nil, os.Stdout, nil)
}

func (t *TestRunner) runAndStreamParallelGinkgoSuite() bool {
	completions := make(chan bool)
	writers := make([]*logWriter, t.numCPU)

	for cpu := 0; cpu < t.numCPU; cpu++ {
		config.GinkgoConfig.ParallelNode = cpu + 1
		config.GinkgoConfig.ParallelTotal = t.numCPU

		ginkgoArgs := config.BuildFlagArgs("ginkgo", config.GinkgoConfig, config.DefaultReporterConfig)

		writers[cpu] = newLogWriter(fmt.Sprintf("[%d]", cpu+1))

		go t.run(ginkgoArgs, nil, writers[cpu], completions)
	}

	passed := true

	for cpu := 0; cpu < t.numCPU; cpu++ {
		passed = <-completions && passed
	}

	for _, writer := range writers {
		writer.Close()
	}

	os.Stdout.Sync()

	return passed
}

func (t *TestRunner) runParallelGinkgoSuite() bool {
	result := make(chan bool)
	completions := make(chan bool)
	reports := make([]*bytes.Buffer, t.numCPU)

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

	for cpu := 0; cpu < t.numCPU; cpu++ {
		config.GinkgoConfig.ParallelNode = cpu + 1
		config.GinkgoConfig.ParallelTotal = t.numCPU

		ginkgoArgs := config.BuildFlagArgs("ginkgo", config.GinkgoConfig, config.DefaultReporterConfig)

		env := os.Environ()
		env = append(env, fmt.Sprintf("GINKGO_REMOTE_REPORTING_SERVER=%s", serverAddress))

		reports[cpu] = &bytes.Buffer{}
		go t.run(ginkgoArgs, env, reports[cpu], completions)
	}

	for cpu := 0; cpu < t.numCPU; cpu++ {
		<-completions
	}

	//all test processes are done, at this point
	//we should be able to wait for the aggregator to tell us that it's done

	var passed = false
	select {
	case passed = <-result:
		fmt.Println("")
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

		for _, report := range reports {
			fmt.Print(report.String())
		}

		os.Stdout.Sync()
	}

	return passed
}

func (t *TestRunner) run(ginkgoArgs []string, env []string, stream io.Writer, completions chan bool) bool {
	var err error
	defer func() {
		if completions != nil {
			completions <- (err == nil)
		}
	}()

	args := []string{"-test.timeout=24h"}
	if t.cover {
		args = append(args, "--test.coverprofile="+t.suite.PackageName+".coverprofile")
	}

	args = append(args, ginkgoArgs...)

	cmd := exec.Command(t.compiledArtifact(), args...)
	t.registerCmd(cmd)
	defer t.unregisterCmd(cmd)

	cmd.Env = env
	cmd.Dir = t.suite.Path
	cmd.Stderr = stream
	cmd.Stdout = stream

	err = cmd.Start()
	if err != nil {
		fmt.Printf("Failed to run test suite!\n\t%s", err.Error())
		return false
	}

	err = cmd.Wait()

	return err == nil
}

func (t *TestRunner) registerCmd(cmd *exec.Cmd) {
	t.lock.Lock()
	t.executedCommands = append(t.executedCommands, cmd)
	t.lock.Unlock()
}

func (t *TestRunner) unregisterCmd(cmd *exec.Cmd) {
	t.lock.Lock()
	for commandIndex, executedCommand := range t.executedCommands {
		if executedCommand == cmd {
			t.executedCommands[commandIndex] = t.executedCommands[len(t.executedCommands)-1]
			t.executedCommands = t.executedCommands[0 : len(t.executedCommands)-1]
			break
		}
	}
	t.lock.Unlock()
}
