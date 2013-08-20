package godescribe

import (
	"fmt"
	"reflect"
)

type runnableNode struct {
	isAsync      bool
	asyncFunc    func(Done)
	syncFunc     func()
	codeLocation CodeLocation
}

func newRunnableNode(body interface{}, codeLocation CodeLocation) *runnableNode {
	bodyType := reflect.TypeOf(body)
	if bodyType.Kind() != reflect.Func {
		panic(fmt.Sprintf("Expected a function but got something else at %v", codeLocation))
	}

	numberOfArguments := bodyType.NumIn()

	if numberOfArguments > 1 {
		panic(fmt.Sprintf("Too many arguments to function at %v", codeLocation))
	}

	if numberOfArguments == 0 {
		return &runnableNode{
			isAsync:      false,
			asyncFunc:    nil,
			syncFunc:     body.(func()),
			codeLocation: codeLocation,
		}
	} else {
		if !(bodyType.In(0).Kind() == reflect.Chan && bodyType.In(0).Elem().Kind() == reflect.Interface) {
			panic(fmt.Sprintf("Must pass a Done channel to function at %v", codeLocation))
		}

		return &runnableNode{
			isAsync:      true,
			asyncFunc:    body.(func(Done)),
			syncFunc:     nil,
			codeLocation: codeLocation,
		}
	}
}

type beforeEachNode struct {
	*runnableNode
}

func newBeforeEachNode(body interface{}, codeLocation CodeLocation) *beforeEachNode {
	return &beforeEachNode{
		runnableNode: newRunnableNode(body, codeLocation),
	}
}

type justBeforeEachNode struct {
	*runnableNode
}

func newJustBeforeEachNode(body interface{}, codeLocation CodeLocation) *justBeforeEachNode {
	return &justBeforeEachNode{
		runnableNode: newRunnableNode(body, codeLocation),
	}
}

type afterEachNode struct {
	*runnableNode
}

func newAfterEachNode(body interface{}, codeLocation CodeLocation) *afterEachNode {
	return &afterEachNode{
		runnableNode: newRunnableNode(body, codeLocation),
	}
}

type itNode struct {
	*runnableNode

	flag flagType
	text string
}

func (node *itNode) isContainerNode() bool {
	return false
}

func (node *itNode) isItNode() bool {
	return true
}
