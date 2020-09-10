package generators

import (
	"go/build"
	"os"
	"path/filepath"
	"strings"

	"github.com/onsi/ginkgo/ginkgo/command"
)

type GeneratorsConfig struct {
	Agouti, NoDot, Internal bool
	CustomTemplate          string
}

func getPackageAndFormattedName() (string, string, string) {
	path, err := os.Getwd()
	command.AbortIfError("Could not get current working diretory:", err)

	dirName := strings.Replace(filepath.Base(path), "-", "_", -1)
	dirName = strings.Replace(dirName, " ", "_", -1)

	pkg, err := build.ImportDir(path, 0)
	packageName := pkg.Name
	if err != nil {
		packageName = dirName
	}

	formattedName := prettifyPackageName(filepath.Base(path))
	return packageName, dirName, formattedName
}

func prettifyPackageName(name string) string {
	name = strings.Replace(name, "-", " ", -1)
	name = strings.Replace(name, "_", " ", -1)
	name = strings.Title(name)
	name = strings.Replace(name, " ", "", -1)
	return name
}

func determinePackageName(name string, internal bool) string {
	if internal {
		return name
	}

	return name + "_test"
}
