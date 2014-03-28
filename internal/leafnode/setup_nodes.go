package leafnode

import (
	"github.com/onsi/ginkgo/types"
	"time"
)

type SetupNode struct {
	runner   *runner
	nodeType types.ExampleComponentType
}

func (node *SetupNode) Run() (outcome types.ExampleState, failure types.ExampleFailure) {
	return node.runner.run()
}

func (node *SetupNode) Type() types.ExampleComponentType {
	return node.nodeType
}

func (node *SetupNode) CodeLocation() types.CodeLocation {
	return node.runner.codeLocation
}

func NewBeforeEachNode(body interface{}, codeLocation types.CodeLocation, timeout time.Duration) *SetupNode {
	return &SetupNode{
		runner:   newRunner(body, codeLocation, timeout),
		nodeType: types.ExampleComponentTypeBeforeEach,
	}
}

func NewAfterEachNode(body interface{}, codeLocation types.CodeLocation, timeout time.Duration) *SetupNode {
	return &SetupNode{
		runner:   newRunner(body, codeLocation, timeout),
		nodeType: types.ExampleComponentTypeAfterEach,
	}
}

func NewJustBeforeEachNode(body interface{}, codeLocation types.CodeLocation, timeout time.Duration) *SetupNode {
	return &SetupNode{
		runner:   newRunner(body, codeLocation, timeout),
		nodeType: types.ExampleComponentTypeJustBeforeEach,
	}
}
