package outline

import (
	"go/ast"
	"go/token"
	"strconv"
)

const (
	// undefinedTextAlt is used if the spec/container text cannot be derived
	undefinedTextAlt = "undefined"
)

// ginkgoMetadata holds useful bits of information for every entry in the outline
type ginkgoMetadata struct {
	// Name is the spec or container function name, e.g. `Describe` or `It`
	Name string `json:"name"`

	// Text is the `text` argument passed to specs, and some containers
	Text string `json:"text"`

	// Start is the position of first character of the spec or container block
	Start token.Pos `json:"start"`

	// End is the position of first character immediately after the spec or container block
	End token.Pos `json:"end"`

	Spec    bool `json:"spec"`
	Focused bool `json:"focused"`
	Pending bool `json:"pending"`
}

// ginkgoNode is used to construct the outline as a tree
type ginkgoNode struct {
	ginkgoMetadata
	Nodes []*ginkgoNode `json:"nodes,omitempty"`
}

type walkFunc func(n *ginkgoNode)

func (n *ginkgoNode) PreOrder(f walkFunc) {
	f(n)
	for _, m := range n.Nodes {
		m.PreOrder(f)
	}
}

func (n *ginkgoNode) PostOrder(f walkFunc) {
	for _, m := range n.Nodes {
		m.PostOrder(f)
	}
	f(n)
}

// PropagateInheritedProperties propagates the Pending and Focused properties
// through the subtree rooted at n.
func (n *ginkgoNode) PropagateInheritedProperties() {
	n.PreOrder(func(thisNode *ginkgoNode) {
		for _, descendantNode := range thisNode.Nodes {
			if thisNode.Pending {
				descendantNode.Pending = true
				descendantNode.Focused = false
			}
			if thisNode.Focused && !descendantNode.Pending {
				descendantNode.Focused = true
			}
		}
	})
}

// BackpropagateUnfocus propagates the Focused property through the subtree
// rooted at n. It applies the rule described in the Ginkgo docs:
// > Nested programmatically focused specs follow a simple rule: if a
// > leaf-node is marked focused, any of its ancestor nodes that are marked
// > focus will be unfocused.
func (n *ginkgoNode) BackpropagateUnfocus() {
	focusedSpecInSubtreeStack := []bool{}
	n.PostOrder(func(thisNode *ginkgoNode) {
		if thisNode.Spec {
			focusedSpecInSubtreeStack = append(focusedSpecInSubtreeStack, thisNode.Focused)
			return
		}
		focusedSpecInSubtree := false
		for range thisNode.Nodes {
			focusedSpecInSubtree = focusedSpecInSubtree || focusedSpecInSubtreeStack[len(focusedSpecInSubtreeStack)-1]
			focusedSpecInSubtreeStack = focusedSpecInSubtreeStack[0 : len(focusedSpecInSubtreeStack)-1]
		}
		focusedSpecInSubtreeStack = append(focusedSpecInSubtreeStack, focusedSpecInSubtree)
		if focusedSpecInSubtree {
			thisNode.Focused = false
		}
	})

}

func ginkgoIdentNameFromCallExpr(ce *ast.CallExpr, ginkgoImportName string) (string, bool) {
	switch ex := ce.Fun.(type) {
	case *ast.Ident:
		if ginkgoImportName != "." {
			return "", false
		}
		return ex.Name, true
	case *ast.SelectorExpr:
		pkgID, ok := ex.X.(*ast.Ident)
		if !ok {
			return "", false
		}
		// A package identifier is top-level, so Obj must be nil
		if pkgID.Obj != nil {
			return "", false
		}
		if ginkgoImportName != pkgID.Name {
			return "", false
		}
		if ex.Sel == nil {
			return "", false
		}
		return ex.Sel.Name, true
	default:
		return "", false
	}
}

// ginkgoNodeFromCallExpr derives an outline entry from a go AST subtree
// corresponding to a Ginkgo container or spec.
func ginkgoNodeFromCallExpr(ce *ast.CallExpr, ginkgoImportName string) (*ginkgoNode, bool) {
	identName, ok := ginkgoIdentNameFromCallExpr(ce, ginkgoImportName)
	if !ok {
		return nil, false
	}

	n := ginkgoNode{}
	n.Name = identName
	n.Start = ce.Pos()
	n.End = ce.End()

	switch identName {
	case "It", "Measure", "Specify":
		n.Spec = true
		n.Text = textOrAltFromCallExpr(ce, undefinedTextAlt)
	case "FIt", "FMeasure", "FSpecify":
		n.Spec = true
		n.Focused = true
		n.Text = textOrAltFromCallExpr(ce, undefinedTextAlt)
	case "PIt", "PMeasure", "PSpecify", "XIt", "XMeasure", "XSpecify":
		n.Spec = true
		n.Pending = true
		n.Text = textOrAltFromCallExpr(ce, undefinedTextAlt)
	case "Context", "Describe", "When":
		n.Text = textOrAltFromCallExpr(ce, undefinedTextAlt)
	case "FContext", "FDescribe", "FWhen":
		n.Focused = true
		n.Text = textOrAltFromCallExpr(ce, undefinedTextAlt)
	case "PContext", "PDescribe", "PWhen", "XContext", "XDescribe", "XWhen":
		n.Pending = true
		n.Text = textOrAltFromCallExpr(ce, undefinedTextAlt)
	case "By":
	case "AfterEach", "BeforeEach":
	case "JustAfterEach", "JustBeforeEach":
	case "AfterSuite", "BeforeSuite":
	case "SynchronizedAfterSuite", "SynchronizedBeforeSuite":
	default:
		return nil, false
	}
	return &n, true
}

// textOrAltFromCallExpr tries to derive the "text" of a Ginkgo spec or
// container. If it cannot derive it, it returns the alt text.
func textOrAltFromCallExpr(ce *ast.CallExpr, alt string) string {
	text, defined := textFromCallExpr(ce)
	if !defined {
		return alt
	}
	return text
}

// textFromCallExpr tries to derive the "text" of a Ginkgo spec or container. If
// it cannot derive it, it returns false.
func textFromCallExpr(ce *ast.CallExpr) (string, bool) {
	if len(ce.Args) < 1 {
		return "", false
	}
	text, ok := ce.Args[0].(*ast.BasicLit)
	if !ok {
		return "", false
	}
	switch text.Kind {
	case token.CHAR, token.STRING:
		// For token.CHAR and token.STRING, Value is quoted
		unquoted, err := strconv.Unquote(text.Value)
		if err != nil {
			// If unquoting fails, just use the raw Value
			return text.Value, true
		}
		return unquoted, true
	default:
		return text.Value, true
	}
}
