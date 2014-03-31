package suite

import (
	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/internal/containernode"
	"github.com/onsi/ginkgo/internal/example"
	"github.com/onsi/ginkgo/internal/failer"
	"github.com/onsi/ginkgo/internal/leafnodes"
	"github.com/onsi/ginkgo/internal/specrunner"
	"github.com/onsi/ginkgo/internal/writer"
	"github.com/onsi/ginkgo/reporters"
	"github.com/onsi/ginkgo/types"
	"math/rand"
	"time"
)

type ginkgoTestingT interface {
	Fail()
}

type Suite struct {
	topLevelContainer *containernode.ContainerNode
	currentContainer  *containernode.ContainerNode
	containerIndex    int
	runner            *specrunner.SpecRunner
	failer            *failer.Failer
}

func New(failer *failer.Failer) *Suite {
	topLevelContainer := containernode.New("[Top Level]", types.FlagTypeNone, types.CodeLocation{})

	return &Suite{
		topLevelContainer: topLevelContainer,
		currentContainer:  topLevelContainer,
		failer:            failer,
		containerIndex:    1,
	}
}

func (suite *Suite) Run(t ginkgoTestingT, description string, reporters []reporters.Reporter, writer writer.WriterInterface, config config.GinkgoConfigType) bool {
	if config.ParallelTotal < 1 {
		panic("ginkgo.parallel.total must be >= 1")
	}

	if config.ParallelNode > config.ParallelTotal || config.ParallelNode < 1 {
		panic("ginkgo.parallel.node is one-indexed and must be <= ginkgo.parallel.total")
	}

	r := rand.New(rand.NewSource(config.RandomSeed))
	suite.topLevelContainer.Shuffle(r)
	examples := suite.generateExamples(description, config)
	suite.runner = specrunner.New(description, examples, reporters, writer, config)

	success := suite.runner.Run()
	if !success {
		t.Fail()
	}
	return success
}

func (suite *Suite) generateExamples(description string, config config.GinkgoConfigType) *example.Examples {
	examplesSlice := []*example.Example{}
	for _, collatedNodes := range suite.topLevelContainer.Collate() {
		examplesSlice = append(examplesSlice, example.New(collatedNodes.Subject, collatedNodes.Containers))
	}

	examples := example.NewExamples(examplesSlice)

	if config.RandomizeAllSpecs {
		examples.Shuffle(rand.New(rand.NewSource(config.RandomSeed)))
	}

	examples.ApplyFocus(description, config.FocusString, config.SkipString)

	if config.SkipMeasurements {
		examples.SkipMeasurements()
	}

	if config.ParallelTotal > 1 {
		examples.TrimForParallelization(config.ParallelTotal, config.ParallelNode)
	}

	return examples
}

func (suite *Suite) CurrentRunningExampleSummary() (*types.ExampleSummary, bool) {
	return suite.runner.CurrentExampleSummary()
}

func (suite *Suite) PushContainerNode(text string, body func(), flag types.FlagType, codeLocation types.CodeLocation) {
	container := containernode.New(text, flag, codeLocation)
	suite.currentContainer.PushContainerNode(container)

	previousContainer := suite.currentContainer
	suite.currentContainer = container
	suite.containerIndex++

	body()

	suite.containerIndex--
	suite.currentContainer = previousContainer
}

func (suite *Suite) PushItNode(text string, body interface{}, flag types.FlagType, codeLocation types.CodeLocation, timeout time.Duration) {
	suite.currentContainer.PushSubjectNode(leafnodes.NewItNode(text, body, flag, codeLocation, timeout, suite.failer, suite.containerIndex))
}

func (suite *Suite) PushMeasureNode(text string, body interface{}, flag types.FlagType, codeLocation types.CodeLocation, samples int) {
	suite.currentContainer.PushSubjectNode(leafnodes.NewMeasureNode(text, body, flag, codeLocation, samples, suite.failer, suite.containerIndex))
}

func (suite *Suite) PushBeforeEachNode(body interface{}, codeLocation types.CodeLocation, timeout time.Duration) {
	suite.currentContainer.PushSetupNode(leafnodes.NewBeforeEachNode(body, codeLocation, timeout, suite.failer, suite.containerIndex))
}

func (suite *Suite) PushJustBeforeEachNode(body interface{}, codeLocation types.CodeLocation, timeout time.Duration) {
	suite.currentContainer.PushSetupNode(leafnodes.NewJustBeforeEachNode(body, codeLocation, timeout, suite.failer, suite.containerIndex))
}

func (suite *Suite) PushAfterEachNode(body interface{}, codeLocation types.CodeLocation, timeout time.Duration) {
	suite.currentContainer.PushSetupNode(leafnodes.NewAfterEachNode(body, codeLocation, timeout, suite.failer, suite.containerIndex))
}
