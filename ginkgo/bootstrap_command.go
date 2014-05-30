package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/onsi/ginkgo/ginkgo/nodot"
)

func BuildBootstrapCommand() *Command {
	var noDot bool
	flagSet := flag.NewFlagSet("bootstrap", flag.ExitOnError)
	flagSet.BoolVar(&noDot, "nodot", false, "If set, bootstrap will generate a bootstrap file that does not . import ginkgo and gomega")

	return &Command{
		Name:         "bootstrap",
		FlagSet:      flagSet,
		UsageCommand: "ginkgo bootstrap <FLAGS>",
		Usage: []string{
			"Bootstrap a test suite for the current package",
			"Accepts the following flags:",
		},
		Command: func(args []string, additionalArgs []string) {
			generateBootstrap(noDot)
		},
	}
}

var bootstrapText = `package {{.Package}}_test

import (
	{{.GinkgoImport}}
	{{.GomegaImport}}

	"testing"
)

func Test{{.FormattedPackage}}(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "{{.FormattedPackage}} Suite")
}
`

type bootstrapData struct {
	Package          string
	FormattedPackage string
	GinkgoImport     string
	GomegaImport     string
}

func getPackage() string {
	workingDir, err := os.Getwd()
	if err != nil {
		complainAndQuit("Could not find package: " + err.Error())
	}
	packageName := filepath.Base(workingDir)
	return strings.Replace(packageName, "-", "_", -1)
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

func generateBootstrap(noDot bool) {
	packageName := getPackage()
	formattedPackage := strings.Replace(strings.Title(strings.Replace(packageName, "_", " ", -1)), " ", "", -1)
	data := bootstrapData{
		Package:          packageName,
		FormattedPackage: formattedPackage,
		GinkgoImport:     `. "github.com/onsi/ginkgo"`,
		GomegaImport:     `. "github.com/onsi/gomega"`,
	}

	if noDot {
		data.GinkgoImport = `"github.com/onsi/ginkgo"`
		data.GomegaImport = `"github.com/onsi/gomega"`
	}

	targetFile := fmt.Sprintf("%s_suite_test.go", packageName)
	if fileExists(targetFile) {
		fmt.Printf("%s already exists.\n\n", targetFile)
		os.Exit(1)
	} else {
		fmt.Printf("Generating ginkgo test suite bootstrap for %s in:\n\t%s\n", packageName, targetFile)
	}

	f, err := os.Create(targetFile)
	if err != nil {
		complainAndQuit("Could not create file: " + err.Error())
		panic(err.Error())
	}
	defer f.Close()

	bootstrapTemplate, err := template.New("bootstrap").Parse(bootstrapText)
	if err != nil {
		panic(err.Error())
	}

	buf := &bytes.Buffer{}
	bootstrapTemplate.Execute(buf, data)

	if noDot {
		contents, err := nodot.ApplyNoDot(buf.Bytes())
		if err != nil {
			complainAndQuit("Failed to import nodot declarations: " + err.Error())
		}
		fmt.Println("To update the nodot declarations in the future, switch to this directory and run:\n\tginkgo nodot")
		buf = bytes.NewBuffer(contents)
	}

	buf.WriteTo(f)

	goFmt(targetFile)
}
