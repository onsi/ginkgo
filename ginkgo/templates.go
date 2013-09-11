package main

import (
	"fmt"
	"os"
	"path"
	"strings"
	"text/template"
)

var bootstrapText = `package {{.Package}}

import (
    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"

    "testing"
)

func BootstrapTest(t *testing.T) {
    RegisterFailHandler(Fail)
    RunSpecs(t, "{{.PackageTitleCase}} Suite")
}
`

var specText = `package {{.Package}}

import (
    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"
)

var _ = Describe("{{.Subject}}", func() {

})
`

type bootstrapData struct {
	Package          string
	PackageTitleCase string
}

type specData struct {
	Package string
	Subject string
}

func generateBootstrap() {
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

func generateSpec(subject string) {
	packageName := getPackage()
	if subject == "" {
		subject = packageName
	}

	formattedSubject := strings.Replace(strings.Title(strings.Replace(subject, "_", " ", -1)), " ", "", -1)

	data := specData{
		Package: packageName,
		Subject: formattedSubject,
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

func getPackage() string {
	workingDir, err := os.Getwd()
	if err != nil {
		panic(err.Error())
	}
	return path.Base(workingDir)
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
