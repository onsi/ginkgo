package internal_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/onsi/ginkgo/v2/internal"
)

var _ = Describe("Trees (TreeNode and TreeNodes)", func() {
	Describe("TreeNodes methods", func() {
		var n1, n2, n3 Node
		var childNode Node
		var treeNodes TreeNodes
		BeforeEach(func() {
			n1, n2, n3 = N(), N(), N()
			childNode = N()
			treeNodes = TreeNodes{
				TN(n1,
					TN(childNode),
				),
				TN(n2),
				TN(n3),
			}
		})

		Describe("treenodes.Nodes", func() {
			It("returns the root node of each node in the treenodes slice", func() {
				Ω(treeNodes.Nodes()).Should(Equal(Nodes{n1, n2, n3}))
			})
		})

		Describe("treenodes.WithId", func() {
			Context("when a tree with a root node with a matching id is found", func() {
				It("returns that tree", func() {
					Ω(treeNodes.WithID(n2.ID)).Should(Equal(TN(n2)))
				})
			})

			Context("when the id matches a child node's id", func() {
				It("returns an empty tree as children are not included in the match", func() {
					Ω(treeNodes.WithID(childNode.ID)).Should(BeNil())
				})
			})

			Context("when the id cannot be found", func() {
				It("returns an empty tree", func() {
					Ω(treeNodes.WithID(1000000)).Should(BeZero()) //pretty sure it's a safe bet we don't ever get to 1_000_000 nodes in this test ;)
				})
			})
		})

		Describe("AppendChild", func() {
			It("appends the passed in child treenode to the parent's children and sets the child's parent", func() {
				existingChildNode1 := N()
				existingChildNode2 := N()
				treeNode := TN(N(),
					TN(existingChildNode1),
					TN(existingChildNode2),
				)
				newChildNode := N()
				childTreeNode := &TreeNode{Node: newChildNode}
				treeNode.AppendChild(childTreeNode)
				Ω(treeNode.Children.Nodes()).Should(Equal(Nodes{existingChildNode1, existingChildNode2, newChildNode}))
				Ω(childTreeNode.Parent).Should(Equal(treeNode))
			})
		})

		Describe("ParentChain", func() {
			It("returns the chain of parent nodes", func() {
				grandparent := N()
				parent := N()
				aunt := N()
				child := N()
				sibling := N()
				tree := TN(Node{}, TN(
					grandparent,
					TN(parent, TN(child), TN(sibling)),
					TN(aunt),
				))
				childTree := tree.Children[0].Children[0].Children[0]
				Ω(childTree.Node).Should(Equal(child))
				Ω(childTree.AncestorNodeChain()).Should(Equal(Nodes{grandparent, parent, child}))
			})
		})
	})

	Describe("GenerateSpecsFromTreeRoot", func() {
		var tree *TreeNode
		BeforeEach(func() {
			tree = &TreeNode{}
		})

		Context("when the tree is empty", func() {
			It("returns an empty set of tests", func() {
				Ω(internal.GenerateSpecsFromTreeRoot(tree)).Should(BeEmpty())
			})
		})

		Context("when the tree has no Its", func() {
			BeforeEach(func() {
				tree = TN(Node{},
					TN(N(ntBef)),
					TN(N(ntCon),
						TN(N(ntBef)),
						TN(N(ntAf)),
					),
					TN(N(ntCon),
						TN(N(ntCon),
							TN(N(ntBef)),
							TN(N(ntAf)),
						),
					),
					TN(N(ntAf)),
				)
			})

			It("returns an empty set of tests", func() {
				Ω(internal.GenerateSpecsFromTreeRoot(tree)).Should(BeEmpty())
			})
		})

		Context("when the tree has nodes in it", func() {
			var tests Specs
			BeforeEach(func() {
				tree = TN(Node{},
					TN(N(ntBef, "Bef #0")),
					TN(N(ntIt, "It #1")),
					TN(N(ntCon, "Container #1"),
						TN(N(ntBef, "Bef #1")),
						TN(N(ntAf, "Af #1")),
						TN(N(ntIt, "It #2")),
					),
					TN(N(ntCon, "Container #2"),
						TN(N(ntBef, "Bef #2")),
						TN(N(ntCon, "Nested Container"),
							TN(N(ntBef, "Bef #4")),
							TN(N(ntIt, "It #3")),
							TN(N(ntIt, "It #4")),
							TN(N(ntAf, "Af #2")),
						),
						TN(N(ntIt, "It #5")),
						TN(N(ntCon, "A Container With No Its"),
							TN(N(ntBef, "Bef #5")),
						),
						TN(N(ntAf, "Af #3")),
					),
					TN(N(ntIt, "It #6")),
					TN(N(ntAf, "Af #4")),
				)

				tests = internal.GenerateSpecsFromTreeRoot(tree)
			})

			It("constructs a flattened set of tests", func() {
				Ω(tests).Should(HaveLen(6))
				expectedTexts := [][]string{
					{"Bef #0", "It #1", "Af #4"},
					{"Bef #0", "Container #1", "Bef #1", "Af #1", "It #2", "Af #4"},
					{"Bef #0", "Container #2", "Bef #2", "Nested Container", "Bef #4", "It #3", "Af #2", "Af #3", "Af #4"},
					{"Bef #0", "Container #2", "Bef #2", "Nested Container", "Bef #4", "It #4", "Af #2", "Af #3", "Af #4"},
					{"Bef #0", "Container #2", "Bef #2", "It #5", "Af #3", "Af #4"},
					{"Bef #0", "It #6", "Af #4"},
				}
				for i, expectedText := range expectedTexts {
					Ω(tests[i].Nodes.Texts()).Should(Equal(expectedText))
				}
			})

			It("ensures each node as the correct nesting level", func() {
				extpectedNestingLevels := [][]int{
					{0, 0, 0},
					{0, 0, 1, 1, 1, 0},
					{0, 0, 1, 1, 2, 2, 2, 1, 0},
					{0, 0, 1, 1, 2, 2, 2, 1, 0},
					{0, 0, 1, 1, 1, 0},
					{0, 0, 0},
				}
				for i, expectedNestingLevels := range extpectedNestingLevels {
					for j, expectedNestingLevel := range expectedNestingLevels {
						Ω(tests[i].Nodes[j].NestingLevel).Should(Equal(expectedNestingLevel))
					}
				}
			})
		})
	})
})
