package ginkgo

import (
	. "github.com/onsi/gomega"
	"math/rand"
)

func init() {
	Describe("Container Node", func() {
		var (
			codeLocation CodeLocation
			container    *containerNode
		)

		BeforeEach(func() {
			codeLocation = generateCodeLocation(0)
			container = newContainerNode("description text", flagTypeFocused, codeLocation)
		})

		Describe("creating a container node", func() {
			It("stores off the passed in properties", func() {
				Ω(container.text).Should(Equal("description text"))
				Ω(container.flag).Should(Equal(flagTypeFocused))
				Ω(container.codeLocation).Should(Equal(codeLocation))
			})
		})

		Describe("appending", func() {
			Describe("it nodes", func() {
				It("can append container nodes and it nodes", func() {
					itA := newItNode("itA", func() {}, flagTypeNone, generateCodeLocation(0), 0)
					itB := newItNode("itB", func() {}, flagTypeNone, generateCodeLocation(0), 0)
					subContainer := newContainerNode("subcontainer", flagTypeNone, generateCodeLocation(0))
					container.pushItNode(itA)
					container.pushContainerNode(subContainer)
					container.pushItNode(itB)
					Ω(container.itAndContainerNodes).Should(Equal([]node{
						itA,
						subContainer,
						itB,
					}))
				})
			})

			Describe("other runnable nodes", func() {
				var (
					runnableA *runnableNode
					runnableB *runnableNode
				)

				BeforeEach(func() {
					runnableA = newRunnableNode(func() {}, generateCodeLocation(0), 0)
					runnableB = newRunnableNode(func() {}, generateCodeLocation(0), 0)
				})

				It("can append multiple beforeEach nodes", func() {
					container.pushBeforeEachNode(runnableA)
					container.pushBeforeEachNode(runnableB)
					Ω(container.beforeEachNodes).Should(Equal([]*runnableNode{
						runnableA,
						runnableB,
					}))
				})

				It("can append multiple justBeforeEach nodes", func() {
					container.pushJustBeforeEachNode(runnableA)
					container.pushJustBeforeEachNode(runnableB)
					Ω(container.justBeforeEachNodes).Should(Equal([]*runnableNode{
						runnableA,
						runnableB,
					}))
				})

				It("can append multiple afterEach nodes", func() {
					container.pushAfterEachNode(runnableA)
					container.pushAfterEachNode(runnableB)
					Ω(container.afterEachNodes).Should(Equal([]*runnableNode{
						runnableA,
						runnableB,
					}))
				})
			})
		})

		Describe("generating examples", func() {
			var (
				itA          *itNode
				itB          *itNode
				subContainer *containerNode
				subItA       *itNode
				subItB       *itNode
			)

			BeforeEach(func() {
				itA = newItNode("itA", func() {}, flagTypeNone, generateCodeLocation(0), 0)
				itB = newItNode("itB", func() {}, flagTypeNone, generateCodeLocation(0), 0)
				subContainer = newContainerNode("subcontainer", flagTypeNone, generateCodeLocation(0))
				subItA = newItNode("subItA", func() {}, flagTypeNone, generateCodeLocation(0), 0)
				subItB = newItNode("subItB", func() {}, flagTypeNone, generateCodeLocation(0), 0)

				container.pushItNode(itA)
				container.pushContainerNode(subContainer)
				container.pushItNode(itB)

				subContainer.pushItNode(subItA)
				subContainer.pushItNode(subItB)
			})

			It("generates an example for each It in the hierarchy", func() {
				examples := container.generateExamples()
				Ω(examples).Should(HaveLen(4))

				Ω(examples[0].it).Should(Equal(itA))
				Ω(examples[0].containers).Should(Equal([]*containerNode{container}))

				Ω(examples[1].it).Should(Equal(subItA))
				Ω(examples[1].containers).Should(Equal([]*containerNode{container, subContainer}))

				Ω(examples[2].it).Should(Equal(subItB))
				Ω(examples[2].containers).Should(Equal([]*containerNode{container, subContainer}))

				Ω(examples[3].it).Should(Equal(itB))
				Ω(examples[3].containers).Should(Equal([]*containerNode{container}))
			})

			It("ignores containers in the hierarchy that are empty", func() {
				emptyContainer := newContainerNode("empty container", flagTypeNone, generateCodeLocation(0))
				emptyContainer.pushBeforeEachNode(newRunnableNode(func() {}, generateCodeLocation(0), 0))

				container.pushContainerNode(emptyContainer)
				examples := container.generateExamples()
				Ω(examples).Should(HaveLen(4))
			})
		})

		Describe("shuffling", func() {
			It("shuffles the top level It and Container nodes using the passed in randomizer", func() {
				r := rand.New(rand.NewSource(3))
				itA := newItNode("itA", func() {}, flagTypeNone, generateCodeLocation(0), 0)
				itB := newItNode("itB", func() {}, flagTypeNone, generateCodeLocation(0), 0)
				subContainerA := newContainerNode("subcontainerA", flagTypeNone, generateCodeLocation(0))
				subContainerB := newContainerNode("subcontainerB", flagTypeNone, generateCodeLocation(0))
				itC := newItNode("itC", func() {}, flagTypeNone, generateCodeLocation(0), 0)

				container.pushItNode(itA)
				container.pushContainerNode(subContainerA)
				container.pushItNode(itB)
				container.pushContainerNode(subContainerB)
				container.pushItNode(itC)

				container.shuffle(r)

				Ω(container.itAndContainerNodes).Should(Equal([]node{itB, subContainerA, itC, itA, subContainerB}), "Since we start with an explicit random seed, the expected randomization should be fixed across test runs.")
			})
		})
	})
}
