package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

func BuildGenerateCommand() *Command {
	return &Command{
		Name:         "generate",
		FlagSet:      flag.NewFlagSet("generate", flag.ExitOnError),
		UsageCommand: "ginkgo generate <filename>",
		Usage: []string{
			"Generate a test file named filename_test.go",
			"If the optional <filename> argument is omitted, a file named after the package in the current directory will be created.",
		},
		Command: generateSpec,
	}
}

var specText = `package {{.Package}}_test

import (
	. "{{.PackageImportPath}}"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("{{.Subject}}", func() {

})
`

type specData struct {
	Package           string
	Subject           string
	PackageImportPath string
}

func generateSpec(args []string) {
	subject := ""
	if len(args) > 0 {
		subject = args[0]
	}

	packageName := getPackage()
	if subject == "" {
		subject = packageName
	} else {
		subject = strings.Split(subject, ".go")[0]
		subject = strings.Split(subject, "_test")[0]
	}

	formattedSubject := strings.Replace(strings.Title(strings.Replace(subject, "_", " ", -1)), " ", "", -1)

	data := specData{
		Package:           packageName,
		Subject:           formattedSubject,
		PackageImportPath: getPackageImportPath(),
	}

	targetFile := fmt.Sprintf("%s_test.go", subject)
	if fileExists(targetFile) {
		fmt.Printf("%s already exists.\n\n", targetFile)
		os.Exit(1)
	} else {
		fmt.Printf("Generating ginkgo test for %s in:\n\t%s\n\n", data.Subject, targetFile)
	}

	f, err := os.Create(targetFile)
	if err != nil {
		panic(err.Error())
	}
	defer f.Close()

	specTemplate, err := template.New("spec").Parse(specText)
	if err != nil {
		panic(err.Error())
	}

	specTemplate.Execute(f, data)
}

func getPackageImportPath() string {
	workingDir, err := os.Getwd()
	if err != nil {
		panic(err.Error())
	}
	sep := string(filepath.Separator)
	paths := strings.Split(workingDir, sep+"src"+sep)
	if len(paths) == 1 {
		fmt.Printf("\nCouldn't identify package import path.\n\n\tginkgo generate\n\nMust be run within a package directory under $GOPATH/src/...\nYou're going to have to change UNKNOWN_PACKAGE_PATH in the generated file...\n\n")
		return "UNKNOWN_PACKAGE_PATH"
	}
	return filepath.ToSlash(paths[len(paths)-1])
}
