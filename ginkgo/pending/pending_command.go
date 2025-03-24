package pending

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"slices"
	"strings"

	"golang.org/x/tools/go/ast/inspector"

	"github.com/onsi/ginkgo/v2/ginkgo/command"
	"github.com/onsi/ginkgo/v2/ginkgo/internal"
	"github.com/onsi/ginkgo/v2/types"
)

func BuildPendingCommand() command.Command {
	var cliConfig = types.NewDefaultCLIConfig()

	flags, err := types.BuildLabelsCommandFlagSet(&cliConfig)
	if err != nil {
		panic(err)
	}
	return command.Command{
		Name:     "pending",
		Usage:    "ginkgo pending",
		Flags:    flags,
		ShortDoc: "Recursively search any pending or excluded tests tests under the current directory",
		DocLink:  "filtering-specs",
		Command: func(args []string, _ []string) {
			pendingSpecs(args, cliConfig)
		},
	}
}

var prefixes = []string{
	"PDescribe", "PContext", "PIt", "PDescribeTable", "PEntry", "PSpecify", "PWhen",
	"XDescribe", "XContext", "XIt", "XDescribeTable", "XEntry", "XSpecify", "XWhen",
}

func pendingSpecs(args []string, cliConfig types.CLIConfig) {
	fmt.Println("Scanning for pending...")
	suites := internal.FindSuites(args, cliConfig, false)
	if len(suites) == 0 {
		command.AbortWith("Found no test suites")
	}
	for _, suite := range suites {
		res := fetchPendingFromPackage(suite.Path)
		if len(res) > 0 {
			fmt.Printf("%s:\n", suite.PackageName)
			for _, v := range res {
				fmt.Printf("\t%s : %s\n", v.pos.String(), v.name)
			}

		}
	}
}

type out struct {
	pos  token.Position
	name string
}

func fetchPendingFromPackage(packagePath string) []out {
	fset := token.NewFileSet()
	parsedPackages, err := parser.ParseDir(fset, packagePath, nil, 0)
	command.AbortIfError("Failed to parse package source:", err)

	files := []*ast.File{}
	hasTestPackage := false
	for key, pkg := range parsedPackages {
		if strings.HasSuffix(key, "_test") {
			hasTestPackage = true
			for _, file := range pkg.Files {
				files = append(files, file)
			}
		}
	}
	if !hasTestPackage {
		for _, pkg := range parsedPackages {
			for _, file := range pkg.Files {
				files = append(files, file)
			}
		}
	}

	res := []out{}
	ispr := inspector.New(files)
	ispr.Preorder([]ast.Node{&ast.CallExpr{}}, func(n ast.Node) {
		if c, ok := n.(*ast.CallExpr); ok {
			if i, ok := c.Fun.(*ast.Ident); ok {
				if slices.Contains(prefixes, i.Name) {
					pos := fset.Position(c.Pos())
					res = append(res, out{pos: pos, name: c.Args[0].(*ast.BasicLit).Value})
				}
			}

		}
	})
	return res
}
