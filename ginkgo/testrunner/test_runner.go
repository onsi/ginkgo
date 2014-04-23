package testrunner

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/ginkgo/testsuite"
	"github.com/onsi/ginkgo/internal/remote"
	"github.com/onsi/ginkgo/reporters/stenographer"
)

type TestRunner struct {
	suite *testsuite.TestSuite

	numCPU         int
	parallelStream bool
	race           bool
	cover          bool
	additionalArgs []string
}

func New(suite *testsuite.TestSuite, numCPU int, parallelStream bool, race bool, cover bool, additionalArgs []string) *TestRunner {
	return &TestRunner{
		suite:          suite,
		numCPU:         numCPU,
		parallelStream: parallelStream,
		race:           race,
		cover:          cover,
		additionalArgs: additionalArgs,
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
	os.Remove(t.compiledArtifact())
}

func (t *TestRunner) compiledArtifact() string {
	compiledArtifact, _ := filepath.Abs(filepath.Join(t.suite.Path, fmt.Sprintf("%s.test", t.suite.PackageName)))
	return compiledArtifact
}

func (t *TestRunner) runSerialGinkgoSuite() bool {
	ginkgoArgs := config.BuildFlagArgs("ginkgo", config.GinkgoConfig, config.DefaultReporterConfig)
	return t.run(t.cmd(ginkgoArgs, os.Stdout, 1), nil)
}

func (t *TestRunner) runGoTestSuite() bool {
	return t.run(t.cmd([]string{"-test.v"}, os.Stdout, 1), nil)
}

func (t *TestRunner) runAndStreamParallelGinkgoSuite() bool {
	completions := make(chan bool)
	writers := make([]*logWriter, t.numCPU)

	server, err := remote.NewServer(t.numCPU)
	if err != nil {
		panic("Failed to start parallel spec server")
	}

	server.Start()
	defer server.Close()

	for cpu := 0; cpu < t.numCPU; cpu++ {
		config.GinkgoConfig.ParallelNode = cpu + 1
		config.GinkgoConfig.ParallelTotal = t.numCPU
		config.GinkgoConfig.SyncHost = server.Address()

		ginkgoArgs := config.BuildFlagArgs("ginkgo", config.GinkgoConfig, config.DefaultReporterConfig)

		writers[cpu] = newLogWriter(os.Stdout, cpu+1)

		cmd := t.cmd(ginkgoArgs, writers[cpu], cpu+1)

		server.RegisterAlive(cpu+1, func() bool {
			if cmd.ProcessState == nil {
				return true
			}
			return !cmd.ProcessState.Exited()
		})

		go t.run(cmd, completions)
	}

	passed := true

	for cpu := 0; cpu < t.numCPU; cpu++ {
		passed = <-completions && passed
	}

	for _, writer := range writers {
		writer.Close()
	}

	os.Stdout.Sync()

	if t.cover {
		t.combineCoverprofiles()
	}

	return passed
}

func (t *TestRunner) runParallelGinkgoSuite() bool {
	result := make(chan bool)
	completions := make(chan bool)
	writers := make([]*logWriter, t.numCPU)
	reports := make([]*bytes.Buffer, t.numCPU)

	stenographer := stenographer.New(!config.DefaultReporterConfig.NoColor)
	aggregator := remote.NewAggregator(t.numCPU, result, config.DefaultReporterConfig, stenographer)

	server, err := remote.NewServer(t.numCPU)
	if err != nil {
		panic("Failed to start parallel spec server")
	}
	server.RegisterReporters(aggregator)
	server.Start()
	defer server.Close()

	for cpu := 0; cpu < t.numCPU; cpu++ {
		config.GinkgoConfig.ParallelNode = cpu + 1
		config.GinkgoConfig.ParallelTotal = t.numCPU
		config.GinkgoConfig.SyncHost = server.Address()
		config.GinkgoConfig.StreamHost = server.Address()

		ginkgoArgs := config.BuildFlagArgs("ginkgo", config.GinkgoConfig, config.DefaultReporterConfig)

		reports[cpu] = &bytes.Buffer{}
		writers[cpu] = newLogWriter(reports[cpu], cpu+1)

		cmd := t.cmd(ginkgoArgs, writers[cpu], cpu+1)

		server.RegisterAlive(cpu+1, func() bool {
			if cmd.ProcessState == nil {
				return true
			}
			return !cmd.ProcessState.Exited()
		})

		go t.run(cmd, completions)
	}

	passed := true

	for cpu := 0; cpu < t.numCPU; cpu++ {
		passed = <-completions && passed
	}

	//all test processes are done, at this point
	//we should be able to wait for the aggregator to tell us that it's done

	select {
	case <-result:
		fmt.Println("")
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

		for _, writer := range writers {
			writer.Close()
		}

		for _, report := range reports {
			fmt.Print(report.String())
		}

		os.Stdout.Sync()
	}

	if t.cover {
		t.combineCoverprofiles()
	}

	return passed
}

func (t *TestRunner) cmd(ginkgoArgs []string, stream io.Writer, node int) *exec.Cmd {
	args := []string{"-test.timeout=24h"}
	if t.cover {
		coverprofile := "--test.coverprofile=" + t.suite.PackageName + ".coverprofile"
		if t.numCPU > 1 {
			coverprofile = fmt.Sprintf("%s.%d", coverprofile, node)
		}
		args = append(args, coverprofile)
	}

	args = append(args, ginkgoArgs...)
	args = append(args, t.additionalArgs...)

	cmd := exec.Command(t.compiledArtifact(), args...)

	cmd.Dir = t.suite.Path
	cmd.Stderr = stream
	cmd.Stdout = stream

	return cmd
}

func (t *TestRunner) run(cmd *exec.Cmd, completions chan bool) bool {
	var err error
	defer func() {
		if completions != nil {
			completions <- (err == nil)
		}
	}()

	err = cmd.Start()
	if err != nil {
		fmt.Printf("Failed to run test suite!\n\t%s", err.Error())
		return false
	}

	err = cmd.Wait()

	return err == nil
}

func (t *TestRunner) combineCoverprofiles() {
	profiles := []string{}
	for cpu := 1; cpu <= t.numCPU; cpu++ {
		coverFile := fmt.Sprintf("%s.coverprofile.%d", t.suite.PackageName, cpu)
		coverFile = filepath.Join(t.suite.Path, coverFile)
		coverProfile, err := ioutil.ReadFile(coverFile)
		os.Remove(coverFile)

		if err == nil {
			profiles = append(profiles, string(coverProfile))
		}
	}

	if len(profiles) != t.numCPU {
		return
	}

	lines := map[string]int{}

	for _, coverProfile := range profiles {
		for _, line := range strings.Split(string(coverProfile), "\n")[1:] {
			if len(line) == 0 {
				continue
			}
			components := strings.Split(line, " ")
			count, _ := strconv.Atoi(components[len(components)-1])
			prefix := strings.Join(components[0:len(components)-1], " ")
			lines[prefix] += count
		}
	}

	output := []string{"mode: atomic"}
	for line, count := range lines {
		output = append(output, fmt.Sprintf("%s %d", line, count))
	}
	finalOutput := strings.Join(output, "\n")
	ioutil.WriteFile(filepath.Join(t.suite.Path, fmt.Sprintf("%s.coverprofile", t.suite.PackageName)), []byte(finalOutput), 0666)
}
