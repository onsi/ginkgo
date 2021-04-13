package internal

import (
	"fmt"
	"time"

	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/formatter"
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

	writer WriterInterface
	outputInterceptor
	currentSpecReport types.SpecReport
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

func (suite *Suite) Run(description string, failer *Failer, reporter reporters.Reporter, writer WriterInterface, outputInterceptor OutputInterceptor, interruptHandler InterruptHandlerInterface, config config.GinkgoConfigType) (bool, bool) {
	if suite.phase != PhaseBuildTree {
		panic("cannot run before building the tree = call suite.BuildTree() first")
	}
	tree := ApplyNestedFocusPolicyToTree(suite.tree)
	specs := GenerateSpecsFromTreeRoot(tree)
	specs = ShuffleSpecs(specs, config)
	specs, hasProgrammaticFocus := ApplyFocusToSpecs(specs, description, config)

	suite.phase = PhaseRun

	success := suite.runSpecs(description, specs, failer, reporter, writer, outputInterceptor, interruptHandler, config)
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
						err = types.GinkgoErrors.CaughtPanicDuringABuildPhase(e, node.CodeLocation)
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
func (suite *Suite) CurrentSpecReport() types.SpecReport {
	report := suite.currentSpecReport
	if suite.writer != nil {
		report.CapturedGinkgoWriterOutput = string(suite.writer.Bytes())
	}
	return report
}

func (suite *Suite) runSpecs(description string, specs Specs, failer *Failer, reporter reporters.Reporter, writer WriterInterface, outputInterceptor OutputInterceptor, interruptHandler InterruptHandlerInterface, config config.GinkgoConfigType) bool {
	suite.writer = writer

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
		report := types.SpecReport{LeafNodeType: beforeSuiteNode.NodeType, LeafNodeLocation: beforeSuiteNode.CodeLocation}
		reporter.WillRun(report)

		report = suite.runSuiteNode(report, beforeSuiteNode, failer, interruptStatus.Channel, writer, outputInterceptor, config)
		reporter.DidRun(report)

		if report.State != types.SpecStatePassed {
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

			suite.currentSpecReport = types.SpecReport{
				NodeTexts:        spec.Nodes.WithType(types.NodeTypesForContainerAndIt...).Texts(),
				NodeLocations:    spec.Nodes.WithType(types.NodeTypesForContainerAndIt...).CodeLocations(),
				LeafNodeLocation: spec.FirstNodeWithType(types.NodeTypeIt).CodeLocation,
				LeafNodeType:     types.NodeTypeIt,
			}

			if (config.FailFast && !suitePassed) || interruptHandler.Status().Interrupted {
				spec.Skip = true
			}

			if spec.Skip {
				suite.currentSpecReport.State = types.SpecStateSkipped
				if spec.Nodes.HasNodeMarkedPending() {
					suite.currentSpecReport.State = types.SpecStatePending
				}
			}

			reporter.WillRun(suite.currentSpecReport)

			if !spec.Skip {
				//runSpec updates suite.currentSpecReport directly
				suite.runSpec(spec, failer, interruptHandler, writer, outputInterceptor, config)
			}

			//send the spec report to any attached ReportAFterEach blocks - this will update sutie.currentSpecReport of failures occur in these blocks
			suite.reportAfterEach(suite.currentSpecReport, spec, failer, interruptHandler, writer, outputInterceptor, config)
			reporter.DidRun(suite.currentSpecReport)

			switch suite.currentSpecReport.State {
			case types.SpecStatePassed:
				suiteSummary.NumberOfPassedSpecs += 1
				if suite.currentSpecReport.NumAttempts > 1 {
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

			suite.currentSpecReport = types.SpecReport{}
		}

		if specs.HasAnySpecsMarkedPending() && config.FailOnPending {
			suitePassed = false
		}
	} else {
		suiteSummary.NumberOfSkippedSpecs = len(specs)
	}

	if !afterSuiteNode.IsZero() && numSpecsThatWillBeRun > 0 {
		report := types.SpecReport{LeafNodeType: afterSuiteNode.NodeType, LeafNodeLocation: afterSuiteNode.CodeLocation}
		reporter.WillRun(report)

		report = suite.runSuiteNode(report, afterSuiteNode, failer, interruptHandler.Status().Channel, writer, outputInterceptor, config)
		reporter.DidRun(report)
		if report.State != types.SpecStatePassed {
			suitePassed = false
		}
	}

	suiteSummary.SuiteSucceeded = suitePassed
	suiteSummary.RunTime = time.Since(suiteStartTime)
	reporter.SpecSuiteDidEnd(suiteSummary)

	return suitePassed
}

// runSpec(spec) mutates currentSpecReport.  this is ugly
// but it allows the user to call CurrentGinkgoSpecDescription and get
// an up-to-date state of the spec **from within a running spec**
func (suite *Suite) runSpec(spec Spec, failer *Failer, interruptHandler InterruptHandlerInterface, writer WriterInterface, outputInterceptor OutputInterceptor, config config.GinkgoConfigType) {
	if config.DryRun {
		suite.currentSpecReport.State = types.SpecStatePassed
		return
	}

	writer.Truncate()
	outputInterceptor.StartInterceptingOutput()
	t := time.Now()
	maxAttempts := max(1, config.FlakeAttempts)

	for attempt := 0; attempt < maxAttempts; attempt++ {
		suite.currentSpecReport.NumAttempts = attempt + 1

		if attempt > 0 {
			fmt.Fprintf(writer, "\nGinkgo: Attempt #%d Failed.  Retrying...\n", attempt)
		}

		interruptStatus := interruptHandler.Status()
		deepestNestingLevelAttained := -1
		nodes := spec.Nodes.WithType(types.NodeTypeBeforeEach).SortedByAscendingNestingLevel()
		nodes = nodes.CopyAppend(spec.Nodes.WithType(types.NodeTypeJustBeforeEach).SortedByAscendingNestingLevel()...)
		nodes = nodes.CopyAppend(spec.Nodes.WithType(types.NodeTypeIt)...)

		for _, node := range nodes {
			deepestNestingLevelAttained = max(deepestNestingLevelAttained, node.NestingLevel)
			suite.currentSpecReport.State, suite.currentSpecReport.Failure = suite.runNode(node, failer, interruptStatus.Channel, spec.Nodes.BestTextFor(node), writer, config)
			suite.currentSpecReport.RunTime = time.Since(t)
			if suite.currentSpecReport.State != types.SpecStatePassed {
				break
			}
		}

		cleanUpNodes := spec.Nodes.WithType(types.NodeTypeJustAfterEach).SortedByDescendingNestingLevel()
		cleanUpNodes = cleanUpNodes.CopyAppend(spec.Nodes.WithType(types.NodeTypeAfterEach).SortedByDescendingNestingLevel()...)
		cleanUpNodes = cleanUpNodes.WithinNestingLevel(deepestNestingLevelAttained)
		for _, node := range cleanUpNodes {
			state, failure := suite.runNode(node, failer, interruptHandler.Status().Channel, spec.Nodes.BestTextFor(node), writer, config)
			suite.currentSpecReport.RunTime = time.Since(t)
			if suite.currentSpecReport.State == types.SpecStatePassed {
				suite.currentSpecReport.State = state
				suite.currentSpecReport.Failure = failure
			}
		}

		suite.currentSpecReport.RunTime = time.Since(t)
		suite.currentSpecReport.CapturedGinkgoWriterOutput = string(writer.Bytes())
		suite.currentSpecReport.CapturedStdOutErr = outputInterceptor.StopInterceptingAndReturnOutput()

		if suite.currentSpecReport.State == types.SpecStatePassed {
			return
		}
		if interruptHandler.Status().Interrupted {
			return
		}
	}
}

func (suite *Suite) reportAfterEach(report types.SpecReport, spec Spec, failer *Failer, interruptHandler InterruptHandlerInterface, writer WriterInterface, outputInterceptor OutputInterceptor, config config.GinkgoConfigType) {
	nodes := spec.Nodes.WithType(types.NodeTypeReportAfterEach).SortedByDescendingNestingLevel()
	if len(nodes) == 0 {
		return
	}

	writer.Truncate()
	outputInterceptor.StartInterceptingOutput()
	for _, node := range nodes {
		node.Body = func() {
			node.ReportAfterEachBody(report)
		}
		interruptHandler.SetInterruptMessage(formatter.Fiw(0, formatter.COLS,
			"{{yellow}}Ginkgo received an interrupt signal but is currently running a ReportAfterEach node.  To avoid an invalid report the ReportAfterEach node will not be interrupted however subsequent tests will be skipped.{{/}}\n\n{{bold}}The running ReportAfterEach node is at:\n%s.{{/}}",
			node.CodeLocation,
		))
		state, failure := suite.runNode(node, failer, nil, spec.Nodes.BestTextFor(node), writer, config)
		interruptHandler.ClearInterruptMessage()
		if suite.currentSpecReport.State == types.SpecStatePassed {
			suite.currentSpecReport.State = state
			suite.currentSpecReport.Failure = failure
		}
	}
	suite.currentSpecReport.CapturedGinkgoWriterOutput += string(writer.Bytes())
	suite.currentSpecReport.CapturedStdOutErr += outputInterceptor.StopInterceptingAndReturnOutput()
}

func (suite *Suite) runSuiteNode(report types.SpecReport, node Node, failer *Failer, interruptChannel chan interface{}, writer WriterInterface, outputInterceptor OutputInterceptor, config config.GinkgoConfigType) types.SpecReport {
	if config.DryRun {
		report.State = types.SpecStatePassed
		return report
	}

	writer.Truncate()
	outputInterceptor.StartInterceptingOutput()
	t := time.Now()
	report.State, report.Failure = suite.runNode(node, failer, interruptChannel, "", writer, config)
	report.RunTime = time.Since(t)
	report.CapturedGinkgoWriterOutput = string(writer.Bytes())
	report.CapturedStdOutErr = outputInterceptor.StopInterceptingAndReturnOutput()

	return report
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
			Message:   interruptMessageWithStackTraces(),
			Location:  node.CodeLocation,
			NodeType:  node.NodeType,
			NodeIndex: failureNestingLevel,
		}
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
