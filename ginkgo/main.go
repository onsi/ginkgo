package main

import (
	"fmt"
	"github.com/onsi/ginkgo/config"
	"path/filepath"
	"regexp"

	"bytes"
	"flag"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
)

var numCPU int
var recurse bool
var runMagicI bool
var race bool
var cover bool
var reports []*bytes.Buffer
var executedCommands []*exec.Cmd

type suite struct {
	path        string
	packageName string
	isGinkgo    bool
}

func init() {
	config.Flags("", false)

	flag.IntVar(&(numCPU), "nodes", 1, "The number of parallel test nodes to run")
	flag.BoolVar(&(recurse), "r", false, "Find and run test suites under the current directory recursively")
	flag.BoolVar(&(runMagicI), "i", false, "Run go test -i first, then run the test suite")
	flag.BoolVar(&(race), "race", false, "Run tests with race detection enabled")
	flag.BoolVar(&(cover), "cover", false, "Run tests with coverage analysis, will generate coverage profiles with the package name in the current directory")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of ginkgo:\n\n")
		fmt.Fprintf(os.Stderr, "ginkgo\n  Run the tests in the current directory.  The following flags are available:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "ginkgo bootstrap\n  Bootstrap a test suite for the current package.\n\n")
		fmt.Fprintf(os.Stderr, "ginkgo generate <SUBJECT>\n  Generate a test file for SUBJECT, the file will be named SUBJECT_test.go\n  If omitted, a file named after the package will be created.\n\n")
		fmt.Fprintf(os.Stderr, "ginkgo version\n  Print ginkgo's version.\n\n")
		fmt.Fprintf(os.Stderr, "ginkgo help\n  Print this usage information.\n")
	}

	flag.Parse()
}

func main() {
	if flag.NArg() > 0 {
		handleSubcommands(flag.Args())
	}

	executedCommands = make([]*exec.Cmd, 0)
	reports = make([]*bytes.Buffer, 0)
	registerSignalHandler()

	passed := true

	suites := findSuitesInDir(".", recurse)

	for _, suite := range suites {
		passed = passed && runSuite(suite)
	}

	if passed {
		os.Exit(0)
	} else {
		fmt.Printf("\nTest Suite Failed\n")
		os.Exit(1)
	}
}

func handleSubcommands(args []string) {
	if args[0] == "bootstrap" {
		generateBootstrap()
		os.Exit(0)
	} else if args[0] == "generate" {
		subject := ""
		if len(args) > 1 {
			subject = args[1]
		}
		generateSpec(subject)
		os.Exit(0)
	} else if args[0] == "help" {
		flag.Usage()
		os.Exit(0)
	} else if args[0] == "version" {
		fmt.Printf("Ginkgo V%s\n", config.VERSION)
		os.Exit(0)
	} else {
		fmt.Printf("Unkown command %s\n\n", args[0])
		flag.Usage()

		os.Exit(1)
	}
}

func findSuitesInDir(dir string, recurse bool) []suite {
	suites := []suite{}
	files, _ := ioutil.ReadDir(dir)
	re := regexp.MustCompile(`_test\.go$`)
	for _, file := range files {
		if !file.IsDir() && re.Match([]byte(file.Name())) {
			suites = append(suites, suite{path: dir, packageName: packageNameForDir(dir), isGinkgo: filesHaveGinkgoSuite(dir, files)})
			break
		}
	}

	if recurse {
		re = regexp.MustCompile(`^\.`)
		for _, file := range files {
			if file.IsDir() && !re.Match([]byte(file.Name())) {
				suites = append(suites, findSuitesInDir(dir+"/"+file.Name(), recurse)...)
			}
		}
	}

	return suites
}

func packageNameForDir(dir string) string {
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

func runSuite(suite suite) bool {
	completions := make(chan bool)

	if runMagicI {
		runGoI(suite.path, race)
	}

	if suite.isGinkgo {
		for cpu := 0; cpu < numCPU; cpu++ {
			config.GinkgoConfig.ParallelNode = cpu + 1
			config.GinkgoConfig.ParallelTotal = numCPU

			args := config.BuildFlagArgs("ginkgo", config.GinkgoConfig, config.DefaultReporterConfig)
			if race {
				args = append([]string{"--race"}, args...)
			}
			if cover {
				args = append([]string{"--cover", "--coverprofile=" + suite.packageName + ".coverprofile"})
			}

			var writer io.Writer
			if numCPU > 1 {
				buffer := new(bytes.Buffer)
				reports = append(reports, buffer)
				writer = buffer
			} else {
				writer = os.Stdout
			}

			go runCommand(suite.path, args, writer, completions)
		}

		passed := true

		for cpu := 0; cpu < numCPU; cpu++ {
			passed = <-completions && passed
		}

		if numCPU > 1 {
			printToScreen()
		}

		return passed
	} else {
		args := []string{}
		if race {
			args = append(args, "--race")
		}
		if cover {
			args = append([]string{"--cover", "--coverprofile=" + suite.packageName + ".out"})
		}
		go runCommand(suite.path, args, os.Stdout, completions)
		return <-completions
	}
}

func printToScreen() {
	for _, report := range reports {
		fmt.Print(report.String())
	}
	os.Stdout.Sync()
}

func runGoI(path string, race bool) {
	args := []string{"test", "-i"}
	if race {
		args = append(args, "-race")
	}
	args = append(args, path)
	cmd := exec.Command("go", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("go test -i %s failed with:\n%s", path, output)
		os.Exit(1)
	}
}

func runCommand(path string, args []string, stream io.Writer, completions chan bool) {
	args = append([]string{"test", "-v", "-timeout=24h", path}, args...)

	cmd := exec.Command("go", args...)
	executedCommands = append(executedCommands, cmd)

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
	completions <- (err == nil)
}

func registerSignalHandler() {
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, os.Kill)

		select {
		case sig := <-c:
			for _, cmd := range executedCommands {
				cmd.Process.Signal(sig)
			}
			os.Exit(1)
		}
	}()
}
