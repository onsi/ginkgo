package main

import (
	"fmt"
	"github.com/onsi/ginkgo/config"
	"regexp"

	"bytes"
	"flag"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
)

var numCPU int
var recurse bool
var reports []*bytes.Buffer

func init() {
	config.Flags("", false)

	flag.IntVar(&(numCPU), "nodes", 1, "The number of parallel test nodes to run")
	flag.BoolVar(&(recurse), "r", false, "Find test suites under the current directory recursively")

	flag.Parse()
}

func main() {
	reports = make([]*bytes.Buffer, 0)

	passed := true

	dirs := []string{"."}
	if recurse {
		dirs = findSuitesInDir(".")
	}

	for i, dir := range dirs {
		if i > 0 {
			fmt.Print("\n~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~\n")
		}
		passed = passed && runSuiteAtPath(dir)
	}

	if passed {
		os.Exit(0)
	} else {
		os.Exit(1)
	}
}

func findSuitesInDir(dir string) []string {
	dirs := []string{}
	files, _ := ioutil.ReadDir(dir)
	re := regexp.MustCompile(`_test\.go$`)
	for _, file := range files {
		if !file.IsDir() && re.Match([]byte(file.Name())) {
			dirs = append(dirs, dir)
			break
		}
	}

	re = regexp.MustCompile(`^\.`)
	for _, file := range files {
		if file.IsDir() && !re.Match([]byte(file.Name())) {
			dirs = append(dirs, findSuitesInDir(dir+"/"+file.Name())...)
		}
	}

	return dirs
}

func runSuiteAtPath(path string) bool {
	fmt.Printf("\nRunning suite at %s\n\n", path)

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
	args = append([]string{"test", "-v", path}, args...)

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
