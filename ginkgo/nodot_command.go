package main

import (
	"bufio"
	"flag"
	"os"
	"path/filepath"
	"regexp"

	"github.com/onsi/ginkgo/ginkgo/nodot"
)

func BuildNodotCommand() *Command {
	return &Command{
		Name:         "nodot",
		FlagSet:      flag.NewFlagSet("bootstrap", flag.ExitOnError),
		UsageCommand: "ginkgo nodot",
		Usage: []string{
			"Update the nodot declarations in your test suite",
			"Any missing declarations (from, say, a recently added matcher) will be added to your bootstrap file.",
			"If you've renamed a declaration, that name will be honored and not overwritten.",
		},
		Command: updateNodot,
	}
}

func updateNodot(args []string, additionalArgs []string) {
	suiteFile, perm := findSuiteFile()

	data, err := os.ReadFile(suiteFile)
	if err != nil {
		complainAndQuit("Failed to update nodot declarations: " + err.Error())
	}

	content, err := nodot.ApplyNoDot(data)
	if err != nil {
		complainAndQuit("Failed to update nodot declarations: " + err.Error())
	}
	os.WriteFile(suiteFile, content, perm)

	goFmt(suiteFile)
}

func findSuiteFile() (string, os.FileMode) {
	workingDir, err := os.Getwd()
	if err != nil {
		complainAndQuit("Could not find suite file for nodot: " + err.Error())
	}

	files, err := os.ReadDir(workingDir)
	if err != nil {
		complainAndQuit("Could not find suite file for nodot: " + err.Error())
	}

	re := regexp.MustCompile(`RunSpecs\(|RunSpecsWithDefaultAndCustomReporters\(|RunSpecsWithCustomReporters\(`)

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		path := filepath.Join(workingDir, file.Name())
		f, err := os.Open(path)
		if err != nil {
			complainAndQuit("Could not find suite file for nodot: " + err.Error())
		}
		defer f.Close()

		if re.MatchReader(bufio.NewReader(f)) {
			return path, file.Type()
		}
	}

	complainAndQuit("Could not find a suite file for nodot: you need a bootstrap file that call's Ginkgo's RunSpecs() command.\nTry running ginkgo bootstrap first.")

	return "", 0
}
