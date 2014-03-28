package containernode_test

import (
	"github.com/onsi/ginkgo/internal/leafnodes"
	"github.com/onsi/ginkgo/internal/types"
	"math/rand"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/ginkgo/internal/codelocation"
	. "github.com/onsi/ginkgo/internal/containernode"
	"github.com/onsi/ginkgo/types"
)

var _ = Describe("Container Node", func() {
	var (
		codeLocation types.CodeLocation
		container    *ContainerNode
	)

	BeforeEach(func() {
		codeLocation = codelocation.New(0)
		container = New("description text", internaltypes.FlagTypeFocused, codeLocation)
	})

	Describe("creating a container node", func() {
		It("can answer questions about itself", func() {
			Ω(container.Text()).Should(Equal("description text"))
			Ω(container.Flag()).Should(Equal(internaltypes.FlagTypeFocused))
			Ω(container.CodeLocation()).Should(Equal(codeLocation))
		})
	})

	Describe("pushing setup nodes", func() {
		It("can append multiple before each nodes", func() {
			a := leafnodes.NewBeforeEachNode(func() {}, codelocation.New(0), 0, nil, 0)
			b := leafnodes.NewBeforeEachNode(func() {}, codelocation.New(0), 0, nil, 0)

			container.PushBeforeEachNode(a)
			container.PushBeforeEachNode(b)

			Ω(container.BeforeEachNodes()).Should(Equal([]internaltypes.BasicNode{a, b}))
		})

		It("can append multiple after each nodes", func() {
			a := leafnodes.NewAfterEachNode(func() {}, codelocation.New(0), 0, nil, 0)
			b := leafnodes.NewAfterEachNode(func() {}, codelocation.New(0), 0, nil, 0)

			container.PushAfterEachNode(a)
			container.PushAfterEachNode(b)

			Ω(container.AfterEachNodes()).Should(Equal([]internaltypes.BasicNode{a, b}))
		})

		It("can append multiple just before each nodes", func() {
			a := leafnodes.NewJustBeforeEachNode(func() {}, codelocation.New(0), 0, nil, 0)
			b := leafnodes.NewJustBeforeEachNode(func() {}, codelocation.New(0), 0, nil, 0)

			container.PushJustBeforeEachNode(a)
			container.PushJustBeforeEachNode(b)

			Ω(container.JustBeforeEachNodes()).Should(Equal([]internaltypes.BasicNode{a, b}))
		})
	})

	Context("With appended containers and subject nodes", func() {
		var (
			itA, itB, innerItA, innerItB internaltypes.SubjectNode
			innerContainer               *ContainerNode
		)

		BeforeEach(func() {
			itA = leafnodes.NewItNode("Banana", func() {}, internaltypes.FlagTypeNone, codelocation.New(0), 0, nil, 0)
			itB = leafnodes.NewItNode("Apple", func() {}, internaltypes.FlagTypeNone, codelocation.New(0), 0, nil, 0)

			innerItA = leafnodes.NewItNode("inner A", func() {}, internaltypes.FlagTypeNone, codelocation.New(0), 0, nil, 0)
			innerItB = leafnodes.NewItNode("inner B", func() {}, internaltypes.FlagTypeNone, codelocation.New(0), 0, nil, 0)

			innerContainer = New("Orange", internaltypes.FlagTypeNone, codelocation.New(0))

			container.PushSubjectNode(itA)
			container.PushContainerNode(innerContainer)
			innerContainer.PushSubjectNode(innerItA)
			innerContainer.PushSubjectNode(innerItB)
			container.PushSubjectNode(itB)
		})

		Describe("Collating", func() {
			It("should return a collated set of containers and subject nodes in the correct order", func() {
				collated := container.Collate()
				Ω(collated).Should(HaveLen(4))

				Ω(collated[0]).Should(Equal(CollatedNodes{
					Containers: []*ContainerNode{container},
					Subject:    itA,
				}))

				Ω(collated[1]).Should(Equal(CollatedNodes{
					Containers: []*ContainerNode{container, innerContainer},
					Subject:    innerItA,
				}))

				Ω(collated[2]).Should(Equal(CollatedNodes{
					Containers: []*ContainerNode{container, innerContainer},
					Subject:    innerItB,
				}))

				Ω(collated[3]).Should(Equal(CollatedNodes{
					Containers: []*ContainerNode{container},
					Subject:    itB,
				}))
			})
		})

		Describe("Shuffling", func() {
			var unshuffledCollation []CollatedNodes
			BeforeEach(func() {
				unshuffledCollation = container.Collate()

				r := rand.New(rand.NewSource(17))
				container.Shuffle(r)
			})

			It("should sort, and then shuffle, the top level contents of the container", func() {
				shuffledCollation := container.Collate()
				Ω(shuffledCollation).Should(HaveLen(len(unshuffledCollation)))
				Ω(shuffledCollation).ShouldNot(Equal(unshuffledCollation))

				for _, entry := range unshuffledCollation {
					Ω(shuffledCollation).Should(ContainElement(entry))
				}

				innerAIndex, innerBIndex := 0, 0
				for i, entry := range shuffledCollation {
					if entry.Subject == innerItA {
						innerAIndex = i
					} else if entry.Subject == innerItB {
						innerBIndex = i
					}
				}

				Ω(innerAIndex).Should(Equal(innerBIndex - 1))
			})
		})
	})
})
