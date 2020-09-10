package internal

import "github.com/onsi/ginkgo/types"

type TreeNode struct {
	Node     Node
	Children TreeNodes
}

type TreeNodes []TreeNode

func (tn TreeNodes) Nodes() Nodes {
	out := Nodes{}
	for _, treeNode := range tn {
		out = append(out, treeNode.Node)
	}
	return out
}

func (tn TreeNodes) WithID(id uint) TreeNode {
	for _, treeNode := range tn {
		if treeNode.Node.ID == id {
			return treeNode
		}
	}

	return TreeNode{}
}

func AppendTreeNodeChild(treeNode TreeNode, child TreeNode) TreeNode {
	treeNode.Children = append(treeNode.Children, child)
	return treeNode
}

func GenerateSpecsFromTreeRoot(tree TreeNode) Specs {
	var walkTree func(nestingLevel int, lNodes Nodes, rNodes Nodes, trees TreeNodes) Specs
	walkTree = func(nestingLevel int, lNodes Nodes, rNodes Nodes, trees TreeNodes) Specs {
		tests := Specs{}

		nodes := Nodes{}
		for _, node := range trees.Nodes() {
			node.NestingLevel = nestingLevel
			nodes = append(nodes, node)
		}

		itsAndContainers := nodes.WithType(types.NodeTypesForContainerAndIt...)
		for _, node := range itsAndContainers {
			leftNodes, rightNodes := nodes.SplitAround(node)
			leftNodes = leftNodes.WithoutType(types.NodeTypesForContainerAndIt...)
			rightNodes = rightNodes.WithoutType(types.NodeTypesForContainerAndIt...)

			leftNodes = lNodes.CopyAppend(leftNodes...)
			rightNodes = rightNodes.CopyAppend(rNodes...)

			if node.NodeType.Is(types.NodeTypeIt) {
				tests = append(tests, Spec{Nodes: leftNodes.CopyAppend(node).CopyAppend(rightNodes...)})
			} else {
				treeNode := trees.WithID(node.ID)
				tests = append(tests, walkTree(nestingLevel+1, leftNodes.CopyAppend(node), rightNodes, treeNode.Children)...)
			}
		}

		return tests
	}

	return walkTree(0, Nodes{}, Nodes{}, tree.Children)
}
