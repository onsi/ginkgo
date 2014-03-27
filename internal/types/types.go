package internaltypes

import (
	"github.com/onsi/ginkgo/types"
)

type GinkgoTestDescription struct {
	ComponentTexts []string
	FullTestText   string
	TestText       string

	IsMeasurement bool

	FileName   string
	LineNumber int
}

type GinkgoTestingT interface {
	Fail()
}

type NodeType uint

const (
	NodeTypeInvalid NodeType = iota
	NodeTypeContainer
	NodeTypeIt
	NodeTypeMeasure
)

type FlagType uint

const (
	FlagTypeNone FlagType = iota
	FlagTypeFocused
	FlagTypePending
)

type Outcome uint

const (
	OutcomeInvalid Outcome = iota
	OutcomePanicked
	OutcomeTimedOut
	OutcomeCompleted
)

type FailureData struct {
	Message        string
	CodeLocation   types.CodeLocation
	ForwardedPanic interface{}
}
