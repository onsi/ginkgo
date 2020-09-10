package internal

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/onsi/ginkgo/formatter"
	"github.com/onsi/ginkgo/ginkgo/command"
)

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func GoFmt(path string) {
	out, err := exec.Command("go", "fmt", path).CombinedOutput()
	if err != nil {
		command.AbortIfError(fmt.Sprintf("Could not fmt:\n%s\n", string(out)), err)
	}
}

func PluralizedWord(singular, plural string, count int) string {
	if count == 1 {
		return singular
	}
	return plural
}

func FailedSuitesReport(suites []TestSuite, f formatter.Formatter) string {
	out := ""
	out += "There were failures detected in the following suites:\n"

	maxPackageNameLength := 0
	for _, suite := range suites {
		if len(suite.PackageName) > maxPackageNameLength {
			maxPackageNameLength = len(suite.PackageName)
		}
	}

	packageNameFormatter := fmt.Sprintf("%%%ds", maxPackageNameLength)
	for _, suite := range suites {
		out += f.Fi(1, "{{red}}"+packageNameFormatter+" {{gray}}%s{{/}}\n", suite.PackageName, suite.Path)
	}
	return out
}
