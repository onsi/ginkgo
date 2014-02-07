package convert

import (
	"fmt"
	"go/ast"
	"strings"
	"unicode"
)

/*
 * Creates a func init() node
 */
func createInitBlock() *ast.FuncDecl {
	blockStatement := &ast.BlockStmt{List: []ast.Stmt{}}
	fieldList := &ast.FieldList{}
	funcType := &ast.FuncType{Params: fieldList}
	ident := &ast.Ident{Name: "init"}

	return &ast.FuncDecl{Name: ident, Type: funcType, Body: blockStatement}
}

/*
 * Creates a Describe("Testing with ginkgo", func() { }) node
 */
func createDescribeBlock() *ast.ExprStmt {
	blockStatement := &ast.BlockStmt{List: []ast.Stmt{}}

	fieldList := &ast.FieldList{}
	funcType := &ast.FuncType{Params: fieldList}
	funcLit := &ast.FuncLit{Type: funcType, Body: blockStatement}
	basicLit := &ast.BasicLit{Kind: 9, Value: "\"Testing with Ginkgo\""}
	describeIdent := &ast.Ident{Name: "Describe"}
	callExpr := &ast.CallExpr{Fun: describeIdent, Args: []ast.Expr{basicLit, funcLit}}

	return &ast.ExprStmt{X: callExpr}
}

/*
 * Convenience function to return the name of the *testing.T param
 * for a Test function that will be rewritten. This is useful because
 * we will want to replace the usage of this named *testing.T inside the
 * body of the function with a GinktoT.
 */
func namedTestingTArg(node *ast.FuncDecl) string {
	return node.Type.Params.List[0].Names[0].Name // *exhale*
}

/*
 * Convenience function to return the block statement node for a Describe statement
 */
func blockStatementFromDescribe(desc *ast.ExprStmt) *ast.BlockStmt {
	var funcLit *ast.FuncLit
	var found = false

	for _, node := range desc.X.(*ast.CallExpr).Args {
		switch node := node.(type) {
		case *ast.FuncLit:
			found = true
			funcLit = node
			break
		}
	}

	if !found {
		panic("Error finding ast.FuncLit inside describe statement. Somebody done goofed.")
	}

	return funcLit.Body
}

/* convenience function for creating an It("TestNameHere")
 * with all the body of the test function inside the anonymous
 * func passed to It()
 */
func createItStatementForTestFunc(testFunc *ast.FuncDecl) *ast.ExprStmt {
	blockStatement := &ast.BlockStmt{List: testFunc.Body.List}
	fieldList := &ast.FieldList{}
	funcType := &ast.FuncType{Params: fieldList}
	funcLit := &ast.FuncLit{Type: funcType, Body: blockStatement}

	testName := rewriteTestName(testFunc.Name.Name)
	basicLit := &ast.BasicLit{Kind: 9, Value: fmt.Sprintf("\"%s\"", testName)}
	itBlockIdent := &ast.Ident{Name: "It"}
	callExpr := &ast.CallExpr{Fun: itBlockIdent, Args: []ast.Expr{basicLit, funcLit}}
	return &ast.ExprStmt{X: callExpr}
}

/*
* rewrite test names to be human readable
* eg: rewrites "TestSomethingAmazing" as "something amazing"
*/
func rewriteTestName(testName string) string {
	nameComponents := []string{}
	currentString := ""
	indexOfTest := strings.Index(testName, "Test")
	if indexOfTest != 0 {
		return testName
	}

	testName = strings.Replace(testName, "Test", "", 1)
	first, rest := testName[0], testName[1:]
	testName = string(unicode.ToLower(rune(first))) + rest

	for _, rune := range testName {
		if unicode.IsUpper(rune) {
  		nameComponents = append(nameComponents, currentString)
			currentString = string(unicode.ToLower(rune))
		} else {
			currentString += string(rune)
		}
	}

	return strings.Join(append(nameComponents, currentString), " ")
}
