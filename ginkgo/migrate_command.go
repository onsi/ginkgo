package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"sync"
)

func BuildMigrateCommand() *Command {
	return &Command{
		Name:         "migrate",
		FlagSet:      flag.NewFlagSet("migrate", flag.ExitOnError),
		UsageCommand: "ginkgo migrate",
		Usage: []string{
			"Recursively files under the current directory for migration issues",
		},
		Command: migrateSpecs,
	}
}

func migrateSpecs([]string, []string) {
	fmt.Println("Scanning for migration issues...")

	goFiles := make(chan string)
	go func() {
		listDirectory(goFiles, ".")
		close(goFiles)
	}()

	const workers = 10
	wg := sync.WaitGroup{}
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func() {
			for path := range goFiles {
				scanFileForMigrationIssues(path)
			}
			wg.Done()
		}()
	}

	wg.Wait()
}

func scanFileForMigrationIssues(path string) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Printf("error reading file '%s': %s\n", path, err.Error())
		return
	}

	ast, err := parser.ParseFile(token.NewFileSet(), path, bytes.NewReader(data), 0)
	if err != nil {
		fmt.Printf("error parsing file '%s': %s\n", path, err.Error())
		return
	}

	measure := scanASTForMigrationIssues(ast)

	if measure {
		fmt.Printf("Found deprecated `Measure()` in file: %s\n", path)
	}
}

func scanASTForMigrationIssues(file *ast.File) (measure bool) {
	ast.Inspect(file, func(n ast.Node) bool {
		if c, ok := n.(*ast.CallExpr); ok {
			if i, ok := c.Fun.(*ast.Ident); ok {
				if isMeasure(i.Name) {
					measure = true
				}
			}
		}

		return true
	})

	return measure
}

func isMeasure(name string) bool {
	switch name {
	case "Measure", "FMeasure", "XMeasure", "ginkgo.Measure":
		return true
	default:
		return false
	}
}
