package ginkgo

import (
	"github.com/onsi/ginkgo/types"
	. "github.com/onsi/gomega"
	"math/rand"
	"sort"
)

func init() {
	Describe("Container Node", func() {
		var (
			codeLocation types.CodeLocation
			container    *containerNode
		)

		BeforeEach(func() {
			codeLocation = types.GenerateCodeLocation(0)
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
					itA := newItNode("itA", func() {}, flagTypeNone, types.GenerateCodeLocation(0), 0)
					itB := newItNode("itB", func() {}, flagTypeNone, types.GenerateCodeLocation(0), 0)
					subContainer := newContainerNode("subcontainer", flagTypeNone, types.GenerateCodeLocation(0))
					container.pushSubjectNode(itA)
					container.pushContainerNode(subContainer)
					container.pushSubjectNode(itB)
					Ω(container.subjectAndContainerNodes).Should(Equal([]node{
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
					runnableA = newRunnableNode(func() {}, types.GenerateCodeLocation(0), 0)
					runnableB = newRunnableNode(func() {}, types.GenerateCodeLocation(0), 0)
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
				itA = newItNode("itA", func() {}, flagTypeNone, types.GenerateCodeLocation(0), 0)
				itB = newItNode("itB", func() {}, flagTypeNone, types.GenerateCodeLocation(0), 0)
				subContainer = newContainerNode("subcontainer", flagTypeNone, types.GenerateCodeLocation(0))
				subItA = newItNode("subItA", func() {}, flagTypeNone, types.GenerateCodeLocation(0), 0)
				subItB = newItNode("subItB", func() {}, flagTypeNone, types.GenerateCodeLocation(0), 0)

				container.pushSubjectNode(itA)
				container.pushContainerNode(subContainer)
				container.pushSubjectNode(itB)

				subContainer.pushSubjectNode(subItA)
				subContainer.pushSubjectNode(subItB)
			})

			It("generates an example for each It in the hierarchy", func() {
				examples := container.generateExamples()
				Ω(examples).Should(HaveLen(4))

				Ω(examples[0].subject).Should(Equal(itA))
				Ω(examples[0].containers).Should(Equal([]*containerNode{container}))

				Ω(examples[1].subject).Should(Equal(subItA))
				Ω(examples[1].containers).Should(Equal([]*containerNode{container, subContainer}))

				Ω(examples[2].subject).Should(Equal(subItB))
				Ω(examples[2].containers).Should(Equal([]*containerNode{container, subContainer}))

				Ω(examples[3].subject).Should(Equal(itB))
				Ω(examples[3].containers).Should(Equal([]*containerNode{container}))
			})

			It("ignores containers in the hierarchy that are empty", func() {
				emptyContainer := newContainerNode("empty container", flagTypeNone, types.GenerateCodeLocation(0))
				emptyContainer.pushBeforeEachNode(newRunnableNode(func() {}, types.GenerateCodeLocation(0), 0))

				container.pushContainerNode(emptyContainer)
				examples := container.generateExamples()
				Ω(examples).Should(HaveLen(4))
			})
		})

		Describe("shuffling the container", func() {
			texts := func(container *containerNode) []string {
				texts := make([]string, 0)
				for _, node := range container.subjectAndContainerNodes {
					texts = append(texts, node.getText())
				}
				return texts
			}

			BeforeEach(func() {
				itA := newItNode("Banana", func() {}, flagTypeNone, types.GenerateCodeLocation(0), 0)
				itB := newItNode("Apple", func() {}, flagTypeNone, types.GenerateCodeLocation(0), 0)
				itC := newItNode("Orange", func() {}, flagTypeNone, types.GenerateCodeLocation(0), 0)
				containerA := newContainerNode("Cucumber", flagTypeNone, types.GenerateCodeLocation(0))
				containerB := newContainerNode("Airplane", flagTypeNone, types.GenerateCodeLocation(0))

				container.pushSubjectNode(itA)
				container.pushContainerNode(containerA)
				container.pushSubjectNode(itB)
				container.pushContainerNode(containerB)
				container.pushSubjectNode(itC)
			})

			It("should be sortable", func() {
				sort.Sort(container)
				Ω(texts(container)).Should(Equal([]string{"Airplane", "Apple", "Banana", "Cucumber", "Orange"}))
			})

			It("shuffles all the examples after sorting them", func() {
				container.shuffle(rand.New(rand.NewSource(17)))
				expectedOrder := shuffleStrings([]string{"Airplane", "Apple", "Banana", "Cucumber", "Orange"}, 17)
				Ω(texts(container)).Should(Equal(expectedOrder), "The permutation should be the same across test runs")
			})
		})
	})
}
