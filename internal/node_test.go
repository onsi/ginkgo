package internal_test

import (
	"reflect"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/types"
	. "github.com/onsi/gomega"

	"github.com/onsi/ginkgo/internal"
)

var _ = Describe("UniqueNodeID", func() {
	It("returns a unique id every time it's called", func() {
		Ω(internal.UniqueNodeID()).ShouldNot(Equal(internal.UniqueNodeID()))
	})
})

var _ = Describe("Partitioning Decorations", func() {
	It("separates out decorations and non-decorations", func() {
		type Foo struct {
			A int
		}
		decorations, remaining := internal.PartitionDecorations(
			Offset(3),
			Foo{3},
			types.NewCustomCodeLocation("hey there"),
			"hey there",
			Focus,
			2.0,
			Pending,
			nil,
			1,
			[]interface{}{Focus, Pending, []interface{}{Offset(2), FlakeAttempts(2)}},
			[]interface{}{1, 2, 3.1, nil},
			[]interface{}{},
			FlakeAttempts(1),
			true,
		)

		Ω(decorations).Should(Equal([]interface{}{
			Offset(3),
			types.NewCustomCodeLocation("hey there"),
			Focus,
			Pending,
			[]interface{}{Focus, Pending, []interface{}{Offset(2), FlakeAttempts(2)}},
			FlakeAttempts(1),
		}))

		Ω(remaining).Should(Equal([]interface{}{
			Foo{3},
			"hey there",
			2.0,
			nil,
			1,
			[]interface{}{1, 2, 3.1, nil},
			[]interface{}{},
			true,
		}))
	})
})

var _ = Describe("Construcing nodes", func() {
	var dt *types.DeprecationTracker
	var didRun bool
	var body func()
	BeforeEach(func() {
		dt = types.NewDeprecationTracker()
		didRun = false
		body = func() { didRun = true }
	})

	ExpectAllWell := func(errors []error) {
		ExpectWithOffset(1, errors).Should(BeEmpty())
		ExpectWithOffset(1, dt.DidTrackDeprecations()).Should(BeFalse())
	}

	Describe("happy path", func() {
		It("creates a node with a non-zero id", func() {
			node, errors := internal.NewNode(dt, ntIt, "text", body, cl, Focus)
			Ω(node.ID).Should(BeNumerically(">", 0))
			Ω(node.NodeType).Should(Equal(ntIt))
			Ω(node.Text).Should(Equal("text"))
			node.Body()
			Ω(didRun).Should(BeTrue())
			Ω(node.CodeLocation).Should(Equal(cl))
			Ω(node.MarkedFocus).Should(BeTrue())
			Ω(node.MarkedPending).Should(BeFalse())
			Ω(node.NestingLevel).Should(Equal(-1))
			ExpectAllWell(errors)
		})
	})

	Describe("Assigning CodeLocation", func() {
		Context("with nothing explicitly specified ", func() {
			It("assumes a base-offset of 2", func() {
				cl := types.NewCodeLocation(1)
				node, errors := internal.NewNode(dt, ntIt, "text", body)
				Ω(node.CodeLocation.FileName).Should(Equal(cl.FileName))
				ExpectAllWell(errors)
			})
		})

		Context("specifying code locations", func() {
			It("uses the last passed-in code location", func() {
				cl2 := types.NewCustomCodeLocation("hi")
				node, errors := internal.NewNode(dt, ntIt, "text", body, cl, cl2)
				Ω(node.CodeLocation).Should(Equal(cl2))
				ExpectAllWell(errors)
			})
		})

		Context("specifying offets", func() {
			It("takes the offset and adds it to the base-offset of 2 to compute the code location", func() {
				cl := types.NewCodeLocation(2)
				cl2 := types.NewCustomCodeLocation("hi")
				node, errors := internal.NewNode(dt, ntIt, "text", body, cl2, Offset(1))
				//note that Offset overrides cl2
				Ω(node.CodeLocation.FileName).Should(Equal(cl.FileName))
				ExpectAllWell(errors)
			})
		})
	})

	Describe("ignoring deprecated timeouts", func() {
		It("ignores any float64s", func() {
			node, errors := internal.NewNode(dt, ntIt, "text", body, 3.141, 2.71)
			node.Body()
			Ω(didRun).Should(BeTrue())
			ExpectAllWell(errors)
		})
	})

	Describe("the Focus and Pending decorations", func() {
		It("the node is neither Focused nor Pending by default", func() {
			node, errors := internal.NewNode(dt, ntIt, "text", body)
			Ω(node.MarkedFocus).Should(BeFalse())
			Ω(node.MarkedPending).Should(BeFalse())
			ExpectAllWell(errors)
		})
		It("marks the node as focused", func() {
			node, errors := internal.NewNode(dt, ntIt, "text", body, Focus)
			Ω(node.MarkedFocus).Should(BeTrue())
			Ω(node.MarkedPending).Should(BeFalse())
			ExpectAllWell(errors)
		})
		It("marks the node as pending", func() {
			node, errors := internal.NewNode(dt, ntIt, "text", body, Pending)
			Ω(node.MarkedFocus).Should(BeFalse())
			Ω(node.MarkedPending).Should(BeTrue())
			ExpectAllWell(errors)
		})
		It("errors when both Focus and Pending are set", func() {
			node, errors := internal.NewNode(dt, ntIt, "text", body, cl, Focus, Pending)
			Ω(node).Should(BeZero())
			Ω(errors).Should(ConsistOf(types.GinkgoErrors.InvalidDeclarationOfFocusedAndPending(cl, ntIt)))
		})
		It("allows containers to be marked", func() {
			node, errors := internal.NewNode(dt, ntCon, "text", body, Focus)
			Ω(node.MarkedFocus).Should(BeTrue())
			Ω(node.MarkedPending).Should(BeFalse())
			ExpectAllWell(errors)

			node, errors = internal.NewNode(dt, ntCon, "text", body, Pending)
			Ω(node.MarkedFocus).Should(BeFalse())
			Ω(node.MarkedPending).Should(BeTrue())
			ExpectAllWell(errors)
		})
		It("does not allow non-container/it nodes to be marked", func() {
			node, errors := internal.NewNode(dt, ntBef, "", body, cl, Focus)
			Ω(node).Should(BeZero())
			Ω(errors).Should(ConsistOf(types.GinkgoErrors.InvalidDecorationForNodeType(cl, ntBef, "Focus")))

			node, errors = internal.NewNode(dt, ntAf, "", body, cl, Pending)
			Ω(node).Should(BeZero())
			Ω(errors).Should(ConsistOf(types.GinkgoErrors.InvalidDecorationForNodeType(cl, ntAf, "Pending")))

			Ω(dt.DidTrackDeprecations()).Should(BeFalse())
		})
	})

	Describe("The FlakeAttempts decoration", func() {
		It("is zero by default", func() {
			node, errors := internal.NewNode(dt, ntIt, "text", body)
			Ω(node).ShouldNot(BeZero())
			Ω(node.FlakeAttempts).Should(Equal(0))
			ExpectAllWell(errors)
		})
		It("sets the FlakeAttempts field", func() {
			node, errors := internal.NewNode(dt, ntIt, "text", body, FlakeAttempts(2))
			Ω(node.FlakeAttempts).Should(Equal(2))
			ExpectAllWell(errors)
		})
		It("can be applied to containers", func() {
			node, errors := internal.NewNode(dt, ntCon, "text", body, FlakeAttempts(2))
			Ω(node.FlakeAttempts).Should(Equal(2))
			ExpectAllWell(errors)
		})
		It("cannot be applied to non-container/it nodes", func() {
			node, errors := internal.NewNode(dt, ntBef, "", body, cl, FlakeAttempts(2))
			Ω(node).Should(BeZero())
			Ω(errors).Should(ConsistOf(types.GinkgoErrors.InvalidDecorationForNodeType(cl, ntBef, "FlakeAttempts")))
			Ω(dt.DidTrackDeprecations()).Should(BeFalse())
		})
	})

	Describe("passing in functions", func() {
		It("works when a single function is passed in", func() {
			node, errors := internal.NewNode(dt, ntIt, "text", body, cl)
			node.Body()
			Ω(didRun).Should(BeTrue())
			ExpectAllWell(errors)
		})

		It("allows deprecated async functions and registers a deprecation warning", func() {
			node, errors := internal.NewNode(dt, ntIt, "text", func(done Done) {
				didRun = true
				Ω(done).ShouldNot(BeNil())
				close(done)
			}, cl)
			node.Body()
			Ω(didRun).Should(BeTrue())
			Ω(errors).Should(BeEmpty())
			Ω(dt.DeprecationsReport()).Should(ContainSubstring(types.Deprecations.Async().Message))
		})

		It("errors if more than one function is provided", func() {
			node, errors := internal.NewNode(dt, ntIt, "text", body, body, cl)
			Ω(node).Should(BeZero())
			Ω(errors).Should(ConsistOf(types.GinkgoErrors.MultipleBodyFunctions(cl, ntIt)))
			Ω(dt.DidTrackDeprecations()).Should(BeFalse())
		})

		It("errors if the function has a return value", func() {
			f := func() string { return "" }
			node, errors := internal.NewNode(dt, ntIt, "text", f, cl)
			Ω(node).Should(BeZero())
			Ω(errors).Should(ConsistOf(types.GinkgoErrors.InvalidBodyType(reflect.TypeOf(f), cl, ntIt)))
			Ω(dt.DidTrackDeprecations()).Should(BeFalse())
		})

		It("errors if the function takes more than one argument", func() {
			f := func(Done, string) {}
			node, errors := internal.NewNode(dt, ntIt, "text", f, cl)
			Ω(node).Should(BeZero())
			Ω(errors).Should(ConsistOf(types.GinkgoErrors.InvalidBodyType(reflect.TypeOf(f), cl, ntIt)))
			Ω(dt.DidTrackDeprecations()).Should(BeFalse())
		})

		It("errors if the function takes one argument and that argument is not the deprecated Done channel", func() {
			f := func(chan interface{}) {}
			node, errors := internal.NewNode(dt, ntIt, "text", f, cl)
			Ω(node).Should(BeZero())
			Ω(errors).Should(ConsistOf(types.GinkgoErrors.InvalidBodyType(reflect.TypeOf(f), cl, ntIt)))
			Ω(dt.DidTrackDeprecations()).Should(BeFalse())
		})

		It("errors if no function is passed in", func() {
			node, errors := internal.NewNode(dt, ntIt, "text", cl)
			Ω(node).Should(BeZero())
			Ω(errors).Should(ConsistOf(types.GinkgoErrors.MissingBodyFunction(cl, ntIt)))
			Ω(dt.DidTrackDeprecations()).Should(BeFalse())
		})

		It("is ok if no function is passed in but it is marked pending", func() {
			node, errors := internal.NewNode(dt, ntIt, "text", cl, Pending)
			Ω(node.IsZero()).Should(BeFalse())
			ExpectAllWell(errors)
		})
	})

	Describe("non-recognized decorations", func() {
		It("errors when a non-recognized decoration is provided", func() {
			node, errors := internal.NewNode(dt, ntIt, "text", cl, body, Focus, "aardvark", 5)
			Ω(node).Should(BeZero())
			Ω(errors).Should(ConsistOf(
				types.GinkgoErrors.UnknownDecoration(cl, ntIt, "aardvark"),
				types.GinkgoErrors.UnknownDecoration(cl, ntIt, 5),
			))
			Ω(dt.DidTrackDeprecations()).Should(BeFalse())
		})
	})

	Describe("when decorations are nested in slices", func() {
		It("unrolls them first", func() {
			node, errors := internal.NewNode(dt, ntIt, "text", []interface{}{body, []interface{}{Focus, FlakeAttempts(3)}, FlakeAttempts(2)})
			Ω(node.FlakeAttempts).Should(Equal(2))
			Ω(node.MarkedFocus).Should(BeTrue())
			node.Body()
			Ω(didRun).Should(BeTrue())
			ExpectAllWell(errors)
		})
	})
})

var _ = Describe("Node", func() {
	Describe("The other node constructors", func() {
		Describe("NewSynchronizedBeforeSuiteNode", func() {
			It("returns a correctly configured node", func() {
				var ranNode1, ranAllNodes bool
				node1Body := func() []byte { ranNode1 = true; return nil }
				allNodesBody := func(_ []byte) { ranAllNodes = true }
				node, errors := internal.NewSynchronizedBeforeSuiteNode(node1Body, allNodesBody, cl)
				Ω(errors).Should(BeEmpty())
				Ω(node.ID).Should(BeNumerically(">", 0))
				Ω(node.NodeType).Should(Equal(types.NodeTypeSynchronizedBeforeSuite))

				node.SynchronizedBeforeSuiteNode1Body()
				Ω(ranNode1).Should(BeTrue())

				node.SynchronizedBeforeSuiteAllNodesBody(nil)
				Ω(ranAllNodes).Should(BeTrue())

				Ω(node.CodeLocation).Should(Equal(cl))
				Ω(node.NestingLevel).Should(Equal(0))
			})
		})

		Describe("NewSynchronizedAfterSuiteNode", func() {
			It("returns a correctly configured node", func() {
				var ranNode1, ranAllNodes bool
				allNodesBody := func() { ranAllNodes = true }
				node1Body := func() { ranNode1 = true }

				node, errors := internal.NewSynchronizedAfterSuiteNode(allNodesBody, node1Body, cl)
				Ω(errors).Should(BeEmpty())
				Ω(node.ID).Should(BeNumerically(">", 0))
				Ω(node.NodeType).Should(Equal(types.NodeTypeSynchronizedAfterSuite))

				node.SynchronizedAfterSuiteAllNodesBody()
				Ω(ranAllNodes).Should(BeTrue())

				node.SynchronizedAfterSuiteNode1Body()
				Ω(ranNode1).Should(BeTrue())

				Ω(node.CodeLocation).Should(Equal(cl))
				Ω(node.NestingLevel).Should(Equal(0))
			})
		})

		Describe("NewReportAfterSuiteNode", func() {
			It("returns a correctly configured node", func() {
				var didRun bool
				body := func(types.Report) { didRun = true }
				node, errors := internal.NewReportAfterSuiteNode("my custom report", body, cl)
				Ω(errors).Should(BeEmpty())
				Ω(node.Text).Should(Equal("my custom report"))
				Ω(node.ID).Should(BeNumerically(">", 0))
				Ω(node.NodeType).Should(Equal(types.NodeTypeReportAfterSuite))

				node.ReportAfterSuiteBody(types.Report{})
				Ω(didRun).Should(BeTrue())

				Ω(node.CodeLocation).Should(Equal(cl))
				Ω(node.NestingLevel).Should(Equal(0))
			})
		})

		Describe("NewReportAfterEachNode", func() {
			It("returns a correctly configured node", func() {
				var didRun bool
				body := func(types.SpecReport) { didRun = true }

				node, errors := internal.NewReportAfterEachNode(body, cl)
				Ω(errors).Should(BeEmpty())
				Ω(node.ID).Should(BeNumerically(">", 0))
				Ω(node.NodeType).Should(Equal(types.NodeTypeReportAfterEach))

				node.ReportAfterEachBody(types.SpecReport{})
				Ω(didRun).Should(BeTrue())

				Ω(node.CodeLocation).Should(Equal(cl))
				Ω(node.NestingLevel).Should(Equal(-1))
			})
		})
	})

	Describe("IsZero()", func() {
		It("returns true if the node is zero", func() {
			Ω(Node{}.IsZero()).Should(BeTrue())
		})

		It("returns false if the node is non-zero", func() {
			node, errors := internal.NewNode(nil, ntIt, "hummus", func() {}, cl)
			Ω(errors).Should(BeEmpty())
			Ω(node.IsZero()).Should(BeFalse())
		})
	})
})

var _ = Describe("Nodes", func() {
	Describe("CopyAppend", func() {
		var n1, n2, n3, n4 Node

		BeforeEach(func() {
			n1, n2, n3, n4 = N(), N(), N(), N()
		})

		It("appends the passed in nodes and returns the result", func() {
			result := Nodes{n1, n2}.CopyAppend(n3, n4)
			Ω(result).Should(Equal(Nodes{n1, n2, n3, n4}))
		})

		It("makes a copy, leaving the original untouched", func() {
			original := Nodes{n1, n2}
			original.CopyAppend(n3, n4)
			Ω(original).Should(Equal(Nodes{n1, n2}))
		})
	})

	Describe("SplitAround", func() {
		var nodes Nodes

		BeforeEach(func() {
			nodes = Nodes{N(), N(), N(), N(), N()}
		})

		Context("when the pivot is a member of nodes", func() {
			Context("when the pivot is not at one of the ends", func() {
				It("returns the correct left and right nodes", func() {
					left, right := nodes.SplitAround(nodes[2])
					Ω(left).Should(Equal(Nodes{nodes[0], nodes[1]}))
					Ω(right).Should(Equal(Nodes{nodes[3], nodes[4]}))
				})
			})

			Context("when the pivot is the first member", func() {
				It("returns an empty left nodes and the complete right nodes", func() {
					left, right := nodes.SplitAround(nodes[0])
					Ω(left).Should(BeEmpty())
					Ω(right).Should(Equal(Nodes{nodes[1], nodes[2], nodes[3], nodes[4]}))

				})
			})

			Context("when the pivot is the last member", func() {
				It("returns an empty right nodes and the complete left nodes", func() {
					left, right := nodes.SplitAround(nodes[4])
					Ω(left).Should(Equal(Nodes{nodes[0], nodes[1], nodes[2], nodes[3]}))
					Ω(right).Should(BeEmpty())
				})
			})
		})

		Context("when the pivot is not in nodes", func() {
			It("returns an empty right nodes and the complete left nodes", func() {
				left, right := nodes.SplitAround(N())
				Ω(left).Should(Equal(nodes))
				Ω(right).Should(BeEmpty())
			})
		})
	})

	Describe("FirstNodeWithType", func() {
		var nodes Nodes

		BeforeEach(func() {
			nodes = Nodes{N(ntCon), N("bef1", ntBef), N("bef2", ntBef), N(ntIt), N(ntAf)}
		})

		Context("when there is a matching node", func() {
			It("returns the first node that matches one of the requested node types", func() {
				Ω(nodes.FirstNodeWithType(ntAf, ntIt, ntBef).Text).Should(Equal("bef1"))
			})
		})
		Context("when there is no matching node", func() {
			It("returns an empty node", func() {
				Ω(nodes.FirstNodeWithType(ntJusAf)).Should(BeZero())
			})
		})
	})

	Describe("Filtering By NodeType", func() {
		var nCon, nBef1, nBef2, nIt, nAf Node
		var nodes Nodes

		BeforeEach(func() {
			nCon = N(ntCon)
			nBef1 = N(ntBef)
			nBef2 = N(ntBef)
			nIt = N(ntIt)
			nAf = N(ntAf)
			nodes = Nodes{nCon, nBef1, nBef2, nIt, nAf}
		})

		Describe("WithType", func() {
			Context("when there are matching nodes", func() {
				It("returns them while preserving order", func() {
					Ω(nodes.WithType(ntIt, ntBef)).Should(Equal(Nodes{nBef1, nBef2, nIt}))
				})
			})

			Context("when there are no matching nodes", func() {
				It("returns an empty Nodes{}", func() {
					Ω(nodes.WithType(ntJusAf)).Should(BeEmpty())
				})
			})
		})

		Describe("WithoutType", func() {
			Context("when there are matching nodes", func() {
				It("does not include them in the result", func() {
					Ω(nodes.WithoutType(ntIt, ntBef)).Should(Equal(Nodes{nCon, nAf}))
				})
			})

			Context("when no nodes match", func() {
				It("doesn't elide any nodes", func() {
					Ω(nodes.WithoutType(ntJusAf)).Should(Equal(nodes))
				})
			})
		})
	})

	Describe("SortedByDescendingNestingLevel", func() {
		var n0A, n0B, n1A, n1B, n1C, n2A, n2B Node
		var nodes Nodes
		BeforeEach(func() {
			n0A = N(NestingLevel(0))
			n0B = N(NestingLevel(0))
			n1A = N(NestingLevel(1))
			n1B = N(NestingLevel(1))
			n1C = N(NestingLevel(1))
			n2A = N(NestingLevel(2))
			n2B = N(NestingLevel(2))
			nodes = Nodes{n0A, n0B, n1A, n1B, n1C, n2A, n2B}
		})

		It("returns copy sorted by descending nesting level, preserving order within nesting level", func() {
			Ω(nodes.SortedByDescendingNestingLevel()).Should(Equal(Nodes{n2A, n2B, n1A, n1B, n1C, n0A, n0B}))
			Ω(nodes).Should(Equal(Nodes{n0A, n0B, n1A, n1B, n1C, n2A, n2B}), "original nodes should not have been modified")
		})
	})

	Describe("SortedByAscendingNestingLevel", func() {
		var n0A, n0B, n1A, n1B, n1C, n2A, n2B Node
		var nodes Nodes
		BeforeEach(func() {
			n0A = N(NestingLevel(0))
			n0B = N(NestingLevel(0))
			n1A = N(NestingLevel(1))
			n1B = N(NestingLevel(1))
			n1C = N(NestingLevel(1))
			n2A = N(NestingLevel(2))
			n2B = N(NestingLevel(2))
			nodes = Nodes{n2A, n1A, n1B, n0A, n2B, n0B, n1C}
		})

		It("returns copy sorted by ascending nesting level, preserving order within nesting level", func() {
			Ω(nodes.SortedByAscendingNestingLevel()).Should(Equal(Nodes{n0A, n0B, n1A, n1B, n1C, n2A, n2B}))
			Ω(nodes).Should(Equal(Nodes{n2A, n1A, n1B, n0A, n2B, n0B, n1C}), "original nodes should not have been modified")
		})
	})

	Describe("WithinNestingLevel", func() {
		var n0, n1, n2a, n2b, n3 Node
		var nodes Nodes
		BeforeEach(func() {
			n0 = N(NestingLevel(0))
			n1 = N(NestingLevel(1))
			n2a = N(NestingLevel(2))
			n3 = N(NestingLevel(3))
			n2b = N(NestingLevel(2))
			nodes = Nodes{n0, n1, n2a, n3, n2b}
		})

		It("returns nodes, in order, with nesting level equal to or less than the requested level", func() {
			Ω(nodes.WithinNestingLevel(-1)).Should(BeEmpty())
			Ω(nodes.WithinNestingLevel(0)).Should(Equal(Nodes{n0}))
			Ω(nodes.WithinNestingLevel(1)).Should(Equal(Nodes{n0, n1}))
			Ω(nodes.WithinNestingLevel(2)).Should(Equal(Nodes{n0, n1, n2a, n2b}))
			Ω(nodes.WithinNestingLevel(3)).Should(Equal(Nodes{n0, n1, n2a, n3, n2b}))
		})
	})

	Describe("Texts", func() {
		var nodes Nodes
		BeforeEach(func() {
			nodes = Nodes{N("the first node"), N(""), N("2"), N("c"), N("")}
		})

		It("returns a string slice containing the individual node text strings in order", func() {
			Ω(nodes.Texts()).Should(Equal([]string{"the first node", "", "2", "c", ""}))
		})
	})

	Describe("CodeLocation", func() {
		var nodes Nodes
		var cl1, cl2 types.CodeLocation
		BeforeEach(func() {
			cl1 = types.NewCodeLocation(0)
			cl2 = types.NewCodeLocation(0)
			nodes = Nodes{N(cl1), N(cl2), N()}
		})

		It("returns a types.CodeLocation sice containing the individual node code locations in order", func() {
			Ω(nodes.CodeLocations()).Should(Equal([]types.CodeLocation{cl1, cl2, cl}))
		})
	})

	Describe("BestTextFor", func() {
		var nIt, nBef1, nBef2 Node
		var nodes Nodes
		BeforeEach(func() {
			nIt = N("an it", ntIt, NestingLevel(2))
			nBef1 = N(ntBef, NestingLevel(2))
			nBef2 = N(ntBef, NestingLevel(4))
			nodes = Nodes{
				N("the root container", ntCon, NestingLevel(0)),
				N("the inner container", ntCon, NestingLevel(1)),
				nBef1,
				nIt,
				nBef2,
			}
		})

		Context("when the passed in node has text", func() {
			It("returns that text", func() {
				Ω(nodes.BestTextFor(nIt)).Should(Equal("an it"))
			})
		})

		Context("when the node has no text", func() {
			Context("and there is a node one-nesting-level-up with text", func() {
				It("returns that node's text", func() {
					Ω(nodes.BestTextFor(nBef1)).Should(Equal("the inner container"))
				})
			})

			Context("and there is no node one-nesting-level up with text", func() {
				It("returns empty string", func() {
					Ω(nodes.BestTextFor(nBef2)).Should(Equal(""))
				})
			})
		})
	})

	Describe("HasNodeMarkedPending", func() {
		Context("when there is a node marked pending", func() {
			It("returns true", func() {
				nodes := Nodes{N(), N(), N(Pending), N()}
				Ω(nodes.HasNodeMarkedPending()).Should(BeTrue())
			})
		})

		Context("when there is no node marked pending", func() {
			It("returns false", func() {
				nodes := Nodes{N(), N(), N()}
				Ω(nodes.HasNodeMarkedPending()).Should(BeFalse())
			})
		})
	})

	Describe("HasNodeMarkedFocus", func() {
		Context("when there is a node marked focus", func() {
			It("returns true", func() {
				nodes := Nodes{N(), N(), N(Focus), N()}
				Ω(nodes.HasNodeMarkedFocus()).Should(BeTrue())
			})
		})

		Context("when there is no node marked focus", func() {
			It("returns false", func() {
				nodes := Nodes{N(), N(), N()}
				Ω(nodes.HasNodeMarkedFocus()).Should(BeFalse())
			})
		})
	})
})
