package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
)

type testSuite struct {
	path        string
	packageName string
	isGinkgo    bool
}

func newSuite(dir string, files []os.FileInfo) testSuite {
	dir = "." + string(filepath.Separator) + filepath.Clean(dir)
	return testSuite{
		path:        dir,
		packageName: packageNameForSuite(dir),
		isGinkgo:    filesHaveGinkgoSuite(dir, files),
	}
}

func suitesInDir(dir string, recurse bool) []testSuite {
	suites := []testSuite{}
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
				suites = append(suites, suitesInDir(dir+"/"+file.Name(), recurse)...)
			}
		}
	}

	return suites
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
