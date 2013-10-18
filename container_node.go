package ginkgo

import (
	"github.com/onsi/ginkgo/types"
	"math/rand"
	"sort"
)

type containerNode struct {
	flag         flagType
	text         string
	codeLocation types.CodeLocation

	beforeEachNodes          []*runnableNode
	justBeforeEachNodes      []*runnableNode
	afterEachNodes           []*runnableNode
	subjectAndContainerNodes []node
}

func newContainerNode(text string, flag flagType, codeLocation types.CodeLocation) *containerNode {
	return &containerNode{
		text:         text,
		flag:         flag,
		codeLocation: codeLocation,
	}
}

func (container *containerNode) shuffle(r *rand.Rand) {
	sort.Sort(container)
	permutation := r.Perm(len(container.subjectAndContainerNodes))
	shuffledNodes := make([]node, len(container.subjectAndContainerNodes))
	for i, j := range permutation {
		shuffledNodes[i] = container.subjectAndContainerNodes[j]
	}
	container.subjectAndContainerNodes = shuffledNodes
}

func (node *containerNode) generateExamples() []*example {
	examples := make([]*example, 0)

	for _, containerOrSubject := range node.subjectAndContainerNodes {
		if containerOrSubject.nodeType() == nodeTypeContainer {
			container := containerOrSubject.(*containerNode)
			examples = append(examples, container.generateExamples()...)
		} else {
			subject, ok := containerOrSubject.(exampleSubject)
			if ok {
				examples = append(examples, newExample(subject))
			}
		}
	}

	for _, example := range examples {
		example.addContainerNode(node)
	}

	return examples
}

func (node *containerNode) pushContainerNode(container *containerNode) {
	node.subjectAndContainerNodes = append(node.subjectAndContainerNodes, container)
}

func (node *containerNode) pushSubjectNode(subject exampleSubject) {
	node.subjectAndContainerNodes = append(node.subjectAndContainerNodes, subject)
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

func (node *containerNode) nodeType() nodeType {
	return nodeTypeContainer
}

func (node *containerNode) getText() string {
	return node.text
}

//sort.Interface

func (node *containerNode) Len() int {
	return len(node.subjectAndContainerNodes)
}

func (node *containerNode) Less(i, j int) bool {
	return node.subjectAndContainerNodes[i].getText() < node.subjectAndContainerNodes[j].getText()
}

func (node *containerNode) Swap(i, j int) {
	node.subjectAndContainerNodes[i], node.subjectAndContainerNodes[j] = node.subjectAndContainerNodes[j], node.subjectAndContainerNodes[i]
}
