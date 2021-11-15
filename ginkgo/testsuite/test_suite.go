package testsuite

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

type TestSuite struct {
	Path        string
	PackageName string
	IsGinkgo    bool
	Precompiled bool
}

func PrecompiledTestSuite(path string) (TestSuite, error) {
	info, err := os.Stat(path)
	if err != nil {
		return TestSuite{}, err
	}

	if info.IsDir() {
		return TestSuite{}, errors.New("this is a directory, not a file")
	}

	if (runtime.GOOS != "windows" && filepath.Ext(path) != ".test") ||
		(runtime.GOOS == "windows" && !strings.HasSuffix(path, ".test.exe")) {
		return TestSuite{}, errors.New("this is not a .test (.test.exe) binary")
	}

	if runtime.GOOS != "windows" && info.Mode()&0111 == 0 {
		return TestSuite{}, errors.New("this is not executable")
	}

	dir := relPath(filepath.Dir(path))
	var packageName string
	if strings.HasSuffix(path, ".test.exe") {
		packageName = strings.TrimSuffix(filepath.Base(path), ".test.exe")
	} else {
		packageName = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	}

	return TestSuite{
		Path:        dir,
		PackageName: packageName,
		IsGinkgo:    true,
		Precompiled: true,
	}, nil
}

func SuitesInDir(dir string, recurse bool) []TestSuite {
	suites := []TestSuite{}

	if vendorExperimentCheck(dir) {
		return suites
	}

	files, _ := os.ReadDir(dir)
	re := regexp.MustCompile(`^[^._].*_test\.go$`)
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

func relPath(dir string) string {
	dir, _ = filepath.Abs(dir)
	cwd, _ := os.Getwd()
	dir, _ = filepath.Rel(cwd, filepath.Clean(dir))

	if string(dir[0]) != "." {
		dir = "." + string(filepath.Separator) + dir
	}

	return dir
}

func New(dir string, files []os.DirEntry) TestSuite {
	return TestSuite{
		Path:        relPath(dir),
		PackageName: packageNameForSuite(dir),
		IsGinkgo:    filesHaveGinkgoSuite(dir, files),
	}
}

func packageNameForSuite(dir string) string {
	path, _ := filepath.Abs(dir)
	return filepath.Base(path)
}

func filesHaveGinkgoSuite(dir string, files []os.DirEntry) bool {
	reTestFile := regexp.MustCompile(`_test\.go$`)
	reGinkgo := regexp.MustCompile(`package ginkgo|\/ginkgo"`)

	for _, file := range files {
		if !file.IsDir() && reTestFile.Match([]byte(file.Name())) {
			contents, _ := os.ReadFile(dir + "/" + file.Name())
			if reGinkgo.Match(contents) {
				return true
			}
		}
	}

	return false
}
