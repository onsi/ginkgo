package nodot

import (
	"bufio"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	"github.com/onsi/ginkgo/ginkgo/command"
	"github.com/onsi/ginkgo/ginkgo/internal"
)

func BuildNodotCommand() command.Command {
	return command.Command{
		Name:     "nodot",
		Usage:    "ginkgo nodot",
		ShortDoc: "Update the nodot declarations in your test suite",
		Documentation: `If you've bootstrapped a test suite using ginkgo bootstrap -nodot, run ginkgo nodot in a suite's direcotry to update the declarations in the generated bootstrap file.

Any missing declarations (from, say, a recently added Gomega matcher) will be added to your bootstrap file.

If you've renamed a declaration, that name will be honored and not overwritten.`,
		DocLink: "avoiding-dot-imports",
		Command: func(_ []string, _ []string) {
			updateNodot()
		},
	}
}

func updateNodot() {
	suiteFile, perm := findSuiteFile()

	data, err := ioutil.ReadFile(suiteFile)
	command.AbortIfError("Failed to read boostrap file:", err)

	content, err := ApplyNoDot(data)
	command.AbortIfError("Failed to update nodot declarations:", err)

	ioutil.WriteFile(suiteFile, content, perm)

	internal.GoFmt(suiteFile)
}

func findSuiteFile() (string, os.FileMode) {
	workingDir, err := os.Getwd()
	command.AbortIfError("Could not find suite file for nodot:", err)

	files, err := ioutil.ReadDir(workingDir)
	command.AbortIfError("Could not find suite file for nodot:", err)

	re := regexp.MustCompile(`RunSpecs\(|RunSpecsWithDefaultAndCustomReporters\(|RunSpecsWithCustomReporters\(`)

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		path := filepath.Join(workingDir, file.Name())
		f, err := os.Open(path)
		command.AbortIfError("Could not find suite file for nodot:", err)
		defer f.Close()

		if re.MatchReader(bufio.NewReader(f)) {
			return path, file.Mode()
		}
	}

	command.AbortWith("Could not find a suite file for nodot: you need a bootstrap file that call's Ginkgo's RunSpecs() command.\nTry running ginkgo bootstrap first.")
	return "", 0
}
