package leafnodes

import (
	"github.com/onsi/ginkgo/types"
)

type BasicNode interface {
	Type() types.ExampleComponentType
	Run() (types.ExampleState, types.ExampleFailure)
	CodeLocation() types.CodeLocation
}

type SubjectNode interface {
	BasicNode

	Text() string
	Flag() types.FlagType
	Samples() int
}
