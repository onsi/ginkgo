package internal

import (
	"fmt"
	"time"

	"github.com/onsi/ginkgo/formatter"
	"github.com/onsi/ginkgo/internal/interrupt_handler"
	"github.com/onsi/ginkgo/internal/parallel_support"
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
	tree               *TreeNode
	topLevelContainers Nodes

	phase Phase

	suiteNodes   Nodes
	cleanupNodes Nodes

	writer            WriterInterface
	currentSpecReport types.SpecReport
	currentNode       Node

	client parallel_support.Client
}

func NewSuite() *Suite {
	return &Suite{
		tree:  &TreeNode{},
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

func (suite *Suite) Run(description string, suitePath string, failer *Failer, reporter reporters.Reporter, writer WriterInterface, outputInterceptor OutputInterceptor, interruptHandler interrupt_handler.InterruptHandlerInterface, suiteConfig types.SuiteConfig) (bool, bool) {
	if suite.phase != PhaseBuildTree {
		panic("cannot run before building the tree = call suite.BuildTree() first")
	}
	ApplyNestedFocusPolicyToTree(suite.tree)
	specs := GenerateSpecsFromTreeRoot(suite.tree)
	specs, hasProgrammaticFocus := ApplyFocusToSpecs(specs, description, suiteConfig)

	suite.phase = PhaseRun
	if suiteConfig.ParallelTotal > 1 {
		suite.client = parallel_support.NewClient(suiteConfig.ParallelHost)
	}

	success := suite.runSpecs(description, suitePath, hasProgrammaticFocus, specs, failer, reporter, writer, outputInterceptor, interruptHandler, suiteConfig)
	return success, hasProgrammaticFocus
}

/*
  Tree Construction methods

  PushNode is used during PhaseBuildTopLevel and PhaseBuildTree
*/

func (suite *Suite) PushNode(node Node) error {
	if node.NodeType.Is(types.NodeTypeCleanupInvalid, types.NodeTypeCleanupAfterEach, types.NodeTypeCleanupAfterAll, types.NodeTypeCleanupAfterSuite) {
		return suite.pushCleanupNode(node)
	}

	if node.NodeType.Is(types.NodeTypeBeforeSuite, types.NodeTypeAfterSuite, types.NodeTypeSynchronizedBeforeSuite, types.NodeTypeSynchronizedAfterSuite, types.NodeTypeReportAfterSuite) {
		return suite.pushSuiteNode(node)
	}

	if suite.phase == PhaseRun {
		return types.GinkgoErrors.PushingNodeInRunPhase(node.NodeType, node.CodeLocation)
	}

	if node.MarkedSerial {
		firstOrderedNode := suite.tree.AncestorNodeChain().FirstNodeMarkedOrdered()
		if !firstOrderedNode.IsZero() && !firstOrderedNode.MarkedSerial {
			return types.GinkgoErrors.InvalidSerialNodeInNonSerialOrderedContainer(node.CodeLocation, node.NodeType)
		}
	}

	if node.MarkedOrdered {
		firstOrderedNode := suite.tree.AncestorNodeChain().FirstNodeMarkedOrdered()
		if !firstOrderedNode.IsZero() {
			return types.GinkgoErrors.InvalidNestedOrderedContainer(node.CodeLocation)
		}
	}

	if node.NodeType.Is(types.NodeTypeBeforeAll, types.NodeTypeAfterAll) && !suite.tree.Node.MarkedOrdered {
		return types.GinkgoErrors.SetupNodeNotInOrderedContainer(node.CodeLocation, node.NodeType)
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
			suite.tree = &TreeNode{Node: node}
			parentTree.AppendChild(suite.tree)
			err := func() (err error) {
				defer func() {
					if e := recover(); e != nil {
						err = types.GinkgoErrors.CaughtPanicDuringABuildPhase(e, node.CodeLocation)
					}
				}()
				node.Body()
				return err
			}()
			suite.tree = parentTree
			return err
		}
	} else {
		suite.tree.AppendChild(&TreeNode{Node: node})
		return nil
	}

	return nil
}

func (suite *Suite) pushSuiteNode(node Node) error {
	if suite.phase == PhaseBuildTree {
		return types.GinkgoErrors.SuiteNodeInNestedContext(node.NodeType, node.CodeLocation)
	}

	if suite.phase == PhaseRun {
		return types.GinkgoErrors.SuiteNodeDuringRunPhase(node.NodeType, node.CodeLocation)
	}

	switch node.NodeType {
	case types.NodeTypeBeforeSuite, types.NodeTypeSynchronizedBeforeSuite:
		existingBefores := suite.suiteNodes.WithType(types.NodeTypeBeforeSuite, types.NodeTypeSynchronizedBeforeSuite)
		if len(existingBefores) > 0 {
			return types.GinkgoErrors.MultipleBeforeSuiteNodes(node.NodeType, node.CodeLocation, existingBefores[0].NodeType, existingBefores[0].CodeLocation)
		}
	case types.NodeTypeAfterSuite, types.NodeTypeSynchronizedAfterSuite:
		existingAfters := suite.suiteNodes.WithType(types.NodeTypeAfterSuite, types.NodeTypeSynchronizedAfterSuite)
		if len(existingAfters) > 0 {
			return types.GinkgoErrors.MultipleAfterSuiteNodes(node.NodeType, node.CodeLocation, existingAfters[0].NodeType, existingAfters[0].CodeLocation)
		}
	}

	suite.suiteNodes = append(suite.suiteNodes, node)
	return nil
}

func (suite *Suite) pushCleanupNode(node Node) error {
	if suite.phase != PhaseRun || suite.currentNode.IsZero() {
		return types.GinkgoErrors.PushingCleanupNodeDuringTreeConstruction(node.CodeLocation)
	}

	switch suite.currentNode.NodeType {
	case types.NodeTypeBeforeSuite, types.NodeTypeSynchronizedBeforeSuite, types.NodeTypeAfterSuite, types.NodeTypeSynchronizedAfterSuite:
		node.NodeType = types.NodeTypeCleanupAfterSuite
	case types.NodeTypeBeforeAll, types.NodeTypeAfterAll:
		node.NodeType = types.NodeTypeCleanupAfterAll
	case types.NodeTypeReportBeforeEach, types.NodeTypeReportAfterEach, types.NodeTypeReportAfterSuite:
		return types.GinkgoErrors.PushingCleanupInReportingNode(node.CodeLocation, suite.currentNode.NodeType)
	case types.NodeTypeCleanupInvalid, types.NodeTypeCleanupAfterEach, types.NodeTypeCleanupAfterAll, types.NodeTypeCleanupAfterSuite:
		return types.GinkgoErrors.PushingCleanupInCleanupNode(node.CodeLocation)
	default:
		node.NodeType = types.NodeTypeCleanupAfterEach
	}

	node.NestingLevel = suite.currentNode.NestingLevel
	suite.cleanupNodes = append(suite.cleanupNodes, node)

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

func (suite *Suite) AddReportEntry(entry ReportEntry) error {
	if suite.phase != PhaseRun {
		return types.GinkgoErrors.AddReportEntryNotDuringRunPhase(entry.Location)
	}
	suite.currentSpecReport.ReportEntries = append(suite.currentSpecReport.ReportEntries, entry)
	return nil
}

func (suite *Suite) runSpecs(description string, suitePath string, hasProgrammaticFocus bool, specs Specs, failer *Failer, reporter reporters.Reporter, writer WriterInterface, outputInterceptor OutputInterceptor, interruptHandler interrupt_handler.InterruptHandlerInterface, suiteConfig types.SuiteConfig) bool {
	suite.writer = writer

	numSpecsThatWillBeRun := specs.CountWithoutSkip()

	report := types.Report{
		SuitePath:                 suitePath,
		SuiteDescription:          description,
		SuiteConfig:               suiteConfig,
		SuiteHasProgrammaticFocus: hasProgrammaticFocus,
		PreRunStats: types.PreRunStats{
			TotalSpecs:       len(specs),
			SpecsThatWillRun: numSpecsThatWillBeRun,
		},
		StartTime: time.Now(),
	}

	reporter.SuiteWillBegin(report)
	if suiteConfig.ParallelTotal > 1 {
		suite.client.PostSuiteWillBegin(report)
	}

	report.SuiteSucceeded = true

	processSpecReport := func(specReport types.SpecReport) {
		reporter.DidRun(specReport)
		if suiteConfig.ParallelTotal > 1 {
			suite.client.PostDidRun(specReport)
		}
		if specReport.State.Is(types.SpecStateFailureStates...) {
			report.SuiteSucceeded = false
		}
		report.SpecReports = append(report.SpecReports, specReport)
	}

	interruptStatus := interruptHandler.Status()
	beforeSuiteNode := suite.suiteNodes.FirstNodeWithType(types.NodeTypeBeforeSuite, types.NodeTypeSynchronizedBeforeSuite)
	if !beforeSuiteNode.IsZero() && !interruptStatus.Interrupted && numSpecsThatWillBeRun > 0 {
		suite.currentSpecReport = types.SpecReport{
			LeafNodeType:       beforeSuiteNode.NodeType,
			LeafNodeLocation:   beforeSuiteNode.CodeLocation,
			GinkgoParallelNode: suiteConfig.ParallelNode,
		}
		reporter.WillRun(suite.currentSpecReport)
		suite.runSuiteNode(beforeSuiteNode, failer, interruptStatus.Channel, interruptHandler, writer, outputInterceptor, suiteConfig)
		processSpecReport(suite.currentSpecReport)
	}

	suiteAborted := false
	if report.SuiteSucceeded {
		groupedSpecIndices, serialGroupedSpecIndices := OrderSpecs(specs, suiteConfig)
		nextIndex := MakeNextIndexCounter(suiteConfig)

		for {
			groupedSpecIdx, err := nextIndex()
			if err != nil {
				report.SpecialSuiteFailureReasons = append(report.SpecialSuiteFailureReasons, fmt.Sprintf("Failed to iterate over specs:\n%s", err.Error()))
				report.SuiteSucceeded = false
				break
			}
			if groupedSpecIdx >= len(groupedSpecIndices) {
				if suiteConfig.ParallelNode == 1 && len(serialGroupedSpecIndices) > 0 {
					groupedSpecIndices, serialGroupedSpecIndices, nextIndex = serialGroupedSpecIndices, GroupedSpecIndices{}, MakeIncrementingIndexCounter()
					suite.client.BlockUntilNonprimaryNodesHaveFinished()
					continue
				}
				break
			}

			specIndices := groupedSpecIndices[groupedSpecIdx]
			groupSucceeded := true

			firstToRun, lastToRun := -1, -1
			for _, idx := range specIndices {
				if !specs[idx].Skip {
					if firstToRun == -1 {
						firstToRun = idx
					}
					lastToRun = idx
				}
			}

			for _, idx := range specIndices {
				spec := specs[idx]

				suite.currentSpecReport = types.SpecReport{
					ContainerHierarchyTexts:     spec.Nodes.WithType(types.NodeTypeContainer).Texts(),
					ContainerHierarchyLocations: spec.Nodes.WithType(types.NodeTypeContainer).CodeLocations(),
					ContainerHierarchyLabels:    spec.Nodes.WithType(types.NodeTypeContainer).Labels(),
					LeafNodeLocation:            spec.FirstNodeWithType(types.NodeTypeIt).CodeLocation,
					LeafNodeType:                types.NodeTypeIt,
					LeafNodeText:                spec.FirstNodeWithType(types.NodeTypeIt).Text,
					LeafNodeLabels:              []string(spec.FirstNodeWithType(types.NodeTypeIt).Labels),
					GinkgoParallelNode:          suiteConfig.ParallelNode,
				}

				skipReason := ""
				if (suiteConfig.FailFast && !report.SuiteSucceeded) || interruptHandler.Status().Interrupted || suiteAborted {
					spec.Skip = true
				}
				if !groupSucceeded {
					spec.Skip, skipReason = true, "Spec skipped because an earlier spec in an ordered container failed"
				}

				if spec.Skip {
					if spec.Nodes.HasNodeMarkedPending() {
						suite.currentSpecReport.State = types.SpecStatePending
					} else {
						suite.currentSpecReport.State = types.SpecStateSkipped
						if skipReason != "" {
							suite.currentSpecReport.Failure = suite.failureForLeafNodeWithMessage(spec.FirstNodeWithType(types.NodeTypeIt), skipReason)
						}
					}
				}

				reporter.WillRun(suite.currentSpecReport)
				//send the spec report to any attached ReportBeforeEach blocks - this will update suite.currentSpecReport if failures occur in these blocks
				suite.reportEach(spec, types.NodeTypeReportBeforeEach, failer, interruptHandler, writer, outputInterceptor, suiteConfig)

				if !spec.Skip && !suite.currentSpecReport.State.Is(types.SpecStateFailureStates...) {
					setupAllNodeTypesToInclude := types.NodeTypes{}
					if firstToRun == idx {
						setupAllNodeTypesToInclude = append(setupAllNodeTypesToInclude, types.NodeTypeBeforeAll)
					}
					if lastToRun == idx {
						setupAllNodeTypesToInclude = append(setupAllNodeTypesToInclude, types.NodeTypeAfterAll)
					}
					//runSpec updates suite.currentSpecReport directly
					suite.runSpec(spec, setupAllNodeTypesToInclude, failer, interruptHandler, writer, outputInterceptor, suiteConfig)
				}

				//send the spec report to any attached ReportAfterEach blocks - this will update suite.currentSpecReport if failures occur in these blocks
				suite.reportEach(spec, types.NodeTypeReportAfterEach, failer, interruptHandler, writer, outputInterceptor, suiteConfig)
				processSpecReport(suite.currentSpecReport)
				if suite.currentSpecReport.State.Is(types.SpecStateFailureStates...) {
					groupSucceeded = false
				}
				if suite.currentSpecReport.State == types.SpecStateAborted {
					suiteAborted = true
				}
				if suiteConfig.ParallelTotal > 1 && (suiteAborted || (suiteConfig.FailFast && !report.SuiteSucceeded)) {
					suite.client.PostAbort()
				}
				suite.currentSpecReport = types.SpecReport{}
			}
		}

		if specs.HasAnySpecsMarkedPending() && suiteConfig.FailOnPending {
			report.SpecialSuiteFailureReasons = append(report.SpecialSuiteFailureReasons, "Detected pending specs and --fail-on-pending is set")
			report.SuiteSucceeded = false
		}
	}

	afterSuiteNode := suite.suiteNodes.FirstNodeWithType(types.NodeTypeAfterSuite, types.NodeTypeSynchronizedAfterSuite)
	if !afterSuiteNode.IsZero() && numSpecsThatWillBeRun > 0 {
		suite.currentSpecReport = types.SpecReport{
			LeafNodeType:       afterSuiteNode.NodeType,
			LeafNodeLocation:   afterSuiteNode.CodeLocation,
			GinkgoParallelNode: suiteConfig.ParallelNode,
		}
		reporter.WillRun(suite.currentSpecReport)
		suite.runSuiteNode(afterSuiteNode, failer, interruptHandler.Status().Channel, interruptHandler, writer, outputInterceptor, suiteConfig)
		processSpecReport(suite.currentSpecReport)
	}

	afterSuiteCleanup := suite.cleanupNodes.WithType(types.NodeTypeCleanupAfterSuite).Reverse()
	suite.cleanupNodes = suite.cleanupNodes.WithoutType(types.NodeTypeCleanupAfterSuite)
	if len(afterSuiteCleanup) > 0 {
		for _, cleanupNode := range afterSuiteCleanup {
			suite.currentSpecReport = types.SpecReport{
				LeafNodeType:       cleanupNode.NodeType,
				LeafNodeLocation:   cleanupNode.CodeLocation,
				GinkgoParallelNode: suiteConfig.ParallelNode,
			}
			reporter.WillRun(suite.currentSpecReport)
			suite.runSuiteNode(cleanupNode, failer, interruptHandler.Status().Channel, interruptHandler, writer, outputInterceptor, suiteConfig)
			processSpecReport(suite.currentSpecReport)
		}
	}

	interruptStatus = interruptHandler.Status()
	if interruptStatus.Interrupted {
		report.SpecialSuiteFailureReasons = append(report.SpecialSuiteFailureReasons, interruptStatus.Cause.String())
		report.SuiteSucceeded = false
	}
	report.EndTime = time.Now()
	report.RunTime = report.EndTime.Sub(report.StartTime)

	if suiteConfig.ParallelNode == 1 {
		for _, node := range suite.suiteNodes.WithType(types.NodeTypeReportAfterSuite) {
			suite.currentSpecReport = types.SpecReport{
				LeafNodeType:       node.NodeType,
				LeafNodeLocation:   node.CodeLocation,
				LeafNodeText:       node.Text,
				GinkgoParallelNode: suiteConfig.ParallelNode,
			}
			reporter.WillRun(suite.currentSpecReport)
			suite.runReportAfterSuiteNode(node, report, failer, interruptHandler, writer, outputInterceptor, suiteConfig)
			processSpecReport(suite.currentSpecReport)
		}
	}

	reporter.SuiteDidEnd(report)
	if suiteConfig.ParallelTotal > 1 {
		suite.client.PostSuiteDidEnd(report)
	}

	return report.SuiteSucceeded
}

// runSpec(spec) mutates currentSpecReport.  this is ugly
// but it allows the user to call CurrentGinkgoSpecDescription and get
// an up-to-date state of the spec **from within a running spec**
func (suite *Suite) runSpec(spec Spec, setupAllNodeTypesToRun types.NodeTypes, failer *Failer, interruptHandler interrupt_handler.InterruptHandlerInterface, writer WriterInterface, outputInterceptor OutputInterceptor, suiteConfig types.SuiteConfig) {
	if suiteConfig.DryRun {
		suite.currentSpecReport.State = types.SpecStatePassed
		return
	}

	writer.Truncate()
	outputInterceptor.StartInterceptingOutput()
	suite.currentSpecReport.StartTime = time.Now()
	maxAttempts := max(1, spec.FlakeAttempts())
	if suiteConfig.FlakeAttempts > 0 {
		maxAttempts = suiteConfig.FlakeAttempts
	}

	for attempt := 0; attempt < maxAttempts; attempt++ {
		suite.currentSpecReport.NumAttempts = attempt + 1

		if attempt > 0 {
			fmt.Fprintf(writer, "\nGinkgo: Attempt #%d Failed.  Retrying...\n", attempt)
		}

		interruptStatus := interruptHandler.Status()
		deepestNestingLevelAttained := -1
		var nodes Nodes
		if setupAllNodeTypesToRun.Contains(types.NodeTypeBeforeAll) && attempt == 0 {
			nodes = nodes.CopyAppend(spec.Nodes.WithType(types.NodeTypeBeforeAll)...)
		}
		nodes = nodes.CopyAppend(spec.Nodes.WithType(types.NodeTypeBeforeEach)...).SortedByAscendingNestingLevel()
		nodes = nodes.CopyAppend(spec.Nodes.WithType(types.NodeTypeJustBeforeEach).SortedByAscendingNestingLevel()...)
		nodes = nodes.CopyAppend(spec.Nodes.WithType(types.NodeTypeIt)...)

		for _, node := range nodes {
			deepestNestingLevelAttained = max(deepestNestingLevelAttained, node.NestingLevel)
			suite.currentSpecReport.State, suite.currentSpecReport.Failure = suite.runNode(node, failer, interruptStatus.Channel, interruptHandler, spec.Nodes.BestTextFor(node), writer, suiteConfig)
			suite.currentSpecReport.RunTime = time.Since(suite.currentSpecReport.StartTime)
			if suite.currentSpecReport.State != types.SpecStatePassed {
				break
			}
		}

		// pull out some shared code so we aren't repeating ourselves down below. this just runs after and cleanup nodes
		runAfterAndCleanupNodes := func(afterAndCleanupNodes Nodes) {
			for _, node := range afterAndCleanupNodes {
				state, failure := suite.runNode(node, failer, interruptHandler.Status().Channel, interruptHandler, spec.Nodes.BestTextFor(node), writer, suiteConfig)
				suite.currentSpecReport.RunTime = time.Since(suite.currentSpecReport.StartTime)
				if suite.currentSpecReport.State == types.SpecStatePassed || state == types.SpecStateAborted {
					suite.currentSpecReport.State = state
					suite.currentSpecReport.Failure = failure
				}
			}
		}

		var ranAfterAllNodes bool
		afterNodes := spec.Nodes.WithType(types.NodeTypeJustAfterEach).SortedByDescendingNestingLevel()
		if setupAllNodeTypesToRun.Contains(types.NodeTypeAfterAll) || (suite.currentSpecReport.State != types.SpecStatePassed && attempt == maxAttempts-1) {
			ranAfterAllNodes = true
			afterNodes = afterNodes.CopyAppend(spec.Nodes.WithType(types.NodeTypeAfterEach).CopyAppend(spec.Nodes.WithType(types.NodeTypeAfterAll)...).SortedByDescendingNestingLevel()...)
		} else {
			afterNodes = afterNodes.CopyAppend(spec.Nodes.WithType(types.NodeTypeAfterEach).SortedByDescendingNestingLevel()...)
		}
		afterNodes = afterNodes.WithinNestingLevel(deepestNestingLevelAttained)
		runAfterAndCleanupNodes(afterNodes)

		if (suite.currentSpecReport.State != types.SpecStatePassed && attempt == maxAttempts-1) && !ranAfterAllNodes {
			runAfterAndCleanupNodes(spec.Nodes.WithType(types.NodeTypeAfterAll).WithinNestingLevel(deepestNestingLevelAttained))
		}

		cleanupNodes := suite.cleanupNodes.WithType(types.NodeTypeCleanupAfterEach).Reverse()
		suite.cleanupNodes = suite.cleanupNodes.WithoutType(types.NodeTypeCleanupAfterEach)
		if setupAllNodeTypesToRun.Contains(types.NodeTypeAfterAll) || (suite.currentSpecReport.State != types.SpecStatePassed && attempt == maxAttempts-1) {
			cleanupNodes = append(cleanupNodes, suite.cleanupNodes.WithType(types.NodeTypeCleanupAfterAll).Reverse()...)
			suite.cleanupNodes = suite.cleanupNodes.WithoutType(types.NodeTypeCleanupAfterAll)
		}
		runAfterAndCleanupNodes(cleanupNodes)

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

func (suite *Suite) reportEach(spec Spec, nodeType types.NodeType, failer *Failer, interruptHandler interrupt_handler.InterruptHandlerInterface, writer WriterInterface, outputInterceptor OutputInterceptor, suiteConfig types.SuiteConfig) {
	nodes := spec.Nodes.WithType(nodeType)
	if nodeType == types.NodeTypeReportAfterEach {
		nodes = nodes.SortedByDescendingNestingLevel()
	}
	if nodeType == types.NodeTypeReportBeforeEach {
		nodes = nodes.SortedByAscendingNestingLevel()
	}
	if len(nodes) == 0 {
		return
	}

	for _, node := range nodes {
		writer.Truncate()
		outputInterceptor.StartInterceptingOutput()
		report := suite.currentSpecReport
		node.Body = func() {
			node.ReportEachBody(report)
		}
		interruptHandler.SetInterruptPlaceholderMessage(formatter.Fiw(0, formatter.COLS,
			"{{yellow}}Ginkgo received an interrupt signal but is currently running a %s node.  To avoid an invalid report the %s node will not be interrupted however subsequent tests will be skipped.{{/}}\n\n{{bold}}The running %s node is at:\n%s.{{/}}",
			nodeType, nodeType, nodeType,
			node.CodeLocation,
		))
		state, failure := suite.runNode(node, failer, nil, nil, spec.Nodes.BestTextFor(node), writer, suiteConfig)
		interruptHandler.ClearInterruptPlaceholderMessage()
		// If the spec is not in a failure state (i.e. it's Passed/Skipped/Pending) and the reporter has failed, override the state.
		// Also, if the reporter is every aborted - always override the state to propagate the abort
		if (!suite.currentSpecReport.State.Is(types.SpecStateFailureStates...) && state.Is(types.SpecStateFailureStates...)) || state.Is(types.SpecStateAborted) {
			suite.currentSpecReport.State = state
			suite.currentSpecReport.Failure = failure
		}
		suite.currentSpecReport.CapturedGinkgoWriterOutput += string(writer.Bytes())
		suite.currentSpecReport.CapturedStdOutErr += outputInterceptor.StopInterceptingAndReturnOutput()
	}
}

func (suite *Suite) runSuiteNode(node Node, failer *Failer, interruptChannel chan interface{}, interruptHandler interrupt_handler.InterruptHandlerInterface, writer WriterInterface, outputInterceptor OutputInterceptor, suiteConfig types.SuiteConfig) {
	if suiteConfig.DryRun {
		suite.currentSpecReport.State = types.SpecStatePassed
		return
	}

	writer.Truncate()
	outputInterceptor.StartInterceptingOutput()
	suite.currentSpecReport.StartTime = time.Now()

	var err error
	switch node.NodeType {
	case types.NodeTypeBeforeSuite, types.NodeTypeAfterSuite, types.NodeTypeCleanupAfterSuite:
		suite.currentSpecReport.State, suite.currentSpecReport.Failure = suite.runNode(node, failer, interruptChannel, interruptHandler, "", writer, suiteConfig)
	case types.NodeTypeSynchronizedBeforeSuite:
		var data []byte
		var runAllNodes bool
		if suiteConfig.ParallelNode == 1 {
			if suiteConfig.ParallelTotal > 1 {
				outputInterceptor.StopInterceptingAndReturnOutput()
				outputInterceptor.StartInterceptingOutputAndForwardTo(suite.client)
			}
			node.Body = func() { data = node.SynchronizedBeforeSuiteNode1Body() }
			suite.currentSpecReport.State, suite.currentSpecReport.Failure = suite.runNode(node, failer, interruptChannel, interruptHandler, "", writer, suiteConfig)
			if suiteConfig.ParallelTotal > 1 {
				suite.currentSpecReport.CapturedStdOutErr += outputInterceptor.StopInterceptingAndReturnOutput()
				outputInterceptor.StartInterceptingOutput()
				if suite.currentSpecReport.State.Is(types.SpecStatePassed) {
					err = suite.client.PostSynchronizedBeforeSuiteSucceeded(data)
				} else {
					err = suite.client.PostSynchronizedBeforeSuiteFailed()
				}
			}
			runAllNodes = suite.currentSpecReport.State.Is(types.SpecStatePassed) && err == nil
		} else {
			data, err = suite.client.BlockUntilSynchronizedBeforeSuiteData()
			runAllNodes = err == nil
		}
		if runAllNodes {
			node.Body = func() { node.SynchronizedBeforeSuiteAllNodesBody(data) }
			suite.currentSpecReport.State, suite.currentSpecReport.Failure = suite.runNode(node, failer, interruptChannel, interruptHandler, "", writer, suiteConfig)
		}
	case types.NodeTypeSynchronizedAfterSuite:
		node.Body = node.SynchronizedAfterSuiteAllNodesBody
		suite.currentSpecReport.State, suite.currentSpecReport.Failure = suite.runNode(node, failer, interruptChannel, interruptHandler, "", writer, suiteConfig)
		if suiteConfig.ParallelNode == 1 {
			if suiteConfig.ParallelTotal > 1 {
				err = suite.client.BlockUntilNonprimaryNodesHaveFinished()
			}
			if err == nil {
				if suiteConfig.ParallelTotal > 1 {
					suite.currentSpecReport.CapturedStdOutErr += outputInterceptor.StopInterceptingAndReturnOutput()
					outputInterceptor.StartInterceptingOutputAndForwardTo(suite.client)
				}

				node.Body = node.SynchronizedAfterSuiteNode1Body
				state, failure := suite.runNode(node, failer, interruptChannel, interruptHandler, "", writer, suiteConfig)
				if suite.currentSpecReport.State.Is(types.SpecStatePassed) {
					suite.currentSpecReport.State, suite.currentSpecReport.Failure = state, failure
				}
			}
		}
	}

	if err != nil && !suite.currentSpecReport.State.Is(types.SpecStateFailureStates...) {
		suite.currentSpecReport.State, suite.currentSpecReport.Failure = types.SpecStateFailed, suite.failureForLeafNodeWithMessage(node, err.Error())
	}

	suite.currentSpecReport.EndTime = time.Now()
	suite.currentSpecReport.RunTime = suite.currentSpecReport.EndTime.Sub(suite.currentSpecReport.StartTime)
	suite.currentSpecReport.CapturedGinkgoWriterOutput = string(writer.Bytes())
	suite.currentSpecReport.CapturedStdOutErr += outputInterceptor.StopInterceptingAndReturnOutput()

	return
}

func (suite *Suite) runReportAfterSuiteNode(node Node, report types.Report, failer *Failer, interruptHandler interrupt_handler.InterruptHandlerInterface, writer WriterInterface, outputInterceptor OutputInterceptor, suiteConfig types.SuiteConfig) {
	if suiteConfig.DryRun {
		suite.currentSpecReport.State = types.SpecStatePassed
		return
	}

	writer.Truncate()
	outputInterceptor.StartInterceptingOutput()
	suite.currentSpecReport.StartTime = time.Now()

	if suiteConfig.ParallelTotal > 1 {
		aggregatedReport, err := suite.client.BlockUntilAggregatedNonprimaryNodesReport()
		if err != nil {
			suite.currentSpecReport.State, suite.currentSpecReport.Failure = types.SpecStateFailed, suite.failureForLeafNodeWithMessage(node, err.Error())
			return
		}
		report = report.Add(aggregatedReport)
	}

	node.Body = func() { node.ReportAfterSuiteBody(report) }
	interruptHandler.SetInterruptPlaceholderMessage(formatter.Fiw(0, formatter.COLS,
		"{{yellow}}Ginkgo received an interrupt signal but is currently running a ReportAfterSuite node.  To avoid an invalid report the ReportAfterSuite node will not be interrupted.{{/}}\n\n{{bold}}The running ReportAfterSuite node is at:\n%s.{{/}}",
		node.CodeLocation,
	))
	suite.currentSpecReport.State, suite.currentSpecReport.Failure = suite.runNode(node, failer, nil, nil, "", writer, suiteConfig)
	interruptHandler.ClearInterruptPlaceholderMessage()

	suite.currentSpecReport.EndTime = time.Now()
	suite.currentSpecReport.RunTime = suite.currentSpecReport.EndTime.Sub(suite.currentSpecReport.StartTime)
	suite.currentSpecReport.CapturedGinkgoWriterOutput = string(writer.Bytes())
	suite.currentSpecReport.CapturedStdOutErr = outputInterceptor.StopInterceptingAndReturnOutput()

	return
}

func (suite *Suite) runNode(node Node, failer *Failer, interruptChannel chan interface{}, interruptHandler interrupt_handler.InterruptHandlerInterface, text string, writer WriterInterface, suiteConfig types.SuiteConfig) (types.SpecState, types.Failure) {
	suite.currentNode = node
	defer func() {
		suite.currentNode = Node{}
	}()

	if suiteConfig.EmitSpecProgress {
		if text == "" {
			text = "TOP-LEVEL"
		}
		s := fmt.Sprintf("[%s] %s\n  %s\n", node.NodeType.String(), text, node.CodeLocation.String())
		writer.Write([]byte(s))
	}

	var failure types.Failure
	failure.FailureNodeType, failure.FailureNodeLocation = node.NodeType, node.CodeLocation
	if node.NodeType.Is(types.NodeTypeIt) || node.NodeType.Is(types.NodeTypesForSuiteLevelNodes...) {
		failure.FailureNodeContext = types.FailureNodeIsLeafNode
	} else if node.NestingLevel <= 0 {
		failure.FailureNodeContext = types.FailureNodeAtTopLevel
	} else {
		failure.FailureNodeContext, failure.FailureNodeContainerIndex = types.FailureNodeInContainer, node.NestingLevel-1
	}

	outcomeC := make(chan types.SpecState)
	failureC := make(chan types.Failure)

	go func() {
		finished := false
		defer func() {
			if e := recover(); e != nil || !finished {
				failer.Panic(types.NewCodeLocationWithStackTrace(2), e)
			}

			outcome, failureFromRun := failer.Drain()
			outcomeC <- outcome
			failureC <- failureFromRun
		}()

		node.Body()
		finished = true
	}()

	select {
	case outcome := <-outcomeC:
		failureFromRun := <-failureC
		if outcome == types.SpecStatePassed {
			return outcome, types.Failure{}
		}
		failure.Message, failure.Location, failure.ForwardedPanic = failureFromRun.Message, failureFromRun.Location, failureFromRun.ForwardedPanic
		return outcome, failure
	case <-interruptChannel:
		failure.Message, failure.Location = interruptHandler.InterruptMessageWithStackTraces(), node.CodeLocation
		return types.SpecStateInterrupted, failure
	}
}

func (suite *Suite) failureForLeafNodeWithMessage(node Node, message string) types.Failure {
	return types.Failure{
		Message:             message,
		Location:            node.CodeLocation,
		FailureNodeContext:  types.FailureNodeIsLeafNode,
		FailureNodeType:     node.NodeType,
		FailureNodeLocation: node.CodeLocation,
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
