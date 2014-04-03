package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

func BuildBootstrapCommand() *Command {
	return &Command{
		Name:         "bootstrap",
		FlagSet:      flag.NewFlagSet("bootstrap", flag.ExitOnError),
		UsageCommand: "ginkgo bootstrap",
		Usage:        []string{"Bootstrap a test suite for the current package"},
		Command:      generateBootstrap,
	}
}

var bootstrapText = `package {{.Package}}_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func Test{{.PackageTitleCase}}(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "{{.PackageTitleCase}} Suite")
}
`

type bootstrapData struct {
	Package          string
	PackageTitleCase string
}

func getPackage() string {
	workingDir, err := os.Getwd()
	if err != nil {
		panic(err.Error())
	}
	return filepath.Base(workingDir)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

func generateBootstrap(args []string) {
	packageName := getPackage()
	data := bootstrapData{
		Package:          packageName,
		PackageTitleCase: strings.Title(packageName),
	}

	targetFile := fmt.Sprintf("%s_suite_test.go", packageName)
	if fileExists(targetFile) {
		fmt.Printf("%s already exists.\n\n", targetFile)
		os.Exit(1)
	} else {
		fmt.Printf("Generating ginkgo test suite bootstrap for %s in:\n\t%s\n\n", packageName, targetFile)
	}

	f, err := os.Create(targetFile)
	if err != nil {
		panic(err.Error())
	}
	defer f.Close()

	bootstrapTemplate, err := template.New("bootstrap").Parse(bootstrapText)
	if err != nil {
		panic(err.Error())
	}

	bootstrapTemplate.Execute(f, data)
}
