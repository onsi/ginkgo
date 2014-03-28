package leafnode

import (
	"github.com/onsi/ginkgo/internal/types"
	"github.com/onsi/ginkgo/types"
	"time"
)

type ItNode struct {
	runner *runner

	flag internaltypes.FlagType
	text string
}

func NewItNode(text string, body interface{}, flag internaltypes.FlagType, codeLocation types.CodeLocation, timeout time.Duration) *ItNode {
	return &ItNode{
		runner: newRunner(body, codeLocation, timeout),
		flag:   flag,
		text:   text,
	}
}

func (node *ItNode) Run() (outcome types.ExampleState, failure types.ExampleFailure) {
	return node.runner.run()
}

func (node *ItNode) Type() types.ExampleComponentType {
	return types.ExampleComponentTypeIt
}

func (node *ItNode) Text() string {
	return node.text
}

func (node *ItNode) Flag() internaltypes.FlagType {
	return node.flag
}

func (node *ItNode) CodeLocation() types.CodeLocation {
	return node.runner.codeLocation
}

func (node *ItNode) Samples() int {
	return 1
}
