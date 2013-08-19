package godescribe

type node interface {
	//flesh me out!  both containerNodes and itNodes need to satisfy this (but not runnables)
	isContainerNode() bool
	isExampleNode() bool
}

type containerNode struct {
	flag         flagType
	cType        containerType
	text         string
	codeLocation CodeLocation

	beforeEachNodes     []*runnableNode
	justBeforeEachNodes []*runnableNode
	afterEachNodes      []*runnableNode
	itAndContainerNodes []*node
}

func newContainerNode(text string, cType containerType, flag flagType, codeLocation CodeLocation) *containerNode {
	return &containerNode{
		text:         text,
		cType:        cType,
		flag:         flag,
		codeLocation: codeLocation,
	}
}
