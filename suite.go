package ginkgo

import (
	"math/rand"
	"testing"
)

type flagType uint

const (
	flagTypeNone flagType = iota
	flagTypeFocused
	flagTypePending
)

type failureData struct {
	message        string
	codeLocation   CodeLocation
	forwardedPanic interface{}
}

type suite struct {
	topLevelContainer *containerNode
	currentContainer  *containerNode
	exampleCollection *exampleCollection
}

func newSuite() *suite {
	topLevelContainer := newContainerNode("", flagTypeNone, CodeLocation{})

	return &suite{
		topLevelContainer: topLevelContainer,
		currentContainer:  topLevelContainer,
	}
}

func (suite *suite) run(t *testing.T, description string, randomSeed int64, randomizeAllExamples bool, reporter Reporter) {
	reporter.RandomizationStrategy(randomSeed, randomizeAllExamples)
	r := rand.New(rand.NewSource(randomSeed))
	suite.topLevelContainer.shuffle(r)
	suite.exampleCollection = newExampleCollection(t, description, suite.topLevelContainer.generateExamples(), reporter)
	if randomizeAllExamples {
		suite.exampleCollection.shuffle(r)
	}

	suite.exampleCollection.run()
}

func (suite *suite) fail(message string, callerSkip int) {
	codeLocation, _ := generateCodeLocation(callerSkip + 3)
	suite.exampleCollection.fail(failureData{
		message:      message,
		codeLocation: codeLocation,
	})
}

func (suite *suite) pushContainerNode(text string, body func(), flag flagType, codeLocation CodeLocation) {
	container := newContainerNode(text, flag, codeLocation)
	suite.currentContainer.pushContainerNode(container)

	previousContainer := suite.currentContainer
	suite.currentContainer = container

	body()

	suite.currentContainer = previousContainer
}

func (suite *suite) pushItNode(text string, body interface{}, flag flagType, codeLocation CodeLocation) {
	suite.currentContainer.pushItNode(newItNode(text, body, flag, codeLocation))
}

func (suite *suite) pushBeforeEachNode(body interface{}, codeLocation CodeLocation) {
	suite.currentContainer.pushBeforeEachNode(newRunnableNode(body, codeLocation))
}

func (suite *suite) pushJustBeforeEachNode(body interface{}, codeLocation CodeLocation) {
	suite.currentContainer.pushJustBeforeEachNode(newRunnableNode(body, codeLocation))
}

func (suite *suite) pushAfterEachNode(body interface{}, codeLocation CodeLocation) {
	suite.currentContainer.pushAfterEachNode(newRunnableNode(body, codeLocation))
}
