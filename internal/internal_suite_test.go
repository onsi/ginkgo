package internal_test

import (
	"context"
	"reflect"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"

	"github.com/onsi/ginkgo/v2/internal"
)

func TestInternal(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Internal Suite", AroundNode(func(ctx context.Context, body func(ctx context.Context)) {
		body(context.WithValue(ctx, "suite", "internal"))
	}))
}

type Node = internal.Node
type Nodes = internal.Nodes
type NodeType = types.NodeType
type TreeNode = internal.TreeNode
type TreeNodes = internal.TreeNodes
type Spec = internal.Spec
type Specs = internal.Specs

var ntIt = types.NodeTypeIt
var ntCon = types.NodeTypeContainer
var ntAf = types.NodeTypeAfterEach
var ntBef = types.NodeTypeBeforeEach
var ntJusAf = types.NodeTypeJustAfterEach
var ntJusBef = types.NodeTypeJustBeforeEach

type NestingLevel int

// convenience helper to quickly make nodes
// assumes they are correctly configured and no errors occur
func N(args ...any) Node {
	nodeType, text, nestingLevel, hasBody := types.NodeTypeIt, "", -1, false
	remainingArgs := []any{cl}
	for _, arg := range args {
		switch t := reflect.TypeOf(arg); {
		case t == reflect.TypeOf(NestingLevel(1)):
			nestingLevel = int(arg.(NestingLevel))
		case t == reflect.TypeOf(text):
			text = arg.(string)
		case t == reflect.TypeOf(nodeType):
			nodeType = arg.(types.NodeType)
		case t.Kind() == reflect.Func:
			hasBody = true
			remainingArgs = append(remainingArgs, arg)
		default:
			remainingArgs = append(remainingArgs, arg)
		}
	}
	//the hasBody dance is necessary to (a) make sure internal.NewNode is happy (it requires a body) and (b) to then nil out the resulting body to ensure node comparisons work
	//as reflect.DeepEqual cannot compare functions.  Even by pointer.  'Cause.  You know.
	if !hasBody {
		remainingArgs = append(remainingArgs, func() {})
	}
	node, errors := internal.NewNode(nil, nodeType, text, remainingArgs...)
	if nestingLevel != -1 {
		node.NestingLevel = nestingLevel
	}
	ExpectWithOffset(1, errors).Should(BeEmpty())
	if !hasBody {
		node.Body = nil
	}
	return node
}

// convenience helper to quickly make tree nodes
func TN(node Node, children ...*TreeNode) *TreeNode {
	tn := &TreeNode{Node: node}
	for _, child := range children {
		tn.AppendChild(child)
	}
	return tn
}

// convenience helper to quickly make specs
func S(nodes ...Node) Spec {
	return Spec{Nodes: nodes}
}

// convenience helper to quickly make code locations
func CL(options ...any) types.CodeLocation {
	cl = types.NewCodeLocation(1)
	for _, option := range options {
		if reflect.TypeOf(option).Kind() == reflect.String {
			cl.FileName = option.(string)
		} else if reflect.TypeOf(option).Kind() == reflect.Int {
			cl.LineNumber = option.(int)
		}
	}
	return cl
}

func mustFindNodeWithText(tree *TreeNode, text string) Node {
	node := findNodeWithText(tree, text)
	ExpectWithOffset(1, node).ShouldNot(BeZero(), "Failed to find node in tree with text '%s'", text)
	return node
}

func findNodeWithText(tree *TreeNode, text string) Node {
	if tree.Node.Text == text {
		return tree.Node
	}
	for _, tn := range tree.Children {
		n := findNodeWithText(tn, text)
		if !n.IsZero() {
			return n
		}
	}
	return Node{}
}

var cl types.CodeLocation
var _ = BeforeEach(func() {
	cl = types.NewCodeLocation(0)
})
