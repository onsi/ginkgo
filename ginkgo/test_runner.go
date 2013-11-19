package main

import (
	"bytes"
	"fmt"
	"github.com/onsi/ginkgo/config"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
)

type suite struct {
	path        string
	packageName string
	isGinkgo    bool
}

func newSuite(dir string, files []os.FileInfo) suite {
	return suite{
		path:        dir,
		packageName: packageNameForSuite(dir),
		isGinkgo:    filesHaveGinkgoSuite(dir, files),
	}
}

func packageNameForSuite(dir string) string {
	path, _ := filepath.Abs(dir)
	return filepath.Base(path)
}

func filesHaveGinkgoSuite(dir string, files []os.FileInfo) bool {
	reTestFile := regexp.MustCompile(`_test\.go$`)
	reGinkgo := regexp.MustCompile(`package ginkgo|\/ginkgo"`)
	for _, file := range files {
		if !file.IsDir() && reTestFile.Match([]byte(file.Name())) {
			contents, _ := ioutil.ReadFile(dir + "/" + file.Name())
			if reGinkgo.Match(contents) {
				return true
			}
		}
	}

	return false
}

type testRunner struct {
	numCPU           int
	recurse          bool
	runMagicI        bool
	race             bool
	cover            bool
	executedCommands []*exec.Cmd
	reports          []*bytes.Buffer
}

func newTestRunner(numCPU int, recurse bool, runMagicI bool, race bool, cover bool) *testRunner {
	return &testRunner{
		numCPU:           numCPU,
		recurse:          recurse,
		runMagicI:        runMagicI,
		race:             race,
		cover:            cover,
		executedCommands: []*exec.Cmd{},
		reports:          []*bytes.Buffer{},
	}
}

func (t *testRunner) run() bool {
	t.registerSignalHandler()

	suites := t.findSuitesInDir(".", t.recurse)

	for _, suite := range suites {
		if !t.runSuite(suite) {
			return false
		}
	}

	return true
}

func (t *testRunner) findSuitesInDir(dir string, recurse bool) []suite {
	suites := []suite{}
	files, _ := ioutil.ReadDir(dir)
	re := regexp.MustCompile(`_test\.go$`)
	for _, file := range files {
		if !file.IsDir() && re.Match([]byte(file.Name())) {
			suites = append(suites, newSuite(dir, files))
			break
		}
	}

	if recurse {
		re = regexp.MustCompile(`^\.`)
		for _, file := range files {
			if file.IsDir() && !re.Match([]byte(file.Name())) {
				suites = append(suites, t.findSuitesInDir(dir+"/"+file.Name(), recurse)...)
			}
		}
	}

	return suites
}

func (t *testRunner) runSuite(suite suite) bool {
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

func (t *testRunner) runGoI(suite suite) {
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

func (t *testRunner) runParallelGinkgoSuite(suite suite) bool {
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

func (t *testRunner) runSerialGinkgoSuite(suite suite) bool {
	args := config.BuildFlagArgs("ginkgo", config.GinkgoConfig, config.DefaultReporterConfig)
	args = append(args, t.commonArgs(suite)...)
	return t.runCommand(suite.path, args, os.Stdout, nil)
}

func (t *testRunner) runGoTestSuite(suite suite) bool {
	args := t.commonArgs(suite)
	return t.runCommand(suite.path, args, os.Stdout, nil)
}

func (t *testRunner) commonArgs(suite suite) []string {
	args := []string{}
	if t.race {
		args = append(args, "--race")
	}
	if t.cover {
		args = append([]string{"--cover", "--coverprofile=" + suite.packageName + ".out"})
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
