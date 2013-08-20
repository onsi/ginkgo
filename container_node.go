package godescribe

type node interface {
	//flesh me out!  both containerNodes and itNodes need to satisfy this (but not runnables)
	isContainerNode() bool
	isItNode() bool
}

type containerNode struct {
	flag         flagType
	cType        containerType
	text         string
	codeLocation CodeLocation

	beforeEachNodes     []*beforeEachNode
	justBeforeEachNodes []*justBeforeEachNode
	afterEachNodes      []*afterEachNode
	itAndContainerNodes []node
}

func newContainerNode(text string, cType containerType, flag flagType, codeLocation CodeLocation) *containerNode {
	return &containerNode{
		text:         text,
		cType:        cType,
		flag:         flag,
		codeLocation: codeLocation,
	}
}

func (node *containerNode) pushContainerNode(container *containerNode) {
	node.itAndContainerNodes = append(node.itAndContainerNodes, container)
}

func (node *containerNode) pushItNode(it *itNode) {
	node.itAndContainerNodes = append(node.itAndContainerNodes, it)
}

func (node *containerNode) pushBeforeEachNode(beforeEach *beforeEachNode) {
	node.beforeEachNodes = append(node.beforeEachNodes, beforeEach)
}

func (node *containerNode) pushJustBeforeEachNode(justBeforeEach *justBeforeEachNode) {
	node.justBeforeEachNodes = append(node.justBeforeEachNodes, justBeforeEach)
}

func (node *containerNode) pushAfterEachNode(afterEach *afterEachNode) {
	node.afterEachNodes = append(node.afterEachNodes, afterEach)
}

func (node *containerNode) isContainerNode() bool {
	return true
}

func (node *containerNode) isItNode() bool {
	return false
}
