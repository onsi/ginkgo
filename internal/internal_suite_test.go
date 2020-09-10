package internal_test

import (
	"reflect"
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/types"
	. "github.com/onsi/gomega"

	"github.com/onsi/ginkgo/internal"
)

func TestInternal(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Internal Suite")
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
type MarkedPending bool
type MarkedFocus bool

// convenience helper to quickly make nodes
func N(options ...interface{}) Node {
	node := internal.NewNode(types.NodeTypeIt, "", nil, cl, false, false)
	for _, option := range options {
		if reflect.TypeOf(option).Kind() == reflect.String {
			node.Text = option.(string)
		} else if reflect.TypeOf(option) == reflect.TypeOf(types.NodeTypeInvalid) {
			node.NodeType = option.(NodeType)
		} else if reflect.TypeOf(option) == reflect.TypeOf(NestingLevel(1)) {
			node.NestingLevel = int(option.(NestingLevel))
		} else if reflect.TypeOf(option) == reflect.TypeOf(cl) {
			node.CodeLocation = option.(types.CodeLocation)
		} else if reflect.TypeOf(option) == reflect.TypeOf(MarkedFocus(true)) {
			node.MarkedFocus = bool(option.(MarkedFocus))
		} else if reflect.TypeOf(option) == reflect.TypeOf(MarkedPending(true)) {
			node.MarkedPending = bool(option.(MarkedPending))
		} else if reflect.TypeOf(option).Kind() == reflect.Func {
			node.Body = option.(func())
		}
	}
	return node
}

// convenience helper to quickly make tree nodes
func TN(node Node, children ...TreeNode) TreeNode {
	return TreeNode{
		Node:     node,
		Children: TreeNodes(children),
	}
}

// convenience helper to quickly make specs
func S(nodes ...Node) Spec {
	return Spec{Nodes: nodes}
}

// convenience helper to quickly make code locations
func CL(options ...interface{}) types.CodeLocation {
	cl = types.NewCodeLocation(0)
	for _, option := range options {
		if reflect.TypeOf(option).Kind() == reflect.String {
			cl.FileName = option.(string)
		} else if reflect.TypeOf(option).Kind() == reflect.Int {
			cl.LineNumber = option.(int)
		}
	}
	return cl
}

func mustFindNodeWithText(tree TreeNode, text string) Node {
	node := findNodeWithText(tree, text)
	ExpectWithOffset(1, node).ShouldNot(BeZero(), "Failed to find node in tree with text '%s'", text)
	return node
}

func findNodeWithText(tree TreeNode, text string) Node {
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
