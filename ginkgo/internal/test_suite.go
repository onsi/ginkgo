package internal

import (
	"errors"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/onsi/ginkgo/config"
)

type TestSuite struct {
	Path        string
	PackageName string
	IsGinkgo    bool
	Precompiled bool

	PathToCompiledTest   string
	CompilationError     error
	Passed               bool
	HasProgrammaticFocus bool
}

func (ts TestSuite) NamespacedName() string {
	name := relPath(ts.Path)
	name = strings.TrimLeft(name, "."+string(filepath.Separator))
	name = strings.ReplaceAll(name, string(filepath.Separator), "_")
	name = strings.ReplaceAll(name, " ", "_")
	if name == "" {
		return ts.PackageName
	}
	return name
}

func FindSuites(args []string, cliConfig config.GinkgoCLIConfigType, allowPrecompiled bool) ([]TestSuite, []string) {
	suites := []TestSuite{}

	if len(args) > 0 {
		for _, arg := range args {
			if allowPrecompiled {
				suite, err := precompiledTestSuite(arg)
				if err == nil {
					suites = append(suites, suite)
					continue
				}
			}
			recurseForSuite := cliConfig.Recurse
			if strings.HasSuffix(arg, "/...") && arg != "/..." {
				arg = arg[:len(arg)-4]
				recurseForSuite = true
			}
			suites = append(suites, suitesInDir(arg, recurseForSuite)...)
		}
	} else {
		suites = suitesInDir(".", cliConfig.Recurse)
	}

	skippedPackages := []string{}
	if cliConfig.SkipPackage != "" {
		skipFilters := strings.Split(cliConfig.SkipPackage, ",")
		filteredSuites := []TestSuite{}
		for _, suite := range suites {
			skip := false
			for _, skipFilter := range skipFilters {
				if strings.Contains(suite.Path, skipFilter) {
					skip = true
					break
				}
			}
			if skip {
				skippedPackages = append(skippedPackages, suite.Path)
			} else {
				filteredSuites = append(filteredSuites, suite)
			}
		}
		suites = filteredSuites
	}

	return suites, skippedPackages
}

func precompiledTestSuite(path string) (TestSuite, error) {
	info, err := os.Stat(path)
	if err != nil {
		return TestSuite{}, err
	}

	if info.IsDir() {
		return TestSuite{}, errors.New("this is a directory, not a file")
	}

	if filepath.Ext(path) != ".test" {
		return TestSuite{}, errors.New("this is not a .test binary")
	}

	if info.Mode()&0111 == 0 {
		return TestSuite{}, errors.New("this is not executable")
	}

	dir := relPath(filepath.Dir(path))
	packageName := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))

	path, err = filepath.Abs(path)
	if err != nil {
		return TestSuite{}, err
	}

	return TestSuite{
		Path:               dir,
		PackageName:        packageName,
		IsGinkgo:           true,
		Precompiled:        true,
		PathToCompiledTest: path,
	}, nil
}

func suitesInDir(dir string, recurse bool) []TestSuite {
	suites := []TestSuite{}

	if path.Base(dir) == "vendor" {
		return suites
	}

	files, _ := ioutil.ReadDir(dir)
	re := regexp.MustCompile(`^[^._].*_test\.go$`)
	for _, file := range files {
		if !file.IsDir() && re.Match([]byte(file.Name())) {
			suite := TestSuite{
				Path:        relPath(dir),
				PackageName: packageNameForSuite(dir),
				IsGinkgo:    filesHaveGinkgoSuite(dir, files),
			}
			suites = append(suites, suite)
			break
		}
	}

	if recurse {
		re = regexp.MustCompile(`^[._]`)
		for _, file := range files {
			if file.IsDir() && !re.Match([]byte(file.Name())) {
				suites = append(suites, suitesInDir(dir+"/"+file.Name(), recurse)...)
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
