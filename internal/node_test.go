package internal_test

import (
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

var _ = Describe("Node", func() {
	Describe("The primary NewNode constructor", func() {
		It("creates a node with a non-zero id", func() {
			var didRun bool
			node := internal.NewNode(ntIt, "hummus", func() { didRun = true }, cl, true, false)
			Ω(node.ID).Should(BeNumerically(">", 0))
			Ω(node.NodeType).Should(Equal(ntIt))
			Ω(node.Text).Should(Equal("hummus"))
			node.Body()
			Ω(didRun).Should(BeTrue())
			Ω(node.CodeLocation).Should(Equal(cl))
			Ω(node.MarkedFocus).Should(BeTrue())
			Ω(node.MarkedPending).Should(BeFalse())
			Ω(node.NestingLevel).Should(Equal(-1))
		})
	})

	Describe("The other node constructors", func() {
		Describe("NewSynchronizedBeforeSuiteNode", func() {
			It("returns a correctly configured node", func() {
				var ranNode1, ranAllNodes bool
				node1Body := func() []byte { ranNode1 = true; return nil }
				allNodesBody := func(_ []byte) { ranAllNodes = true }
				node := internal.NewSynchronizedBeforeSuiteNode(node1Body, allNodesBody, cl)
				Ω(node.ID).Should(BeNumerically(">", 0))
				Ω(node.NodeType).Should(Equal(types.NodeTypeSynchronizedBeforeSuite))

				node.SynchronizedBeforeSuiteNode1Body()
				Ω(ranNode1).Should(BeTrue())

				node.SynchronizedBeforeSuiteAllNodesBody(nil)
				Ω(ranAllNodes).Should(BeTrue())

				Ω(node.CodeLocation).Should(Equal(cl))
				Ω(node.NestingLevel).Should(Equal(-1))
			})
		})

		Describe("NewSynchronizedAfterSuiteNode", func() {
			It("returns a correctly configured node", func() {
				var ranNode1, ranAllNodes bool
				allNodesBody := func() { ranAllNodes = true }
				node1Body := func() { ranNode1 = true }

				node := internal.NewSynchronizedAfterSuiteNode(allNodesBody, node1Body, cl)
				Ω(node.ID).Should(BeNumerically(">", 0))
				Ω(node.NodeType).Should(Equal(types.NodeTypeSynchronizedAfterSuite))

				node.SynchronizedAfterSuiteAllNodesBody()
				Ω(ranAllNodes).Should(BeTrue())

				node.SynchronizedAfterSuiteNode1Body()
				Ω(ranNode1).Should(BeTrue())

				Ω(node.CodeLocation).Should(Equal(cl))
				Ω(node.NestingLevel).Should(Equal(-1))
			})
		})

		Describe("NewReportAfterSuiteNode", func() {
			It("returns a correctly configured node", func() {
				var didRun bool
				body := func(types.Report) { didRun = true }
				node := internal.NewReportAfterSuiteNode(body, cl)
				Ω(node.ID).Should(BeNumerically(">", 0))
				Ω(node.NodeType).Should(Equal(types.NodeTypeReportAfterSuite))

				node.ReportAfterSuiteBody(types.Report{})
				Ω(didRun).Should(BeTrue())

				Ω(node.CodeLocation).Should(Equal(cl))
				Ω(node.NestingLevel).Should(Equal(-1))
			})
		})

		Describe("NewReportAfterEachNode", func() {
			It("returns a correctly configured node", func() {
				var didRun bool
				body := func(types.SpecReport) { didRun = true }

				node := internal.NewReportAfterEachNode(body, cl)
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
			node := internal.NewNode(ntIt, "hummus", func() {}, cl, false, false)
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
				nodes := Nodes{N(), N(), N(MarkedPending(true)), N()}
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
				nodes := Nodes{N(), N(), N(MarkedFocus(true)), N()}
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
