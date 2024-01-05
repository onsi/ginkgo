package dsl_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDSL(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "DSL Suite")
}

func ExtractSymbols(f *ast.File) []string {
	symbols := []string{}
	for _, decl := range f.Decls {
		names := []string{}
		switch v := decl.(type) {
		case *ast.FuncDecl:
			if v.Recv == nil {
				names = append(names, v.Name.Name)
			}
		case *ast.GenDecl:
			switch v.Tok {
			case token.TYPE:
				s := v.Specs[0].(*ast.TypeSpec)
				names = append(names, s.Name.Name)
			case token.CONST, token.VAR:
				s := v.Specs[0].(*ast.ValueSpec)
				for _, n := range s.Names {
					names = append(names, n.Name)
				}
			}
		}
		for _, name := range names {
			if ast.IsExported(name) {
				symbols = append(symbols, name)
			}
		}
	}
	return symbols
}

var _ = It("ensures complete coverage of the core dsl", func() {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, "../", nil, 0)
	Ω(err).ShouldNot(HaveOccurred())
	expectedSymbols := []string{}
	for fn, file := range pkgs["ginkgo"].Files {
		if fn == "../deprecated_dsl.go" {
			continue
		}
		expectedSymbols = append(expectedSymbols, ExtractSymbols(file)...)
	}

	actualSymbols := []string{}
	for _, pkg := range []string{"core", "reporting", "decorators", "table"} {
		pkgs, err := parser.ParseDir(fset, "./"+pkg, nil, 0)
		Ω(err).ShouldNot(HaveOccurred())
		for _, file := range pkgs[pkg].Files {
			actualSymbols = append(actualSymbols, ExtractSymbols(file)...)
		}
	}

	Ω(actualSymbols).Should(ConsistOf(expectedSymbols))
})
