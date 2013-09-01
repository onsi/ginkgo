package ginkgo

import (
	"math/rand"
	"regexp"
	"time"
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
	topLevelContainer := newContainerNode("[Top Level]", flagTypeNone, CodeLocation{})

	return &suite{
		topLevelContainer: topLevelContainer,
		currentContainer:  topLevelContainer,
	}
}

func (suite *suite) run(t testingT, description string, reporter Reporter, config GinkoConfigType) {
	reporter.RandomizationStrategy(*(config.RandomSeed), *(config.RandomizeAllSpecs))
	r := rand.New(rand.NewSource(*(config.RandomSeed)))
	suite.topLevelContainer.shuffle(r)

	var re *regexp.Regexp
	if *(config.FocusString) != "" {
		re = regexp.MustCompile(*config.FocusString)
	}

	suite.exampleCollection = newExampleCollection(t, description, suite.topLevelContainer.generateExamples(), re, reporter)
	if *(config.RandomizeAllSpecs) {
		suite.exampleCollection.shuffle(r)
	}

	suite.exampleCollection.run()
}

func (suite *suite) fail(message string, callerSkip int) {
	if suite.exampleCollection != nil {
		suite.exampleCollection.fail(failureData{
			message:      message,
			codeLocation: generateCodeLocation(callerSkip + 2),
		})
	}
}

func (suite *suite) pushContainerNode(text string, body func(), flag flagType, codeLocation CodeLocation) {
	container := newContainerNode(text, flag, codeLocation)
	suite.currentContainer.pushContainerNode(container)

	previousContainer := suite.currentContainer
	suite.currentContainer = container

	body()

	suite.currentContainer = previousContainer
}

func (suite *suite) pushItNode(text string, body interface{}, flag flagType, codeLocation CodeLocation, timeout time.Duration) {
	suite.currentContainer.pushItNode(newItNode(text, body, flag, codeLocation, timeout))
}

func (suite *suite) pushBeforeEachNode(body interface{}, codeLocation CodeLocation, timeout time.Duration) {
	suite.currentContainer.pushBeforeEachNode(newRunnableNode(body, codeLocation, timeout))
}

func (suite *suite) pushJustBeforeEachNode(body interface{}, codeLocation CodeLocation, timeout time.Duration) {
	suite.currentContainer.pushJustBeforeEachNode(newRunnableNode(body, codeLocation, timeout))
}

func (suite *suite) pushAfterEachNode(body interface{}, codeLocation CodeLocation, timeout time.Duration) {
	suite.currentContainer.pushAfterEachNode(newRunnableNode(body, codeLocation, timeout))
}
