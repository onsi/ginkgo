package ginkgo

import (
	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/types"

	"math/rand"
	"time"
)

type failureData struct {
	message        string
	codeLocation   types.CodeLocation
	forwardedPanic interface{}
}

type suite struct {
	topLevelContainer *containerNode
	currentContainer  *containerNode
	exampleCollection *exampleCollection
}

func newSuite() *suite {
	topLevelContainer := newContainerNode("[Top Level]", flagTypeNone, types.CodeLocation{})

	return &suite{
		topLevelContainer: topLevelContainer,
		currentContainer:  topLevelContainer,
	}
}

func (suite *suite) run(t GinkgoTestingT, description string, reporters []Reporter, config config.GinkgoConfigType) bool {
	r := rand.New(rand.NewSource(config.RandomSeed))
	suite.topLevelContainer.shuffle(r)

	if config.ParallelTotal < 1 {
		panic("ginkgo.parallel.total must be >= 1")
	}

	if config.ParallelNode > config.ParallelTotal || config.ParallelNode < 1 {
		panic("ginkgo.parallel.node is one-indexed and must be <= ginkgo.parallel.total")
	}

	suite.exampleCollection = newExampleCollection(t, description, suite.topLevelContainer.generateExamples(), reporters, config)

	return suite.exampleCollection.run()
}

func (suite *suite) fail(message string, callerSkip int) {
	if suite.exampleCollection != nil {
		suite.exampleCollection.fail(failureData{
			message:      message,
			codeLocation: types.GenerateCodeLocation(callerSkip + 2),
		})
	}
}

func (suite *suite) currentGinkgoTestDescription() GinkgoTestDescription {
	return suite.exampleCollection.currentGinkgoTestDescription()
}

func (suite *suite) pushContainerNode(text string, body func(), flag flagType, codeLocation types.CodeLocation) {
	container := newContainerNode(text, flag, codeLocation)
	suite.currentContainer.pushContainerNode(container)

	previousContainer := suite.currentContainer
	suite.currentContainer = container

	body()

	suite.currentContainer = previousContainer
}

func (suite *suite) pushItNode(text string, body interface{}, flag flagType, codeLocation types.CodeLocation, timeout time.Duration) {
	suite.currentContainer.pushSubjectNode(newItNode(text, body, flag, codeLocation, timeout))
}

func (suite *suite) pushMeasureNode(text string, body func(Benchmarker), flag flagType, codeLocation types.CodeLocation, samples int) {
	suite.currentContainer.pushSubjectNode(newMeasureNode(text, body, flag, codeLocation, samples))
}

func (suite *suite) pushBeforeEachNode(body interface{}, codeLocation types.CodeLocation, timeout time.Duration) {
	suite.currentContainer.pushBeforeEachNode(newRunnableNode(body, codeLocation, timeout))
}

func (suite *suite) pushJustBeforeEachNode(body interface{}, codeLocation types.CodeLocation, timeout time.Duration) {
	suite.currentContainer.pushJustBeforeEachNode(newRunnableNode(body, codeLocation, timeout))
}

func (suite *suite) pushAfterEachNode(body interface{}, codeLocation types.CodeLocation, timeout time.Duration) {
	suite.currentContainer.pushAfterEachNode(newRunnableNode(body, codeLocation, timeout))
}
