package internal

import (
	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/internal/codelocation"
	"github.com/onsi/ginkgo/internal/containernode"
	"github.com/onsi/ginkgo/internal/leafnode"
	"github.com/onsi/ginkgo/internal/measurenode"
	internaltypes "github.com/onsi/ginkgo/internal/types"
	"github.com/onsi/ginkgo/reporters"
	"github.com/onsi/ginkgo/types"

	"math/rand"
	"time"
)

type Suite struct {
	topLevelContainer *containernode.ContainerNode
	currentContainer  *containernode.ContainerNode
	exampleCollection *exampleCollection
}

func NewSuite() *Suite {
	topLevelContainer := containernode.New("[Top Level]", internaltypes.FlagTypeNone, types.CodeLocation{})

	return &Suite{
		topLevelContainer: topLevelContainer,
		currentContainer:  topLevelContainer,
	}
}

func (suite *Suite) Run(t internaltypes.GinkgoTestingT, description string, reporters []reporters.Reporter, writer ginkgoWriterInterface, config config.GinkgoConfigType) bool {
	r := rand.New(rand.NewSource(config.RandomSeed))
	suite.topLevelContainer.Shuffle(r)

	if config.ParallelTotal < 1 {
		panic("ginkgo.parallel.total must be >= 1")
	}

	if config.ParallelNode > config.ParallelTotal || config.ParallelNode < 1 {
		panic("ginkgo.parallel.node is one-indexed and must be <= ginkgo.parallel.total")
	}

	suite.exampleCollection = newExampleCollection(t, description, suite.topLevelContainer.GenerateExamples(), reporters, writer, config)

	return suite.exampleCollection.run()
}

func (suite *Suite) Fail(message string, callerSkip int) {
	if suite.exampleCollection != nil {
		suite.exampleCollection.fail(types.ExampleFailure{
			Message:  message,
			Location: codelocation.New(callerSkip + 2),
		})
	}
}

func (suite *Suite) CurrentGinkgoTestDescription() internaltypes.GinkgoTestDescription {
	return suite.exampleCollection.currentGinkgoTestDescription()
}

func (suite *Suite) PushContainerNode(text string, body func(), flag internaltypes.FlagType, codeLocation types.CodeLocation) {
	container := containernode.New(text, flag, codeLocation)
	suite.currentContainer.PushContainerNode(container)

	previousContainer := suite.currentContainer
	suite.currentContainer = container

	body()

	suite.currentContainer = previousContainer
}

func (suite *Suite) PushItNode(text string, body interface{}, flag internaltypes.FlagType, codeLocation types.CodeLocation, timeout time.Duration) {
	suite.currentContainer.PushSubjectNode(leafnode.NewItNode(text, body, flag, codeLocation, timeout))
}

func (suite *Suite) PushMeasureNode(text string, body interface{}, flag internaltypes.FlagType, codeLocation types.CodeLocation, samples int) {
	suite.currentContainer.PushSubjectNode(measurenode.New(text, body, flag, codeLocation, samples))
}

func (suite *Suite) PushBeforeEachNode(body interface{}, codeLocation types.CodeLocation, timeout time.Duration) {
	suite.currentContainer.PushBeforeEachNode(leafnode.NewBeforeEachNode(body, codeLocation, timeout))
}

func (suite *Suite) PushJustBeforeEachNode(body interface{}, codeLocation types.CodeLocation, timeout time.Duration) {
	suite.currentContainer.PushJustBeforeEachNode(leafnode.NewJustBeforeEachNode(body, codeLocation, timeout))
}

func (suite *Suite) PushAfterEachNode(body interface{}, codeLocation types.CodeLocation, timeout time.Duration) {
	suite.currentContainer.PushAfterEachNode(leafnode.NewAfterEachNode(body, codeLocation, timeout))
}
