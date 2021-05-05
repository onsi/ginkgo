package internal

import (
	"reflect"
	"sort"

	"sync"

	"github.com/onsi/ginkgo/types"
)

var _global_node_id_counter = uint(0)
var _global_id_mutex = &sync.Mutex{}

func UniqueNodeID() uint {
	//There's a reace in the internal integration tests if we don't make
	//accessing _global_node_id_counter safe across goroutines.
	_global_id_mutex.Lock()
	defer _global_id_mutex.Unlock()
	_global_node_id_counter += 1
	return _global_node_id_counter
}

type Node struct {
	ID       uint
	NodeType types.NodeType

	Text         string
	Body         func()
	CodeLocation types.CodeLocation
	NestingLevel int

	SynchronizedBeforeSuiteNode1Body    func() []byte
	SynchronizedBeforeSuiteAllNodesBody func([]byte)

	SynchronizedAfterSuiteAllNodesBody func()
	SynchronizedAfterSuiteNode1Body    func()

	ReportAfterEachBody  func(types.SpecReport)
	ReportAfterSuiteBody func(types.Report)

	MarkedFocus   bool
	MarkedPending bool
	FlakeAttempts int
}

// Decoration Types
type focusType bool
type pendingType bool

const Focus = focusType(true)
const Pending = pendingType(true)

type FlakeAttempts uint
type Offset uint
type Done chan<- interface{} // Deprecated Done Channel for asynchronous testing

func NewNode(deprecationTracker *types.DeprecationTracker, nodeType types.NodeType, text string, args ...interface{}) (Node, []error) {
	baseOffset := 2
	node := Node{
		ID:           UniqueNodeID(),
		NodeType:     nodeType,
		Text:         text,
		CodeLocation: types.NewCodeLocation(baseOffset),
		NestingLevel: -1,
	}
	errors := []error{}
	appendErrorIf := func(predicate bool, err error) bool {
		if predicate {
			errors = append(errors, err)
		}
		return predicate
	}

	args = unrollInterfaceSlice(args)

	remainingArgs := []interface{}{}
	//First get the CodeLocation up-to-date
	for _, arg := range args {
		switch t := reflect.TypeOf(arg); {
		case t == reflect.TypeOf(Offset(0)):
			node.CodeLocation = types.NewCodeLocation(baseOffset + int(arg.(Offset)))
		case t == reflect.TypeOf(types.CodeLocation{}):
			node.CodeLocation = arg.(types.CodeLocation)
		default:
			remainingArgs = append(remainingArgs, arg)
		}
	}

	trackedFunctionError := false
	args = remainingArgs
	remainingArgs = []interface{}{}
	//now process the rest of the args
	for _, arg := range args {
		switch t := reflect.TypeOf(arg); {
		case t == reflect.TypeOf(float64(0)):
			break //ignore deprecated timeouts
		case t == reflect.TypeOf(Focus):
			node.MarkedFocus = bool(arg.(focusType))
			appendErrorIf(!nodeType.Is(types.NodeTypesForContainerAndIt...), types.GinkgoErrors.InvalidDecorationForNodeType(node.CodeLocation, nodeType, "Focus"))
		case t == reflect.TypeOf(Pending):
			node.MarkedPending = bool(arg.(pendingType))
			appendErrorIf(!nodeType.Is(types.NodeTypesForContainerAndIt...), types.GinkgoErrors.InvalidDecorationForNodeType(node.CodeLocation, nodeType, "Pending"))
		case t == reflect.TypeOf(FlakeAttempts(0)):
			node.FlakeAttempts = int(arg.(FlakeAttempts))
			appendErrorIf(!nodeType.Is(types.NodeTypesForContainerAndIt...), types.GinkgoErrors.InvalidDecorationForNodeType(node.CodeLocation, nodeType, "FlakeAttempts"))
		case t.Kind() == reflect.Func:
			if appendErrorIf(node.Body != nil, types.GinkgoErrors.MultipleBodyFunctions(node.CodeLocation, nodeType)) {
				trackedFunctionError = true
				break
			}
			isValid := (t.NumOut() == 0) && (t.NumIn() <= 1) && (t.NumIn() == 0 || t.In(0) == reflect.TypeOf(make(Done)))
			if appendErrorIf(!isValid, types.GinkgoErrors.InvalidBodyType(t, node.CodeLocation, nodeType)) {
				trackedFunctionError = true
				break
			}
			if t.NumIn() == 0 {
				node.Body = arg.(func())
			} else {
				deprecationTracker.TrackDeprecation(types.Deprecations.Async(), node.CodeLocation)
				deprecatedAsyncBody := arg.(func(Done))
				node.Body = func() { deprecatedAsyncBody(make(Done)) }
			}
		default:
			remainingArgs = append(remainingArgs, arg)
		}
	}

	//validations
	appendErrorIf(node.MarkedPending && node.MarkedFocus, types.GinkgoErrors.InvalidDeclarationOfFocusedAndPending(node.CodeLocation, nodeType))
	appendErrorIf(node.Body == nil && !node.MarkedPending && !trackedFunctionError, types.GinkgoErrors.MissingBodyFunction(node.CodeLocation, nodeType))
	for _, arg := range remainingArgs {
		errors = append(errors, types.GinkgoErrors.UnknownDecoration(node.CodeLocation, nodeType, arg))
	}

	if len(errors) > 0 {
		return Node{}, errors
	}

	return node, errors
}

func NewSynchronizedBeforeSuiteNode(node1Body func() []byte, allNodesBody func([]byte), codeLocation types.CodeLocation) (Node, []error) {
	return Node{
		ID:                                  UniqueNodeID(),
		NodeType:                            types.NodeTypeSynchronizedBeforeSuite,
		SynchronizedBeforeSuiteNode1Body:    node1Body,
		SynchronizedBeforeSuiteAllNodesBody: allNodesBody,
		CodeLocation:                        codeLocation,
	}, nil
}

func NewSynchronizedAfterSuiteNode(allNodesBody func(), node1Body func(), codeLocation types.CodeLocation) (Node, []error) {
	return Node{
		ID:                                 UniqueNodeID(),
		NodeType:                           types.NodeTypeSynchronizedAfterSuite,
		SynchronizedAfterSuiteAllNodesBody: allNodesBody,
		SynchronizedAfterSuiteNode1Body:    node1Body,
		CodeLocation:                       codeLocation,
	}, nil
}

func NewReportAfterEachNode(body func(types.SpecReport), codeLocation types.CodeLocation) (Node, []error) {
	return Node{
		ID:                  UniqueNodeID(),
		NodeType:            types.NodeTypeReportAfterEach,
		ReportAfterEachBody: body,
		CodeLocation:        codeLocation,
		NestingLevel:        -1,
	}, nil
}

func NewReportAfterSuiteNode(text string, body func(types.Report), codeLocation types.CodeLocation) (Node, []error) {
	return Node{
		ID:                   UniqueNodeID(),
		Text:                 text,
		NodeType:             types.NodeTypeReportAfterSuite,
		ReportAfterSuiteBody: body,
		CodeLocation:         codeLocation,
	}, nil
}

func (n Node) IsZero() bool {
	return n.ID == 0
}

/* Nodes */
type Nodes []Node

func (n Nodes) CopyAppend(nodes ...Node) Nodes {
	out := Nodes{}
	for _, node := range n {
		out = append(out, node)
	}
	for _, node := range nodes {
		out = append(out, node)
	}
	return out
}

func (n Nodes) SplitAround(pivot Node) (Nodes, Nodes) {
	left := Nodes{}
	right := Nodes{}
	found := false
	for _, node := range n {
		if node.ID == pivot.ID {
			found = true
			continue
		}
		if found {
			right = append(right, node)
		} else {
			left = append(left, node)
		}
	}

	return left, right
}

func (n Nodes) FirstNodeWithType(nodeTypes ...types.NodeType) Node {
	for _, node := range n {
		if node.NodeType.Is(nodeTypes...) {
			return node
		}
	}
	return Node{}
}

func (n Nodes) WithType(nodeTypes ...types.NodeType) Nodes {
	out := Nodes{}
	for _, node := range n {
		if node.NodeType.Is(nodeTypes...) {
			out = append(out, node)
		}
	}
	return out
}

func (n Nodes) WithoutType(nodeTypes ...types.NodeType) Nodes {
	out := Nodes{}
	for _, node := range n {
		if !node.NodeType.Is(nodeTypes...) {
			out = append(out, node)
		}
	}
	return out
}

func (n Nodes) WithinNestingLevel(deepestNestingLevel int) Nodes {
	out := Nodes{}
	for _, node := range n {
		if node.NestingLevel <= deepestNestingLevel {
			out = append(out, node)
		}
	}
	return out
}

func (n Nodes) SortedByDescendingNestingLevel() Nodes {
	out := make(Nodes, len(n))
	copy(out, n)
	sort.SliceStable(out, func(i int, j int) bool {
		return out[i].NestingLevel > out[j].NestingLevel
	})

	return out
}

func (n Nodes) SortedByAscendingNestingLevel() Nodes {
	out := make(Nodes, len(n))
	copy(out, n)
	sort.SliceStable(out, func(i int, j int) bool {
		return out[i].NestingLevel < out[j].NestingLevel
	})

	return out
}

func (n Nodes) Texts() []string {
	out := []string{}
	for _, node := range n {
		out = append(out, node.Text)
	}
	return out
}

func (n Nodes) CodeLocations() []types.CodeLocation {
	out := []types.CodeLocation{}
	for _, node := range n {
		out = append(out, node.CodeLocation)
	}
	return out
}

func (n Nodes) BestTextFor(node Node) string {
	if node.Text != "" {
		return node.Text
	}
	parentNestingLevel := node.NestingLevel - 1
	for _, node := range n {
		if node.Text != "" && node.NestingLevel == parentNestingLevel {
			return node.Text
		}
	}

	return ""
}

func (n Nodes) HasNodeMarkedPending() bool {
	for _, node := range n {
		if node.MarkedPending {
			return true
		}
	}
	return false
}

func (n Nodes) HasNodeMarkedFocus() bool {
	for _, node := range n {
		if node.MarkedFocus {
			return true
		}
	}
	return false
}

func unrollInterfaceSlice(args interface{}) []interface{} {
	v := reflect.ValueOf(args)
	if v.Kind() != reflect.Slice {
		return []interface{}{args}
	}
	out := []interface{}{}
	for i := 0; i < v.Len(); i++ {
		el := reflect.ValueOf(v.Index(i).Interface())
		if el.Kind() == reflect.Slice {
			out = append(out, unrollInterfaceSlice(el.Interface())...)
		} else {
			out = append(out, v.Index(i).Interface())
		}
	}
	return out
}
