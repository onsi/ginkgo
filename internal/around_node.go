package internal

import (
	"github.com/onsi/ginkgo/v2/types"
)

func ComputeAroundNodes(specs Specs) Specs {
	out := Specs{}
	for _, spec := range specs {
		nodes := Nodes{}
		currentNestingLevel := 0
		aroundNodes := []AroundNode{}
		nestingLevelIndices := []int{}
		for _, node := range spec.Nodes {
			switch node.NodeType {
			case types.NodeTypeContainer:
				currentNestingLevel = node.NestingLevel + 1
				nestingLevelIndices = append(nestingLevelIndices, len(aroundNodes))
				aroundNodes = append(aroundNodes, node.AroundNodes...)
				nodes = append(nodes, node)
			default:
				if currentNestingLevel > node.NestingLevel {
					currentNestingLevel = node.NestingLevel
					aroundNodes = aroundNodes[:nestingLevelIndices[currentNestingLevel]]
				}
				nodeAroundNodes := []AroundNode{}
				nodeAroundNodes = append(nodeAroundNodes, aroundNodes...)
				nodeAroundNodes = append(nodeAroundNodes, node.AroundNodes...)
				node.AroundNodes = nodeAroundNodes
				nodes = append(nodes, node)
			}
		}
		spec.Nodes = nodes
		out = append(out, spec)
	}
	return out
}
