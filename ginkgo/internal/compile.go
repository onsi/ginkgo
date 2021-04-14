package internal

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/onsi/ginkgo/types"
)

func CompileSuite(suite TestSuite, goFlagsConfig types.GoFlagsConfig) TestSuite {
	if suite.PathToCompiledTest != "" {
		return suite
	}

	suite.CompilationError = nil

	path, err := filepath.Abs(filepath.Join(suite.Path, suite.PackageName+".test"))
	if err != nil {
		suite.CompilationError = fmt.Errorf("Failed to compute compilation target path:\n%s", err.Error())
		return suite
	}

	args, err := types.GenerateGoTestCompileArgs(goFlagsConfig, path, suite.Path)
	if err != nil {
		suite.CompilationError = fmt.Errorf("Failed to generate go test compile flags:\n%s", err.Error())
		return suite
	}

	cmd := exec.Command("go", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if len(output) > 0 {
			suite.CompilationError = fmt.Errorf("Failed to compile %s:\n\n%s", suite.PackageName, output)
		} else {
			suite.CompilationError = fmt.Errorf("Failed to compile %s\n%s", suite.PackageName, err.Error())
		}
		return suite
	}

	if len(output) > 0 {
		fmt.Println(string(output))
	}

	if !FileExists(path) {
		suite.CompilationError = fmt.Errorf("Failed to compile %s:\nOutput file %s could not be found", suite.PackageName, path)
		return suite
	}

	suite.PathToCompiledTest = path
	return suite
}

func Cleanup(goFlagsConfig types.GoFlagsConfig, suites ...TestSuite) {
	if goFlagsConfig.BinaryMustBePreserved() {
		return
	}
	for _, suite := range suites {
		if !suite.Precompiled {
			os.Remove(suite.PathToCompiledTest)
		}
	}
}
