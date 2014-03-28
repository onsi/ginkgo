package internaltypes

import (
	"github.com/onsi/ginkgo/types"
)

type GinkgoTestingT interface {
	Fail()
}

type FlagType uint

const (
	FlagTypeNone FlagType = iota
	FlagTypeFocused
	FlagTypePending
)

type BasicNode interface {
	Type() types.ExampleComponentType
	Run() (types.ExampleState, types.ExampleFailure)
	CodeLocation() types.CodeLocation
}

type SubjectNode interface {
	BasicNode

	Text() string
	Flag() FlagType
	Samples() int
}
