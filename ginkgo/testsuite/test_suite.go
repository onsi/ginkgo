package testsuite

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
)

type TestSuite struct {
	Path        string
	PackageName string
	IsGinkgo    bool
}

func SuitesInDir(dir string, recurse bool) []TestSuite {
	suites := []TestSuite{}
	files, _ := ioutil.ReadDir(dir)
	re := regexp.MustCompile(`_test\.go$`)
	for _, file := range files {
		if !file.IsDir() && re.Match([]byte(file.Name())) {
			suites = append(suites, New(dir, files))
			break
		}
	}

	if recurse {
		re = regexp.MustCompile(`^[._]`)
		for _, file := range files {
			if file.IsDir() && !re.Match([]byte(file.Name())) {
				suites = append(suites, SuitesInDir(dir+"/"+file.Name(), recurse)...)
			}
		}
	}

	return suites
}

func New(dir string, files []os.FileInfo) TestSuite {
	dir, _ = filepath.Abs(dir)
	cwd, _ := os.Getwd()
	dir, _ = filepath.Rel(cwd, filepath.Clean(dir))
	dir = "." + string(filepath.Separator) + dir

	return TestSuite{
		Path:        dir,
		PackageName: packageNameForSuite(dir),
		IsGinkgo:    filesHaveGinkgoSuite(dir, files),
	}
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
