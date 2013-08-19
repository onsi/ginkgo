package godescribe

import (
	"testing"
)

type containerType uint

const (
	containerTypeInvalid containerType = iota
	containerTypeTopLevel
	containerTypeDescribe
	containerTypeContext
)

type flagType uint

const (
	flagTypeNone flagType = iota
	flagTypePending
	flagTypeFocused
)

type suite struct {
	topLevelContainer *containerNode
	currentContainer  *containerNode
}

func newSuite() *suite {
	topLevelContainer := newContainerNode("", containerTypeTopLevel, flagTypeNone)

	return &suite{
		topLevelContainer: topLevelContainer,
		currentContainer:  topLevelContainer,
	}
}

func (suite *suite) run(t *testing.T, description string, randomSeed int64, randomizeAllExamples bool, reporter Reporter) {
	//randomize the suite
	//process any focus (if any focused, mark all non-focussed as skipped)

	//generate summary report
	//run each example (& send report)
	//generate summary report
}

func (suite *suite) fail(message string, callerSkip int) {
	//somehow without panicking?
}

func (suite *suite) pushContainerNode(text string, body func(), conType containerType, flag flagType) {

}

func (suite *suite) pushExampleNode(text string, body interface{}, flag flagType) {

}

func (suite *suite) pushBeforeEachNode(body interface{}) {

}

func (suite *suite) pushJustBeforeEachNode(body interface{}) {

}

func (suite *suite) pushAfterEachNode(body interface{}) {

}
