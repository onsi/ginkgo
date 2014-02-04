package convert

import (
	"go/ast"
)

func newGinkgoTFromIdent(ident *ast.Ident) *ast.CallExpr {
	return &ast.CallExpr{
		Lparen: ident.NamePos + 1,
		Rparen: ident.NamePos + 2,
		Fun:    &ast.Ident{Name: "GinkgoT"},
	}
}

func newGinkgoTestingT() *ast.Ident {
	return &ast.Ident{Name: "GinkgoTestingT"}
}
