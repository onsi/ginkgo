package internal

import (
	"github.com/onsi/ginkgo/types"
)

type node interface {
	nodeType() nodeType
	getText() string
}

type exampleSubject interface {
	node

	run() (runOutcome, failureData)
	getFlag() FlagType
	getCodeLocation() types.CodeLocation
}

type FlagType uint

const (
	FlagTypeNone FlagType = iota
	FlagTypeFocused
	FlagTypePending
)

type runOutcome uint

const (
	runOutcomeInvalid runOutcome = iota
	runOutcomePanicked
	runOutcomeTimedOut
	runOutcomeCompleted
)

type nodeType uint

const (
	nodeTypeInvalid nodeType = iota
	nodeTypeContainer
	nodeTypeIt
	nodeTypeMeasure
)
