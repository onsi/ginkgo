package internal

import (
	"fmt"
	"time"

	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/reporters"
	"github.com/onsi/ginkgo/types"
)

type Phase uint

const (
	PhaseBuildTopLevel Phase = iota
	PhaseBuildTree
	PhaseRun
)

type Suite struct {
	tree               TreeNode
	topLevelContainers Nodes

	phase Phase

	beforeSuiteNodeBuilder SuiteNodeBuilder
	afterSuiteNodeBuilder  SuiteNodeBuilder

	currentSpecSummary types.Summary
}

func NewSuite() *Suite {
	return &Suite{
		phase: PhaseBuildTopLevel,
	}
}

func (suite *Suite) BuildTree() error {
	// During PhaseBuildTopLevel, the top level containers are stored in suite.topLevelCotainers and entered
	// We now enter PhaseBuildTree where these top level containers are entered and added to the spec tree
	suite.phase = PhaseBuildTree
	for _, topLevelContainer := range suite.topLevelContainers {
		err := suite.PushNode(topLevelContainer)
		if err != nil {
			return err
		}
	}
	return nil
}

func (suite *Suite) Run(description string, failer *Failer, reporter reporters.Reporter, writer WriterInterface, interruptHandler InterruptHandlerInterface, config config.GinkgoConfigType) (bool, bool) {
	if suite.phase != PhaseBuildTree {
		panic("cannot run before building the tree = call suite.BuildTree() first")
	}
	tree := ApplyNestedFocusPolicyToTree(suite.tree)
	specs := GenerateSpecsFromTreeRoot(tree)
	specs = ShuffleSpecs(specs, config)
	specs, hasProgrammaticFocus := ApplyFocusToSpecs(specs, description, config)

	suite.phase = PhaseRun
	if interruptHandler == nil {
		interruptHandler = NewInterruptHandler()
	}
	success := suite.runSpecs(description, specs, failer, reporter, writer, interruptHandler, config)
	return success, hasProgrammaticFocus
}

/*
  Tree Construction methods

  PushNode and PushSuiteNodeBuilder are used during PhaseBuildTopLevel and PhaseBuildTree
*/

func (suite *Suite) PushNode(node Node) error {
	if suite.phase == PhaseRun {
		return types.GinkgoErrors.PushingNodeInRunPhase(node.NodeType, node.CodeLocation)
	}

	if node.NodeType == types.NodeTypeContainer {
		// During PhaseBuildTopLevel we only track the top level containers without entering them
		// We only enter the top level container nodes during PhaseBuildTree
		//
		// This ensures the tree is only constructed after `go spec` has called `flag.Parse()` and gives
		// the user an opportunity to load configuration information in the `TestX` go spec hook just before `RunSpecs`
		// is invoked.  This makes the lifecycle easier to reason about and solves issues like #693.
		if suite.phase == PhaseBuildTopLevel {
			suite.topLevelContainers = append(suite.topLevelContainers, node)
			return nil
		}
		if suite.phase == PhaseBuildTree {
			parentTree := suite.tree
			suite.tree = TreeNode{Node: node}
			err := func() (err error) {
				defer func() {
					if e := recover(); e != nil {
						err = types.GinkgoErrors.CaughtPanicDuringABuildPhase(node.CodeLocation)
					}
				}()
				node.Body()
				return err
			}()
			suite.tree = AppendTreeNodeChild(parentTree, suite.tree)
			return err
		}
	} else {
		suite.tree = AppendTreeNodeChild(suite.tree, TreeNode{Node: node})
		return nil
	}

	return nil
}

func (suite *Suite) PushSuiteNodeBuilder(nodeBuilder SuiteNodeBuilder) error {
	if suite.phase == PhaseBuildTree {
		return types.GinkgoErrors.SetupNodeInNestedContext(nodeBuilder.NodeType, nodeBuilder.CodeLocation)
	}

	if suite.phase == PhaseRun {
		return types.GinkgoErrors.SetupNodeDuringRunPhase(nodeBuilder.NodeType, nodeBuilder.CodeLocation)
	}

	switch nodeBuilder.NodeType {
	case types.NodeTypeBeforeSuite, types.NodeTypeSynchronizedBeforeSuite:
		if suite.beforeSuiteNodeBuilder.NodeType != types.NodeTypeInvalid {
			return types.GinkgoErrors.MultipleBeforeSuiteNodes(nodeBuilder.NodeType, nodeBuilder.CodeLocation, suite.beforeSuiteNodeBuilder.NodeType, suite.beforeSuiteNodeBuilder.CodeLocation)
		}
		suite.beforeSuiteNodeBuilder = nodeBuilder
	case types.NodeTypeAfterSuite, types.NodeTypeSynchronizedAfterSuite:
		if suite.afterSuiteNodeBuilder.NodeType != types.NodeTypeInvalid {
			return types.GinkgoErrors.MultipleAfterSuiteNodes(nodeBuilder.NodeType, nodeBuilder.CodeLocation, suite.beforeSuiteNodeBuilder.NodeType, suite.beforeSuiteNodeBuilder.CodeLocation)
		}
		suite.afterSuiteNodeBuilder = nodeBuilder
	default:
		panic("invalid configuration of SuiteNodeBuilder")
	}

	return nil
}

/*
  Spec Running methods - used during PhaseRun
*/
func (suite *Suite) CurrentSpecSummary() (types.Summary, bool) {
	if suite.currentSpecSummary.State == types.SpecStateInvalid {
		return types.Summary{}, false
	}
	return suite.currentSpecSummary, true
}

func (suite *Suite) runSpecs(description string, specs Specs, failer *Failer, reporter reporters.Reporter, writer WriterInterface, interruptHandler InterruptHandlerInterface, config config.GinkgoConfigType) bool {
	suiteStartTime := time.Now()

	beforeSuiteNode := suite.beforeSuiteNodeBuilder.BuildNode(config, failer)
	afterSuiteNode := suite.afterSuiteNodeBuilder.BuildNode(config, failer)

	numSpecsThatWillBeRun := specs.CountWithoutSkip()
	suiteSummary := types.SuiteSummary{
		SuiteDescription:           description,
		NumberOfTotalSpecs:         len(specs),
		NumberOfSpecsThatWillBeRun: numSpecsThatWillBeRun,
	}

	reporter.SpecSuiteWillBegin(config, suiteSummary)

	suitePassed := true

	interruptStatus := interruptHandler.Status()
	if !beforeSuiteNode.IsZero() && !interruptStatus.Interrupted && numSpecsThatWillBeRun > 0 {
		summary := types.Summary{LeafNodeType: beforeSuiteNode.NodeType, LeafNodeLocation: beforeSuiteNode.CodeLocation}
		reporter.WillRun(summary)

		summary = suite.runSuiteNode(summary, beforeSuiteNode, failer, interruptStatus.Channel, writer, config)
		reporter.DidRun(summary)

		if summary.State != types.SpecStatePassed {
			suitePassed = false
		}
	}

	if suitePassed {
		nextIndex := MakeNextIndexCounter(config)

		for {
			idx, err := nextIndex()
			if err != nil {
				fmt.Println("failed to iterate over specs:\n" + err.Error())
				suitePassed = false
				break
			}
			if idx >= len(specs) {
				break
			}

			spec := specs[idx]

			suite.currentSpecSummary = types.Summary{
				NodeTexts:        spec.Nodes.WithType(types.NodeTypesForContainerAndIt...).Texts(),
				NodeLocations:    spec.Nodes.WithType(types.NodeTypesForContainerAndIt...).CodeLocations(),
				LeafNodeLocation: spec.FirstNodeWithType(types.NodeTypeIt).CodeLocation,
				LeafNodeType:     types.NodeTypeIt,
			}

			if (config.FailFast && !suitePassed) || interruptHandler.Status().Interrupted {
				spec.Skip = true
			}

			if spec.Skip {
				suite.currentSpecSummary.State = types.SpecStateSkipped
				if spec.Nodes.HasNodeMarkedPending() {
					suite.currentSpecSummary.State = types.SpecStatePending
				}
			}

			reporter.WillRun(suite.currentSpecSummary)

			if !spec.Skip {
				//runSpec updates suite.currentSpecSummary directly
				suite.runSpec(spec, failer, interruptHandler, writer, config)
			}

			reporter.DidRun(suite.currentSpecSummary)

			switch suite.currentSpecSummary.State {
			case types.SpecStatePassed:
				suiteSummary.NumberOfPassedSpecs += 1
				if suite.currentSpecSummary.NumAttempts > 1 {
					suiteSummary.NumberOfFlakedSpecs += 1
				}
			case types.SpecStateFailed, types.SpecStatePanicked, types.SpecStateInterrupted:
				suitePassed = false
				suiteSummary.NumberOfFailedSpecs += 1
			case types.SpecStatePending:
				suiteSummary.NumberOfPendingSpecs += 1
				suiteSummary.NumberOfSkippedSpecs += 1
			case types.SpecStateSkipped:
				suiteSummary.NumberOfSkippedSpecs += 1
			default:
			}

			suite.currentSpecSummary = types.Summary{}
		}

		if specs.HasAnySpecsMarkedPending() && config.FailOnPending {
			suitePassed = false
		}
	} else {
		suiteSummary.NumberOfSkippedSpecs = len(specs)
	}

	if !afterSuiteNode.IsZero() && numSpecsThatWillBeRun > 0 {
		summary := types.Summary{LeafNodeType: afterSuiteNode.NodeType, LeafNodeLocation: afterSuiteNode.CodeLocation}
		reporter.WillRun(summary)

		summary = suite.runSuiteNode(summary, afterSuiteNode, failer, interruptHandler.Status().Channel, writer, config)
		reporter.DidRun(summary)
		if summary.State != types.SpecStatePassed {
			suitePassed = false
		}
	}

	suiteSummary.SuiteSucceeded = suitePassed
	suiteSummary.RunTime = time.Since(suiteStartTime)
	reporter.SpecSuiteDidEnd(suiteSummary)

	return suitePassed
}

// runSpec(spec) mutates currentSpecSummary.  this is ugly
// but it allows the user to call CurrentGinkgoSpecDescription and get
// an up-to-date state of the spec **from within a running spec**
func (suite *Suite) runSpec(spec Spec, failer *Failer, interruptHandler InterruptHandlerInterface, writer WriterInterface, config config.GinkgoConfigType) {
	if config.DryRun {
		suite.currentSpecSummary.State = types.SpecStatePassed
		return
	}

	writer.Truncate()
	t := time.Now()
	maxAttempts := max(1, config.FlakeAttempts)

	for attempt := 0; attempt < maxAttempts; attempt++ {
		suite.currentSpecSummary.NumAttempts = attempt + 1

		if attempt > 0 {
			fmt.Fprintf(writer, "\nGinkgo: Attempt #%d Failed.  Retrying...\n", attempt)
		}

		interruptStatus := interruptHandler.Status()
		deepestNestingLevelAttained := -1
		nodes := spec.Nodes.WithType(types.NodeTypeBeforeEach).CopyAppend(spec.Nodes.WithType(types.NodeTypeJustBeforeEach)...).CopyAppend(spec.Nodes.WithType(types.NodeTypeIt)...)
		for _, node := range nodes {
			deepestNestingLevelAttained = max(deepestNestingLevelAttained, node.NestingLevel)
			suite.currentSpecSummary.State, suite.currentSpecSummary.Failure = suite.runNode(node, failer, interruptStatus.Channel, spec.Nodes.BestTextFor(node), writer, config)
			suite.currentSpecSummary.RunTime = time.Since(t)
			if suite.currentSpecSummary.State != types.SpecStatePassed {
				break
			}
		}

		cleanUpNodes := spec.Nodes.WithType(types.NodeTypeJustAfterEach).SortedByDescendingNestingLevel()
		cleanUpNodes = cleanUpNodes.CopyAppend(spec.Nodes.WithType(types.NodeTypeAfterEach).SortedByDescendingNestingLevel()...)
		cleanUpNodes = cleanUpNodes.WithinNestingLevel(deepestNestingLevelAttained)
		for _, node := range cleanUpNodes {
			state, failure := suite.runNode(node, failer, interruptHandler.Status().Channel, spec.Nodes.BestTextFor(node), writer, config)
			suite.currentSpecSummary.RunTime = time.Since(t)
			if suite.currentSpecSummary.State == types.SpecStatePassed {
				suite.currentSpecSummary.State = state
				suite.currentSpecSummary.Failure = failure
			}
		}

		suite.currentSpecSummary.RunTime = time.Since(t)
		suite.currentSpecSummary.CapturedGinkgoWriterOutput = string(writer.Bytes())

		if suite.currentSpecSummary.State == types.SpecStatePassed {
			return
		}
		if interruptHandler.Status().Interrupted {
			return
		}
	}
}

func (suite *Suite) runNode(node Node, failer *Failer, interruptChannel chan interface{}, text string, writer WriterInterface, config config.GinkgoConfigType) (types.SpecState, types.Failure) {
	if config.EmitSpecProgress {
		if text == "" {
			text = "TOP-LEVEL"
		}
		s := fmt.Sprintf("[%s] %s\n  %s\n", node.NodeType.String(), text, node.CodeLocation.String())
		writer.Write([]byte(s))
	}

	failureNestingLevel := node.NestingLevel - 1
	if node.NodeType.Is(types.NodeTypeIt) {
		failureNestingLevel = node.NestingLevel
	}

	outcomeC := make(chan types.SpecState)
	failureC := make(chan types.Failure)

	go func() {
		finished := false
		defer func() {
			if e := recover(); e != nil || !finished {
				failer.Panic(types.NewCodeLocation(2), e)
			}

			outcome, failure := failer.Drain()
			if outcome != types.SpecStatePassed {
				failure.NodeType = node.NodeType
				failure.NodeIndex = failureNestingLevel
			}

			outcomeC <- outcome
			failureC <- failure
		}()

		node.Body()
		finished = true
	}()

	select {
	case outcome := <-outcomeC:
		failure := <-failureC
		return outcome, failure
	case <-interruptChannel:
		return types.SpecStateInterrupted, types.Failure{
			Message:   "interrupted by user",
			Location:  node.CodeLocation,
			NodeType:  node.NodeType,
			NodeIndex: failureNestingLevel,
		}
	}
}

func (suite *Suite) runSuiteNode(summary types.Summary, node Node, failer *Failer, interruptChannel chan interface{}, writer WriterInterface, config config.GinkgoConfigType) types.Summary {
	if config.DryRun {
		summary.State = types.SpecStatePassed
		return summary
	}

	writer.Truncate()
	t := time.Now()
	summary.State, summary.Failure = suite.runNode(node, failer, interruptChannel, "", writer, config)
	summary.RunTime = time.Since(t)
	summary.CapturedGinkgoWriterOutput = string(writer.Bytes())

	return summary
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
