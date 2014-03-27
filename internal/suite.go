package internal

import (
	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/internal/codelocation"
	internaltypes "github.com/onsi/ginkgo/internal/types"
	"github.com/onsi/ginkgo/reporters"
	"github.com/onsi/ginkgo/types"

	"math/rand"
	"time"
)

type failureData struct {
	message        string
	codeLocation   types.CodeLocation
	forwardedPanic interface{}
}

type Suite struct {
	topLevelContainer *containerNode
	currentContainer  *containerNode
	exampleCollection *exampleCollection
}

func NewSuite() *Suite {
	topLevelContainer := newContainerNode("[Top Level]", FlagTypeNone, types.CodeLocation{})

	return &Suite{
		topLevelContainer: topLevelContainer,
		currentContainer:  topLevelContainer,
	}
}

func (suite *Suite) Run(t internaltypes.GinkgoTestingT, description string, reporters []reporters.Reporter, writer ginkgoWriterInterface, config config.GinkgoConfigType) bool {
	r := rand.New(rand.NewSource(config.RandomSeed))
	suite.topLevelContainer.shuffle(r)

	if config.ParallelTotal < 1 {
		panic("ginkgo.parallel.total must be >= 1")
	}

	if config.ParallelNode > config.ParallelTotal || config.ParallelNode < 1 {
		panic("ginkgo.parallel.node is one-indexed and must be <= ginkgo.parallel.total")
	}

	suite.exampleCollection = newExampleCollection(t, description, suite.topLevelContainer.generateExamples(), reporters, writer, config)

	return suite.exampleCollection.run()
}

func (suite *Suite) Fail(message string, callerSkip int) {
	if suite.exampleCollection != nil {
		suite.exampleCollection.fail(failureData{
			message:      message,
			codeLocation: codelocation.New(callerSkip + 2),
		})
	}
}

func (suite *Suite) CurrentGinkgoTestDescription() internaltypes.GinkgoTestDescription {
	return suite.exampleCollection.currentGinkgoTestDescription()
}

func (suite *Suite) PushContainerNode(text string, body func(), flag FlagType, codeLocation types.CodeLocation) {
	container := newContainerNode(text, flag, codeLocation)
	suite.currentContainer.pushContainerNode(container)

	previousContainer := suite.currentContainer
	suite.currentContainer = container

	body()

	suite.currentContainer = previousContainer
}

func (suite *Suite) PushItNode(text string, body interface{}, flag FlagType, codeLocation types.CodeLocation, timeout time.Duration) {
	suite.currentContainer.pushSubjectNode(newItNode(text, body, flag, codeLocation, timeout))
}

func (suite *Suite) PushMeasureNode(text string, body interface{}, flag FlagType, codeLocation types.CodeLocation, samples int) {
	suite.currentContainer.pushSubjectNode(newMeasureNode(text, body, flag, codeLocation, samples))
}

func (suite *Suite) PushBeforeEachNode(body interface{}, codeLocation types.CodeLocation, timeout time.Duration) {
	suite.currentContainer.pushBeforeEachNode(newRunnableNode(body, codeLocation, timeout))
}

func (suite *Suite) PushJustBeforeEachNode(body interface{}, codeLocation types.CodeLocation, timeout time.Duration) {
	suite.currentContainer.pushJustBeforeEachNode(newRunnableNode(body, codeLocation, timeout))
}

func (suite *Suite) PushAfterEachNode(body interface{}, codeLocation types.CodeLocation, timeout time.Duration) {
	suite.currentContainer.pushAfterEachNode(newRunnableNode(body, codeLocation, timeout))
}
