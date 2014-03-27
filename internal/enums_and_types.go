package internal

import (
	"github.com/onsi/ginkgo/internal/types"
	"github.com/onsi/ginkgo/types"
)

type node interface {
	Type() internaltypes.NodeType
	Text() string
}

type exampleSubject interface {
	node

	Run() (internaltypes.Outcome, internaltypes.FailureData)
	Flag() internaltypes.FlagType
	CodeLocation() types.CodeLocation
}
