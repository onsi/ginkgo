package internal

import (
	"fmt"
	"time"

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

func (suite *Suite) Run(description string, suitePath string, failer *Failer, reporter reporters.Reporter, writer WriterInterface, outputInterceptor OutputInterceptor, interruptHandler InterruptHandlerInterface, suiteConfig types.SuiteConfig) (bool, bool) {
	if suite.phase != PhaseBuildTree {
		panic("cannot run before building the tree = call suite.BuildTree() first")
	}
	tree := ApplyNestedFocusPolicyToTree(suite.tree)
	specs := GenerateSpecsFromTreeRoot(tree)
	specs = ShuffleSpecs(specs, suiteConfig)
	specs, hasProgrammaticFocus := ApplyFocusToSpecs(specs, description, suiteConfig)

	suite.phase = PhaseRun

	success := suite.runSpecs(description, suitePath, specs, failer, reporter, writer, outputInterceptor, interruptHandler, suiteConfig)
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
		// the user an opportunity to load suiteConfiguration information in the `TestX` go spec hook just before `RunSpecs`
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
		panic("invalid suiteConfiguration of SuiteNodeBuilder")
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

func (suite *Suite) runSpecs(description string, suitePath string, specs Specs, failer *Failer, reporter reporters.Reporter, writer WriterInterface, outputInterceptor OutputInterceptor, interruptHandler InterruptHandlerInterface, suiteConfig types.SuiteConfig) bool {
	suite.writer = writer

	beforeSuiteNode := suite.beforeSuiteNodeBuilder.BuildNode(suiteConfig, failer)
	afterSuiteNode := suite.afterSuiteNodeBuilder.BuildNode(suiteConfig, failer)

	numSpecsThatWillBeRun := specs.CountWithoutSkip()

	report := types.Report{
		SuitePath:        suitePath,
		SuiteDescription: description,
		SuiteConfig:      suiteConfig,
		PreRunStats: types.PreRunStats{
			TotalSpecs:       len(specs),
			SpecsThatWillRun: numSpecsThatWillBeRun,
		},
		StartTime: time.Now(),
	}

	reporter.SpecSuiteWillBegin(report)

	suitePassed := true

	interruptStatus := interruptHandler.Status()
	if !beforeSuiteNode.IsZero() && !interruptStatus.Interrupted && numSpecsThatWillBeRun > 0 {
		specReport := types.SpecReport{LeafNodeType: beforeSuiteNode.NodeType, LeafNodeLocation: beforeSuiteNode.CodeLocation, GinkgoParallelNode: suiteConfig.ParallelNode}
		reporter.WillRun(specReport)

		specReport = suite.runSuiteNode(specReport, beforeSuiteNode, failer, interruptStatus.Channel, writer, outputInterceptor, suiteConfig)
		reporter.DidRun(specReport)

		if specReport.State != types.SpecStatePassed {
			suitePassed = false
		}
		report.SpecReports = append(report.SpecReports, specReport)
	}

	if suitePassed {
		nextIndex := MakeNextIndexCounter(suiteConfig)

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
				NodeTexts:          spec.Nodes.WithType(types.NodeTypesForContainerAndIt...).Texts(),
				NodeLocations:      spec.Nodes.WithType(types.NodeTypesForContainerAndIt...).CodeLocations(),
				LeafNodeLocation:   spec.FirstNodeWithType(types.NodeTypeIt).CodeLocation,
				LeafNodeType:       types.NodeTypeIt,
				GinkgoParallelNode: suiteConfig.ParallelNode,
			}

			if (suiteConfig.FailFast && !suitePassed) || interruptHandler.Status().Interrupted {
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
				suite.runSpec(spec, failer, interruptHandler, writer, outputInterceptor, suiteConfig)
			}

			//send the spec report to any attached ReportAFterEach blocks - this will update sutie.currentSpecReport of failures occur in these blocks
			suite.reportAfterEach(suite.currentSpecReport, spec, failer, interruptHandler, writer, outputInterceptor, suiteConfig)
			reporter.DidRun(suite.currentSpecReport)

			if suite.currentSpecReport.State.Is(types.SpecStateFailureStates...) {
				suitePassed = false
			}
			report.SpecReports = append(report.SpecReports, suite.currentSpecReport)
			suite.currentSpecReport = types.SpecReport{}
		}

		if specs.HasAnySpecsMarkedPending() && suiteConfig.FailOnPending {
			suitePassed = false
		}
	}

	if !afterSuiteNode.IsZero() && numSpecsThatWillBeRun > 0 {
		specReport := types.SpecReport{LeafNodeType: afterSuiteNode.NodeType, LeafNodeLocation: afterSuiteNode.CodeLocation, GinkgoParallelNode: suiteConfig.ParallelNode}
		reporter.WillRun(specReport)

		specReport = suite.runSuiteNode(specReport, afterSuiteNode, failer, interruptHandler.Status().Channel, writer, outputInterceptor, suiteConfig)
		reporter.DidRun(specReport)
		if specReport.State != types.SpecStatePassed {
			suitePassed = false
		}
		report.SpecReports = append(report.SpecReports, specReport)
	}

	report.SuiteSucceeded = suitePassed
	report.EndTime = time.Now()
	report.RunTime = report.EndTime.Sub(report.StartTime)
	reporter.SpecSuiteDidEnd(report)

	return suitePassed
}

// runSpec(spec) mutates currentSpecReport.  this is ugly
// but it allows the user to call CurrentGinkgoSpecDescription and get
// an up-to-date state of the spec **from within a running spec**
func (suite *Suite) runSpec(spec Spec, failer *Failer, interruptHandler InterruptHandlerInterface, writer WriterInterface, outputInterceptor OutputInterceptor, suiteConfig types.SuiteConfig) {
	if suiteConfig.DryRun {
		suite.currentSpecReport.State = types.SpecStatePassed
		return
	}

	writer.Truncate()
	outputInterceptor.StartInterceptingOutput()
	suite.currentSpecReport.StartTime = time.Now()
	maxAttempts := max(1, suiteConfig.FlakeAttempts)

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
			suite.currentSpecReport.State, suite.currentSpecReport.Failure = suite.runNode(node, failer, interruptStatus.Channel, spec.Nodes.BestTextFor(node), writer, suiteConfig)
			suite.currentSpecReport.RunTime = time.Since(suite.currentSpecReport.StartTime)
			if suite.currentSpecReport.State != types.SpecStatePassed {
				break
			}
		}

		cleanUpNodes := spec.Nodes.WithType(types.NodeTypeJustAfterEach).SortedByDescendingNestingLevel()
		cleanUpNodes = cleanUpNodes.CopyAppend(spec.Nodes.WithType(types.NodeTypeAfterEach).SortedByDescendingNestingLevel()...)
		cleanUpNodes = cleanUpNodes.WithinNestingLevel(deepestNestingLevelAttained)
		for _, node := range cleanUpNodes {
			state, failure := suite.runNode(node, failer, interruptHandler.Status().Channel, spec.Nodes.BestTextFor(node), writer, suiteConfig)
			suite.currentSpecReport.RunTime = time.Since(suite.currentSpecReport.StartTime)
			if suite.currentSpecReport.State == types.SpecStatePassed {
				suite.currentSpecReport.State = state
				suite.currentSpecReport.Failure = failure
			}
		}

		suite.currentSpecReport.EndTime = time.Now()
		suite.currentSpecReport.RunTime = suite.currentSpecReport.EndTime.Sub(suite.currentSpecReport.StartTime)
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

func (suite *Suite) reportAfterEach(report types.SpecReport, spec Spec, failer *Failer, interruptHandler InterruptHandlerInterface, writer WriterInterface, outputInterceptor OutputInterceptor, suiteConfig types.SuiteConfig) {
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
		state, failure := suite.runNode(node, failer, nil, spec.Nodes.BestTextFor(node), writer, suiteConfig)
		interruptHandler.ClearInterruptMessage()
		if suite.currentSpecReport.State == types.SpecStatePassed {
			suite.currentSpecReport.State = state
			suite.currentSpecReport.Failure = failure
		}
	}
	suite.currentSpecReport.CapturedGinkgoWriterOutput += string(writer.Bytes())
	suite.currentSpecReport.CapturedStdOutErr += outputInterceptor.StopInterceptingAndReturnOutput()
}

func (suite *Suite) runSuiteNode(report types.SpecReport, node Node, failer *Failer, interruptChannel chan interface{}, writer WriterInterface, outputInterceptor OutputInterceptor, suiteConfig types.SuiteConfig) types.SpecReport {
	if suiteConfig.DryRun {
		report.State = types.SpecStatePassed
		return report
	}

	writer.Truncate()
	outputInterceptor.StartInterceptingOutput()
	report.StartTime = time.Now()
	report.State, report.Failure = suite.runNode(node, failer, interruptChannel, "", writer, suiteConfig)
	report.EndTime = time.Now()
	report.RunTime = report.EndTime.Sub(report.StartTime)
	report.CapturedGinkgoWriterOutput = string(writer.Bytes())
	report.CapturedStdOutErr = outputInterceptor.StopInterceptingAndReturnOutput()

	return report
}

func (suite *Suite) runNode(node Node, failer *Failer, interruptChannel chan interface{}, text string, writer WriterInterface, suiteConfig types.SuiteConfig) (types.SpecState, types.Failure) {
	if suiteConfig.EmitSpecProgress {
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
