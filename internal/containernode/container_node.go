package containernode

import (
	"github.com/onsi/ginkgo/internal/types"
	"github.com/onsi/ginkgo/types"
	"math/rand"
	"sort"
)

type subjectOrContainerNode struct {
	containerNode *ContainerNode
	subjectNode   internaltypes.SubjectNode
}

func (n subjectOrContainerNode) text() string {
	if n.containerNode != nil {
		return n.containerNode.Text()
	} else {
		return n.subjectNode.Text()
	}
}

type CollatedNodes struct {
	Containers []*ContainerNode
	Subject    internaltypes.SubjectNode
}

type ContainerNode struct {
	text         string
	flag         internaltypes.FlagType
	codeLocation types.CodeLocation

	beforeEachNodes          []internaltypes.BasicNode
	justBeforeEachNodes      []internaltypes.BasicNode
	afterEachNodes           []internaltypes.BasicNode
	subjectAndContainerNodes []subjectOrContainerNode
}

func New(text string, flag internaltypes.FlagType, codeLocation types.CodeLocation) *ContainerNode {
	return &ContainerNode{
		text:         text,
		flag:         flag,
		codeLocation: codeLocation,
	}
}

func (container *ContainerNode) Shuffle(r *rand.Rand) {
	sort.Sort(container)
	permutation := r.Perm(len(container.subjectAndContainerNodes))
	shuffledNodes := make([]subjectOrContainerNode, len(container.subjectAndContainerNodes))
	for i, j := range permutation {
		shuffledNodes[i] = container.subjectAndContainerNodes[j]
	}
	container.subjectAndContainerNodes = shuffledNodes
}

func (node *ContainerNode) Collate() []CollatedNodes {
	return node.collate([]*ContainerNode{})
}

func (node *ContainerNode) collate(enclosingContainers []*ContainerNode) []CollatedNodes {
	collated := make([]CollatedNodes, 0)

	containers := make([]*ContainerNode, len(enclosingContainers))
	copy(containers, enclosingContainers)
	containers = append(containers, node)

	for _, subjectOrContainer := range node.subjectAndContainerNodes {
		if subjectOrContainer.containerNode != nil {
			collated = append(collated, subjectOrContainer.containerNode.collate(containers)...)
		} else {
			collated = append(collated, CollatedNodes{
				Containers: containers,
				Subject:    subjectOrContainer.subjectNode,
			})
		}
	}

	return collated
}

func (node *ContainerNode) PushContainerNode(container *ContainerNode) {
	node.subjectAndContainerNodes = append(node.subjectAndContainerNodes, subjectOrContainerNode{containerNode: container})
}

func (node *ContainerNode) PushSubjectNode(subject internaltypes.SubjectNode) {
	node.subjectAndContainerNodes = append(node.subjectAndContainerNodes, subjectOrContainerNode{subjectNode: subject})
}

func (node *ContainerNode) PushBeforeEachNode(beforeEach internaltypes.BasicNode) {
	node.beforeEachNodes = append(node.beforeEachNodes, beforeEach)
}

func (node *ContainerNode) PushJustBeforeEachNode(justBeforeEach internaltypes.BasicNode) {
	node.justBeforeEachNodes = append(node.justBeforeEachNodes, justBeforeEach)
}

func (node *ContainerNode) PushAfterEachNode(afterEach internaltypes.BasicNode) {
	node.afterEachNodes = append(node.afterEachNodes, afterEach)
}

func (node *ContainerNode) BeforeEachNodes() []internaltypes.BasicNode {
	return node.beforeEachNodes
}

func (node *ContainerNode) AfterEachNodes() []internaltypes.BasicNode {
	return node.afterEachNodes
}

func (node *ContainerNode) JustBeforeEachNodes() []internaltypes.BasicNode {
	return node.justBeforeEachNodes
}

func (node *ContainerNode) Text() string {
	return node.text
}

func (node *ContainerNode) CodeLocation() types.CodeLocation {
	return node.codeLocation
}

func (node *ContainerNode) Flag() internaltypes.FlagType {
	return node.flag
}

//sort.Interface

func (node *ContainerNode) Len() int {
	return len(node.subjectAndContainerNodes)
}

func (node *ContainerNode) Less(i, j int) bool {
	return node.subjectAndContainerNodes[i].text() < node.subjectAndContainerNodes[j].text()
}

func (node *ContainerNode) Swap(i, j int) {
	node.subjectAndContainerNodes[i], node.subjectAndContainerNodes[j] = node.subjectAndContainerNodes[j], node.subjectAndContainerNodes[i]
}
