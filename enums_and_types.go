package ginkgo

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
	getFlag() flagType
	getCodeLocation() types.CodeLocation
}

type flagType uint

const (
	flagTypeNone flagType = iota
	flagTypeFocused
	flagTypePending
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
