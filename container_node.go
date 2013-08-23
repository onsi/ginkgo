package ginkgo

import (
	"math/rand"
)

type node interface {
	//flesh me out!  both containerNodes and itNodes need to satisfy this (but not runnables)
	isContainerNode() bool
	isItNode() bool
}

type containerNode struct {
	flag         flagType
	text         string
	codeLocation CodeLocation

	beforeEachNodes     []*runnableNode
	justBeforeEachNodes []*runnableNode
	afterEachNodes      []*runnableNode
	itAndContainerNodes []node
}

func newContainerNode(text string, flag flagType, codeLocation CodeLocation) *containerNode {
	return &containerNode{
		text:         text,
		flag:         flag,
		codeLocation: codeLocation,
	}
}

func (container *containerNode) shuffle(r *rand.Rand) {
	permutation := r.Perm(len(container.itAndContainerNodes))
	shuffledNodes := make([]node, len(container.itAndContainerNodes))
	for i, j := range permutation {
		shuffledNodes[i] = container.itAndContainerNodes[j]
	}
	container.itAndContainerNodes = shuffledNodes
}

func (node *containerNode) generateExamples() []*example {
	examples := make([]*example, 0)

	for _, containerOrIt := range node.itAndContainerNodes {
		if containerOrIt.isContainerNode() {
			container := containerOrIt.(*containerNode)
			examples = append(examples, container.generateExamples()...)
		} else {
			examples = append(examples, newExample(containerOrIt.(*itNode)))
		}
	}

	for _, example := range examples {
		example.addContainerNode(node)
	}

	return examples
}

func (node *containerNode) pushContainerNode(container *containerNode) {
	node.itAndContainerNodes = append(node.itAndContainerNodes, container)
}

func (node *containerNode) pushItNode(it *itNode) {
	node.itAndContainerNodes = append(node.itAndContainerNodes, it)
}

func (node *containerNode) pushBeforeEachNode(beforeEach *runnableNode) {
	node.beforeEachNodes = append(node.beforeEachNodes, beforeEach)
}

func (node *containerNode) pushJustBeforeEachNode(justBeforeEach *runnableNode) {
	node.justBeforeEachNodes = append(node.justBeforeEachNodes, justBeforeEach)
}

func (node *containerNode) pushAfterEachNode(afterEach *runnableNode) {
	node.afterEachNodes = append(node.afterEachNodes, afterEach)
}

func (node *containerNode) isContainerNode() bool {
	return true
}

func (node *containerNode) isItNode() bool {
	return false
}
