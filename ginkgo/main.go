package main

import (
	"fmt"
	"github.com/onsi/ginkgo/config"

	"bytes"
	"flag"
	"io"
	"os"
	"os/exec"
)

//Add -R option to recursively run all tests under . (search for dirs that contain _test files in them)

var numCPU int
var reports []*bytes.Buffer

func init() {
	config.Flags("", false)

	flag.IntVar(&(numCPU), "nodes", 1, "The number of parallel test nodes to run")

	flag.Parse()
}

func main() {
	reports = make([]*bytes.Buffer, 0)

	passed := runSuiteAtPath(".")

	if passed {
		os.Exit(0)
	} else {
		os.Exit(1)
	}
}

func runSuiteAtPath(path string) bool {
	fmt.Printf("Running suite at %s\n\n", path)

	completions := make(chan bool)
	for cpu := 0; cpu < numCPU; cpu++ {
		config.GinkgoConfig.ParallelNode = cpu + 1
		config.GinkgoConfig.ParallelTotal = numCPU

		args := config.BuildFlagArgs("ginkgo", config.GinkgoConfig, config.DefaultReporterConfig)

		var writer io.Writer
		if numCPU > 1 {
			buffer := new(bytes.Buffer)
			reports = append(reports, buffer)
			writer = buffer
		} else {
			writer = os.Stdout
		}

		go runCommand(path, args, writer, completions)
	}

	passed := true

	for cpu := 0; cpu < numCPU; cpu++ {
		passed = passed && <-completions
	}

	if numCPU > 1 {
		printToScreen()
	}

	return passed
}

func printToScreen() {
	for _, report := range reports {
		fmt.Print(report.String())
	}
	os.Stdout.Sync()
}

func runCommand(path string, args []string, stream io.Writer, completions chan bool) {
	args = append([]string{"test", "-v"}, args...)
	args = append(args, path)

	cmd := exec.Command("go", args...)

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	go io.Copy(stream, stdout)
	go io.Copy(stream, stderr)

	err := cmd.Start()
	if err != nil {
		os.Exit(1)
	}

	err = cmd.Wait()
	completions <- (err == nil)
}
