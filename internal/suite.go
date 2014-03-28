package internal

import (
	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/internal/containernode"
	"github.com/onsi/ginkgo/internal/example"
	"github.com/onsi/ginkgo/internal/failer"
	"github.com/onsi/ginkgo/internal/leafnodes"
	internaltypes "github.com/onsi/ginkgo/internal/types"
	"github.com/onsi/ginkgo/reporters"
	"github.com/onsi/ginkgo/types"

	"math/rand"
	"time"
)

type Suite struct {
	topLevelContainer *containernode.ContainerNode
	currentContainer  *containernode.ContainerNode
	containerIndex    int
	exampleCollection *exampleCollection
	failer            *failer.Failer
}

func NewSuite(failer *failer.Failer) *Suite {
	topLevelContainer := containernode.New("[Top Level]", internaltypes.FlagTypeNone, types.CodeLocation{})

	return &Suite{
		topLevelContainer: topLevelContainer,
		currentContainer:  topLevelContainer,
		failer:            failer,
		containerIndex:    1,
	}
}

func (suite *Suite) Run(t internaltypes.GinkgoTestingT, description string, reporters []reporters.Reporter, writer ginkgoWriterInterface, config config.GinkgoConfigType) bool {
	if config.ParallelTotal < 1 {
		panic("ginkgo.parallel.total must be >= 1")
	}

	if config.ParallelNode > config.ParallelTotal || config.ParallelNode < 1 {
		panic("ginkgo.parallel.node is one-indexed and must be <= ginkgo.parallel.total")
	}

	r := rand.New(rand.NewSource(config.RandomSeed))
	suite.topLevelContainer.Shuffle(r)
	suite.exampleCollection = newExampleCollection(t, description, suite.generateExamples(), reporters, writer, config)

	return suite.exampleCollection.run()
}

func (suite *Suite) generateExamples() []*example.Example {
	examples := []*example.Example{}
	for _, collatedNodes := range suite.topLevelContainer.Collate() {
		examples = append(examples, example.New(collatedNodes.Subject, collatedNodes.Containers))
	}
	return examples
}

func (suite *Suite) CurrentRunningExampleSummary() (*types.ExampleSummary, bool) {
	return suite.exampleCollection.currentExampleSummary()
}

func (suite *Suite) PushContainerNode(text string, body func(), flag internaltypes.FlagType, codeLocation types.CodeLocation) {
	container := containernode.New(text, flag, codeLocation)
	suite.currentContainer.PushContainerNode(container)

	previousContainer := suite.currentContainer
	suite.currentContainer = container
	suite.containerIndex++

	body()

	suite.containerIndex--
	suite.currentContainer = previousContainer
}

func (suite *Suite) PushItNode(text string, body interface{}, flag internaltypes.FlagType, codeLocation types.CodeLocation, timeout time.Duration) {
	suite.currentContainer.PushSubjectNode(leafnodes.NewItNode(text, body, flag, codeLocation, timeout, suite.failer, suite.containerIndex))
}

func (suite *Suite) PushMeasureNode(text string, body interface{}, flag internaltypes.FlagType, codeLocation types.CodeLocation, samples int) {
	suite.currentContainer.PushSubjectNode(leafnodes.NewMeasureNode(text, body, flag, codeLocation, samples, suite.failer, suite.containerIndex))
}

func (suite *Suite) PushBeforeEachNode(body interface{}, codeLocation types.CodeLocation, timeout time.Duration) {
	suite.currentContainer.PushBeforeEachNode(leafnodes.NewBeforeEachNode(body, codeLocation, timeout, suite.failer, suite.containerIndex))
}

func (suite *Suite) PushJustBeforeEachNode(body interface{}, codeLocation types.CodeLocation, timeout time.Duration) {
	suite.currentContainer.PushJustBeforeEachNode(leafnodes.NewJustBeforeEachNode(body, codeLocation, timeout, suite.failer, suite.containerIndex))
}

func (suite *Suite) PushAfterEachNode(body interface{}, codeLocation types.CodeLocation, timeout time.Duration) {
	suite.currentContainer.PushAfterEachNode(leafnodes.NewAfterEachNode(body, codeLocation, timeout, suite.failer, suite.containerIndex))
}
